package task

import (
	"errors"
	"netfs/internal/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server/database"
	"netfs/internal/server/volume"
	"time"
)

// The name of the task table in the database.
const TaskTable = "task"

const (
	Status database.RecordField = iota
	Type
	Payload
)

// If the type is unknown.
var ErrUnknownTaskType = errors.New("unknown a type of the task")

// Tasks executor configuration.
type TaskExecuteConfig struct {
	Copy               TaskCopyConfig
	MaxAvailableTasks  int16
	TasksWaitingSecond uint64
}

// Task fields.
type TaskRecordField uint8

// Status of the task.
type TaskStatus uint8

// Type of the task.
type TaskType uint8

const (
	Waiting TaskStatus = iota
	Stopped
	Failed
	Running
	Completed

	Copy TaskType = iota
)

// The context for the task.
type TaskExecuteContext struct {
	Volumes   volume.VolumeManager
	Transport transport.TransportSender
	Config    TaskExecuteConfig
}

// The task abstraction.
type Task interface {
	TaskType() TaskType
	Init(TaskExecuteContext) error
	BeforeExecute(TaskExecuteContext) error
	Execute(TaskExecuteContext) error
	AfterExecute(TaskExecuteContext) error
}

// Converts a database record to a task.
func TaskFromRecord(record database.Record) (Task, error) {
	if TaskType(record.GetUint8(Type)) == Copy {
		return CopyTaskFromRecord(record)
	}
	return nil, ErrUnknownTaskType
}

// Converts a task to a database record.
func TaskToRecord(table database.Table, task Task) (database.Record, error) {
	if task.TaskType() == Copy {
		return CopyTaskToRecord(table, task.(*CopyTask))
	}
	return nil, ErrUnknownTaskType
}

// The tasks executor.
type TaskExecutor interface {
	Submit(Task) (uint64, error)
	Start() error
	Stop() error
}

func NewTaskExecutor(config TaskExecuteConfig, db database.Database, volumes volume.VolumeManager, transport transport.TransportSender, log *logger.Logger) (TaskExecutor, error) {
	return &taskExecutor{config: config, log: log, db: db, volumes: volumes, transport: transport, cancel: make(chan bool)}, nil
}

type taskExecutor struct {
	config    TaskExecuteConfig
	log       *logger.Logger
	db        database.Database
	volumes   volume.VolumeManager
	transport transport.TransportSender
	cancel    chan bool
}

func (exec *taskExecutor) Submit(task Task) (uint64, error) {
	var taskId uint64
	var record database.Record

	log := exec.log
	err := task.Init(TaskExecuteContext{Config: exec.config, Transport: exec.transport, Volumes: exec.volumes})
	if err == nil {
		table := exec.db.Table(TaskTable)
		if record, err = TaskToRecord(table, task); err == nil {
			err = table.Set(record)
		}
	}

	if err != nil {
		log.Error("task: [%s] is failed: [%s]", task, err.Error())
	} else {
		taskId = record.GetRecordId()
		log.Info("task: [%v] is waiting", task)
	}
	return taskId, err
}

func (exec *taskExecutor) Start() error {
	exec.log.Info("task executor is staring")

	maxAvailableTasks := exec.config.MaxAvailableTasks
	tasksWaitingSecond := time.Duration(exec.config.TasksWaitingSecond)

	table := exec.db.Table(TaskTable)
	go func(cancel chan bool) {
		available := maxAvailableTasks
		complete := make(chan Task)
		ctx := TaskExecuteContext{Config: exec.config, Transport: exec.transport, Volumes: exec.volumes}
		condition := database.Eq(Status, []byte{byte(Waiting)})

		cancelled := false
		for {
			select {
			case cancelled = <-cancel:
				if cancelled {
					exec.log.Info("tasks: [%d] are cancelled", maxAvailableTasks-available)
				}
			case task := <-complete:
				exec.log.Info("task: [%s] is completed", task)

				err := task.AfterExecute(ctx)
				if record, convErr := TaskToRecord(table, task); convErr == nil {
					err = errors.Join(err, table.Set(record))
				}

				if err != nil {
					exec.log.Error("task [%s] after execution error: [%s]", task, err.Error())
				}
				available++
			default:
				if !cancelled && available > 0 {
					records, err := table.Get(condition, database.Limit(available))
					if err == nil {
						exec.log.Info("tasks: [%d] are found", len(records))

						if available == maxAvailableTasks && len(records) == 0 {
							exec.log.Info("task executor is waiting: [%d] seconds", tasksWaitingSecond)
							time.Sleep(tasksWaitingSecond * time.Second)
						}

						for _, record := range records {
							var task Task
							task, err = TaskFromRecord(record)
							if err == nil {
								err = task.BeforeExecute(ctx)
							}

							if err == nil {
								record, err = TaskToRecord(table, task)
							}

							if err == nil {
								err = table.Set(record)
							}

							if err == nil {
								available--
								go func(complete chan Task) {
									if err := task.Execute(ctx); err != nil {
										exec.log.Error("task [%s] execution error: [%s]", task, err.Error())
									}
									complete <- task
								}(complete)
							} else {
								exec.log.Error("task [%s] before execution error: [%s]", task, err.Error())
							}
						}
					}
				}
			}

			if cancelled && available == maxAvailableTasks {
				return
			}
		}
	}(exec.cancel)
	return nil
}

func (exec *taskExecutor) Stop() error {
	exec.cancel <- true
	exec.log.Info("task executor is stopped")
	return nil
}
