package api

import (
	"netfs/api/transport"
	"strconv"
)

// Status of the task.
type TaskStatus uint8

const (
	Waiting TaskStatus = iota
	Cancelled
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

// Cancels the current task.
func (tsk *RemoteTask) Cancel(client transport.TransportSender) error {
	params := []string{Endpoints.FileCopyStop.Id, strconv.Itoa(tsk.Id)}
	req, err := client.NewRequest(tsk.Host.IP, Endpoints.FileCopyStop.Name, params, nil, nil)

	if err == nil {
		_, err = client.Send(req)
	}
	return err
}
