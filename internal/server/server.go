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
	"path/filepath"
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
	srv.receiver.Receive(api.Endpoints.FileChildren.Name, srv.FileChildrenHandle)
	srv.receiver.Receive(api.Endpoints.FileCreate, srv.FileCreateHandle)
	srv.receiver.Receive(api.Endpoints.FileWrite.Name, srv.FileWriteHandle)
	srv.receiver.Receive(api.Endpoints.FileRemove.Name, srv.FileRemoveHandle)
	srv.receiver.Receive(api.Endpoints.FileCopyStart, srv.FileCopyStartHandle)
	srv.receiver.Receive(api.Endpoints.FileCopyStatus.Name, srv.FileCopyStatusHandle)
	srv.receiver.Receive(api.Endpoints.FileCopyStop.Name, srv.FileCopyCancelHandle)
	srv.receiver.Receive(api.Endpoints.Volume.Name, srv.VolumeHandle)
	srv.receiver.Receive(api.Endpoints.VolumeChildren.Name, srv.VolumeChildrenHandle)

	// TODO. Remove it, for test only
	path, _ := filepath.Abs("D:/andrey/documents")
	vl, _ := srv.volumes.Create(api.VolumeInfo{Name: "Test", LocalPath: "test:/", OsPath: path})
	vl.ReIndex()

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

// The function handles request and returns volumes.
func (srv *Server) VolumeHandle(req transport.Request) ([]byte, any, error) {
	var err error
	var result []api.VolumeInfo

	volumeId, _ := req.ParamUInt64(api.Endpoints.Volume.VolumeId)
	if volumeId != 0 {
		srv.log.Info("VolumeHandle() volumeId: %v", volumeId)

		var volume volume.Volume
		if volume, err = srv.volumes.Volume(volumeId); err == nil {
			result = []api.VolumeInfo{volume.Info()}
		}
	} else {
		if volumes, err := srv.volumes.Volumes(); err == nil {
			result = make([]api.VolumeInfo, len(volumes))
			for index, volume := range volumes {
				result[index] = volume.Info()
			}
		}
	}

	if err != nil {
		srv.log.Error("VolumeHandle() err: %v", err)
	}
	srv.log.Info("VolumeHandle() volumes: %v", result)
	return nil, result, err
}

// The function handles request and returns elements of the volume.
func (srv *Server) VolumeChildrenHandle(req transport.Request) ([]byte, any, error) {
	var children []api.FileInfo

	volumeId, volumeIdErr := req.ParamUInt64(api.Endpoints.VolumeChildren.VolumeId)
	skip, skipErr := req.ParamInt(api.Endpoints.VolumeChildren.Skip)
	limit, limitErr := req.ParamInt(api.Endpoints.VolumeChildren.Limit)

	err := errors.Join(volumeIdErr, skipErr, limitErr)
	if err == nil {
		srv.log.Info("VolumeChildren() volumeId: %v, skip: %v, limit: %v", volumeId, skip, limit)

		var volume volume.Volume
		if volume, err = srv.volumes.Volume(volumeId); err == nil {
			children, err = volume.Children(0, skip, limit)
		}
	}

	if err != nil {
		srv.log.Error("VolumeChildren() err: %v", err)
	}
	srv.log.Info("VolumeChildren() children: %v", len(children))
	return nil, children, err
}

// The function handles request and returns information about the file.
func (srv *Server) FileInfoHandle(req transport.Request) ([]byte, any, error) {
	var info *api.FileInfo

	volumeId, volumeIdErr := req.ParamUInt64(api.Endpoints.FileInfo.VolumeId)
	fileId, fileIdErr := req.ParamUInt64(api.Endpoints.FileInfo.FileId)

	err := errors.Join(volumeIdErr, fileIdErr)
	if err == nil {
		srv.log.Info("FileInfoHandle() volumeId: %v, fileId: %v", volumeId, fileId)

		var volume volume.Volume
		if volume, err = srv.volumes.Volume(volumeId); err == nil {
			info, err = volume.File(fileId)
		}
	}

	if err != nil {
		srv.log.Error("FileInfoHandle() err: %v", err)
	}
	srv.log.Info("FileInfoHandle() info: %v", *info)
	return nil, info, err
}

// The function handles request and returns children of the directory.
func (srv *Server) FileChildrenHandle(req transport.Request) ([]byte, any, error) {
	var children []api.FileInfo

	volumeId, volumeIdErr := req.ParamUInt64(api.Endpoints.FileChildren.VolumeId)
	fileId, fileIdErr := req.ParamUInt64(api.Endpoints.FileChildren.FileId)
	skip, skipErr := req.ParamInt(api.Endpoints.FileChildren.Skip)
	limit, limitErr := req.ParamInt(api.Endpoints.FileChildren.Limit)

	err := errors.Join(volumeIdErr, fileIdErr, skipErr, limitErr)
	if err == nil {
		srv.log.Info("FileChildrenHandle() volumeId: %v, fileId: %v, skip: %v, limit: %v", volumeId, fileId, skip, limit)

		var volume volume.Volume
		if volume, err = srv.volumes.Volume(volumeId); err == nil {
			children, err = volume.Children(fileId, skip, limit)
		}
	}

	if err != nil {
		srv.log.Error("FileChildrenHandle() err: %v", err)
	}
	srv.log.Info("FileChildrenHandle() children: %v", len(children))
	return nil, children, err
}

// The function handles request and creates a new file or directory by api.FileInfo.
func (srv *Server) FileCreateHandle(req transport.Request) ([]byte, any, error) {
	info := &api.FileInfo{}
	_, err := req.Body(info)
	if err == nil {
		srv.log.Info("FileCreateHandle() file: %v", *info)

		var vl volume.Volume
		vl, err = srv.volumes.Volume(info.VolumeId)
		if err == nil {
			info, err = vl.Create(info)
		}
	}

	if err != nil {
		srv.log.Error("FileCreateHandle() err: %v", err)
	} else {
		srv.log.Info("FileCreateHandle() created: %v", *info)
	}
	return nil, info, err
}

// The function handles request and writes data to a file.
func (srv *Server) FileWriteHandle(req transport.Request) ([]byte, any, error) {
	volumeId, volumeIdErr := req.ParamUInt64(api.Endpoints.FileInfo.VolumeId)
	fileId, fileIdErr := req.ParamUInt64(api.Endpoints.FileInfo.FileId)

	err := errors.Join(volumeIdErr, fileIdErr)
	if err == nil {
		srv.log.Info("FileWriteHandle() volumeId: %v, fileId: %v", volumeId, fileId)

		var vl volume.Volume
		if vl, err = srv.volumes.Volume(volumeId); err == nil {
			err = vl.Write(fileId, req.RawBody())
		}
	}

	if err != nil {
		srv.log.Error("FileWriteHandle() err: %v", err)
	}
	return nil, nil, err
}

// The function handles request and removes the file.
func (srv *Server) FileRemoveHandle(req transport.Request) ([]byte, any, error) {
	volumeId, volumeIdErr := req.ParamUInt64(api.Endpoints.FileInfo.VolumeId)
	fileId, fileIdErr := req.ParamUInt64(api.Endpoints.FileInfo.FileId)

	err := errors.Join(volumeIdErr, fileIdErr)
	if err == nil {
		srv.log.Info("FileRemoveHandle() volumeId: %v, fileId: %v", volumeId, fileId)

		var vl volume.Volume
		if vl, err = srv.volumes.Volume(volumeId); err == nil {
			err = vl.Remove(fileId)
		}
	}

	if err != nil {
		srv.log.Error("FileRemoveHandle() err: %v", err)
	}
	return nil, nil, err
}

// The function handles request and starts a new task to copy the file or directory.
func (srv *Server) FileCopyStartHandle(req transport.Request) ([]byte, any, error) {
	var result api.RemoteTask

	files := &[]api.RemoteFile{}
	_, err := req.Body(files)
	if err == nil {
		srv.log.Info("FileCopyStartHandle() files: %v", *files)

		var copyTask *task.CopyTask
		copyTask, err = task.NewCopyTask((*files)[0], (*files)[1])
		if err == nil {
			srv.log.Info("FileCopyStartHandle() task: %v", *copyTask)

			var taskId int
			if taskId, err = srv.tasks.SetTask(copyTask); err == nil {
				result = api.RemoteTask{Id: taskId, Status: copyTask.Status, Host: srv.network.LocalHost()}
			}
		}
	}

	if err != nil {
		srv.log.Error("FileCopyStartHandle() err: %v", err)
	} else {
		srv.log.Info("FileCopyStartHandle() task: %v", result)
	}
	return nil, result, err
}

// The function handles request and returns status of the task.
func (srv *Server) FileCopyStatusHandle(req transport.Request) ([]byte, any, error) {
	var result api.RemoteTask

	taskId, err := req.ParamUInt64(api.Endpoints.FileCopyStatus.TaskId)
	if err == nil {
		srv.log.Info("FileCopyStatusHandle() taskId: %v", taskId)

		var task task.Task
		if task, err = srv.tasks.GetTask(int(taskId)); err == nil { // TODO. int(taskId)
			result = api.RemoteTask{Id: task.TaskId(), Status: task.TaskStatus(), Host: srv.network.LocalHost()}
		}
	}

	if err != nil {
		srv.log.Error("FileCopyStatusHandle() err: %v", err)
	} else {
		srv.log.Info("FileCopyStatusHandle() task: %v", result)
	}
	return nil, result, err
}

// The function handles request and stops the task.
func (srv *Server) FileCopyCancelHandle(req transport.Request) ([]byte, any, error) {
	taskId, err := req.ParamUInt64(api.Endpoints.FileCopyStop.TaskId)
	if err == nil {
		srv.log.Info("FileCopyCancelHandle() taskId: %v", taskId)
		err = srv.tasks.CancelTask(int(taskId)) // TODO. int(taskId)
	}

	if err != nil {
		srv.log.Error("FileCopyCancelHandle() err: %v", err)
	}
	return nil, nil, err
}
