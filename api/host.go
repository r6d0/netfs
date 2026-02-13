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

// The function returns volumes of the host.
func (host RemoteHost) Volumes(client transport.TransportSender) ([]RemoteVolume, error) {
	req, err := client.NewRequest(host.IP, Endpoints.Volume, nil, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			volumes := []VolumeInfo{}
			if _, err = res.Body(&volumes); err == nil {
				result := make([]RemoteVolume, len(volumes))
				for index, volume := range volumes {
					result[index] = RemoteVolume{Info: volume, Host: host}
				}
				return result, nil
			}
		}
	}
	return nil, err
}

// The function returns information about a task by id.
func (host RemoteHost) Task(client transport.TransportSender, taskId int) (*RemoteTask, error) {
	params := []string{Endpoints.FileCopyStatus.TaskId, strconv.Itoa(taskId)}
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
