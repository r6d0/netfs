package netfs

import "net/http"

// Server API.
var API = struct {
	// Stops the server.
	Stop string
	// Information about file.
	FileInfo struct {
		URL    string
		Method string
		Path   string
	}
	// Information about host.
	Host string
	// Create file or directory.
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
	// Starting a file or directory copy operation.
	FileCopyStatus struct {
		URL    string
		Method string
		Id     string
		Status string
	}
}{
	Stop: "/netfs/api/server/stop",
	Host: "/netfs/api/host",
	FileInfo: struct {
		URL    string
		Method string
		Path   string
	}{URL: "/netfs/api/file/info", Method: http.MethodGet, Path: "path"},
	FileCreate: struct {
		URL         string
		Method      string
		ContentType string
	}{URL: "/netfs/api/file/create", Method: http.MethodPost, ContentType: "application/octet-stream"},
	FileWrite: struct {
		URL         string
		Method      string
		ContentType string
		Path        string
	}{URL: "/netfs/api/file/write", Method: http.MethodPost, ContentType: "application/octet-stream", Path: "path"},
	FileCopyStart: struct {
		URL         string
		Method      string
		ContentType string
	}{URL: "/netfs/api/file/copy/start", Method: http.MethodPost, ContentType: "application/octet-stream"},
	FileCopyStatus: struct {
		URL    string
		Method string
		Id     string
		Status string
	}{URL: "/netfs/api/file/copy/status", Method: http.MethodGet, Id: "id", Status: "status"},
}
