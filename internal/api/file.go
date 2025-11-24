package api

import (
	"netfs/internal/api/transport"
)

// Type of file.
type FileType byte

const (
	FILE FileType = iota
	DIRECTORY
)

// Returns a string representation of the file type.
func (fileType FileType) String() string {
	if fileType == FILE {
		return "f"
	}
	return "d"
}

type FileInfo struct {
	FileName string
	FilePath string
	FileType FileType
	FileSize int64
}

// Information about file.
type RemoteFile struct {
	Host RemoteHost
	Info FileInfo
}

// Writes data to remote file.
func (file RemoteFile) Write(client transport.TransportSender, data []byte) error {
	return client.SendRawBody(file.Host.IP, API.FileWrite(file.Info.FilePath), data)
}

// Creates file or directory on remote host.
func (file RemoteFile) Create(client transport.TransportSender) error {
	return client.SendBody(file.Host.IP, API.FileCreate(), file)
}

// Copies the current file to the target file.
func (file RemoteFile) CopyTo(client transport.TransportSender, target RemoteFile) error {
	return client.SendBody(file.Host.IP, API.FileCopyStart(), []RemoteFile{file, target})
}
