package server_test

import (
	"bytes"
	"fmt"
	"netfs/internal/api"
	"netfs/internal/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"os"
	"path/filepath"
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
	host, err := network.Host(network.LocalIP())
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
	info, err := network.LocalHost().File(network.Transport(), "root:/myfile.txt")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}
	fmt.Println(info)
	srv.Stop()
}

func TestFileCreateHandleSuccess(t *testing.T) {
	osPath, _ := filepath.Abs("./dir1")
	defer os.RemoveAll(osPath)

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
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCreateHandleSuccess", FilePath: "root:/dir1/TestFileCreateHandleSuccess", FileType: api.FILE}}
	err = file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(filepath.Join(osPath, "TestFileCreateHandleSuccess"))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	srv.Stop()
}

func TestFileRemoveHandleSuccess(t *testing.T) {
	osPath, _ := filepath.Abs("./dir1")
	defer os.RemoveAll(osPath)

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
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCreateHandleSuccess", FilePath: "root:/dir1/TestFileCreateHandleSuccess", FileType: api.FILE}}
	file.Create(network.Transport())

	err = file.Remove(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(filepath.Join(osPath, "TestFileCreateHandleSuccess"))
	if err == nil {
		t.Fatal("error should be not nil, but err is nil")
	}

	srv.Stop()
}

func TestFileCopyStartHandleSuccess(t *testing.T) {
	osPath, _ := filepath.Abs("./dir1")
	osCopyPath, _ := filepath.Abs("./dir2")
	defer os.RemoveAll(osPath)
	defer os.RemoveAll(osCopyPath)

	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
		Log:      logger.LoggerConfig{Level: logger.Info},
		Database: database.DatabaseConfig{Path: "./"},
		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 1},
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
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStartHandleSuccess", FilePath: "root:/dir1/TestFileCopyStartHandleSuccess", FileType: api.FILE}}
	err = file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	generated := generate(100)
	err = file.Write(network.Transport(), generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var task *api.RemoteTask
	copyFile := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStartHandleSuccess", FilePath: "root:/dir2/TestFileCopyStartHandleSuccess", FileType: api.FILE}}
	task, err = file.CopyTo(network.Transport(), copyFile)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}
	if task == nil {
		t.Fatal("task should be not nil")
	}
	if task.Status != api.Waiting {
		t.Fatalf("status should be [%d]", api.Waiting)
	}
	time.Sleep(5 * time.Second)

	_, err = os.Stat(filepath.Join(osCopyPath, "TestFileCopyStartHandleSuccess"))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var data []byte
	data, err = os.ReadFile(filepath.Join(osCopyPath, "TestFileCopyStartHandleSuccess"))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(data, generated) {
		t.Fatalf("data should be equal")
	}

	srv.Stop()
}

func TestFileCopyStatusHandleSuccess(t *testing.T) {
	osPath, _ := filepath.Abs("./dir1")
	osCopyPath, _ := filepath.Abs("./dir2")
	defer os.RemoveAll(osPath)
	defer os.RemoveAll(osCopyPath)

	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
		Log:      logger.LoggerConfig{Level: logger.Info},
		Database: database.DatabaseConfig{Path: "./"},
		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 1},
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
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStatusHandleSuccess", FilePath: "root:/dir1/TestFileCopyStatusHandleSuccess", FileType: api.FILE}}
	err = file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	generated := generate(100)
	err = file.Write(network.Transport(), generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var task *api.RemoteTask
	copyFile := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStatusHandleSuccess", FilePath: "root:/dir2/TestFileCopyStatusHandleSuccess", FileType: api.FILE}}
	task, err = file.CopyTo(network.Transport(), copyFile)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}
	if task == nil {
		t.Fatal("task should be not nil")
	}
	if task.Status != api.Waiting {
		t.Fatalf("status should be [%d]", api.Waiting)
	}
	time.Sleep(5 * time.Second)

	task, err = host.Task(network.Transport(), task.Id)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}
	if task.Status != api.Completed {
		t.Fatalf("status should be [%d], but status is [%d]", api.Completed, task.Status)
	}

	srv.Stop()
}

func TestFileCopyCancelHandleSuccess(t *testing.T) {
	osPath, _ := filepath.Abs("./dir1")
	osCopyPath, _ := filepath.Abs("./dir2")
	defer os.RemoveAll(osPath)
	defer os.RemoveAll(osCopyPath)

	config := server.ServerConfig{
		Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
		Log:      logger.LoggerConfig{Level: logger.Info},
		Database: database.DatabaseConfig{Path: "./"},
		Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 1000},
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
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStopHandleSuccess", FilePath: "root:/dir1/TestFileCopyStopHandleSuccess", FileType: api.FILE}}
	err = file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	generated := generate(100)
	err = file.Write(network.Transport(), generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var task *api.RemoteTask
	copyFile := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStopHandleSuccess", FilePath: "root:/dir2/TestFileCopyStopHandleSuccess", FileType: api.FILE}}
	task, err = file.CopyTo(network.Transport(), copyFile)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	err = task.Cancel(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	task, _ = host.Task(network.Transport(), task.Id)
	if task.Status != api.Cancelled {
		t.Fatalf("status should be [%d], but status is [%d]", api.Cancelled, task.Status)
	}

	_, err = host.File(network.Transport(), copyFile.Info.FilePath)
	if err == nil {
		t.Fatal("error should be not nil, but err is nil")
	}

	srv.Stop()
}

func generate(size int) []byte {
	result := make([]byte, size)
	for i := range size {
		result[i] = byte(1)
	}
	return result
}
