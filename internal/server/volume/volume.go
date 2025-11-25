package volume

import (
	"errors"
	"netfs/internal/api"
	"netfs/internal/server/database"
	"strings"
)

const volumeSeparator = ":"

// The name of the volume table in the database.
const VolumeTable = "volume"

// Fields of the 'volume' table.
const (
	VolumeName database.RecordField = iota
	VolumePath
	VolumePerm
)

// The name of the files table in the database.
const VolumeFileTable = "volume_file"

// Fields of the 'volume_file' table.
const (
	FileName database.RecordField = iota
	FilePath
	FileSize
	FileType
)

// If the file is not found.
var ErrFileNotFound = errors.New("file is not found")

// If the volume is not found.
var ErrVolumeNotFound = errors.New("volume is not found")

// If read operation is not permitted.
var ErrReadIsNotPermitted = errors.New("read operation is not permitted")

// If write operation is not permitted.
var ErrWriteIsNotPermitted = errors.New("write operation is not permitted")

// Volume permitions.
type VolumePermition uint8

const (
	Read VolumePermition = 1 << iota
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
	id   uint64
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
		records, err := table.Get(database.Eq(FilePath, []byte(path)))
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
	// TODO. For testing only
	vlTable := db.Table(VolumeTable)
	vlRecord := database.NewRecord(3)
	vlRecord.SetRecordId(vlTable.NextId())
	vlRecord.SetField(VolumeName, []byte("root"))
	vlRecord.SetField(VolumePath, []byte("./"))
	vlRecord.SetUint8(VolumePerm, uint8(Read|Write))
	vlTable.Set(vlRecord)

	flTable := db.Table(VolumeFileTable)
	flRecord := database.NewRecord(5)
	flRecord.SetField(FileName, []byte("myfile.txt"))
	flRecord.SetField(FilePath, []byte("root:/myfile.txt"))
	flRecord.SetUint64(FileSize, 100)
	flRecord.SetUint8(FileType, uint8(api.FILE))
	flTable.Set(flRecord)

	records, err := vlTable.Get()
	if err == nil {
		volumes := make([]Volume, len(records))
		for idx := range records {
			var volume Volume
			if volume, err = volumeFromRecord(db, records[idx]); err == nil {
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

func (mng *volumeManager) Volume(path string) (Volume, error) {
	name := path
	if strings.Contains(path, volumeSeparator) {
		name = strings.Split(path, volumeSeparator)[0]
	}

	for _, volume := range mng.volumes {
		if volume.Name() == name {
			return volume, nil
		}
	}
	return nil, ErrVolumeNotFound
}

func volumeFromRecord(db database.Database, record database.Record) (Volume, error) {
	return &volume{
		id:   record.GetRecordId(),
		name: string(record.GetField(VolumeName)),
		path: string(record.GetField(VolumePath)),
		perm: VolumePermition(record.GetUint8(VolumePerm)),
		db:   db,
	}, nil
}

func fileInfoFromRecord(record database.Record) (*api.FileInfo, error) {
	return &api.FileInfo{
		FileName: string(record.GetField(FileName)),
		FilePath: string(record.GetField(FilePath)),
		FileType: api.FileType(record.GetUint8(FileType)),
		FileSize: int64(record.GetUint64(FileSize)),
	}, nil
}
