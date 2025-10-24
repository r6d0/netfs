package server

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	netfs "netfs/internal"
// 	"netfs/internal/logger"
// 	"netfs/internal/server/task"
// 	"netfs/internal/store"
// 	"os"
// 	"os/signal"
// 	"strconv"
// 	"strings"
// 	"syscall"
// )

// const portSeparator = ":"

// // HTTP server.
// type Server struct {
// 	db         store.Store
// 	httpServer *http.Server
// 	config     *netfs.Config
// 	tasks      *_TasksContext
// 	host       netfs.RemoteHost
// 	stop       chan os.Signal
// 	executor   task.TaskExecutor
// 	log        *logger.Logger
// }

// // Start requests listening. It's blocking current thread.
// func (serv *Server) Start() error {
// 	// Start database.
// 	err := serv.db.Start()

// 	if err == nil {
// 		// Start tasks executor.
// 		err = serv.executor.Start()
// 		if err == nil {
// 			// Run HTTP server.
// 			go func() {
// 				err = serv.httpServer.ListenAndServe()
// 				if err != http.ErrServerClosed {
// 					serv.stop <- syscall.SIGINT
// 				}
// 			}()

// 			// Stop signal waiting.
// 			<-serv.stop

// 			if err == nil {
// 				// HTTP server stopping.
// 				err = serv.httpServer.Shutdown(context.Background())
// 			}

// 			// Tasks stopping.
// 			serv.tasks._Shutdown()
// 		}

// 		// Stop tasks executor.
// 		err = errors.Join(err, serv.executor.Stop())
// 	}

// 	// Stop database.
// 	err = errors.Join(err, serv.db.Stop())
// 	serv.log.Error("server error: [%s]", err)

// 	if err == nil {
// 		serv.log.Info("server is stopped")
// 	}
// 	return err
// }

// // Stop the running server.
// func (serv *Server) Stop() {
// 	serv.stop <- syscall.SIGINT
// }

// // New instance of the netfs server.
// func NewServer(config *netfs.Config) (*Server, error) {
// 	log := logger.NewLogger(logger.LoggerConfig{Level: logger.Info})

// 	network, err := netfs.NewNetwork(config)
// 	if err == nil {
// 		store := store.NewStore(store.StoreConfig{Path: "./data"})

// 		var executor task.TaskExecutor
// 		if executor, err = task.NewTaskExecutor(task.TaskConfig{MaxAvailableTasks: 100, TasksWaitingSecond: 2}, store, log); err == nil {
// 			// Information about server host
// 			var host *netfs.RemoteHost
// 			host, err = network.GetLocalHost()
// 			if err == nil {
// 				server := &Server{config: config, host: *host, db: store, log: log, executor: executor}

// 				// Stop signal
// 				server.stop = make(chan os.Signal, 1)
// 				signal.Notify(server.stop, syscall.SIGINT, syscall.SIGTERM)

// 				// Tasks context
// 				ctx, cancel := context.WithCancel(context.Background())
// 				server.tasks = &_TasksContext{_Context: ctx, _Shutdown: cancel}

// 				// API Registration
// 				mux := http.NewServeMux()
// 				mux.HandleFunc(netfs.API.Stop, server.stopHandle)
// 				mux.HandleFunc(netfs.API.Host, server.hostHandle)
// 				mux.HandleFunc(netfs.API.FileInfo.URL, server.fileInfoHandle)
// 				mux.HandleFunc(netfs.API.FileCreate.URL, server.fileCreateHandle)
// 				mux.HandleFunc(netfs.API.FileWrite.URL, server.fileWriteHandle)
// 				mux.HandleFunc(netfs.API.FileCopyStart.URL, server.fileCopyStartHandle)
// 				mux.HandleFunc(netfs.API.FileCopyStatus.URL, server.fileCopyStatusHandle)

// 				server.httpServer = &http.Server{Addr: portSeparator + strconv.Itoa(int(config.Server.Port)), Handler: mux}
// 				return server, nil
// 			}
// 		}
// 	}
// 	return nil, err
// }

// // Stops the server.
// func (serv *Server) stopHandle(writer http.ResponseWriter, request *http.Request) {
// 	if request.Method == http.MethodGet {
// 		// Only from current host.
// 		if strings.Contains(request.RemoteAddr, serv.host.IP.String()) {
// 			serv.Stop()
// 		} else {
// 			writer.WriteHeader(http.StatusForbidden)
// 		}
// 	} else {
// 		writer.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

// // Returns information about the current host.
// func (serv *Server) hostHandle(writer http.ResponseWriter, request *http.Request) {
// 	if request.Method == http.MethodGet {
// 		data, err := json.Marshal(serv.host)
// 		if err == nil {
// 			_, err = writer.Write(data)
// 		}

// 		if err != nil {
// 			fmt.Println(err) // TODO. Log format

// 			writer.Write([]byte(err.Error()))
// 			writer.WriteHeader(http.StatusInternalServerError)
// 		}
// 	} else {
// 		writer.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

// // Returns information about file or directory.
// func (serv *Server) fileInfoHandle(writer http.ResponseWriter, request *http.Request) {
// 	if request.Method == netfs.API.FileInfo.Method {
// 		query := request.URL.Query()
// 		if path := query.Get(netfs.API.FileInfo.Path); path != "" {
// 			file, err := os.Open(path)
// 			if err == nil {
// 				defer file.Close()

// 				var info os.FileInfo
// 				if info, err = file.Stat(); err == nil {
// 					fileType := netfs.FILE
// 					if info.IsDir() {
// 						fileType = netfs.DIRECTORY
// 					}

// 					var data []byte
// 					data, err = json.Marshal(netfs.RemoteFile{Host: serv.host, Name: info.Name(), Path: path, Type: fileType, Size: uint64(info.Size())})
// 					if err == nil {
// 						writer.Write(data)
// 					}
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
