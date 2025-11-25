package volume_test

import (
	"netfs/internal/api"
	"netfs/internal/server/database"
	"netfs/internal/server/volume"
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
