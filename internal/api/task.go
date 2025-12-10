package api

import (
	"netfs/internal/api/transport"
	"strconv"
)

// Status of the task.
type TaskStatus uint8

const (
	Waiting TaskStatus = iota
	Stopped
	Failed
	Running
	Completed
)

// Netfs server task.
type RemoteTask struct {
	Id       int
	Progress int8
	Status   TaskStatus
	Host     RemoteHost
}

// Refreshes the task data.
func (tsk *RemoteTask) Refresh(client transport.TransportSender) error {
	params := []string{Endpoints.FileCopyStatus.Id, strconv.Itoa(tsk.Id)}
	req, err := client.NewRequest(tsk.Host.IP, Endpoints.FileCopyStatus.Name, params, nil, nil)

	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			_, err = res.Body(tsk)
		}
	}
	return err
}
