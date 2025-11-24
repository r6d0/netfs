package volume

import (
	"errors"
	"netfs/internal/api"
	"netfs/internal/server/database"
)

// The name of the volume table in the database.
const VolumeTable = "volume"

// The name of the files table in the database.
const VolumeFileTable = "volume_file"

// If the file is not found.
var ErrFileNotFound = errors.New("file is not found")

// If the volume is not found.
var ErrVolumeNotFound = errors.New("volume is not found")

// If read operation is not permitted.
var ErrReadIsNotPermitted = errors.New("read operation is not permitted")

// If write operation is not permitted.
var ErrWriteIsNotPermitted = errors.New("write operation is not permitted")

// Fields of the volume_file table.
type VolumeFileRecordField uint8

// Fields of the volume table.
type VolumeRecordField uint8

// Volume permitions.
type VolumePermition uint8

const (
	VolumeName VolumeRecordField = iota
	VolumePath
	VolumePerm

	FileName VolumeFileRecordField = iota
	FilePath
	FileSize
	FileType
	FileParent

	Read VolumePermition = iota
	Write
)

// Abstraction over the file system for working with files.
type Volume interface {
	Name() string
	Path() string
	Perm() VolumePermition
	Info(string) (*api.FileInfo, error)
	Read(string, int64, int64) ([]byte, error)
	Write(string, []byte) error
}

type volume struct {
	name string
	path string
	perm VolumePermition
	db   database.Database
}

func (vl *volume) Name() string {
	return vl.name
}

func (vl *volume) Path() string {
	return vl.path
}

func (vl *volume) Perm() VolumePermition {
	return vl.perm
}

func (vl *volume) Info(path string) (*api.FileInfo, error) {
	if vl.perm&Read != 0 {
		table := vl.db.Table(VolumeFileTable)
		records, err := table.Get(database.Eq(uint8(FilePath), []byte(path)))
		if err == nil {
			if len(records) == 1 {
				return fileInfoFromRecord(records[0])
			}
			err = ErrFileNotFound
		}
		return nil, err
	}
	return nil, ErrReadIsNotPermitted
}

func (vl *volume) Read(path string, offset int64, size int64) ([]byte, error) {
	if vl.perm&Read != 0 {
		// TODO. Read logic
	}
	return nil, ErrReadIsNotPermitted
}

func (vl *volume) Write(path string, data []byte) error {
	if vl.perm&Write != 0 {
		// TODO. Write logic
	}
	return ErrWriteIsNotPermitted
}

// Abstraction for working with volumes.
type VolumeManager interface {
	Volume(string) (Volume, error)
}

// Returns a new instance of the volume manager.
func NewVolumeManager(db database.Database) (VolumeManager, error) {
	table := db.Table(VolumeTable)
	records, err := table.Get()
	if err == nil {
		volumes := make([]Volume, len(records))
		for idx := range records {
			var volume Volume
			if volume, err = volumeFromRecord(records[idx]); err == nil {
				volumes[idx] = volume
			} else {
				break
			}
		}

		if err == nil {
			return &volumeManager{db: db, volumes: volumes}, nil
		}
	}
	return nil, err
}

type volumeManager struct {
	db      database.Database
	volumes []Volume
}

func (mng *volumeManager) Volume(name string) (Volume, error) {
	for _, volume := range mng.volumes {
		if volume.Name() == name {
			return volume, nil
		}
	}
	return nil, ErrVolumeNotFound
}

func volumeFromRecord(record database.Record) (Volume, error) {
	return &volume{
		name: string(record.GetField(uint8(VolumeName))),
		path: string(record.GetField(uint8(VolumePath))),
		perm: VolumePermition(record.GetUint8(uint8(VolumePerm))),
	}, nil
}

func fileInfoFromRecord(record database.Record) (*api.FileInfo, error) {
	return &api.FileInfo{
		FileName: string(record.GetField(uint8(FileName))),
		FilePath: string(record.GetField(uint8(FilePath))),
		FileType: api.FileType(record.GetUint8(uint8(FileType))),
		FileSize: int64(record.GetUint64(uint8(FileSize))),
	}, nil
}
