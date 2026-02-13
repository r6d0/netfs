package api_test

import (
	"netfs/api"
	"netfs/api/transport"
	"time"
)

const testVolumeId = 100
const testFileId = 100
const testFileName = "test_file.txt"

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
		volumeId, err := req.ParamUInt64(api.Endpoints.FileInfo.VolumeId)
		fileId, err := req.ParamUInt64(api.Endpoints.FileInfo.FileId)
		return nil, api.FileInfo{Type: api.FILE, VolumeId: volumeId, Id: fileId}, err
	})
	rec.Start()
}

func afterEach() {
	rec.Stop()
}
