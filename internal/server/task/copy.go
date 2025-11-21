package task

import (
	"encoding/json"
	"errors"
	"io"
	"netfs/internal/api"
	"netfs/internal/server/database"
	"os"
)

type TaskCopyConfig struct {
	BufferSize uint64
}

// The task of copying the file.
type CopyTask struct {
	Id     uint64
	Status TaskStatus
	Type   TaskType
	Offset int64
	Source api.RemoteFile
	Target api.RemoteFile
	Error  string
}

func (task *CopyTask) TaskType() TaskType {
	return task.Type
}

func (task *CopyTask) Init(TaskExecuteContext) error {
	task.Status = Waiting
	return nil
}

func (task *CopyTask) BeforeExecute(TaskExecuteContext) error {
	task.Status = Running
	return nil
}

func (task *CopyTask) Execute(ctx TaskExecuteContext) error {
	config := ctx.Config.Copy
	source := task.Source
	target := task.Target

	err := target.Create(ctx.Transport)
	if err == nil {
		var file *os.File
		if file, err = os.Open(source.Path); err == nil { // TODO. Use volume.OpenFile(source.Path) for reuse the opened file already
			defer file.Close()

			var info os.FileInfo
			if info, err = file.Stat(); err == nil {
				fileSize := uint64(info.Size())
				bufferSize := min(fileSize, config.BufferSize)
				if fileSize > 0 && bufferSize > 0 {
					buffer := make([]byte, bufferSize)

					var size int
					if size, err = file.ReadAt(buffer, task.Offset); err == nil {
						if err = target.Write(ctx.Transport, buffer[:size]); err == nil {
							task.Status = Waiting
							task.Offset += int64(size)
						}
					} else if errors.Is(err, io.EOF) {
						err = nil
						task.Status = Completed
					}
				}
			}
		}
	}

	if err != nil {
		task.Status = Failed
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

// Converts a database record to a task.
func CopyTaskFromRecord(record database.Record) (*CopyTask, error) {
	task := &CopyTask{}
	data := record.GetField(uint8(Payload))
	if err := json.Unmarshal(data, task); err != nil {
		return nil, err
	}
	return task, nil
}

// Converts a task to a database record.
func CopyTaskToRecord(task *CopyTask) (database.Record, error) {
	var record database.Record
	data, err := json.Marshal(task)
	if err == nil {
		record = database.NewRecord(uint8(Payload) + 1)
		record.SetRecordId(task.Id) // TODO. Database must generate id itself.
		record.SetUint64(uint8(Id), task.Id)
		record.SetUint8(uint8(Status), uint8(task.Status))
		record.SetUint8(uint8(Type), uint8(task.Type))
		record.SetField(uint8(Payload), data)
	}
	return record, err
}
