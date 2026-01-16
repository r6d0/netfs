package api_test

import (
	"errors"
	"netfs/api"
	"netfs/api/transport"
	"testing"
	"time"
)

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
	rec.Receive(api.Endpoints.FileInfo.Name, func(transport.Request) ([]byte, any, error) {
		return nil, api.FileInfo{FileName: "test_file.txt", FilePath: "./test_file.txt", FileType: api.FILE, FileSize: 1024}, nil
	})
	rec.Start()
}

func afterEach() {
	rec.Stop()
}

func TestWriteSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileWrite.Name, func(req transport.Request) ([]byte, any, error) {
		if req.Param(api.Endpoints.FileWrite.Path) != "./test_file.txt" {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, nil, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	err := file.Write(network.Transport(), []byte("TEST"))
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestWriteResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileWrite.Name, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	err := file.Write(network.Transport(), []byte("TEST"))
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestCreateSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileCreate, func(transport.Request) ([]byte, any, error) {
		return nil, nil, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	err := file.Create(network.Transport())
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestCreateResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileCreate, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	err := file.Create(network.Transport())
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestCopyToSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileCopyStart, func(transport.Request) ([]byte, any, error) {
		return nil, api.RemoteTask{Id: 1, Status: api.Waiting, Host: local}, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	task, err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{FilePath: "./test_file_1.txt"}})
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}
	if task == nil {
		t.Fatal("task should be not nil")
	}
	if task.Status != api.Waiting {
		t.Fatalf("status should be [%d]", api.Waiting)
	}
}

func TestCopyToResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()
	rec.Receive(api.Endpoints.FileCopyStart, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	_, err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{FilePath: "./test_file_1.txt"}})
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestFileRemoveSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileRemove.Name, func(req transport.Request) ([]byte, any, error) {
		if req.Param(api.Endpoints.FileRemove.Path) != "./test_file.txt" {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, nil, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	err := file.Remove(network.Transport())
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestFileRemoveResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileRemove.Name, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	err := file.Remove(network.Transport())
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestChildrenSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileChildren.Name, func(req transport.Request) ([]byte, any, error) {
		if req.Param(api.Endpoints.FileChildren.Path) != "./test_file.txt" {
			return nil, nil, errors.New("can't submit request")
		}

		if req.Param(api.Endpoints.FileChildren.Skip) != "0" {
			return nil, nil, errors.New("can't submit request")
		}

		if req.Param(api.Endpoints.FileChildren.Limit) != "100" {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, []api.RemoteFile{{Info: api.FileInfo{FilePath: "./test_file_1.txt"}}}, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	children, err := file.Children(network.Transport(), 0, 100)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if len(children) == 0 {
		t.Fatal("children should be not empty")
	}
}

func TestChildrenResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileChildren.Name, func(req transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), "./test_file.txt")
	_, err := file.Children(network.Transport(), 0, 100)
	if err == nil {
		t.Fatal("error should be not nil")
	}
}
