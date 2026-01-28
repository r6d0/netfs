package api_test

import (
	"netfs/api"
	"netfs/api/transport"
	"path/filepath"
	"time"
)

// Do not use with t.Parallel(...)

var network *api.Network
var local api.RemoteHost
var rec transport.TransportReceiver
var config = api.NetworkConfig{Port: 9184, Protocol: transport.HTTP, Timeout: 5 * time.Second}

func beforeEach() {
	rec, _ = transport.NewReceiver(config.Protocol, config.Port)

	network, _ = api.NewNetwork(config)
	local = network.LocalHost()
	rec.Receive(api.Endpoints.ServerHost, func(transport.Request) ([]byte, any, error) {
		return nil, local, nil
	})
	rec.Receive(api.Endpoints.FileInfo.Name, func(req transport.Request) ([]byte, any, error) {
		path, err := req.ParamRequired(api.Endpoints.FileInfo.Path)
		_, name := filepath.Split(path)
		return nil, api.FileInfo{FileName: name, FilePath: path, FileType: api.FILE, FileSize: 1024}, err
	})
	rec.Start()
}

func afterEach() {
	rec.Stop()
}
