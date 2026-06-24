package api

import (
	"netfs/api/transport"
)

// Status of the task.
type TaskStatus uint8

const (
	Failed TaskStatus = iota
	Running
	Cancelled
	Completed
)

// The task identifier.
type TaskId string

// Netfs server task.
type RemoteCopyTask struct {
	Source   RemoteFile
	Target   RemoteFile
	Host     RemoteHost
	Id       TaskId
	Error    error
	Progress int
	Count    int
	Current  int
	Status   TaskStatus
}

// Cancels the current task.
func (tsk *RemoteCopyTask) Cancel(client transport.TransportSender) error {
	params := []string{Endpoints.FileCopyCancel.TaskId, string(tsk.Id)}
	req, err := client.NewRequest(tsk.Host.IP, Endpoints.FileCopyCancel.Name, params, nil, nil)

	if err == nil {
		_, err = client.Send(req)
	}
	return err
}
