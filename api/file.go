package api

import (
	"netfs/api/transport"
	"strconv"
	"strings"
)

var units = [5]string{"B", "KB", "MB", "GB", "TB"}

// Type of file.
type FileType byte

const (
	FILE FileType = 1 << iota
	DIRECTORY
)

// Returns a string representation of the file type.
func (fileType FileType) String() string {
	if fileType == FILE {
		return "f"
	}
	return "d"
}

// Size of the file.
type FileSize int64

// String representation of file size.
func (fileSize FileSize) String() string {
	size := int64(fileSize)
	if size == 0 {
		return "0"
	} else {
		unit := 0
		for unit < len(units) && size >= 1024 {
			size /= 1024
			unit++
		}
		return strings.Join(
			[]string{
				strconv.FormatInt(size, decimalBase),
				units[unit],
			},
			" ",
		)
	}
}

// File identifier.
type FileId string

// Information about file.
type FileInfo struct {
	Id       FileId
	Name     string
	Path     string
	Type     FileType
	Size     FileSize
	ParentId FileId
}

// File on a remote host.
type RemoteFile struct {
	Info FileInfo
	Host RemoteHost
}

// Returns children of the directory.
func (file *RemoteFile) Children(client transport.TransportSender) ([]RemoteFile, error) {
	params := []string{
		Endpoints.FileChildren.FileId, string(file.Info.Id),
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
	params := []string{
		Endpoints.FileWrite.FileId, string(file.Info.Id),
	}
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileWrite.Name, params, data, nil)
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}

// Copies the current file to the target file.
func (file *RemoteFile) CopyTo(client transport.TransportSender, target RemoteFile) (*RemoteCopyTask, error) {
	task := &RemoteCopyTask{Source: *file, Target: target}

	req, err := client.NewRequest(file.Host.IP, Endpoints.FileCopyStart, nil, nil, *task)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			if _, err = res.Body(task); err == nil {
				return task, nil
			}
		}
	}
	return nil, err
}

// Removes the file from the remote host.
func (file *RemoteFile) Remove(client transport.TransportSender) error {
	params := []string{
		Endpoints.FileRemove.FileId, string(file.Info.Id),
	}
	req, err := client.NewRequest(file.Host.IP, Endpoints.FileRemove.Name, params, nil, nil)
	if err == nil {
		_, err = client.Send(req)
	}
	return err
}
