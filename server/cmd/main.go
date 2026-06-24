package main

import (
	"netfs/api"
	"netfs/api/transport"
	server "netfs/server/internal"
	"time"
)

func main() {
	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 8989, Protocol: transport.HTTP, Timeout: time.Second * 5},
		RootList: []string{"c:/", "d:/"},
	}

	srv, err := server.NewServer(config)
	if err != nil {
		panic(err)
	}

	if err = srv.Start(); err != nil {
		panic(err)
	}
}
