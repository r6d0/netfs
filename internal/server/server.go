package server

import (
	"errors"
	"netfs/api"
	"netfs/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"netfs/internal/server/volume"
	"os"
	"os/signal"
	"strconv"
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
	tasks    task.TaskManager
	volumes  volume.VolumeManager
	stop     chan os.Signal
}

// Starts the netfs server.
func (srv *Server) Start() error {
	srv.receiver.Receive(api.Endpoints.ServerStop, srv.StopServerHandle)
	srv.receiver.Receive(api.Endpoints.ServerHost, srv.ServerHostHandle)
	srv.receiver.Receive(api.Endpoints.FileInfo.Name, srv.FileInfoHandle)
	srv.receiver.Receive(api.Endpoints.FileCreate, srv.FileCreateHandle)
	srv.receiver.Receive(api.Endpoints.FileWrite.Name, srv.FileWriteHandle)
	srv.receiver.Receive(api.Endpoints.FileRemove.Name, srv.FileRemoveHandle)
	srv.receiver.Receive(api.Endpoints.FileCopyStart, srv.FileCopyStartHandle)
	srv.receiver.Receive(api.Endpoints.FileCopyStatus.Name, srv.FileCopyStatusHandle)
	srv.receiver.Receive(api.Endpoints.FileCopyStop.Name, srv.FileCopyCancelHandle)

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
				var tasks task.TaskManager
				if tasks, err = task.NewTaskManager(config.Task, db, volumes, network.Transport(), log); err == nil {
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
func (srv *Server) StopServerHandle(req transport.Request) ([]byte, any, error) { // TODO. Check current host.
	return nil, nil, srv.Stop()
}

// Returns information about the current host.
func (srv *Server) ServerHostHandle(req transport.Request) ([]byte, any, error) {
	return nil, srv.network.LocalHost(), nil
}

// Returns information about file.
func (srv *Server) FileInfoHandle(req transport.Request) ([]byte, any, error) { // TODO. add validation
	path := req.Param(api.Endpoints.FileInfo.Path)
	volume, err := srv.volumes.Volume(path)
	if err == nil {
		var info *api.FileInfo
		info, err = volume.Info(path)
		return nil, info, err
	}
	return nil, nil, err
}

// Creates new file or directory by api.FileInfo.
func (srv *Server) FileCreateHandle(req transport.Request) ([]byte, any, error) {
	info := &api.FileInfo{}
	_, err := req.Body(info)
	if err == nil {
		var vl volume.Volume
		vl, err = srv.volumes.Volume(info.FilePath)
		if err == nil {
			err = vl.Create(info)
		}
	}
	return nil, nil, err
}

// Writes data to file.
func (srv *Server) FileWriteHandle(req transport.Request) ([]byte, any, error) {
	path := req.Param(api.Endpoints.FileWrite.Path)
	volume, err := srv.volumes.Volume(path)
	if err == nil {
		err = volume.Write(path, req.RawBody())
	}
	return nil, nil, err
}

// Removes file or directory.
func (srv *Server) FileRemoveHandle(req transport.Request) ([]byte, any, error) {
	path := req.Param(api.Endpoints.FileRemove.Path)
	volume, err := srv.volumes.Volume(path)
	if err == nil {
		err = volume.Remove(path)
	}
	return nil, nil, err
}

// Starts a new task to copy the file or directory.
func (srv *Server) FileCopyStartHandle(req transport.Request) ([]byte, any, error) { // TODO. Check len(files) == 2
	files := &[]api.RemoteFile{}
	_, err := req.Body(files)
	if err == nil {
		var copyTask *task.CopyTask
		copyTask, err = task.NewCopyTask((*files)[0], (*files)[1])
		if err == nil {
			var taskId int
			if taskId, err = srv.tasks.SetTask(copyTask); err == nil {
				return nil, api.RemoteTask{Id: taskId, Status: copyTask.Status, Host: srv.network.LocalHost()}, err
			}
		}
	}
	return nil, nil, err
}

// Returns status of the task.
func (srv *Server) FileCopyStatusHandle(req transport.Request) ([]byte, any, error) {
	param := req.Param(api.Endpoints.FileCopyStatus.Id)
	id, err := strconv.Atoi(param)
	if err == nil {
		var task task.Task
		if task, err = srv.tasks.GetTask(id); err == nil {
			return nil, api.RemoteTask{Id: id, Status: task.TaskStatus(), Host: srv.network.LocalHost()}, nil
		}
	}
	return nil, nil, err
}

// Stops the task.
func (srv *Server) FileCopyCancelHandle(req transport.Request) ([]byte, any, error) {
	param := req.Param(api.Endpoints.FileCopyStop.Id)
	id, err := strconv.Atoi(param)
	if err == nil {
		err = srv.tasks.CancelTask(id)
	}
	return nil, nil, err
}
