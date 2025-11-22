package server_test

import (
	"fmt"
	"netfs/internal/api"
	"netfs/internal/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"testing"
	"time"
)

func TestServerHostHandleSuccess(t *testing.T) {
	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
		Log:      logger.LoggerConfig{Level: logger.Info},
		Database: database.DatabaseConfig{Path: "./"},
		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}},
	}

	srv, err := server.NewServer(config)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	go func() {
		srv.Start()
	}()
	time.Sleep(2 * time.Second)

	network, _ := api.NewNetwork(config.Network)
	host, err := network.GetHost(network.LocalIP())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if host == nil {
		t.Fatalf("host should be not nil")
	}
	fmt.Println(host)

	srv.Stop()
}

func TestStopServerHandleSuccess(t *testing.T) {
	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
		Log:      logger.LoggerConfig{Level: logger.Info},
		Database: database.DatabaseConfig{Path: "./"},
		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}},
	}

	srv, err := server.NewServer(config)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	go func() {
		srv.Start()
	}()
	time.Sleep(2 * time.Second)

	network, _ := api.NewNetwork(config.Network)
	err = network.Transport().Send(network.LocalIP(), api.API.ServerStop())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var host *api.RemoteHost
	host, err = network.GetHost(network.LocalIP())
	if err == nil {
		t.Fatalf("error should be not nil")
	}

	if host != nil {
		t.Fatalf("host should be nil")
	}
	srv.Stop()
}
