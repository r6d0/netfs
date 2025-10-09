package netfs

import (
	"bytes"
	"encoding/json"
	"net/http"
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
	Host   RemoteHost
	Name   string
	Path   string
	Type   RemoteFileType
	Size   uint64
	client *http.Client
}

// Copies file to target.
func (file *RemoteFile) CopyTo(target *RemoteFile) error {
	data, err := json.Marshal([]RemoteFile{*file, *target})
	if err == nil {
		host := target.Host
		_, err = file.client.Post(host.GetURL(API.FileCopyStart.URL), API.FileCopyStart.ContentType, bytes.NewReader(data))
	}
	return err
}
