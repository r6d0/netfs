package api

import "netfs/internal/api/transport"

// Transport API.
var API = struct {
	ServerHost     func() transport.TransportPoint
	ServerStop     func() transport.TransportPoint
	FileInfo       func() transport.TransportPoint
	FileCreate     func() transport.TransportPoint
	FileWrite      func(string) transport.TransportPoint
	FileCopyStart  func() transport.TransportPoint
	FileCopyStatus func() transport.TransportPoint
}{
	ServerHost: func() transport.TransportPoint { return transport.TransportPoint([]string{"/netfs/api/server/host"}) },
	ServerStop: func() transport.TransportPoint { return transport.TransportPoint([]string{"/netfs/api/server/stop"}) },
	FileInfo:   func() transport.TransportPoint { return transport.TransportPoint([]string{"/netfs/api/file/info"}) },
	FileCreate: func() transport.TransportPoint { return transport.TransportPoint([]string{"/netfs/api/file/create"}) },
	FileWrite: func(path string) transport.TransportPoint {
		return transport.TransportPoint([]string{"/netfs/api/file/write", "path", path})
	},
	FileCopyStart: func() transport.TransportPoint {
		return transport.TransportPoint([]string{"/netfs/api/file/copy/start"})
	},
	FileCopyStatus: func() transport.TransportPoint {
		return transport.TransportPoint([]string{"/netfs/api/file/copy/status"})
	},
}
