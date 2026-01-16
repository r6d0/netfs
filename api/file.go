package api

import (
	"netfs/api/transport"
	"strconv"
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

// Information about file.
type FileInfo struct {
	FileName string
	FilePath string
	FileType FileType
	FileSize int64
}

// File on a remote resource.
type RemoteFile struct {
	Host RemoteHost
	Info FileInfo
}

// Returns children of the directory.
func (file *RemoteFile) Children(client transport.TransportSender, skip int, limit int) ([]RemoteFile, error) {
	params := []string{
		Endpoints.FileChildren.Path, file.Info.FilePath,
		Endpoints.FileChildren.Skip, strconv.Itoa(skip),
		Endpoints.FileChildren.Limit, strconv.Itoa(limit),
	}

	req, err := client.NewRequest(file.Host.IP, Endpoints.FileChildren.Name, params, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			files := []FileInfo{}
			if _, err = res.Body(&files); err == nil {
				result := make([]RemoteFile, len(files))
				for index, info := range files {
					result[index] = RemoteFile{Info: info, Host: file.Host}
				}
				return result, nil
			}
		}
	}
	return nil, err
}

// Writes data to remote file.
func (file *RemoteFile) Write(client transport.TransportSender, data []byte) error {
	params := []string{Endpoints.FileWrite.Path, file.Info.FilePath}
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileWrite.Name, params, data, nil)
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}

// Creates file or directory on remote host.
func (file *RemoteFile) Create(client transport.TransportSender) error {
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileCreate, nil, nil, file.Info)
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}

// Copies the current file to the target file.
func (file *RemoteFile) CopyTo(client transport.TransportSender, target RemoteFile) (*RemoteTask, error) {
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileCopyStart, nil, nil, []RemoteFile{*file, target})
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			task := &RemoteTask{}
			if _, err = res.Body(task); err == nil {
				return task, nil
			}
		}
	}
	return nil, err
}

// Removes the file from the remote host.
func (file *RemoteFile) Remove(client transport.TransportSender) error {
	params := []string{Endpoints.FileRemove.Path, file.Info.FilePath}
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileRemove.Name, params, nil, nil)
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}
