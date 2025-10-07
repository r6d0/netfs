package netfs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

// Type of file.
type RemoteFileType byte

const (
	FILE RemoteFileType = iota
	DIRECTORY
)

// Returns a string representation of the file type.
func (fileType RemoteFileType) String() string {
	if fileType == FILE {
		return "f"
	}
	return "d"
}

// Information about file.
type RemoteFile struct {
	Host    RemoteHost
	Name    string
	Path    string
	Type    RemoteFileType
	Size    uint64
	_Client *http.Client
}

// Copies file to target.
func (file *RemoteFile) CopyTo(target *RemoteFile) error {
	data, err := json.Marshal([]RemoteFile{*file, *target})
	if err == nil {
		host := target.Host
		_, err = file._Client.Post(host.GetURL(_API.FileCopyStart.URL), _API.FileCopyStart.ContentType, bytes.NewReader(data))
	}
	return err
}

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

const _COPY_TASK = "task.copy."

type _CopyTaskStatus byte

const (
	_CREATED _CopyTaskStatus = iota
	_DELAYED
	_STOPPED
	_FAILED
	_RUNNING
	_COMPLETED
)

// Asynchronous task for files copying.
type _CopyTask struct {
	_DB         *badger.DB
	_Context    context.Context
	_Callback   chan _CopyTaskStatus
	_BufferSize uint64
	Id          []byte
	Offset      uint64
	Status      _CopyTaskStatus
	Source      RemoteFile
	Target      RemoteFile
	Error       string
}

// Saves task to database.
func (task *_CopyTask) Save() error {
	return task._DB.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(*task)
		if err == nil {
			err = txn.Set(task.Id, data)
		}
		return err
	})
}

// Starts the copying process.
func (task *_CopyTask) Process() {
	var err error
	switch task.Source.Type {
	// Copy directory
	case DIRECTORY:
		err = task.ProcessDirectory()
	// Copy file
	case FILE:
		err = task.ProcessFile()
	}
	task._Callback <- task.Status

	fmt.Println("ERROR: ", err)
}

// Copy directory to remote host.
func (task *_CopyTask) ProcessDirectory() error {
	var err error
	var data []byte
	if data, err = json.Marshal(task.Target); err == nil {
		host := task.Target.Host
		_, err = http.Post(host.GetURL(_API.FileCreate.URL), _API.FileCreate.ContentType, bytes.NewReader(data))
		if err != nil {
			task.Status = _DELAYED
			task.Error = err.Error()
		}
	} else {
		task.Status = _FAILED
		task.Error = err.Error()
	}

	if err == nil {
		task.Status = _COMPLETED
	}
	return errors.Join(err, task.Save())
}

// Copy file to remote host.
func (task *_CopyTask) ProcessFile() error {
	file, err := os.Open(task.Source.Path)
	if err != nil {
		task.Status = _FAILED
		task.Error = err.Error()
	} else {
		defer file.Close()

		var info os.FileInfo
		if info, err = file.Stat(); err == nil {
			url := task.Target.Host.GetURL(_API.FileWrite.URL, _API.FileWrite.Path, task.Target.Path)
			buffer := make([]byte, min(uint64(info.Size()), task._BufferSize))

			size := 0
			for task.Status == _RUNNING {
				select {
				case <-task._Context.Done(): // canceled
					task.Status = _DELAYED
					task.Save()
				default: // in progress
					if size, err = file.Read(buffer); err == nil {
						if _, err = http.Post(url, _API.FileWrite.ContentType, bytes.NewReader(buffer[:size])); err == nil {
							task.Offset += uint64(size)
						} else {
							task.Status = _DELAYED
							task.Error = err.Error()
						}
					} else if errors.Is(err, io.EOF) {
						err = nil
						task.Status = _COMPLETED
					} else {
						task.Status = _FAILED
						task.Error = err.Error()
					}
					err = errors.Join(err, task.Save())
				}
			}
		} else {
			task.Status = _FAILED
			task.Error = err.Error()
		}
	}
	return errors.Join(err, task.Save())
}

// Returns information about file or directory.
func (serv *Server) _FileInfoHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == _API.FileInfo.Method {
		query := request.URL.Query()
		if path := query.Get(_API.FileInfo.Path); path != "" {
			file, err := os.Open(path)
			if err == nil {
				defer file.Close()

				var info os.FileInfo
				if info, err = file.Stat(); err == nil {
					fileType := FILE
					if info.IsDir() {
						fileType = DIRECTORY
					}

					var data []byte
					data, err = json.Marshal(RemoteFile{Host: serv._Host, Name: info.Name(), Path: path, Type: fileType, Size: uint64(info.Size())})
					if err == nil {
						writer.Write(data)
					}
				}
			}

			if err != nil {
				fmt.Println(err)

				writer.Write([]byte(err.Error()))
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Writes the received data to file.
func (serv *Server) _FileWriteHandle(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("_FileWriteHandle")
	if request.Method == _API.FileWrite.Method {
		query := request.URL.Query()
		if path := query.Get(_API.FileWrite.Path); path != "" {
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				defer file.Close()
				defer request.Body.Close()

				var data []byte
				if data, err = io.ReadAll(request.Body); err == nil {
					file.Write(data)
				}
			}

			if err != nil {
				fmt.Println(err)

				writer.Write([]byte(err.Error()))
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Starting a file or directory copy operation.
func (serv *Server) _FileCopyStartHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == _API.FileCopyStart.Method {
		defer request.Body.Close()

		var err error
		var data []byte
		if data, err = io.ReadAll(request.Body); err == nil {
			files := &[]RemoteFile{}
			if err = json.Unmarshal(data, files); err == nil {
				source := (*files)[0]
				target := (*files)[1]
				if source.Type == DIRECTORY {
					prevType := DIRECTORY
					prevName := source.Name
					prevPath := source.Path
					parent := filepath.Dir(source.Path)
					task := _CopyTask{_DB: serv._DB, Status: _CREATED}
					err = filepath.WalkDir(source.Path, func(path string, entry fs.DirEntry, err error) error {
						if strings.Contains(path, prevPath) {
							prevName = entry.Name()
							prevPath = path
							if entry.IsDir() {
								prevType = DIRECTORY
							} else {
								prevType = FILE
							}
						} else {
							newPath := strings.ReplaceAll(prevPath, parent, target.Path)

							task.Id = []byte(_COPY_TASK + newPath)
							task.Source = RemoteFile{Host: source.Host, Name: prevName, Path: prevPath, Type: prevType}
							task.Target = RemoteFile{Host: target.Host, Name: prevName, Path: newPath, Type: prevType}
							err = task.Save()

							prevName = entry.Name()
							prevPath = path
							if entry.IsDir() {
								prevType = DIRECTORY
							} else {
								prevType = FILE
							}
						}
						return err
					})

					newPath := strings.ReplaceAll(prevPath, parent, target.Path)

					task.Id = []byte(_COPY_TASK + newPath)
					task.Source = RemoteFile{Host: source.Host, Name: prevName, Path: prevPath, Type: prevType}
					task.Target = RemoteFile{Host: target.Host, Name: prevName, Path: newPath, Type: prevType}
					err = task.Save()
				} else {
					id := []byte(_COPY_TASK + source.Path)
					task := _CopyTask{_DB: serv._DB, Id: id, Status: _CREATED, Source: source, Target: target}
					err = task.Save()
				}
			}
		}

		if err != nil {
			fmt.Println(err)

			writer.Write([]byte(err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Create directory.
func (serv *Server) _FileCreateHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == _API.FileCreate.Method {
		defer request.Body.Close()

		data, err := io.ReadAll(request.Body)
		if err == nil {
			file := &RemoteFile{}
			if err = json.Unmarshal(data, file); err == nil {
				if file.Type == DIRECTORY {
					err = os.MkdirAll(file.Path, os.ModePerm)
				} else {
					err = errors.New("unsupported file operation")
				}
			}
		}

		if err != nil {
			fmt.Println(err)

			writer.Write([]byte(err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Executes async tasks.
func (serv *Server) _ExecuteCopyTask() {
	go func(ctx context.Context) {
		var err error

		prefix := []byte(_COPY_TASK)
		callback := make(chan _CopyTaskStatus)
		count := serv._Config.TaskCount

		for err == nil {
			select {
			case <-ctx.Done():
				return
			case <-callback:
				count++
			default:
				err = serv._DB.View(func(txn *badger.Txn) error {
					var err error

					it := txn.NewIterator(badger.DefaultIteratorOptions)
					defer it.Close()

					for it.Seek(prefix); err == nil && it.ValidForPrefix(prefix) && count > 0; it.Next() {
						item := it.Item()

						var data []byte
						if data, err = item.ValueCopy(make([]byte, item.ValueSize())); err == nil {
							task := &_CopyTask{}
							if err = json.Unmarshal(data, task); err == nil && task.Status <= _DELAYED {
								fmt.Println(string(item.Key()))

								task._DB = serv._DB
								task._Context = ctx
								task._Callback = callback
								task._BufferSize = serv._Config.BufferSize
								task.Status = _RUNNING
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
	}(serv._Tasks._Context)
}
