package server_test

import (
	"bytes"
	"fmt"
	"netfs/api"
	"netfs/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var config = server.ServerConfig{
	Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 5},
	Log:      logger.LoggerConfig{Level: logger.Info},
	Database: database.DatabaseConfig{Path: "./"},
	Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}},
}

var srv *server.Server

func beforeEach() {
	var err error
	srv, err = server.NewServer(config)
	if err != nil {
		panic(fmt.Sprintf("error should be nil, but err is [%s]", err))
	}
	go func() {
		srv.Start()
	}()
}

func afterEach() {
	srv.Stop()
}

func TestServerHostHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host, err := network.Host(network.LocalIP())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if host == nil {
		t.Fatalf("host should be not nil")
	}
}

func TestFileInfoHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	path := "testvolume:/testdir/testfile.txt"

	network, _ := api.NewNetwork(config.Network)
	info, err := network.LocalHost().File(network.Transport(), path)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if info.Info.FilePath != path {
		t.Fatalf("file: [%s] not found", path)
	}
}

func TestFileChildrenHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	path := "testvolume:/"

	network, _ := api.NewNetwork(config.Network)
	file, err := network.LocalHost().File(network.Transport(), path)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	children, err := file.Children(network.Transport(), 0, 100)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if len(children) == 0 {
		t.Fatal("children should be not empty")
	}
}

func TestFileCreateHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	path := "testvolume:/dir1/TestFileCreateHandleSuccess"

	osPath, _ := filepath.Abs("./dir1")
	defer os.RemoveAll(osPath)

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCreateHandleSuccess", FilePath: path, FileType: api.FILE}}
	err := file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(filepath.Join(osPath, "TestFileCreateHandleSuccess"))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}
}

func TestFileRemoveHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	path := "testvolume:/dir1/TestFileRemoveHandleSuccess"

	osPath, _ := filepath.Abs("./dir1")
	defer os.RemoveAll(osPath)

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileRemoveHandleSuccess", FilePath: path, FileType: api.FILE}}
	err := file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	err = file.Remove(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(filepath.Join(osPath, "TestFileRemoveHandleSuccess"))
	if err == nil {
		t.Fatal("error should be not nil, but err is nil")
	}
}

func TestFileCopyStartHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	originalPath := "testvolume:/TestFileCopyStartHandleSuccess/TestFileCopyStartHandleSuccess.txt"
	copyPath := "testvolume:/TestFileCopyStartHandleSuccessCopy/TestFileCopyStartHandleSuccessCopy.txt"

	osPath, _ := filepath.Abs("./TestFileCopyStartHandleSuccess")
	osCopyPath, _ := filepath.Abs("./TestFileCopyStartHandleSuccessCopy")
	defer os.RemoveAll(osPath)
	defer os.RemoveAll(osCopyPath)

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStartHandleSuccess.txt", FilePath: originalPath, FileType: api.FILE}}
	err := file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	generated := generate(100)
	err = file.Write(network.Transport(), generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var task *api.RemoteTask
	copyFile := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStartHandleSuccessCopy.txt", FilePath: copyPath, FileType: api.FILE}}
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

	_, err = os.Stat(filepath.Join(osCopyPath, copyFile.Info.FileName))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var data []byte
	data, err = os.ReadFile(filepath.Join(osCopyPath, copyFile.Info.FileName))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(data, generated) {
		t.Fatalf("data should be equal")
	}
}

func TestFileCopyStatusHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	originalPath := "testvolume:/TestFileCopyStatusHandleSuccess/TestFileCopyStatusHandleSuccess.txt"
	copyPath := "testvolume:/TestFileCopyStatusHandleSuccessCopy/TestFileCopyStatusHandleSuccessCopy.txt"

	osPath, _ := filepath.Abs("./TestFileCopyStatusHandleSuccess")
	osCopyPath, _ := filepath.Abs("./TestFileCopyStatusHandleSuccessCopy")
	defer os.RemoveAll(osPath)
	defer os.RemoveAll(osCopyPath)

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStatusHandleSuccess.txt", FilePath: originalPath, FileType: api.FILE}}
	err := file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	generated := generate(100)
	err = file.Write(network.Transport(), generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var task *api.RemoteTask
	copyFile := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyStatusHandleSuccessCopy.txt", FilePath: copyPath, FileType: api.FILE}}
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
}

func TestFileCopyCancelHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	originalPath := "testvolume:/TestFileCopyCancelHandleSuccess/TestFileCopyCancelHandleSuccess.txt"
	copyPath := "testvolume:/TestFileCopyCancelHandleSuccessCopy/TestFileCopyCancelHandleSuccessCopy.txt"

	osPath, _ := filepath.Abs("./TestFileCopyCancelHandleSuccess")
	osCopyPath, _ := filepath.Abs("./TestFileCopyCancelHandleSuccessCopy")
	defer os.RemoveAll(osPath)
	defer os.RemoveAll(osCopyPath)

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	file := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyCancelHandleSuccess.txt", FilePath: originalPath, FileType: api.FILE}}
	err := file.Create(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	generated := generate(100)
	err = file.Write(network.Transport(), generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	var task *api.RemoteTask
	copyFile := api.RemoteFile{Host: host, Info: api.FileInfo{FileName: "TestFileCopyCancelHandleSuccessCopy.txt", FilePath: copyPath, FileType: api.FILE}}
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
}

func generate(size int) []byte {
	result := make([]byte, size)
	for i := range size {
		result[i] = byte(1)
	}
	return result
}
