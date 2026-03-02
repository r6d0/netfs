package volume_test

import (
	"bytes"
	"fmt"
	"io"
	"netfs/api"
	"netfs/internal/server/database"
	"netfs/internal/server/volume"
	"os"
	"path/filepath"
	"testing"
)

func TestVolumeCreateSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, err := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if vl == nil {
		t.Fatal("volume should be not nil")
	}

	vl, err = manager.Volume(vl.Info().Id)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if vl == nil {
		t.Fatal("volume should be not nil")
	}
}

func TestVolumesSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})

	vls, err := manager.Volumes()
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if len(vls) == 0 {
		t.Fatal("volumes should be not empty")
	}

	if vls[0].Info().Id != vl.Info().Id {
		t.Fatalf("id should be [%d], but it is [%d]", vl.Info().Id, vls[0].Info().Id)
	}
}

func TestChildrenSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})

	created, _ := vl.Create(&api.FileInfo{Name: "TestChildrenSuccess", Type: api.DIRECTORY, ParentId: vl.Info().Id, VolumeId: vl.Info().Id})
	createdChild, _ := vl.Create(&api.FileInfo{Name: "TestChildrenSuccess.txt", Type: api.FILE, ParentId: created.Id, VolumeId: vl.Info().Id})

	children, err := vl.Children(created.Id, 0, 100)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if len(children) == 0 {
		t.Fatal("children should be not empty")
	}

	if children[0].Id != createdChild.Id {
		t.Fatalf("id should be [%d], but it is [%d]", createdChild.Id, children[0].Id)
	}

	vl.Remove(created.Id)
	vl.Remove(createdChild.Id)
}

func TestReadSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	created, _ := vl.Create(&api.FileInfo{Name: "TestReadSuccess.txt", Type: api.FILE, VolumeId: vl.Info().Id})

	generated := generate(100)
	vl.Write(created.Id, generated)

	data, err := vl.Read(created.Id, 0, int64(len(generated)))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(generated, data) {
		t.Fatalf("the data should be equal to the generated")
	}

	data, err = vl.Read(created.Id, int64(len(generated)/2), int64(len(generated)))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(generated[:len(generated)/2], data) {
		t.Fatalf("the data should be equal to the generated")
	}

	vl.Remove(created.Id)
}

func TestWriteSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	created, _ := vl.Create(&api.FileInfo{Name: "TestWriteSuccess.txt", Type: api.FILE, VolumeId: vl.Info().Id})

	generated := generate(100)
	err := vl.Write(created.Id, generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	file, _ := os.Open(vl.ResolveOsPath(created.Path))
	data, _ := io.ReadAll(file)
	if !bytes.Equal(generated, data) {
		t.Fatalf("the data should be equal to the generated")
	}

	file.Close()
	vl.Remove(created.Id)
}

func TestResolvePathSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	osPath := filepath.Join(vlOsPath, "testfile.txt")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})

	created, _ := vl.Create(&api.FileInfo{Name: "testfile.txt", Type: api.FILE, VolumeId: vl.Info().Id})
	path := vl.ResolveOsPath(created.Path)
	if path != volume.NormalizePath(osPath, false) {
		t.Fatalf("path should be [%s], but it is [%s]", osPath, path)
	}

	vl.Remove(created.Id)
}

func TestCreateDirectorySuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})

	created, err := vl.Create(&api.FileInfo{Name: "TestCreateDirectorySuccess", Type: api.DIRECTORY, VolumeId: vl.Info().Id})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(vl.ResolveOsPath(created.Path))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = vl.File(created.Id)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	createdChild, err := vl.Create(&api.FileInfo{Name: "TestCreateDirectorySuccess2", Type: api.DIRECTORY, VolumeId: vl.Info().Id})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(vl.ResolveOsPath(createdChild.Path))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = vl.File(createdChild.Id)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	vl.Remove(created.Id)
	vl.Remove(createdChild.Id)
}

func TestCreateFileSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})

	created, err := vl.Create(&api.FileInfo{Name: "TestCreateFileSuccess.txt", Type: api.FILE, VolumeId: vl.Info().Id})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(vl.ResolveOsPath(created.Path))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = vl.File(created.Id)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	vl.Remove(created.Id)
}

func TestNormalizePathSuccess(t *testing.T) {
	separator := string(filepath.Separator)
	path := "a:" + separator + "testdir" + separator + "testfile.txt"

	normalized := volume.NormalizePath(path, false)
	if normalized != "a:/testdir/testfile.txt" {
		t.Fatalf("path should be [%s], but path is [%s]", "a:/testdir/testfile.txt", normalized)
	}

	path = "a:" + separator + "testdir" + separator + "testdir2"
	normalized = volume.NormalizePath(path, true)
	if normalized != "a:/testdir/testdir2/" {
		t.Fatalf("path should be [%s], but path is [%s]", "a:/testdir/testdir2/", normalized)
	}
}

func TestReIndexSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./testvolume")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))

	os.Mkdir(vlOsPath, os.ModeAppend)
	for index := range 100 {
		os.MkdirAll(filepath.Join(vlOsPath, fmt.Sprintf("TestReIndexSuccess_%d/TestReIndexSuccess_%d", index, index)), os.ModeAppend)
	}
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", OsPath: vlOsPath, LocalPath: "testvolume:/"})

	err := vl.ReIndex()
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	children, _ := vl.Children(0, 0, 100)
	if len(children) != 100 {
		t.Fatalf("children size should be [100], but size is [%d]", len(children))
	}

	for _, child := range children {
		children, _ := vl.Children(child.Id, 0, 100)
		if len(children) != 1 {
			t.Fatalf("children size should be [1], but size is [%d]", len(children))
		}
	}

	os.RemoveAll(vlOsPath)
}

func generate(size int) []byte {
	result := make([]byte, size)
	for i := range size {
		result[i] = byte(1)
	}
	return result
}
