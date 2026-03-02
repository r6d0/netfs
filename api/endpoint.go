package api

type FileInfoEndpoint struct {
	Name     string
	VolumeId string
	FileId   string
}

type FileWriteEndpoint struct {
	Name     string
	VolumeId string
	FileId   string
}

type FileRemoveEndpoint struct {
	Name     string
	VolumeId string
	FileId   string
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
	Name     string
	VolumeId string
	FileId   string
	Skip     string
	Limit    string
}

type VolumeEndpoint struct {
	Name     string
	VolumeId string
}

type VolumeChildrenEndpoint struct {
	Name     string
	VolumeId string
	Skip     string
	Limit    string
}

var Endpoints = struct {
	ServerHost     string
	ServerStop     string
	FileInfo       FileInfoEndpoint
	FileCreate     string
	FileWrite      FileWriteEndpoint
	FileRemove     FileRemoveEndpoint
	FileCopyStart  string
	FileCopyStatus FileCopyStatusEndpoint
	FileCopyStop   FileCopyCancelEndpoint
	FileChildren   FileChildrenEndpoint
	Volume         VolumeEndpoint
	VolumeChildren VolumeChildrenEndpoint
}{
	ServerHost:     "/netfs/api/server/host",
	ServerStop:     "/netfs/api/server/stop",
	FileInfo:       FileInfoEndpoint{Name: "/netfs/api/file/info", VolumeId: "volumeId", FileId: "fileId"},
	FileCreate:     "/netfs/api/file/create",
	FileWrite:      FileWriteEndpoint{Name: "/netfs/api/file/write", VolumeId: "volumeId", FileId: "fileId"},
	FileRemove:     FileRemoveEndpoint{Name: "/netfs/api/file/remove", VolumeId: "volumeId", FileId: "fileId"},
	FileCopyStart:  "/netfs/api/file/copy/start",
	FileCopyStatus: FileCopyStatusEndpoint{Name: "/netfs/api/file/copy/status", TaskId: "id"},
	FileCopyStop:   FileCopyCancelEndpoint{Name: "/netfs/api/file/copy/cancel", TaskId: "id"},
	FileChildren:   FileChildrenEndpoint{Name: "/netfs/api/file/children", VolumeId: "volumeId", FileId: "fileId", Skip: "skip", Limit: "limit"},
	Volume:         VolumeEndpoint{Name: "/netfs/api/volume", VolumeId: "volumeId"},
	VolumeChildren: VolumeChildrenEndpoint{Name: "/netfs/api/volume/children", VolumeId: "volumeId", Skip: "skip", Limit: "limit"},
}
