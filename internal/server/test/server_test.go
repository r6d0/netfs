package server_test

import (
	"fmt"
	"netfs/api"
	"netfs/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"testing"
	"time"
)

var config = server.ServerConfig{
	Network:  api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 1},
	Log:      logger.LoggerConfig{Level: logger.Info},
	Database: database.DatabaseConfig{Path: "./"},
	Task:     task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}, TasksWaitingSecond: 1},
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

func TestVolumesHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	volumes, err := host.Volumes(network.Transport())

	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if len(volumes) == 0 {
		t.Fatal("volumes should be not empty, but it is empty")
	}
}

func TestFileChildrenHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	volumes, _ := host.Volumes(network.Transport())

	dir, _ := volumes[0].Create(network.Transport(), api.FileInfo{Name: "test", Type: api.DIRECTORY, VolumeId: volumes[0].Info.Id})
	file, _ := volumes[0].Create(network.Transport(), api.FileInfo{Name: "test.txt", Type: api.FILE, VolumeId: volumes[0].Info.Id, ParentId: dir.Info.Id})

	children, err := dir.Children(network.Transport(), 0, 100)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if children[0].Info.Id != file.Info.Id {
		t.Fatalf("file id should be [%d], but file id is [%d]", children[0].Info.Id, file.Info.Id)
	}

	dir.Remove(network.Transport())
}

func TestFileCreateHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	volumes, _ := host.Volumes(network.Transport())

	file, err := volumes[0].Create(network.Transport(), api.FileInfo{Name: "test.txt", Type: api.FILE, VolumeId: volumes[0].Info.Id})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if file == nil {
		t.Fatal("file should be not nil")
	}

	file, _ = volumes[0].File(network.Transport(), file.Info.Id)
	if file == nil {
		t.Fatal("file should be not nil")
	}

	file.Remove(network.Transport())
}

func TestFileCopyStartHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()
	volumes, _ := host.Volumes(network.Transport())

	file, _ := volumes[0].Create(network.Transport(), api.FileInfo{Name: "test.txt", Type: api.FILE, VolumeId: volumes[0].Info.Id})
	_, err := file.CopyTo(network.Transport(), api.RemoteFile{Host: host, Info: api.FileInfo{Name: "test_copy.txt", Type: api.FILE, VolumeId: volumes[0].Info.Id}})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	time.Sleep(2 * time.Second)

	fileId := uint64(0)
	children, _ := volumes[0].Children(network.Transport(), 0, 100)
	for _, child := range children {
		if child.Info.Name == "test_copy.txt" {
			fileId = child.Info.Id
			break
		}
	}
	if fileId == 0 {
		t.Fatal("fileId should be not empty")
	}

	fileCopy, _ := volumes[0].File(network.Transport(), fileId)
	if fileCopy == nil {
		t.Fatal("file should be not nil")
	}

	file.Remove(network.Transport())
	fileCopy.Remove(network.Transport())
}
