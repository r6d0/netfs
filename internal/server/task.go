package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	netfs "netfs/internal"
	"os"

	"github.com/dgraph-io/badger/v4"
)

const _COPY_TASK = "task.copy."

type copyTaskStatus byte

const (
	created copyTaskStatus = iota
	delayed
	stopped
	failed
	running
	completed
)

// Asynchronous task for files copying.
type copyTask struct {
	db         *badger.DB
	context    context.Context
	callback   chan copyTaskStatus
	bufferSize uint64
	Id         []byte
	Offset     uint64
	Status     copyTaskStatus
	Source     netfs.RemoteFile
	Target     netfs.RemoteFile
	Error      string
}

// Saves task to database.
func (task *copyTask) Save() error {
	return task.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(*task)
		if err == nil {
			err = txn.Set(task.Id, data)
		}
		return err
	})
}

// Starts the copying process.
func (task *copyTask) Process() {
	var err error
	switch task.Source.Type {
	// Copy directory
	case netfs.DIRECTORY:
		err = task.processDirectory()
	// Copy file
	case netfs.FILE:
		err = task.ProcessFile()
	}
	task.callback <- task.Status

	fmt.Println("ERROR: ", err)
}

// Copy directory to remote host.
func (task *copyTask) processDirectory() error {
	var err error
	var data []byte
	if data, err = json.Marshal(task.Target); err == nil {
		host := task.Target.Host
		_, err = http.Post(host.GetURL(netfs.API.FileCreate.URL), netfs.API.FileCreate.ContentType, bytes.NewReader(data))
		if err != nil {
			task.Status = delayed
			task.Error = err.Error()
		}
	} else {
		task.Status = failed
		task.Error = err.Error()
	}

	if err == nil {
		task.Status = completed
	}
	return errors.Join(err, task.Save())
}

// Copy file to remote host.
func (task *copyTask) ProcessFile() error {
	file, err := os.Open(task.Source.Path)
	if err != nil {
		task.Status = failed
		task.Error = err.Error()
	} else {
		defer file.Close()

		var info os.FileInfo
		if info, err = file.Stat(); err == nil {
			url := task.Target.Host.GetURL(netfs.API.FileWrite.URL, netfs.API.FileWrite.Path, task.Target.Path)
			buffer := make([]byte, min(uint64(info.Size()), task.bufferSize))

			size := 0
			for task.Status == running {
				select {
				case <-task.context.Done(): // canceled
					task.Status = delayed
					task.Save()
				default: // in progress
					if size, err = file.Read(buffer); err == nil {
						if _, err = http.Post(url, netfs.API.FileWrite.ContentType, bytes.NewReader(buffer[:size])); err == nil {
							task.Offset += uint64(size)
						} else {
							task.Status = delayed
							task.Error = err.Error()
						}
					} else if errors.Is(err, io.EOF) {
						err = nil
						task.Status = completed
					} else {
						task.Status = failed
						task.Error = err.Error()
					}
					err = errors.Join(err, task.Save())
				}
			}
		} else {
			task.Status = failed
			task.Error = err.Error()
		}
	}
	return errors.Join(err, task.Save())
}

// Executes async tasks.
func executeCopyTask(serv *Server) {
	go func(ctx context.Context) {
		var err error

		prefix := []byte(_COPY_TASK)
		callback := make(chan copyTaskStatus)
		count := serv.config.TaskCount

		for err == nil {
			select {
			case <-ctx.Done():
				return
			case <-callback:
				count++
			default:
				err = serv.db.View(func(txn *badger.Txn) error {
					var err error

					it := txn.NewIterator(badger.DefaultIteratorOptions)
					defer it.Close()

					for it.Seek(prefix); err == nil && it.ValidForPrefix(prefix) && count > 0; it.Next() {
						item := it.Item()

						var data []byte
						if data, err = item.ValueCopy(make([]byte, item.ValueSize())); err == nil {
							task := &copyTask{}
							if err = json.Unmarshal(data, task); err == nil && task.Status <= delayed {
								fmt.Println(string(item.Key()))

								task.db = serv.db
								task.context = ctx
								task.callback = callback
								task.bufferSize = serv.config.BufferSize
								task.Status = running
								task.Save()

								fmt.Println("TASK: ", task.Source)
								go task.Process()
								count--
							}
						}
					}
					return err
				})
			}
		}
	}(serv.tasks._Context)
}
