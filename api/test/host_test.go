package api_test

import (
	"errors"
	"netfs/api"
	"netfs/api/transport"
	"testing"
)

func TestFileInfoSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	host, _ := network.Host(local.IP)
	file, err := host.File(network.Transport(), "testvolume:/test_dir/test_file.txt")
	if err != nil {
		t.Fatal("error should be nil")
	}

	if file.Info.FileType != api.FILE {
		t.Fatal("element should be a file")
	}

	if file.Info.FileName != "test_file.txt" {
		t.Fatal("file name should be [test_file.txt]")
	}

	if file.Info.FilePath != "testvolume:/test_dir/test_file.txt" {
		t.Fatal("file path should be [testvolume:/test_dir/test_file.txt]")
	}
}

func TestTaskSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileCopyStatus.Name, func(transport.Request) ([]byte, any, error) {
		return nil, api.RemoteTask{Id: 1, Status: api.Completed, Host: local}, nil
	})

	host, _ := network.Host(local.IP)
	task, err := host.Task(network.Transport(), 1)
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}
	if task.Status != api.Completed {
		t.Fatalf("status should be equal [%d], but status is [%d]", api.Completed, task.Status)
	}
}

func TestTaskResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileCopyStatus.Name, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	_, err := host.Task(network.Transport(), 1)
	if err == nil {
		t.Fatalf("error should be not nil, but error is nil")
	}
}

func TestVolumesSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{{Name: "testvolume", OsPath: "./", LocalPath: "testvolume:/"}}, nil
	})

	host, _ := network.Host(local.IP)
	volumes, err := host.Volumes(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if len(volumes) == 0 {
		t.Fatal("volumes should be not empty")
	}
}

func TestVolumesResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	_, err := host.Volumes(network.Transport())
	if err == nil {
		t.Fatalf("error should be not nil, but error is nil")
	}
}
