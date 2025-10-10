package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	netfs "netfs/internal"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/dgraph-io/badger/v4"
)

const portSeparator = ":"

// HTTP server.
type Server struct {
	db         *badger.DB
	httpServer *http.Server
	config     *netfs.Config
	tasks      *_TasksContext
	host       netfs.RemoteHost
	stop       chan os.Signal
}

// Start requests listening. It's blocking current thread.
func (serv *Server) Start() error {
	// Database connection
	db, err := badger.Open(badger.DefaultOptions(serv.config.Database.Path))
	serv.db = db

	if err == nil {
		// Run HTTP server.
		go func() {
			// Execute async tasks.
			executeCopyTask(serv)

			err = serv.httpServer.ListenAndServe()
			if err != http.ErrServerClosed {
				serv.stop <- syscall.SIGINT
			}
		}()

		// Stop signal waiting.
		<-serv.stop

		if err == nil {
			// HTTP server stopping.
			err = serv.httpServer.Shutdown(context.Background())
			if err != nil {
				log.Fatalln(err) // TODO. Log format
			}
		}

		// Tasks stopping.
		serv.tasks._Shutdown()

		// Database connection closing.
		if serv.db != nil {
			serv.db.Close()
		}
	}
	return err
}

// Stop the running server.
func (serv *Server) Stop() {
	serv.stop <- syscall.SIGINT
}

// New instance of the netfs server.
func NewServer(config *netfs.Config) (*Server, error) {
	network, err := netfs.NewNetwork(config)
	if err == nil {
		// Information about server host
		var host *netfs.RemoteHost
		host, err = network.GetLocalHost()
		if err == nil {
			server := &Server{config: config, host: *host}

			// Stop signal
			server.stop = make(chan os.Signal, 1)
			signal.Notify(server.stop, syscall.SIGINT, syscall.SIGTERM)

			// Tasks context
			ctx, cancel := context.WithCancel(context.Background())
			server.tasks = &_TasksContext{_Context: ctx, _Shutdown: cancel}

			// API Registration
			mux := http.NewServeMux()
			mux.HandleFunc(netfs.API.Stop, server.stopHandle)
			mux.HandleFunc(netfs.API.Host, server.hostHandle)
			mux.HandleFunc(netfs.API.FileInfo.URL, server.fileInfoHandle)
			mux.HandleFunc(netfs.API.FileCreate.URL, server.fileCreateHandle)
			mux.HandleFunc(netfs.API.FileWrite.URL, server.fileWriteHandle)
			mux.HandleFunc(netfs.API.FileCopyStart.URL, server.fileCopyStartHandle)

			server.httpServer = &http.Server{Addr: portSeparator + strconv.Itoa(int(config.Server.Port)), Handler: mux}
			return server, nil
		}
	}
	return nil, err
}

// Stops the server.
func (serv *Server) stopHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		// Only from current host.
		if strings.Contains(request.RemoteAddr, serv.host.IP.String()) {
			serv.Stop()
		} else {
			writer.WriteHeader(http.StatusForbidden)
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Returns information about the current host.
func (serv *Server) hostHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		data, err := json.Marshal(serv.host)
		if err == nil {
			_, err = writer.Write(data)
		}

		if err != nil {
			fmt.Println(err) // TODO. Log format

			writer.Write([]byte(err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Returns information about file or directory.
func (serv *Server) fileInfoHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == netfs.API.FileInfo.Method {
		query := request.URL.Query()
		if path := query.Get(netfs.API.FileInfo.Path); path != "" {
			file, err := os.Open(path)
			if err == nil {
				defer file.Close()

				var info os.FileInfo
				if info, err = file.Stat(); err == nil {
					fileType := netfs.FILE
					if info.IsDir() {
						fileType = netfs.DIRECTORY
					}

					var data []byte
					data, err = json.Marshal(netfs.RemoteFile{Host: serv.host, Name: info.Name(), Path: path, Type: fileType, Size: uint64(info.Size())})
					if err == nil {
						writer.Write(data)
					}
				}
			}

			if err != nil {
				fmt.Println(err) // TODO. Log format

				writer.Write([]byte(err.Error()))
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Create directory.
func (serv *Server) fileCreateHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == netfs.API.FileCreate.Method {
		defer request.Body.Close()

		data, err := io.ReadAll(request.Body)
		if err == nil {
			file := &netfs.RemoteFile{}
			if err = json.Unmarshal(data, file); err == nil {
				if file.Type == netfs.DIRECTORY {
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

// Writes the received data to file.
func (serv *Server) fileWriteHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == netfs.API.FileWrite.Method {
		query := request.URL.Query()
		if path := query.Get(netfs.API.FileWrite.Path); path != "" {
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
				fmt.Println(err) // TODO. Log format

				writer.Write([]byte(err.Error()))
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Starting a file or directory copy operation.
func (serv *Server) fileCopyStartHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == netfs.API.FileCopyStart.Method {
		defer request.Body.Close()

		var err error
		var data []byte
		if data, err = io.ReadAll(request.Body); err == nil {
			files := &[]netfs.RemoteFile{}
			if err = json.Unmarshal(data, files); err == nil {
				source := (*files)[0]
				target := (*files)[1]
				if source.Type == netfs.DIRECTORY {
					prevType := netfs.DIRECTORY
					prevName := source.Name
					prevPath := source.Path
					parent := filepath.Dir(source.Path)
					task := copyTask{db: serv.db, Status: created}
					err = filepath.WalkDir(source.Path, func(path string, entry fs.DirEntry, err error) error {
						if strings.Contains(path, prevPath) {
							prevName = entry.Name()
							prevPath = path
							if entry.IsDir() {
								prevType = netfs.DIRECTORY
							} else {
								prevType = netfs.FILE
							}
						} else {
							newPath := strings.ReplaceAll(prevPath, parent, target.Path)

							task.Id = []byte(_COPY_TASK + newPath)
							task.Source = netfs.RemoteFile{Host: source.Host, Name: prevName, Path: prevPath, Type: prevType}
							task.Target = netfs.RemoteFile{Host: target.Host, Name: prevName, Path: newPath, Type: prevType}
							err = task.Save()

							prevName = entry.Name()
							prevPath = path
							if entry.IsDir() {
								prevType = netfs.DIRECTORY
							} else {
								prevType = netfs.FILE
							}
						}
						return err
					})

					newPath := strings.ReplaceAll(prevPath, parent, target.Path)

					task.Id = []byte(_COPY_TASK + newPath)
					task.Source = netfs.RemoteFile{Host: source.Host, Name: prevName, Path: prevPath, Type: prevType}
					task.Target = netfs.RemoteFile{Host: target.Host, Name: prevName, Path: newPath, Type: prevType}
					err = task.Save()
				} else {
					id := []byte(_COPY_TASK + source.Path)
					task := copyTask{db: serv.db, Id: id, Status: created, Source: source, Target: target}
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

// The context of asynchronous tasks.
type _TasksContext struct {
	_Context  context.Context
	_Shutdown context.CancelFunc
}
