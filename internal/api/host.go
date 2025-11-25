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
	res, err := client.SendRawBodyAndReceive(host.IP, API.FileInfo(), []byte(path), &FileInfo{})
	if err == nil {
		info := res.(*FileInfo)
		return &RemoteFile{Info: *info, Host: host}, nil
	}
	return nil, err
}
