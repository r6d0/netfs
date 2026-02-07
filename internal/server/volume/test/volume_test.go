package volume_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"netfs/api"
	"netfs/internal/server/database"
	"netfs/internal/server/volume"
	"os"
	"path/filepath"
	"testing"
)

func TestVolumeSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	vl, err := manager.Volume("testvolume")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if vl == nil {
		t.Fatal("volume should be not nil")
	}
}

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
}

func TestVolumeErrVolumeNotFound(t *testing.T) {
	db := database.NewDatabase(database.DatabaseConfig{})

	manager, _ := volume.NewVolumeManager(db)
	_, err := manager.Volume("TestVolumeErrVolumeNotFound")
	if errors.Is(volume.ErrVolumeNotFound, err) {
		t.Fatalf("error should be [%s], but err is [%s]", volume.ErrVolumeNotFound, err)
	}
}

func TestFileSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	vl.Create(&api.FileInfo{FileName: "TestFileSuccess.txt", FilePath: "testvolume:/TestFileSuccess.txt", FileType: api.FILE})

	info, err := vl.File("testvolume:/TestFileSuccess.txt")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if info == nil {
		t.Fatal("info should be not nil")
	}

	vl.Remove(info.FilePath)
}

func TestChildrenSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	vl.Create(&api.FileInfo{FileName: "TestChildrenSuccess.txt", FilePath: "testvolume:/TestChildrenSuccess.txt", FileType: api.FILE})

	children, err := vl.Children("testvolume:/", 0, 100)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if len(children) == 0 {
		t.Fatal("children should be not empty")
	}

	vl.Remove("testvolume:/TestChildrenSuccess.txt")
}

func TestReadSuccess(t *testing.T) {
	generated := generate(100)

	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	vl.Create(&api.FileInfo{FileName: "TestReadSuccess.txt", FilePath: "testvolume:/TestReadSuccess.txt", FileType: api.FILE})
	vl.Write("testvolume:/TestReadSuccess.txt", generated)

	data, err := vl.Read("testvolume:/TestReadSuccess.txt", 0, int64(len(generated)))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(generated, data) {
		t.Fatalf("the data should be equal to the generated")
	}

	data, err = vl.Read("testvolume:/TestReadSuccess.txt", int64(len(generated)/2), int64(len(generated)))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(generated[:len(generated)/2], data) {
		t.Fatalf("the data should be equal to the generated")
	}

	vl.Remove("testvolume:/TestReadSuccess.txt")
}

func TestWriteSuccess(t *testing.T) {
	generated := generate(100)

	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	vl.Create(&api.FileInfo{FileName: "TestWriteSuccess.txt", FilePath: "testvolume:/TestWriteSuccess.txt", FileType: api.FILE})

	err := vl.Write("testvolume:/TestWriteSuccess.txt", generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	file, _ := os.Open(vl.ResolveOsPath("testvolume:/TestWriteSuccess.txt"))
	data, _ := io.ReadAll(file)
	if !bytes.Equal(generated, data) {
		t.Fatalf("the data should be equal to the generated")
	}

	file.Close()
	vl.Remove("testvolume:/TestWriteSuccess.txt")
}

func TestResolvePathSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	osPath, _ := filepath.Abs("./TestResolvePathSuccess.txt")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	vl.Create(&api.FileInfo{FileName: "TestResolvePathSuccess.txt", FilePath: "testvolume:/TestResolvePathSuccess.txt", FileType: api.FILE})

	path := vl.ResolveOsPath("testvolume:/TestResolvePathSuccess.txt")
	if path != volume.NormalizePath(osPath, false) {
		t.Fatalf("path should be [%s], but it is [%s]", osPath, path)
	}

	vl.Remove("testvolume:/TestResolvePathSuccess.txt")
}

func TestCreateDirectorySuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})

	path := "testvolume:/TestCreateDirectorySuccess_1/TestCreateDirectorySuccess_2/TestCreateDirectorySuccess_3/"
	err := vl.Create(&api.FileInfo{FileName: "TestCreateDirectorySuccess_3", FilePath: path, FileType: api.DIRECTORY})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(vl.ResolveOsPath(path))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = vl.File(path)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = vl.File("testvolume:/TestCreateDirectorySuccess_1/TestCreateDirectorySuccess_2/")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = vl.File("testvolume:/TestCreateDirectorySuccess_1/")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	vl.Remove("testvolume:/TestCreateDirectorySuccess_1/")
}

func TestCreateFileSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})

	path := "testvolume:/TestCreateFileSuccess_1/TestCreateFileSuccess_2/TestCreateFileSuccess_3/TestCreateFileSuccess.txt"
	err := vl.Create(&api.FileInfo{FileName: "TestCreateFileSuccess.txt", FilePath: path, FileType: api.FILE})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(vl.ResolveOsPath(path))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = vl.File(path)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	vl.Remove("testvolume:/TestCreateFileSuccess_1/")
}

func TestRemoveFileSuccess(t *testing.T) {
	vlOsPath, _ := filepath.Abs("./")
	manager, _ := volume.NewVolumeManager(database.NewDatabase(database.DatabaseConfig{}))
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", LocalPath: "testvolume:/", OsPath: vlOsPath})
	vl.Create(&api.FileInfo{FileName: "TestRemoveFileSuccess.txt", FilePath: "testvolume:/TestRemoveFileSuccess.txt", FileType: api.FILE})

	err := vl.Remove("testvolume:/TestRemoveFileSuccess.txt")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	_, err = os.Stat(vl.ResolveOsPath("testvolume:/TestRemoveFileSuccess.txt"))
	if err == nil {
		t.Fatal("error should be not nil, but err is nil")
	}

	_, err = vl.File("testvolume:/TestRemoveFileSuccess.txt")
	if err == nil {
		t.Fatal("error should be not nil, but err is nil")
	}
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
		os.Mkdir(filepath.Join(vlOsPath, fmt.Sprintf("TestReIndexSuccess_%d", index)), os.ModeAppend)
	}
	vl, _ := manager.Create(api.VolumeInfo{Name: "testvolume", OsPath: vlOsPath, LocalPath: "testvolume:/"})

	err := vl.ReIndex()
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	for index := range 100 {
		_, err = vl.File(fmt.Sprintf("testvolume:/TestReIndexSuccess_%d/", index))
		if err != nil {
			t.Fatalf("error should be nil, but err is [%s]", err)
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
