package main

import (
	"netfs/api"
	"netfs/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"time"
)

func main() {
	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 8989, Protocol: transport.HTTP, Timeout: time.Second * 5},
		Log:      logger.LoggerConfig{Level: logger.Info},
		Database: database.DatabaseConfig{Path: "./"},
		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 5},
	}

	srv, err := server.NewServer(config)
	if err != nil {
		panic(err)
	}

	if err = srv.Start(); err != nil {
		panic(err)
	}
}
