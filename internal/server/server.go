package server

import (
	"errors"
	"netfs/internal/api"
	"netfs/internal/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"netfs/internal/server/volume"
	"os"
	"os/signal"
	"syscall"
)

// The netfs server configuration.
type ServerConfig struct {
	Network  api.NetworkConfig
	Log      logger.LoggerConfig
	Database database.DatabaseConfig
	Task     task.TaskExecuteConfig
}

// The netfs server.
type Server struct {
	db       database.Database
	log      *logger.Logger
	network  *api.Network
	receiver transport.TransportReceiver
	tasks    task.TaskExecutor
	volumes  volume.VolumeManager
	stop     chan os.Signal
}

// Starts the netfs server.
func (srv *Server) Start() error {
	srv.receiver.Receive(api.API.ServerStop(), srv.StopServerHandle)
	srv.receiver.ReceiveRawBodyAndSend(api.API.ServerHost(), srv.ServerHostHandle)
	srv.receiver.ReceiveRawBodyAndSend(api.API.FileInfo(), srv.FileInfoHandle)
	srv.receiver.ReceiveBody(api.API.FileCreate(), func() any { return &api.FileInfo{} }, srv.FileCreateHandle)
	srv.receiver.ReceiveBodyAndSend(api.API.FileCopyStart(), func() any { return []api.RemoteFile{} }, srv.FileCopyStartHandle)

	dbErr := srv.db.Start()
	recErr := srv.receiver.Start()
	tskErr := srv.tasks.Start()

	if dbErr == nil && recErr == nil && tskErr == nil {
		<-srv.stop // Stop signal waiting.
	}

	if recErr == nil {
		srv.receiver.Stop()
	}

	if tskErr == nil {
		srv.tasks.Stop()
	}

	if dbErr == nil {
		srv.db.Stop()
	}
	return errors.Join(dbErr, recErr, tskErr)
}

// Stops the netfs server.
func (srv *Server) Stop() error {
	srv.stop <- syscall.SIGINT
	return nil
}

// New instance of the netfs server.
func NewServer(config ServerConfig) (*Server, error) {
	log := logger.NewLogger(config.Log)
	db := database.NewDatabase(config.Database)

	network, err := api.NewNetwork(config.Network)
	if err == nil {
		var receiver transport.TransportReceiver
		if receiver, err = transport.NewReceiver(config.Network.Protocol, config.Network.Port); err == nil {
			var volumes volume.VolumeManager
			if volumes, err = volume.NewVolumeManager(db); err == nil {
				var tasks task.TaskExecutor
				if tasks, err = task.NewTaskExecutor(config.Task, db, volumes, network.Transport(), log); err == nil {
					stop := make(chan os.Signal, 1)
					signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

					return &Server{db: db, log: log, network: network, receiver: receiver, tasks: tasks, volumes: volumes, stop: stop}, nil
				}
			}
		}
	}
	return nil, err
}

// Stops the server by request from current host.
func (srv *Server) StopServerHandle() error { // TODO. Check current host.
	return srv.Stop()
}

// Returns information about the current host.
func (srv *Server) ServerHostHandle([]byte) (any, error) {
	return srv.network.LocalHost(), nil
}

// Returns information about file.
func (srv *Server) FileInfoHandle(data []byte) (any, error) {
	path := string(data)
	volume, err := srv.volumes.Volume(path)
	if err == nil {
		return volume.Info(path)
	}
	return nil, err
}

// Creates new file or directory by api.FileInfo.
func (srv *Server) FileCreateHandle(data any) error {
	info := data.(*api.FileInfo)
	volume, err := srv.volumes.Volume(info.FilePath)
	if err == nil {
		err = volume.Create(info)
	}
	return err
}

func (srv *Server) FileCopyStartHandle(data any) (any, error) {
	return nil, nil
}

// // Create directory.
// func (serv *Server) fileCreateHandle(writer http.ResponseWriter, request *http.Request) {
// 	if request.Method == netfs.API.FileCreate.Method {
// 		defer request.Body.Close()

// 		data, err := io.ReadAll(request.Body)
// 		if err == nil {
// 			file := &netfs.RemoteFile{}
// 			if err = json.Unmarshal(data, file); err == nil {
// 				if file.Type == netfs.DIRECTORY {
// 					err = os.MkdirAll(file.Path, os.ModePerm)
// 				} else {
// 					err = errors.New("unsupported file operation")
// 				}
// 			}
// 		}

// 		if err != nil {
// 			fmt.Println(err)

// 			writer.Write([]byte(err.Error()))
// 			writer.WriteHeader(http.StatusInternalServerError)
// 		}
// 	} else {
// 		writer.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

// // Writes the received data to file.
// func (serv *Server) fileWriteHandle(writer http.ResponseWriter, request *http.Request) {
// 	if request.Method == netfs.API.FileWrite.Method {
// 		query := request.URL.Query()
// 		if path := query.Get(netfs.API.FileWrite.Path); path != "" {
// 			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 			if err == nil {
// 				defer file.Close()
// 				defer request.Body.Close()

// 				var data []byte
// 				if data, err = io.ReadAll(request.Body); err == nil {
// 					file.Write(data)
// 				}
// 			}

// 			if err != nil {
// 				fmt.Println(err) // TODO. Log format

// 				writer.Write([]byte(err.Error()))
// 				writer.WriteHeader(http.StatusInternalServerError)
// 			}
// 		}
// 	} else {
// 		writer.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

// // Starting a file or directory copy operation.
// func (serv *Server) fileCopyStartHandle(writer http.ResponseWriter, request *http.Request) {
// 	log := serv.log
// 	log.Info("handle url: [%s]", request.URL)

// 	if request.Method == netfs.API.FileCopyStart.Method {
// 		defer request.Body.Close()

// 		var err error
// 		var data []byte
// 		if data, err = io.ReadAll(request.Body); err == nil {
// 			files := &[]netfs.RemoteFile{}
// 			if err = json.Unmarshal(data, files); err == nil {
// 				var copyTask task.Task
// 				if copyTask, err = task.NewCopyTask((*files)[0], (*files)[1]); err == nil {
// 					if err = serv.executor.Submit(copyTask); err == nil {
// 						log.Info("handle response status: [%d]", http.StatusOK)
// 						log.Info("handle response body: [%s]", copyTask.Id())

// 						writer.Write([]byte(copyTask.Id()))
// 						writer.WriteHeader(http.StatusOK)
// 					}
// 				}
// 			}
// 		}

// 		if err != nil {
// 			log.Info("handle error: [%s]", err)
// 			log.Info("handle response status: [%d]", http.StatusInternalServerError)
// 			log.Info("handle response body: [%s]", err.Error())

// 			writer.Write([]byte(err.Error()))
// 			writer.WriteHeader(http.StatusInternalServerError)
// 		}
// 	} else {
// 		writer.WriteHeader(http.StatusMethodNotAllowed)
// 		log.Info("handle response status: [%d]", http.StatusMethodNotAllowed)
// 		log.Info("handle response body: [nil]")
// 	}
// }

// // Returns information about copy operation.
// func (serv *Server) fileCopyStatusHandle(writer http.ResponseWriter, request *http.Request) {
// 	api := netfs.API.FileCopyStatus
// 	if request.Method == api.Method {
// 		defer request.Body.Close()

// 		query := request.URL.Query()
// 		id := query.Get(api.Id)
// 		status := query.Get(api.Status)

// 		if id != "" {
// 			// TODO. Search operations by id
// 		} else if status != "" {
// 			// TODO. Search operations by status
// 		} else {
// 			writer.Write([]byte("request must have [id] or [status] parameter"))
// 			writer.WriteHeader(http.StatusBadRequest)
// 		}
// 	} else {
// 		writer.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

// // The context of asynchronous tasks.
// type _TasksContext struct {
// 	_Context  context.Context
// 	_Shutdown context.CancelFunc
// }
