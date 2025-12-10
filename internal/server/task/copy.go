package task

import (
	"encoding/json"
	"netfs/internal/api"
	"netfs/internal/server/database"
	"netfs/internal/server/volume"
)

type TaskCopyConfig struct {
	BufferSize int64
}

// The task of copying the file.
type CopyTask struct {
	Id     uint64
	Status api.TaskStatus
	Type   TaskType
	Offset int64
	Source api.RemoteFile
	Target api.RemoteFile
	Error  string
}

func (task *CopyTask) TaskId() int {
	return int(task.Id) // TODO. remove casting.
}

func (task *CopyTask) TaskStatus() api.TaskStatus {
	return task.Status
}

func (task *CopyTask) TaskType() TaskType {
	return task.Type
}

func (task *CopyTask) Init(TaskExecuteContext) error {
	task.Status = api.Waiting
	return nil
}

func (task *CopyTask) BeforeExecute(TaskExecuteContext) error {
	task.Status = api.Running
	return nil
}

func (task *CopyTask) Execute(ctx TaskExecuteContext) error {
	config := ctx.Config.Copy
	source := task.Source
	target := task.Target

	err := target.Create(ctx.Transport)
	if err == nil {
		size := config.BufferSize
		path := source.Info.FilePath

		var vl volume.Volume
		if vl, err = ctx.Volumes.Volume(path); err == nil {
			var buffer []byte
			if buffer, err = vl.Read(path, task.Offset, size); err == nil {
				if err = target.Write(ctx.Transport, buffer); err == nil {
					task.Status = api.Waiting
					task.Offset += int64(len(buffer))
				}
			}

			if len(buffer) < int(size) {
				task.Status = api.Completed
			}
		}
	}

	if err != nil {
		task.Status = api.Failed
		task.Error = err.Error()
	}
	return err
}

func (*CopyTask) AfterExecute(TaskExecuteContext) error {
	return nil
}

// Creates a new instance of the task.
func NewCopyTask(source api.RemoteFile, target api.RemoteFile) (*CopyTask, error) {
	return &CopyTask{Type: Copy, Source: source, Target: target}, nil
}

func copyTaskFromRecord(record database.Record) (*CopyTask, error) {
	task := &CopyTask{}
	data := record.GetField(Payload)
	if err := json.Unmarshal(data, task); err != nil {
		return nil, err
	}
	return task, nil
}

func copyTaskToRecord(table database.Table, task *CopyTask) (database.Record, error) {
	if task.Id == 0 {
		task.Id = table.NextId()
	}

	var record database.Record
	data, err := json.Marshal(task)
	if err == nil {
		record = database.NewRecord(3)
		record.SetRecordId(task.Id)
		record.SetUint8(Status, uint8(task.Status))
		record.SetUint8(Type, uint8(task.Type))
		record.SetField(Payload, data)
	}
	return record, err
}
