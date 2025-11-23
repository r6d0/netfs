package volume

import (
	"errors"
	"netfs/internal/server/database"
	"os"
)

// The name of the volume table in the database.
const VolumeTable = "volume"

// If the volume is not found.
var ErrVolumeNotFound = errors.New("volume not found")

// If read operation is not permitted.
var ErrReadIsNotPermitted = errors.New("read operation is not permitted")

// If write operation is not permitted.
var ErrWriteIsNotPermitted = errors.New("write operation is not permitted")

// Volume fields.
type VolumeRecordField uint8

// Volume permitions.
type VolumePermition uint8

const (
	Name VolumeRecordField = iota
	Path
	Perm

	Read VolumePermition = iota
	Write
)

// Abstraction over the file system for working with files.
type Volume interface {
	Name() string
	Path() string
	Perm() VolumePermition
	Info(string) (os.FileInfo, error)
	Write(string, []byte) error
}

type volume struct {
	name string
	path string
	perm VolumePermition
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

func (vl *volume) Info(path string) (os.FileInfo, error) {
	if vl.perm&Read != 0 {
		// TODO. Get info logic
	}
	return nil, ErrReadIsNotPermitted
}

func (vl *volume) Write(path string, data []byte) error {
	if vl.perm&Write != 0 {
		// TODO. Write logic
	}
	return ErrWriteIsNotPermitted
}

// Converts database record to volume instrance.
func VolumeFromRecord(record database.Record) (Volume, error) {
	return &volume{
		name: string(record.GetField(uint8(Name))),
		path: string(record.GetField(uint8(Path))),
		perm: VolumePermition(record.GetUint8(uint8(Perm))),
	}, nil
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
			if volume, err = VolumeFromRecord(records[idx]); err == nil {
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
