package api

import (
	"netfs/api/transport"
	"strconv"
)

// Available the volume operation.
type VolumePermitionType uint8

const (
	View VolumePermitionType = iota
	Read
	Write
)

// Information about the volume.
type VolumeInfo struct {
	// The volume name.
	Name string
	// The path in OS.
	OsPath string
	// The relative path for elements of the volume.
	LocalPath string
}

// The volume on a remote host.
type RemoteVolume struct {
	Info VolumeInfo
	Host RemoteHost
}

// The function returns elements of the current volume.
func (vl *RemoteVolume) Children(client transport.TransportSender, skip int, limit int) ([]RemoteFile, error) {
	parameters := []string{
		Endpoints.VolumeChildren.Volume, vl.Info.Name,
		Endpoints.VolumeChildren.Skip, strconv.Itoa(skip),
		Endpoints.VolumeChildren.Limit, strconv.Itoa(limit),
	}

	req, err := client.NewRequest(vl.Host.IP, Endpoints.VolumeChildren.Name, parameters, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			files := []FileInfo{}
			if _, err = res.Body(&files); err == nil {
				result := make([]RemoteFile, len(files))
				for index, info := range files {
					result[index] = RemoteFile{Info: info, Host: vl.Host}
				}
				return result, nil
			}
		}
	}
	return nil, err
}
