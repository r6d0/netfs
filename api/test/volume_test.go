package api_test

import (
	"errors"
	"netfs/api"
	"netfs/api/transport"
	"testing"
)

func TestVolumeChildrenSuccess(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{{Name: "testvolume", OsPath: "./", LocalPath: "testvolume:/"}}, nil
	})
	rec.Receive(api.Endpoints.VolumeChildren.Name, func(req transport.Request) ([]byte, any, error) {
		if req.Param(api.Endpoints.VolumeChildren.Volume) != "testvolume" {
			return nil, nil, errors.New("can't submit request")
		}

		if req.Param(api.Endpoints.VolumeChildren.Skip) != "0" {
			return nil, nil, errors.New("can't submit request")
		}

		if req.Param(api.Endpoints.VolumeChildren.Limit) != "100" {
			return nil, nil, errors.New("can't submit request")
		}

		return nil, []api.FileInfo{{FileName: "test_file.txt", FilePath: "testvolume:/test_dir/test_file.txt", FileType: api.FILE}}, nil
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	files, err := volumes[0].Children(network.Transport(), 0, 100)
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if len(files) == 0 {
		t.Fatal("files should be not empty")
	}
}

func TestVolumeChildrenResponseError(t *testing.T) {
	beforeEach()
	defer afterEach()

	rec.Receive(api.Endpoints.Volume, func(transport.Request) ([]byte, any, error) {
		return nil, []api.VolumeInfo{{Name: "testvolume", OsPath: "./", LocalPath: "testvolume:/"}}, nil
	})
	rec.Receive(api.Endpoints.VolumeChildren.Name, func(transport.Request) ([]byte, any, error) {
		return nil, nil, errors.New("can't submit request")
	})

	host, _ := network.Host(local.IP)
	volumes, _ := host.Volumes(network.Transport())
	_, err := volumes[0].Children(network.Transport(), 0, 100)
	if err == nil {
		t.Fatalf("error should be not nil, but error is nil")
	}
}
