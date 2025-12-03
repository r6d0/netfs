package api

import (
	"net"
	"netfs/internal/api/transport"
)

// Information about host.
type RemoteHost struct {
	Name string
	IP   net.IP
}

// Returns information about file by path.
func (host RemoteHost) OpenFile(client transport.TransportSender, path string) (*RemoteFile, error) {
	parameters := []string{Endpoints.FileInfo.Path, path}
	req, err := client.NewRequest(host.IP, Endpoints.FileInfo.Name, parameters, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = client.Send(req); err == nil {
			info := &FileInfo{}
			if _, err = res.Body(info); err == nil {
				return &RemoteFile{Info: *info, Host: host}, nil
			}
		}
	}
	return nil, err
}
