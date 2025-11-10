package api

import (
	"netfs/internal/api/transport"
)

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
	Host     RemoteHost
	Name     string
	Path     string
	FileType RemoteFileType
	Size     uint64
}

// Writes data to remote file.
func (file RemoteFile) Write(client transport.Transport, data []byte) error {
	return client.SendRawBody(file.Host.IP, API.FileWrite(file.Path), data)
}

// Creates file or directory on remote host.
func (file RemoteFile) Create(client transport.Transport) error {
	return client.SendBody(file.Host.IP, API.FileCreate(), file)
}

// Copies the current file to the target file.
func (file RemoteFile) CopyTo(client transport.Transport, target RemoteFile) error {
	return client.SendBody(file.Host.IP, API.FileCopyStart(), []RemoteFile{file, target})
}
