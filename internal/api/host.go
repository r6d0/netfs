package api

import (
	"net"
	"netfs/internal/api/transport"
	"strconv"
)

// Information about host.
type RemoteHost struct {
	Name string
	IP   net.IP
}

// Returns information about file by path.
func (host RemoteHost) File(client transport.TransportSender, path string) (*RemoteFile, error) {
	parameters := []string{Endpoints.FileInfo.Path, path}
	req, err := client.NewRequest(host.IP, Endpoints.FileInfo.Name, parameters, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			info := &FileInfo{}
			if _, err = res.Body(info); err == nil {
				return &RemoteFile{Info: *info, Host: host}, nil
			}
		}
	}
	return nil, err
}

// Returns information about task by id.
func (host RemoteHost) Task(client transport.TransportSender, taskId int) (*RemoteTask, error) {
	params := []string{Endpoints.FileCopyStatus.Id, strconv.Itoa(taskId)}
	req, err := client.NewRequest(host.IP, Endpoints.FileCopyStatus.Name, params, nil, nil)

	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			task := &RemoteTask{}
			if _, err = res.Body(task); err == nil {
				return task, nil
			}
		}
	}
	return nil, err
}
