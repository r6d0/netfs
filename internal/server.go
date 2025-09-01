package netfs

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/dgraph-io/badger/v4"
)

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

// HTTP server.
type Server struct {
	_DB     *badger.DB
	_Server *http.Server
	_Config *Config
	_Tasks  *_TasksContext
	_Host   RemoteHost
	_Stop   chan os.Signal
}

// Start requests listening. It's blocking current thread.
func (serv *Server) Start() {
	// Database connection
	db, err := badger.Open(badger.DefaultOptions(serv._Config.Database.Path))
	serv._DB = db

	if err == nil {
		// Run HTTP server.
		go func() {
			// Execute async tasks.
			serv._ExecuteCopyTask()

			err := serv._Server.ListenAndServe()
			if err != http.ErrServerClosed {
				log.Fatalln(err) // TODO. Log format
			}
		}()

		// Stop signal waiting.
		<-serv._Stop

		// HTTP server stopping.
		err := serv._Server.Shutdown(context.Background())
		if err != nil {
			log.Fatalln(err) // TODO. Log format
		}

		// Tasks stopping.
		serv._Tasks._Shutdown()

		// Database connection closing.
		if serv._DB != nil {
			serv._DB.Close()
		}
	}
}

// Stop the running server.
func (serv *Server) Stop() {
	serv._Stop <- syscall.SIGINT
}

// New instance of the netfs server.
func NewServer(config *Config) (*Server, error) {
	// Information about server host
	host, err := _GetLocalHost(config.Server.Protocol, int(config.Server.Port))
	if err == nil {
		server := &Server{_Config: config, _Host: *host}

		// Stop signal
		server._Stop = make(chan os.Signal, 1)
		signal.Notify(server._Stop, syscall.SIGINT, syscall.SIGTERM)

		// Tasks context
		ctx, cancel := context.WithCancel(context.Background())
		server._Tasks = &_TasksContext{_Context: ctx, _Shutdown: cancel}

		// API Registration
		mux := http.NewServeMux()
		mux.HandleFunc(_API.Host, server._HostHandle)
		mux.HandleFunc(_API.FileInfo.URL, server._FileInfoHandle)
		mux.HandleFunc(_API.FileCreate.URL, server._FileCreateHandle)
		mux.HandleFunc(_API.FileWrite.URL, server._FileWriteHandle)
		mux.HandleFunc(_API.FileCopyStart.URL, server._FileCopyStartHandle)

		server._Server = &http.Server{Addr: _PORT_SEPARATOR + strconv.Itoa(int(config.Server.Port)), Handler: mux}
		return server, nil
	}
	return nil, err
}

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

// The context of asynchronous tasks.
type _TasksContext struct {
	_Context  context.Context
	_Shutdown context.CancelFunc
}

// Server API.
var _API = struct {
	// Information about file.
	FileInfo struct {
		URL    string
		Method string
		Path   string
	}
	// Information about host.
	Host string
	// Create directory.
	FileCreate struct {
		URL         string
		Method      string
		ContentType string
	}
	// Write data to file.
	FileWrite struct {
		URL         string
		Method      string
		ContentType string
		Path        string
	}
	// Starting a file or directory copy operation.
	FileCopyStart struct {
		URL         string
		Method      string
		ContentType string
	}
}{
	Host: "/do-sync/api/host",
	FileInfo: struct {
		URL    string
		Method string
		Path   string
	}{URL: "/do-sync/api/file/info", Method: http.MethodGet, Path: "path"},
	FileCreate: struct {
		URL         string
		Method      string
		ContentType string
	}{URL: "/do-sync/api/file/create", Method: http.MethodPost, ContentType: "application/octet-stream"},
	FileWrite: struct {
		URL         string
		Method      string
		ContentType string
		Path        string
	}{URL: "/do-sync/api/file/write", Method: http.MethodPost, Path: "path", ContentType: "application/octet-stream"},
	FileCopyStart: struct {
		URL         string
		Method      string
		ContentType string
	}{URL: "/do-sync/api/file/copy/start", Method: http.MethodPost, ContentType: "application/octet-stream"},
}
