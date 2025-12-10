package api

type FileInfoEndpoint struct {
	Name string
	Path string
}

type FileWriteEndpoint struct {
	Name string
	Path string
}

type FileCopyStatusEndpoint struct {
	Name string
	Id   string
}

type FileCopyStopEndpoint struct {
	Name string
	Id   string
}

var Endpoints = struct {
	ServerHost     string
	ServerStop     string
	FileInfo       FileInfoEndpoint
	FileCreate     string
	FileWrite      FileWriteEndpoint
	FileCopyStart  string
	FileCopyStatus FileCopyStatusEndpoint
	FileCopyStop   FileCopyStopEndpoint
}{
	ServerHost:     "/netfs/api/server/host",
	ServerStop:     "/netfs/api/server/stop",
	FileInfo:       FileInfoEndpoint{Name: "/netfs/api/file/info", Path: "path"},
	FileCreate:     "/netfs/api/file/create",
	FileWrite:      FileWriteEndpoint{Name: "/netfs/api/file/write", Path: "path"},
	FileCopyStart:  "/netfs/api/file/copy/start",
	FileCopyStatus: FileCopyStatusEndpoint{Name: "/netfs/api/file/copy/status", Id: "id"},
	FileCopyStop:   FileCopyStopEndpoint{Name: "/netfs/api/file/copy/stop", Id: "id"},
}
