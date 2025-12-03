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
	parameters := []string{Endpoints.FileWrite.Path, file.Info.FilePath}
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileWrite.Name, parameters, data, nil)
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}

// Creates file or directory on remote host.
func (file RemoteFile) Create(client transport.TransportSender) error {
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileCreate, nil, nil, file.Info)
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}

// Copies the current file to the target file.
func (file RemoteFile) CopyTo(client transport.TransportSender, target RemoteFile) error {
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileCopyStart, nil, nil, []RemoteFile{file, target})
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}
