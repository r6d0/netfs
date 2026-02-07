package volume

import (
	"errors"
	"io"
	"io/fs"
	"netfs/api"
	"netfs/internal/server/database"
	"os"
	"path/filepath"
	"strings"
)

const volumeSeparator = ":"
const pathSeparator = "/"

// The name of the volume table in the database.
const VolumeTable = "volume"

// Fields of the 'volume' table.
const (
	VolumeName database.RecordField = iota
	VolumeOsPath
	VolumeLocalPath
)

// The name of the files table in the database.
const VolumeFileTable = "volume_file"

// Fields of the 'volume_file' table.
const (
	FileName database.RecordField = iota
	FilePath
	FileSize
	FileType
	FileParentPath
	FileVolumeId
)

// If the file is not found.
var ErrFileNotFound = errors.New("file is not found")

// If the volume is not found.
var ErrVolumeNotFound = errors.New("volume is not found")

// Abstraction over the file system for working with files.
type Volume interface {
	// The function returns information about volume.
	Info() api.VolumeInfo
	// The function scans all volume elements.
	ReIndex() error
	// The current volume size.
	Size() int64
	File(string) (*api.FileInfo, error)
	Children(string, int, int) ([]api.FileInfo, error)
	Create(*api.FileInfo) error
	Read(string, int64, int64) ([]byte, error)
	Write(string, []byte) error
	Remove(string) error
	// The function returns the path in OS for the local path.
	ResolveOsPath(string) string
	// The function returns the local path for the path in OS.
	ResolveLocalPath(string) string
}

type volume struct {
	id   uint64
	info api.VolumeInfo
	db   database.Database
}

func (vl *volume) Info() api.VolumeInfo {
	return vl.info
}

func (vl *volume) ReIndex() error {
	table := vl.db.Table(VolumeFileTable)
	err := table.Del(database.EqUInt64(FileVolumeId, vl.id))

	if err == nil {
		err = filepath.WalkDir(vl.info.OsPath, func(path string, dir fs.DirEntry, err error) error {
			if path != vl.info.OsPath {
				var info fs.FileInfo
				if info, err = dir.Info(); err == nil {
					isDir := info.IsDir()
					parent, name := filepath.Split(path)
					normalizedPath := NormalizePath(path, isDir)
					normalizedParentPath := NormalizePath(parent, true)

					record := table.New()
					record.SetString(FileName, name)
					record.SetString(FilePath, vl.ResolveLocalPath(normalizedPath))
					record.SetString(FileParentPath, vl.ResolveLocalPath(normalizedParentPath))
					record.SetUint64(FileSize, uint64(info.Size()))
					record.SetUint64(FileVolumeId, vl.id)
					if isDir {
						record.SetUint8(FileType, uint8(api.DIRECTORY))
					} else {
						record.SetUint8(FileType, uint8(api.FILE))
					}

					err = table.Set(record)
				}
			}
			return err
		})
	}
	return err
}

func (vl *volume) Size() int64 { // TODO. add it
	return 0
}

func (vl *volume) File(path string) (*api.FileInfo, error) {
	table := vl.db.Table(VolumeFileTable)
	record, err := table.Any(database.Eq(FilePath, []byte(path)))
	if record != nil && err == nil {
		return fileInfoFromRecord(record)
	}
	return nil, errors.Join(ErrFileNotFound, err)
}

func (vl *volume) Children(path string, skip int, limit int) ([]api.FileInfo, error) {
	table := vl.db.Table(VolumeFileTable)
	records, err := table.Get(
		database.Eq(FileParentPath, []byte(path)),
		database.Skip(int16(skip)),
		database.Limit(int16(limit)),
	)

	if err == nil {
		result := make([]api.FileInfo, len(records))
		for index, record := range records {
			var info *api.FileInfo
			if info, err = fileInfoFromRecord(record); err == nil {
				result[index] = *info
			} else {
				break
			}
		}

		if err == nil {
			return result, nil
		}
	}
	return nil, err
}

func (vl *volume) Create(info *api.FileInfo) error {
	var err error

	normalized := NormalizePath(info.FilePath, info.FileType == api.DIRECTORY)
	path := vl.ResolveOsPath(normalized)
	if info.FileType == api.DIRECTORY {
		err = os.MkdirAll(path, 0755) // TODO. 0755?
	} else {
		dir, _ := filepath.Split(path)
		if err = os.MkdirAll(dir, 0755); err == nil { // TODO. 0755?
			var file *os.File
			if file, err = os.Create(path); err == nil {
				file.Close()
			}
		}
	}

	if err == nil {
		end := vl.info.LocalPath
		current := filepath.ToSlash(info.FilePath)
		cleanCurrent, _ := strings.CutSuffix(current, pathSeparator)
		current, _ = filepath.Split(cleanCurrent)

		table := vl.db.Table(VolumeFileTable)
		records := []database.Record{fileInfoToRecord(table, current, vl.id, info)}
		for current != end {
			cleanCurrent, _ := strings.CutSuffix(current, pathSeparator)

			next, name := filepath.Split(cleanCurrent)
			fileInfo := &api.FileInfo{FileName: name, FilePath: current, FileType: api.DIRECTORY}
			records = append(records, fileInfoToRecord(table, next, vl.id, fileInfo))
			current = next
		}
		err = table.Set(records...)
	}
	return err
}

func (vl *volume) Read(path string, offset int64, size int64) ([]byte, error) {
	table := vl.db.Table(VolumeFileTable)
	records, err := table.Get(database.Eq(FilePath, []byte(path)))
	if err == nil {
		if len(records) == 1 {
			fileSize := records[0].GetUint64(FileSize)
			fileOsPath := vl.ResolveOsPath(string(records[0].GetField(FilePath)))

			var file *os.File
			if file, err = os.Open(fileOsPath); err == nil {
				defer file.Close() // TODO. Add file to cache

				read := 0
				data := make([]byte, min(size, int64(fileSize)))
				read, err = file.ReadAt(data, offset)
				if err == nil || errors.Is(err, io.EOF) {
					return data[:read], nil
				}
			}
		} else {
			err = ErrFileNotFound
		}
	}
	return nil, err
}

func (vl *volume) Write(path string, data []byte) error {
	table := vl.db.Table(VolumeFileTable)
	records, err := table.Get(database.Eq(FilePath, []byte(path)))
	if err == nil {
		if len(records) == 1 {
			record := records[0]
			fileOsPath := vl.ResolveOsPath(string(record.GetField(FilePath)))

			var file *os.File
			if file, err = os.OpenFile(fileOsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer file.Close() // TODO. Add file to cache

				_, err = file.Write(data)
				if err == nil {
					size := record.GetUint64(FileSize) + uint64(len(data))
					record.SetUint64(FileSize, size)

					err = table.Set(records...)
				}
			}
		} else {
			err = ErrFileNotFound
		}
	}
	return err
}

func (vl *volume) Remove(path string) error {
	table := vl.db.Table(VolumeFileTable)
	records, err := table.Get(database.Eq(FilePath, []byte(path))) // TODO. remove children of directory
	if err == nil {
		if len(records) == 1 {
			if err = table.Del(database.Id(records[0].GetRecordId())); err == nil {
				fileOsPath := vl.ResolveOsPath(string(records[0].GetField(FilePath)))
				err = os.RemoveAll(fileOsPath)
			}
		}
	}
	return err
}

func (vl *volume) ResolveOsPath(original string) string {
	if strings.HasPrefix(original, vl.info.LocalPath) {
		path, _ := strings.CutPrefix(original, vl.info.LocalPath)
		return vl.info.OsPath + path
	}
	return original
}

func (vl *volume) ResolveLocalPath(original string) string {
	if strings.HasPrefix(original, vl.info.OsPath) {
		path, _ := strings.CutPrefix(original, vl.info.OsPath)
		return vl.info.LocalPath + path
	}
	return original
}

// Abstraction for working with volumes.
type VolumeManager interface {
	Create(api.VolumeInfo) (Volume, error)
	Volume(string) (Volume, error)
	Volumes() ([]Volume, error)
}

// Returns a new instance of the volume manager.
func NewVolumeManager(db database.Database) (VolumeManager, error) {
	table := db.Table(VolumeTable)
	records, err := table.Get()
	if err == nil {
		volumes := make([]Volume, len(records))
		for idx := range records {
			var volume Volume
			if volume, err = volumeFromRecord(db, records[idx]); err == nil {
				volume.ReIndex()
				volumes[idx] = volume
			} else {
				break
			}
		}

		if err == nil {
			return &volumeManager{db: db}, nil
		}
	}
	return nil, err
}

type volumeManager struct {
	db database.Database
}

func (mng *volumeManager) Create(info api.VolumeInfo) (Volume, error) {
	table := mng.db.Table(VolumeTable)

	info.OsPath = NormalizePath(info.OsPath, true)
	info.LocalPath = NormalizePath(info.LocalPath, true)

	record := table.New()
	record.SetString(VolumeName, info.Name)
	record.SetString(VolumeOsPath, info.OsPath)
	record.SetString(VolumeLocalPath, info.LocalPath)

	err := table.Set(record)
	if err == nil {
		if err = os.MkdirAll(info.OsPath, 0755); err == nil { // TODO. 0755?
			return &volume{info: info, db: mng.db}, nil
		}
	}
	return nil, err
}

func (mng *volumeManager) Volume(path string) (Volume, error) {
	name := path
	if strings.Contains(path, volumeSeparator) {
		name = strings.Split(path, volumeSeparator)[0]
	}

	table := mng.db.Table(VolumeTable)
	record, err := table.Any(database.Eq(VolumeName, []byte(name)))
	if record != nil && err == nil {
		return &volume{
			id: record.GetRecordId(),
			info: api.VolumeInfo{
				Name:      record.GetString(VolumeName),
				LocalPath: record.GetString(VolumeLocalPath),
				OsPath:    record.GetString(VolumeOsPath),
			},
			db: mng.db,
		}, nil
	}
	return nil, errors.Join(ErrVolumeNotFound, err)
}

func (mng *volumeManager) Volumes() ([]Volume, error) {
	table := mng.db.Table(VolumeTable)
	records, err := table.Get()
	if err == nil {
		volumes := make([]Volume, len(records))
		for index, record := range records {
			volumes[index] = &volume{
				id: record.GetRecordId(),
				info: api.VolumeInfo{
					Name:      record.GetString(VolumeName),
					LocalPath: record.GetString(VolumeLocalPath),
					OsPath:    record.GetString(VolumeOsPath),
				},
				db: mng.db,
			}
		}
		return volumes, nil
	}
	return nil, err
}

func NormalizePath(path string, isDir bool) string {
	path = strings.ReplaceAll(path, string(filepath.Separator), pathSeparator)
	if isDir {
		if !strings.HasSuffix(path, pathSeparator) {
			path = path + pathSeparator
		}
	}
	return path
}

func volumeFromRecord(db database.Database, record database.Record) (Volume, error) {
	return &volume{
		id: record.GetRecordId(),
		info: api.VolumeInfo{
			Name:      string(record.GetField(VolumeName)),
			LocalPath: strings.Join([]string{string(record.GetField(VolumeName)), volumeSeparator, pathSeparator}, ""),
			OsPath:    string(record.GetField(VolumeOsPath)),
		},
		db: db,
	}, nil
}

func fileInfoToRecord(table database.Table, parentPath string, volumeId uint64, info *api.FileInfo) database.Record {
	record := table.New()
	record.SetField(FileName, []byte(info.FileName))
	record.SetField(FilePath, []byte(info.FilePath))
	record.SetUint8(FileType, uint8(info.FileType))
	record.SetUint64(FileSize, uint64(info.FileSize))
	record.SetField(FileParentPath, []byte(parentPath))
	record.SetUint64(FileVolumeId, volumeId)
	return record
}

func fileInfoFromRecord(record database.Record) (*api.FileInfo, error) {
	return &api.FileInfo{
		FileName: string(record.GetField(FileName)),
		FilePath: string(record.GetField(FilePath)),
		FileType: api.FileType(record.GetUint8(FileType)),
		FileSize: int64(record.GetUint64(FileSize)),
	}, nil
}
