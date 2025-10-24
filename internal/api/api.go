package api

// Transport API.
var API = struct {
	ServerHost     string
	ServerStop     string
	FileInfo       string
	FileCreate     string
	FileWrite      string
	FileCopyStart  string
	FileCopyStatus string
}{
	ServerHost:     "/netfs/api/server/host",
	ServerStop:     "/netfs/api/server/stop",
	FileInfo:       "/netfs/api/file/info",
	FileCreate:     "/netfs/api/file/create",
	FileWrite:      "/netfs/api/file/write",
	FileCopyStart:  "/netfs/api/file/copy/start",
	FileCopyStatus: "/netfs/api/file/copy/status",
}
