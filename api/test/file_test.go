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
		return nil, []api.VolumeInfo{api.VolumeInfo{Id: testVolumeId}}, nil
	})
	rec.Receive(api.Endpoints.FileWrite.Name, func(req transport.Request) ([]byte, any, error) {
		volumeId, _ := req.ParamUInt64(api.Endpoints.FileWrite.VolumeId)
		if volumeId != testVolumeId {
			return nil, nil, errors.New("can't submit request")
		}

		fileId, _ := req.ParamUInt64(api.Endpoints.FileWrite.FileId)
		if fileId != testFileId {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, nil, nil
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
	err := file.Write(network.Transport(), []byte("TEST"))
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestWriteResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(req transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{api.VolumeInfo{Id: testVolumeId}}, nil
	})
	rec.Receive(api.Endpoints.FileWrite.Name, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
	err := file.Write(network.Transport(), []byte("TEST"))
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestCopyToSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(req transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{api.VolumeInfo{Id: testVolumeId}}, nil
	})
	rec.Receive(api.Endpoints.FileCopyStart, func(transport.Request) ([]byte, any, error) {
		return nil, api.RemoteTask{Id: 1, Status: api.Waiting, Host: local}, nil
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
	task, err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{Path: "./test_file_1.txt"}})
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

	rec.Receive(api.Endpoints.Volume, func(req transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{api.VolumeInfo{Id: testVolumeId}}, nil
	})
	rec.Receive(api.Endpoints.FileCopyStart, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
	_, err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{Path: "./test_file_1.txt"}})
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestFileRemoveSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(req transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{api.VolumeInfo{Id: testVolumeId}}, nil
	})
	rec.Receive(api.Endpoints.FileRemove.Name, func(req transport.Request) ([]byte, any, error) {
		volumeId, _ := req.ParamUInt64(api.Endpoints.FileWrite.VolumeId)
		if volumeId != testVolumeId {
			return nil, nil, errors.New("can't submit request")
		}

		fileId, _ := req.ParamUInt64(api.Endpoints.FileWrite.FileId)
		if fileId != testFileId {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, nil, nil
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
	err := file.Remove(network.Transport())
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestFileRemoveResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(req transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{api.VolumeInfo{Id: testVolumeId}}, nil
	})
	rec.Receive(api.Endpoints.FileRemove.Name, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
	err := file.Remove(network.Transport())
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestChildrenSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(req transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{api.VolumeInfo{Id: testVolumeId}}, nil
	})
	rec.Receive(api.Endpoints.FileChildren.Name, func(req transport.Request) ([]byte, any, error) {
		volumeId, _ := req.ParamUInt64(api.Endpoints.FileWrite.VolumeId)
		if volumeId != testVolumeId {
			return nil, nil, errors.New("can't submit request")
		}

		fileId, _ := req.ParamUInt64(api.Endpoints.FileWrite.FileId)
		if fileId != testFileId {
			return nil, nil, errors.New("can't submit request")
		}

		if req.Param(api.Endpoints.FileChildren.Skip) != "0" {
			return nil, nil, errors.New("can't submit request")
		}

		if req.Param(api.Endpoints.FileChildren.Limit) != "100" {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, []api.RemoteFile{{Info: api.FileInfo{Path: "./test_file_1.txt"}}}, nil
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
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
	volumes, _ := host.Volumes(network.Transport())
	file, _ := volumes[0].File(network.Transport(), testFileId)
	_, err := file.Children(network.Transport(), 0, 100)
	if err == nil {
		t.Fatal("error should be not nil")
	}
}
