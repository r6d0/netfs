package task

import (
	"encoding/json"
	netfs "netfs/internal"
	"netfs/internal/logger"
	"netfs/internal/store"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const collName = "task"
const collSep = "."

type TaskConfig struct {
	MaxAvailableTasks  uint64
	TasksWaitingSecond uint64
}

// Status of task.
type TaskStatus uint8

// type of task.
type TaskType uint8

const (
	Waiting TaskStatus = iota
	Stopped
	Failed
	Running
	Completed

	Copy TaskType = iota
)

type Task interface {
	Id() string
	Process(chan bool, chan string) error
	Status() TaskStatus
	Type() TaskType
}

type TaskExecutor interface {
	Submit(Task) error
	Start() error
	Stop() error
}

func NewTaskExecutor(config TaskConfig, db store.Store, log *logger.Logger) (TaskExecutor, error) {
	return &taskExecutor{config: config, log: log, db: db, cancel: make(chan bool)}, nil
}

// Creates a new instance of the file copy task.
func NewCopyTask(source netfs.RemoteFile, target netfs.RemoteFile) (Task, error) {
	task := taskCopy{Source: source, Target: target}
	task.TaskId = uuid.NewString()
	task.TaskStatus = Waiting
	task.TaskType = Copy

	return &task, nil
}

// The task of copying the file.
type taskCopy struct {
	TaskId     string
	TaskStatus TaskStatus
	TaskType   TaskType
	Offset     uint64
	Source     netfs.RemoteFile
	Target     netfs.RemoteFile
	Error      string
}

func (task taskCopy) Id() string {
	return task.TaskId
}

func (*taskCopy) Process(cancel chan bool, complete chan string) error {
	return nil
}

func (task *taskCopy) Status() TaskStatus {
	return task.TaskStatus
}

func (task *taskCopy) Type() TaskType {
	return task.TaskType
}

type taskExecutor struct {
	config TaskConfig
	log    *logger.Logger
	db     store.Store
	cancel chan bool
}

func (exec *taskExecutor) Submit(task Task) error {
	log := exec.log
	taskType := strconv.Itoa(int(task.Type()))
	taskStatus := strconv.Itoa(int(task.Status()))

	key := strings.Join([]string{collName, taskStatus, taskType, task.Id()}, collSep)
	log.Info("submit task key: [%s]", key)

	value, err := json.Marshal(task)
	if err == nil {
		log.Info("submit task value: [%s]", string(value))
		err = exec.db.Set([]byte(key), value)
	}

	if err != nil {
		log.Error("submit task error: [%s]", err.Error())
	}
	return err
}

func (exec *taskExecutor) Start() error {
	exec.log.Info("task executor is staring")

	maxAvailableTasks := exec.config.MaxAvailableTasks
	tasksWaitingSecond := time.Duration(exec.config.TasksWaitingSecond)

	go func(cancel chan bool) {
		prefix := strings.Join([]string{collName, strconv.Itoa(int(Waiting))}, collSep)
		available := maxAvailableTasks
		complete := make(chan string)

		cancelled := false
		for {
			select {
			case cancelled = <-cancel:
				if cancelled {
					exec.log.Info("tasks: [%d] are cancelled", maxAvailableTasks-available)
				}
			case id := <-complete:
				exec.log.Info("task: [%s] is completed", id)

				available++
				if cancelled && available == maxAvailableTasks {
					return
				}
			default:
				if !cancelled {
					items, err := exec.db.All([]byte(prefix), available)
					if err == nil {
						exec.log.Info("tasks: [%d] are found", len(items))

						if available == maxAvailableTasks && len(items) == 0 {
							exec.log.Info("task executor is waiting: [%d] seconds", tasksWaitingSecond)
							time.Sleep(tasksWaitingSecond * time.Second)
						}

						for _, item := range items {
							key := string(item.Key)
							exec.log.Info("task: [%s] is preparing to execution", key)

							var taskType int
							if taskType, err = strconv.Atoi(strings.Split(key, collSep)[2]); err == nil {
								var task Task
								if taskType == int(Copy) {
									task = &taskCopy{}
								}

								if err = json.Unmarshal(item.Value, task); err == nil {
									exec.log.Info("task: [%s] has data: [%s]", key, task)
									err = task.Process(cancel, complete)
								}
							}

							if err == nil {
								available--
							} else {
								exec.log.Error("task execution error: [%s]", err.Error())
							}
						}
					} else {
						exec.log.Error("task execution error: [%s]", err.Error())
					}
				}
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
