package task

import (
	"errors"
	"netfs/internal/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server/database"
	"time"
)

// If the type is unknown.
var ErrUnknownTaskType = errors.New("unknown a type of the task")

// Tasks executor configuration.
type TaskExecuteConfig struct {
	Copy               TaskCopyConfig
	MaxAvailableTasks  uint16
	TasksWaitingSecond uint64
}

// Task fields.
type TaskRecordField uint8

// Status of the task.
type TaskStatus uint8

// Type of the task.
type TaskType uint8

const (
	Id TaskRecordField = iota
	Status
	Type
	Payload

	Waiting TaskStatus = iota
	Stopped
	Failed
	Running
	Completed

	Copy TaskType = iota
)

// The context for the task.
type TaskExecuteContext struct {
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
	if TaskType(record.GetUint8(uint8(Type))) == Copy {
		return CopyTaskFromRecord(record)
	}
	return nil, ErrUnknownTaskType
}

// Converts a task to a database record.
func TaskToRecord(task Task) (database.Record, error) {
	if task.TaskType() == Copy {
		return CopyTaskToRecord(task.(*CopyTask))
	}
	return nil, ErrUnknownTaskType
}

// The tasks executor.
type TaskExecutor interface {
	Submit(Task) error
	Start() error
	Stop() error
}

func NewTaskExecutor(config TaskExecuteConfig, db database.Database, transport transport.TransportSender, log *logger.Logger) (TaskExecutor, error) {
	return &taskExecutor{config: config, log: log, db: db, transport: transport, cancel: make(chan bool)}, nil
}

type taskExecutor struct {
	config    TaskExecuteConfig
	log       *logger.Logger
	db        database.Database
	transport transport.TransportSender
	cancel    chan bool
}

func (exec *taskExecutor) Submit(task Task) error {
	log := exec.log
	err := task.Init(TaskExecuteContext{Config: exec.config, Transport: exec.transport})
	if err == nil {
		var record database.Record
		if record, err = TaskToRecord(task); err == nil {
			err = exec.db.Set(record)
		}
	}

	if err != nil {
		log.Error("task: [%s] is failed: [%s]", task, err.Error())
	} else {
		log.Info("task: [%v] is waiting", task)
	}
	return err
}

func (exec *taskExecutor) Start() error {
	exec.log.Info("task executor is staring")

	maxAvailableTasks := exec.config.MaxAvailableTasks
	tasksWaitingSecond := time.Duration(exec.config.TasksWaitingSecond)

	go func(cancel chan bool) {
		available := maxAvailableTasks
		complete := make(chan Task)
		ctx := TaskExecuteContext{Config: exec.config, Transport: exec.transport}
		condition := database.Equals(uint16(Status), []byte{byte(Waiting)})

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
				if record, convErr := TaskToRecord(task); convErr == nil {
					err = errors.Join(err, exec.db.Set(record))
				}

				if err != nil {
					exec.log.Error("task [%s] after execution error: [%s]", task, err.Error())
				}
				available++
			default:
				if !cancelled && available > 0 {
					records, err := exec.db.Get(condition, database.Limit(available))
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
								record, err = TaskToRecord(task)
							}

							if err == nil {
								err = exec.db.Set(record)
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
