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
	// The volume identifier.
	Id uint64
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

// The function creates a file or directory on the remote host.
func (vl *RemoteVolume) Create(client transport.TransportSender, info FileInfo) (*RemoteFile, error) {
	info.VolumeId = vl.Info.Id

	req, err := client.NewRequest(vl.Host.IP, Endpoints.FileCreate, nil, nil, info)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			info := &FileInfo{}
			if _, err = res.Body(info); err == nil {
				return &RemoteFile{Info: *info, Host: vl.Host}, nil
			}
		}
	}
	return nil, err
}

// The function returns information about a file by id.
func (vl *RemoteVolume) File(client transport.TransportSender, fileId uint64) (*RemoteFile, error) {
	params := []string{
		Endpoints.FileInfo.VolumeId, strconv.FormatUint(vl.Info.Id, decimalBase),
		Endpoints.FileInfo.FileId, strconv.FormatUint(fileId, decimalBase),
	}
	req, err := client.NewRequest(vl.Host.IP, Endpoints.FileInfo.Name, params, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			info := &FileInfo{}
			if _, err = res.Body(info); err == nil {
				return &RemoteFile{Info: *info, Host: vl.Host}, nil
			}
		}
	}
	return nil, err
}

// The function returns elements of the current volume.
func (vl *RemoteVolume) Children(client transport.TransportSender, skip int, limit int) ([]RemoteFile, error) {
	parameters := []string{
		Endpoints.VolumeChildren.VolumeId, strconv.FormatUint(vl.Info.Id, 10),
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
