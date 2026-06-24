package api_test

import (
	"errors"
	"netfs/api"
	"netfs/api/transport"
	"testing"
)

func TestWriteSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileWrite.Name, func(req transport.Request) ([]byte, any, error) {
		fileId, _ := req.ParamRequired(api.Endpoints.FileWrite.FileId)
		if api.FileId(fileId) != testFileId {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, nil, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), testFileId)
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
	file, _ := host.File(network.Transport(), testFileId)
	err := file.Write(network.Transport(), []byte("TEST"))
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestCopyToSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileCopyStart, func(req transport.Request) ([]byte, any, error) {
		task := &api.RemoteCopyTask{}
		_, err := req.Body(task)
		if err == nil {
			return nil, api.RemoteCopyTask{Id: api.TaskId("1"), Status: api.Running, Host: local}, nil
		}
		return nil, nil, err
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), testFileId)
	task, err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{Path: "./test_file_1.txt"}})
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}
	if task == nil {
		t.Fatal("task should be not nil")
	}
	if task.Status != api.Running {
		t.Fatalf("status should be [%d]", api.Running)
	}
}

func TestCopyToResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileCopyStart, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), testFileId)
	_, err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{Path: "./test_file_1.txt"}})
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestFileRemoveSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileRemove.Name, func(req transport.Request) ([]byte, any, error) {
		fileId, _ := req.ParamRequired(api.Endpoints.FileWrite.FileId)
		if api.FileId(fileId) != testFileId {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, nil, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), testFileId)
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
	file, _ := host.File(network.Transport(), testFileId)
	err := file.Remove(network.Transport())
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestChildrenSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.FileChildren.Name, func(req transport.Request) ([]byte, any, error) {
		fileId, _ := req.ParamRequired(api.Endpoints.FileWrite.FileId)
		if api.FileId(fileId) != testFileId {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, []api.RemoteFile{{Info: api.FileInfo{Path: "./test_file_1.txt"}}}, nil
	})

	host, _ := network.Host(local.IP)
	file, _ := host.File(network.Transport(), testFileId)
	children, err := file.Children(network.Transport())
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
	file, _ := host.File(network.Transport(), testFileId)
	_, err := file.Children(network.Transport())
	if err == nil {
		t.Fatal("error should be not nil")
	}
}
