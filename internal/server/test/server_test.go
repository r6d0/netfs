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

// func TestStopServerHandleSuccess(t *testing.T) {
// 	config := server.ServerConfig{
// 		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
// 		Log:      logger.LoggerConfig{Level: logger.Info},
// 		Database: database.DatabaseConfig{Path: "./"},
// 		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}},
// 	}

// 	srv, err := server.NewServer(config)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(2 * time.Second)

// 	network, _ := api.NewNetwork(config.Network)
// 	err = network.Transport().Send(network.LocalIP(), api.API.ServerStop())
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	var host *api.RemoteHost
// 	host, err = network.GetHost(network.LocalIP())
// 	if err == nil {
// 		t.Fatalf("error should be not nil")
// 	}

// 	if host != nil {
// 		t.Fatalf("host should be nil")
// 	}
// 	srv.Stop()
// }

func TestFileInfoHandleSuccess(t *testing.T) {
	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
		Log:      logger.LoggerConfig{Level: logger.Info},
		Database: database.DatabaseConfig{Path: "./"},
		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 100},
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
	info, err := network.LocalHost().FileInfo(network.Transport(), "root:/myfile.txt")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}
	fmt.Println(info)
	srv.Stop()
}

// func TestFileCreateHandleSuccess(t *testing.T) {
// 	osPath, _ := filepath.Abs("./dir1")
// 	defer os.RemoveAll(osPath)

// 	config := server.ServerConfig{
// 		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
// 		Log:      logger.LoggerConfig{Level: logger.Info},
// 		Database: database.DatabaseConfig{Path: "./"},
// 		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 100},
// 	}

// 	srv, err := server.NewServer(config)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(2 * time.Second)

// 	network, _ := api.NewNetwork(config.Network)
// 	err = network.Transport().SendBody(network.LocalIP(), api.API.FileCreate(), api.FileInfo{FileName: "TestFileCreateHandleSuccess", FilePath: "root:/dir1/TestFileCreateHandleSuccess", FileType: api.FILE})
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	_, err = os.Stat(filepath.Join(osPath, "TestFileCreateHandleSuccess"))
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	srv.Stop()
// }

// func TestFileCopyStartHandleSuccess(t *testing.T) {
// 	// osPath, _ := filepath.Abs("./dir1")
// 	// defer os.RemoveAll(osPath)

// 	config := server.ServerConfig{
// 		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
// 		Log:      logger.LoggerConfig{Level: logger.Info},
// 		Database: database.DatabaseConfig{Path: "./"},
// 		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 1},
// 	}

// 	srv, err := server.NewServer(config)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(2 * time.Second)

// 	network, _ := api.NewNetwork(config.Network)
// 	transport := network.Transport()

// 	transport.SendBody(network.LocalIP(), api.API.FileCreate(), api.FileInfo{FileName: "TestFileCopyStartHandleSuccess", FilePath: "root:/dir1/TestFileCopyStartHandleSuccess", FileType: api.FILE})
// 	err = transport.SendBody(
// 		network.LocalIP(),
// 		api.API.FileCopyStart(),
// 		[]api.RemoteFile{
// 			{Host: network.LocalHost(), Info: api.FileInfo{FileName: "TestFileCopyStartHandleSuccess", FilePath: "root:/dir1/TestFileCopyStartHandleSuccess", FileType: api.FILE}},
// 			{Host: network.LocalHost(), Info: api.FileInfo{FileName: "TestFileCopyStartHandleSuccessCopy", FilePath: "root:/dir1/TestFileCopyStartHandleSuccessCopy", FileType: api.FILE}},
// 		},
// 	)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	time.Sleep(5 * time.Second)

// 	srv.Stop()
// }
