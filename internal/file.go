package netfs

import (
	"netfs/internal/transport"
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
	return client.SendRawBody(file.Host.IP, API.FileWrite.URL, data)
}

// Creates file or directory on remote host.
func (file RemoteFile) Create(client transport.Transport) error {
	return client.SendBody(file.Host.IP, API.FileCreate.URL, file)
}

// Copies the current file to the target file.
func (file RemoteFile) CopyTo(client transport.Transport, target RemoteFile) error {
	return client.SendBody(file.Host.IP, API.FileCopyStart.URL, []RemoteFile{file, target})
}
