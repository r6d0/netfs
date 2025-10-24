package netfs

import (
	"net"
	"netfs/internal/transport"
)

// Information about host.
type RemoteHost struct {
	Name string
	IP   net.IP
}

// Returns information about file by path.
func (host RemoteHost) OpenFile(client transport.Transport, file RemoteFile) (*RemoteFile, error) {
	res, err := client.SendBodyAndReceive(host.IP, API.FileInfo.URL, file, &RemoteFile{})
	if err == nil {
		return res.(*RemoteFile), nil
	}
	return nil, err
}
