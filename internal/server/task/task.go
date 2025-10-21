package task

import (
	"encoding/json"
	"errors"
	"io"
	netfs "netfs/internal"
	"netfs/internal/logger"
	"netfs/internal/store"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const collName = "task"
const collSep = "."

type TaskCopyConfig struct {
	BufferSize uint64
}

type TaskConfig struct {
	Copy               TaskCopyConfig
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

type TaskProcessContext struct {
	Log      *logger.Logger
	DB       store.Store
	Config   TaskConfig
	Cancel   chan bool
	Complete chan string
}

type Task interface {
	Id() string
	Process(TaskProcessContext) error
	ToWaiting() (store.StoreItem, error)
	ToRunning() (store.StoreItem, error)
	ToFailed(error) (store.StoreItem, error)
	ToCompleted() (store.StoreItem, error)
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
	task.Status = Waiting
	task.Type = Copy

	return &task, nil
}

// The task of copying the file.
type taskCopy struct {
	TaskId string
	Status TaskStatus
	Type   TaskType
	Offset uint64
	Source netfs.RemoteFile
	Target netfs.RemoteFile
	Error  string
}

func (task taskCopy) Id() string {
	return task.TaskId
}

func (task *taskCopy) Process(ctx TaskProcessContext) error {
	go func() {
		log := ctx.Log
		cancel := ctx.Cancel
		db := ctx.DB
		config := ctx.Config.Copy

		var err error
		var file *os.File
		var size int
		var packages uint64
		var buffer []byte
		var item store.StoreItem

		end := false
		cancelled := false
		for !end && !cancelled && err == nil {
			select {
			case cancelled = <-cancel:
				if cancelled {
					log.Info("task: [%s] is cancelled", task.Id())
				}
			default:
				target := task.Target
				if target.Type == netfs.DIRECTORY {
					log.Info("task: [%s] is running", task.Id())

					end = true
					err = target.Create()
				} else {
					source := task.Source

					if file == nil {
						log.Info("task: [%s] is running", task.Id())

						if file, err = os.Open(source.Path); err == nil {
							log.Info("task: [%s]. file: [%s] is opened", task.Id(), source.Path)
							defer file.Close()

							var info os.FileInfo
							if info, err = file.Stat(); err == nil {
								fileSize := uint64(info.Size())
								bufferSize := min(fileSize, config.BufferSize)
								if fileSize > 0 && bufferSize > 0 {
									packages = fileSize / bufferSize
									buffer = make([]byte, bufferSize)
								}
								log.Info("task: [%s]. file: [%s] has [%d] packages to send", task.Id(), source.Path, packages)
							}
						}
					} else {
						if size, err = file.Read(buffer); err == nil {
							if err = target.Write(buffer[:size]); err == nil {
								task.Offset += uint64(size)

								packages--
								log.Info("task: [%s]. package of file: [%s] has been sent. [%d] left", task.Id(), source.Path, packages)
							}
						} else if errors.Is(err, io.EOF) {
							end = true
							err = nil
						}
					}
				}

			}

			if err == nil {
				if item, err = task.ToRunning(); err == nil {
					err = db.Set(item)
				}
			}
		}

		if end && err == nil { // completed
			if item, err = task.ToCompleted(); err == nil {
				if err = db.Set(item); err == nil {
					log.Info("task: [%s] is completed", task.Id())
				}
			}
		} else if cancelled && err == nil { // cancelled
			if item, err = task.ToRunning(); err == nil {
				err = db.Set(item)
			}
		}

		if err != nil {
			if item, err = task.ToFailed(err); err == nil {
				err = errors.Join(err, db.Set(item))
			}

			log.Error("task: [%s] is failed: [%s]", task.Id(), err.Error())
		}
		ctx.Complete <- task.Id()
	}()
	return nil
}

func (task *taskCopy) ToWaiting() (store.StoreItem, error) {
	task.Status = Waiting
	value, err := json.Marshal(task)

	key := strings.Join([]string{collName, strconv.Itoa(int(task.Status)), task.Id()}, collSep)
	return store.StoreItem{Key: []byte(key), Value: value}, err
}

func (task *taskCopy) ToRunning() (store.StoreItem, error) {
	task.Status = Running
	value, err := json.Marshal(task)

	key := strings.Join([]string{collName, strconv.Itoa(int(task.Status)), task.Id()}, collSep)
	return store.StoreItem{Key: []byte(key), Value: value}, err
}

func (task *taskCopy) ToFailed(original error) (store.StoreItem, error) {
	task.Status = Failed
	task.Error = original.Error()
	value, err := json.Marshal(task)

	key := strings.Join([]string{collName, strconv.Itoa(int(task.Status)), task.Id()}, collSep)
	return store.StoreItem{Key: []byte(key), Value: value}, errors.Join(err, original)
}

func (task *taskCopy) ToCompleted() (store.StoreItem, error) {
	task.Status = Completed
	value, err := json.Marshal(task)

	key := strings.Join([]string{collName, strconv.Itoa(int(task.Status)), task.Id()}, collSep)
	return store.StoreItem{Key: []byte(key), Value: value}, err
}

type taskExecutor struct {
	config TaskConfig
	log    *logger.Logger
	db     store.Store
	cancel chan bool
}

func (exec *taskExecutor) Submit(task Task) error {
	log := exec.log
	db := exec.db

	item, err := task.ToWaiting()
	if err == nil {
		err = db.Set(item)
		if err == nil {
			item.Key = []byte(strings.Join([]string{collName, task.Id()}, collSep))
			err = db.Set(item)
		}
	}

	if err != nil {
		log.Error("task: [%s] is failed: [%s]", task.Id(), err.Error())
	} else {
		log.Info("task: [%s] is waiting", task.Id())
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
		ctx := TaskProcessContext{Log: exec.log, Cancel: cancel, Complete: complete}

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
									if err = task.Process(ctx); err == nil {
										err = exec.db.Del(item.Key)
									}
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
