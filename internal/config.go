package netfs

import "net/http"

// Server configuration.
type Config struct {
	// The size of the data transfer buffer.
	BufferSize uint64
	// The size of the asynchronous task pool.
	TaskCount uint64
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
