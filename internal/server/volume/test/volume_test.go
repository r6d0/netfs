package volume_test

import (
	"bytes"
	"io"
	"netfs/internal/api"
	"netfs/internal/server/database"
	"netfs/internal/server/volume"
	"os"
	"path/filepath"
	"testing"
)

func TestNewVolumeManagerSuccess(t *testing.T) {
	db := database.NewDatabase(database.DatabaseConfig{})
	table := db.Table(volume.VolumeTable)

	record := database.NewRecord(3)
	record.SetRecordId(table.NextId())
	record.SetField(volume.VolumeName, []byte("root"))
	record.SetField(volume.VolumePath, []byte("./"))
	record.SetUint8(volume.VolumePerm, uint8(volume.Read|volume.Write))
	table.Set(record)

	manager, err := volume.NewVolumeManager(db)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if manager == nil {
		t.Fatal("manager should be not nil")
	}
}

func TestVolumeSuccess(t *testing.T) {
	db := database.NewDatabase(database.DatabaseConfig{})
	table := db.Table(volume.VolumeTable)

	record := database.NewRecord(3)
	record.SetRecordId(table.NextId())
	record.SetField(volume.VolumeName, []byte("root"))
	record.SetField(volume.VolumePath, []byte("./"))
	record.SetUint8(volume.VolumePerm, uint8(volume.Read|volume.Write))
	table.Set(record)

	manager, _ := volume.NewVolumeManager(db)
	vl, err := manager.Volume("root")
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
	_, err := manager.Volume("root")
	if err != volume.ErrVolumeNotFound {
		t.Fatalf("error should be [%s], but err is [%s]", volume.ErrVolumeNotFound, err)
	}
}

func TestInfoSuccess(t *testing.T) {
	db := database.NewDatabase(database.DatabaseConfig{})

	vlTable := db.Table(volume.VolumeTable)
	vlRecord := database.NewRecord(3)
	vlRecord.SetRecordId(vlTable.NextId())
	vlRecord.SetField(volume.VolumeName, []byte("root"))
	vlRecord.SetField(volume.VolumePath, []byte("./"))
	vlRecord.SetUint8(volume.VolumePerm, uint8(volume.Read|volume.Write))
	vlTable.Set(vlRecord)

	flTable := db.Table(volume.VolumeFileTable)
	flRecord := database.NewRecord(5)
	flRecord.SetField(volume.FileName, []byte("myfile.txt"))
	flRecord.SetField(volume.FilePath, []byte("root:/myfile.txt"))
	flRecord.SetUint64(volume.FileSize, 100)
	flRecord.SetUint8(volume.FileType, uint8(api.FILE))
	flTable.Set(flRecord)

	manager, _ := volume.NewVolumeManager(db)
	vl, _ := manager.Volume("root")

	info, err := vl.Info("root:/myfile.txt")
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if info == nil {
		t.Fatal("info should be not nil")
	}
}

func TestReadSuccess(t *testing.T) {
	generated := generate(100) // 100 bytes
	osPath, _ := filepath.Abs("./TestStartSuccess")
	os.WriteFile(osPath, generated, os.ModeAppend)

	db := database.NewDatabase(database.DatabaseConfig{})

	vlTable := db.Table(volume.VolumeTable)
	vlRecord := database.NewRecord(3)
	vlRecord.SetRecordId(vlTable.NextId())
	vlRecord.SetField(volume.VolumeName, []byte("root"))
	vlRecord.SetField(volume.VolumePath, []byte("./"))
	vlRecord.SetUint8(volume.VolumePerm, uint8(volume.Read|volume.Write))
	vlTable.Set(vlRecord)

	flTable := db.Table(volume.VolumeFileTable)
	flRecord := database.NewRecord(5)
	flRecord.SetRecordId(flTable.NextId())
	flRecord.SetField(volume.FileName, []byte("TestReadSuccess"))
	flRecord.SetField(volume.FilePath, []byte("root:/TestReadSuccess"))
	flRecord.SetUint64(volume.FileSize, uint64(len(generated)))
	flRecord.SetUint8(volume.FileType, uint8(api.FILE))
	flRecord.SetField(volume.FileOsPath, []byte(osPath))
	flTable.Set(flRecord)

	manager, _ := volume.NewVolumeManager(db)
	vl, _ := manager.Volume("root")

	data, err := vl.Read("root:/TestReadSuccess", 0, int64(len(generated)))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(generated, data) {
		t.Fatalf("the data should be equal to the generated")
	}

	data, err = vl.Read("root:/TestReadSuccess", int64(len(generated)/2), int64(len(generated)))
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if !bytes.Equal(generated[:len(generated)/2], data) {
		t.Fatalf("the data should be equal to the generated")
	}

	os.RemoveAll(osPath)
}

func TestWriteSuccess(t *testing.T) {
	generated := generate(100) // 100 bytes
	osPath, _ := filepath.Abs("./TestWriteSuccess")

	db := database.NewDatabase(database.DatabaseConfig{})

	vlTable := db.Table(volume.VolumeTable)
	vlRecord := database.NewRecord(3)
	vlRecord.SetRecordId(vlTable.NextId())
	vlRecord.SetField(volume.VolumeName, []byte("root"))
	vlRecord.SetField(volume.VolumePath, []byte("./"))
	vlRecord.SetUint8(volume.VolumePerm, uint8(volume.Read|volume.Write))
	vlTable.Set(vlRecord)

	flTable := db.Table(volume.VolumeFileTable)
	flRecord := database.NewRecord(5)
	flRecord.SetRecordId(flTable.NextId())
	flRecord.SetField(volume.FileName, []byte("TestWriteSuccess"))
	flRecord.SetField(volume.FilePath, []byte("root:/TestWriteSuccess"))
	flRecord.SetUint64(volume.FileSize, uint64(len(generated)))
	flRecord.SetUint8(volume.FileType, uint8(api.FILE))
	flRecord.SetField(volume.FileOsPath, []byte(osPath))
	flTable.Set(flRecord)

	manager, _ := volume.NewVolumeManager(db)
	vl, _ := manager.Volume("root")

	err := vl.Write("root:/TestWriteSuccess", generated)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	file, _ := os.Open(osPath)
	data, _ := io.ReadAll(file)
	if !bytes.Equal(generated, data) {
		t.Fatalf("the data should be equal to the generated")
	}

	file.Close()
	os.RemoveAll(osPath)
}

func generate(size int) []byte {
	result := make([]byte, size)
	for i := range size {
		result[i] = byte(1)
	}
	return result
}
