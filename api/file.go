package api

import "netfs/api/transport"

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
