package api

import (
	"net"
	"netfs/api/transport"
	"strconv"
)

const rootDirectory = "/"

// Information about host.
type RemoteHost struct {
	Name string
	IP   net.IP
}

// The function returns the root directory of the remote host.
func (host *RemoteHost) Root() *RemoteFile {
	return &RemoteFile{Host: *host, Info: FileInfo{Id: rootDirectory, Path: rootDirectory}}
}

// The function creates a file or directory on the remote host.
func (host *RemoteHost) Create(client transport.TransportSender, info FileInfo, replace bool) (*RemoteFile, error) {
	params := []string{
		Endpoints.FileCreate.Replace, strconv.FormatBool(replace),
	}

	req, err := client.NewRequest(host.IP, Endpoints.FileCreate.Name, params, nil, info)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			info := &FileInfo{}
			if _, err = res.Body(info); err == nil {
				return &RemoteFile{Info: *info, Host: *host}, nil
			}
		}
	}
	return nil, err
}

// The function returns information about a file by id.
func (host *RemoteHost) File(client transport.TransportSender, fileId FileId) (*RemoteFile, error) {
	params := []string{
		Endpoints.FileInfo.FileId, string(fileId),
	}
	req, err := client.NewRequest(host.IP, Endpoints.FileInfo.Name, params, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			info := &FileInfo{}
			if _, err = res.Body(info); err == nil {
				return &RemoteFile{Info: *info, Host: *host}, nil
			}
		}
	}
	return nil, err
}

// The function returns information about all tasks.
func (host RemoteHost) Tasks(client transport.TransportSender) ([]RemoteCopyTask, error) {
	req, err := client.NewRequest(host.IP, Endpoints.FileCopy, nil, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			tasks := []RemoteCopyTask{}
			if _, err = res.Body(&tasks); err == nil {
				return tasks, nil
			}
		}
	}
	return nil, err
}

// The function returns information about a task by id.
func (host RemoteHost) Task(client transport.TransportSender, taskId TaskId) (*RemoteCopyTask, error) {
	params := []string{Endpoints.FileCopyStatus.TaskId, string(taskId)}
	req, err := client.NewRequest(host.IP, Endpoints.FileCopyStatus.Name, params, nil, nil)

	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			task := &RemoteCopyTask{}
			if _, err = res.Body(task); err == nil {
				return task, nil
			}
		}
	}
	return nil, err
}
