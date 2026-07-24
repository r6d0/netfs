package api

type FileInfoEndpoint struct {
	Name   string
	FileId string
}

type FileWriteEndpoint struct {
	Name   string
	FileId string
}

type FileRemoveEndpoint struct {
	Name   string
	FileId string
}

type FileCopyStatusEndpoint struct {
	Name   string
	TaskId string
}

type FileCopyCancelEndpoint struct {
	Name   string
	TaskId string
}

type FileChildrenEndpoint struct {
	Name   string
	FileId string
}

type FileCreateEndpoint struct {
	Name    string
	Replace string
}

var Endpoints = struct {
	ServerHost     string
	ServerStop     string
	FileInfo       FileInfoEndpoint
	FileCreate     FileCreateEndpoint
	FileWrite      FileWriteEndpoint
	FileRemove     FileRemoveEndpoint
	FileCopy       string
	FileCopyStart  string
	FileCopyStatus FileCopyStatusEndpoint
	FileCopyCancel FileCopyCancelEndpoint
	FileChildren   FileChildrenEndpoint
}{
	ServerHost:     "/netfs/api/server/host",
	ServerStop:     "/netfs/api/server/stop",
	FileInfo:       FileInfoEndpoint{Name: "/netfs/api/file/info", FileId: "fileId"},
	FileCreate:     FileCreateEndpoint{Name: "/netfs/api/file/create", Replace: "replace"},
	FileWrite:      FileWriteEndpoint{Name: "/netfs/api/file/write", FileId: "fileId"},
	FileRemove:     FileRemoveEndpoint{Name: "/netfs/api/file/remove", FileId: "fileId"},
	FileCopy:       "/netfs/api/file/copy/all",
	FileCopyStart:  "/netfs/api/file/copy/start",
	FileCopyStatus: FileCopyStatusEndpoint{Name: "/netfs/api/file/copy/status", TaskId: "id"},
	FileCopyCancel: FileCopyCancelEndpoint{Name: "/netfs/api/file/copy/cancel", TaskId: "id"},
	FileChildren:   FileChildrenEndpoint{Name: "/netfs/api/file/children", FileId: "fileId"},
}
