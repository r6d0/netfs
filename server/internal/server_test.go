package server_test

import (
	"fmt"
	"netfs/api"
	"netfs/api/transport"
	server "netfs/server/internal"
	"path/filepath"
	"testing"
	"time"
)

var config = server.ServerConfig{
	Network: api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: time.Second * 1},
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

func TestFileChildrenHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()

	root, _ := filepath.Abs("./")
	dir, _ := host.Create(
		network.Transport(),
		api.FileInfo{Name: "test", Path: filepath.Join(root, "test"), Type: api.DIRECTORY},
	)
	file, _ := host.Create(
		network.Transport(),
		api.FileInfo{
			Name:     "test.txt",
			Type:     api.FILE,
			Path:     filepath.Join(string(dir.Info.Path), "test.txt"),
			ParentId: dir.Info.Id,
		},
	)

	children, err := dir.Children(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if children[0].Info.Id != file.Info.Id {
		t.Fatalf("file id should be [%v], but file id is [%v]", children[0].Info.Id, file.Info.Id)
	}

	dir.Remove(network.Transport())
}

func TestFileCreateHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()

	root, _ := filepath.Abs("./")
	file, err := host.Create(
		network.Transport(),
		api.FileInfo{Name: "test.txt", Path: filepath.Join(root, "test.txt"), Type: api.FILE},
	)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if file == nil {
		t.Fatal("file should be not nil")
	}

	file, _ = host.File(network.Transport(), file.Info.Id)
	if file == nil {
		t.Fatal("file should be not nil")
	}

	file.Remove(network.Transport())
}

func TestFileCreateHandleErrFileAlreadyExists(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()

	root, _ := filepath.Abs("./")
	file, err := host.Create(
		network.Transport(),
		api.FileInfo{Name: "test.txt", Path: filepath.Join(root, "test.txt"), Type: api.FILE},
	)

	_, err = host.Create(
		network.Transport(),
		api.FileInfo{Name: "test.txt", Path: filepath.Join(root, "test.txt"), Type: api.FILE},
	)
	if err == nil {
		t.Fatalf("error should be not nil")
	}

	file.Remove(network.Transport())
}

func TestFileCopyStartHandleSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	network, _ := api.NewNetwork(config.Network)
	host := network.LocalHost()

	root, _ := filepath.Abs("./")
	file, _ := host.Create(
		network.Transport(),
		api.FileInfo{Name: "test.txt", Path: filepath.Join(root, "test.txt"), Type: api.FILE},
	)
	file.Write(network.Transport(), generate(1024))

	_, err := file.CopyTo(
		network.Transport(),
		api.RemoteFile{
			Host: host,
			Info: api.FileInfo{
				Name: "text_copy.txt",
				Path: filepath.Join(root, "test_copy.txt"),
				Type: api.FILE,
			},
		},
	)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	time.Sleep(5 * time.Second)
}

func generate(size int) []byte {
	result := make([]byte, size)
	for i := range size {
		result[i] = byte(1)
	}
	return result
}
