package volume

import (
	"errors"
	"io"
	"io/fs"
	"netfs/api"
	"netfs/internal/collection"
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
	FileParentId
	FileVolumeId
)

// If the file is not found.
var ErrFileNotFound = errors.New("file is not found")

// If file has incorrect path.
var ErrFileHasIncorrectPath = errors.New("file has incorrect path")

// If the volume is not found.
var ErrVolumeNotFound = errors.New("volume is not found")

// Abstraction over the file system for working with files.
type Volume interface {
	// The function returns information about volume.
	Info() api.VolumeInfo
	// The function scans all volume elements.
	ReIndex() error
	Size() int64
	// The function returns information about the file.
	File(uint64) (*api.FileInfo, error)
	// The function returns children of the directory.
	Children(uint64, int, int) ([]api.FileInfo, error)
	// The function creates a new file or directory.
	Create(*api.FileInfo) (*api.FileInfo, error)
	// The function reads data from a file.
	Read(uint64, int64, int64) ([]byte, error)
	// The function writes data to a file.
	Write(uint64, []byte) error
	// The function removes the file.
	Remove(uint64) error
	// The function returns the path in OS for the local path.
	ResolveOsPath(string) string
	// The function returns the local path for the path in OS.
	ResolveLocalPath(string) string
}

type volume struct {
	info api.VolumeInfo
	db   database.Database
}

func (vl *volume) Info() api.VolumeInfo {
	return vl.info
}

func (vl *volume) ReIndex() error {
	table := vl.db.Table(VolumeFileTable)
	err := table.Del(database.EqUInt64(FileVolumeId, vl.info.Id))

	if err == nil {
		current := &api.FileInfo{Path: vl.info.LocalPath}
		stack := collection.Stack[*api.FileInfo]{}

		err = filepath.WalkDir(vl.info.OsPath, func(path string, dir fs.DirEntry, err error) error {
			path = NormalizePath(path, dir.IsDir())
			if path != vl.info.OsPath {
				local := vl.ResolveLocalPath(path)
				info := &api.FileInfo{Id: vl.info.Id, Path: local, Name: dir.Name(), VolumeId: vl.info.Id}

				for !strings.HasPrefix(local, current.Path) {
					current = stack.Pop()
				}
				info.ParentId = current.Id

				if dir.IsDir() {
					info.Type = api.DIRECTORY

					stack.Push(current)
					current = info
				} else {
					info.Type = api.FILE

					var fsInfo fs.FileInfo
					if fsInfo, err = dir.Info(); err == nil {
						info.Size = uint64(fsInfo.Size()) // TODO. Use uint64(fsInfo.Size())
					}
				}

				_, err = vl.Create(info)
			}
			return err
		})
	}
	return err
}

func (vl *volume) Size() int64 { // TODO. add it
	return 0
}

func (vl *volume) File(fileId uint64) (*api.FileInfo, error) {
	table := vl.db.Table(VolumeFileTable)
	record, err := table.Any(database.Id(fileId))
	if record != nil && err == nil {
		return &api.FileInfo{
			Id:       fileId,
			Name:     record.GetString(FileName),
			Path:     record.GetString(FilePath),
			Type:     api.FileType(record.GetUint8(FileType)),
			Size:     record.GetUint64(FileSize),
			ParentId: record.GetUint64(FileParentId),
			VolumeId: vl.info.Id,
		}, nil
	}
	return nil, errors.Join(ErrFileNotFound, err)
}

func (vl *volume) Children(fileId uint64, skip int, limit int) ([]api.FileInfo, error) {
	table := vl.db.Table(VolumeFileTable)
	records, err := table.Get(
		database.EqUInt64(FileParentId, fileId),
		database.Skip(int16(skip)),
		database.Limit(int16(limit)),
	)

	if err == nil {
		result := make([]api.FileInfo, len(records))
		for idx, record := range records {
			result[idx] = api.FileInfo{
				Id:       record.GetRecordId(),
				Name:     record.GetString(FileName),
				Path:     record.GetString(FilePath),
				Type:     api.FileType(record.GetUint8(FileType)),
				Size:     record.GetUint64(FileSize),
				ParentId: record.GetUint64(FileParentId),
				VolumeId: vl.info.Id,
			}
		}

		return result, nil
	}
	return nil, err
}

func (vl *volume) Create(info *api.FileInfo) (*api.FileInfo, error) {
	var err error

	table := vl.db.Table(VolumeFileTable)
	path := info.Path
	if path == "" {
		if info.ParentId > 0 {
			var parent database.Record
			if parent, err = table.Any(database.Id(info.ParentId)); err == nil {
				if parent != nil {
					path = NormalizePath(parent.GetString(FilePath)+info.Name, info.Type == api.DIRECTORY)
				} else {
					err = ErrFileNotFound
				}
			}
		} else {
			path = NormalizePath(vl.info.LocalPath+info.Name, info.Type == api.DIRECTORY)
		}
	}

	if err == nil {
		err = table.Txn(func(table database.Table) error {
			record := table.New()
			record.SetString(FileName, info.Name)
			record.SetString(FilePath, path)
			record.SetUint64(FileSize, info.Size)
			record.SetUint8(FileType, uint8(info.Type))
			record.SetUint64(FileParentId, info.ParentId)
			record.SetUint64(FileVolumeId, info.VolumeId)

			txnErr := table.Set(record)
			if txnErr == nil {
				if info.Type == api.DIRECTORY {
					txnErr = os.Mkdir(vl.ResolveOsPath(path), 0755) // TODO. 0755?
				} else {
					var file *os.File
					if file, txnErr = os.Create(vl.ResolveOsPath(path)); txnErr == nil {
						file.Close()
					}
				}

				if txnErr == nil || errors.Is(txnErr, fs.ErrExist) {
					info.Id = record.GetRecordId()
					info.Path = path
					return nil
				}
			}
			return txnErr
		})
	}
	return info, err
}

func (vl *volume) Read(fileId uint64, offset int64, size int64) ([]byte, error) {
	table := vl.db.Table(VolumeFileTable)
	record, err := table.Any(database.Id(fileId))
	if err == nil {
		if record != nil {
			osPath := vl.ResolveOsPath(record.GetString(FilePath))

			var file *os.File
			if file, err = os.Open(osPath); err == nil {
				defer file.Close() // TODO. Add file to cache

				read := 0
				data := make([]byte, min(size, int64(record.GetUint64(FileSize)))) // TODO. record.GetInt64(FileSize)
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

func (vl *volume) Write(fileId uint64, data []byte) error {
	table := vl.db.Table(VolumeFileTable)
	record, err := table.Any(database.Id(fileId))
	if err == nil {
		if record != nil {
			fileOsPath := vl.ResolveOsPath(record.GetString(FilePath))

			var file *os.File
			if file, err = os.OpenFile(fileOsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer file.Close() // TODO. Add file to cache?

				_, err = file.Write(data)
				if err == nil {
					size := record.GetUint64(FileSize) + uint64(len(data))
					record.SetUint64(FileSize, size)

					err = table.Set(record)
				}
			}
		} else {
			err = ErrFileNotFound
		}
	}
	return err
}

func (vl *volume) Remove(fileId uint64) error {
	table := vl.db.Table(VolumeFileTable)
	record, err := table.Any(database.Id(fileId))
	if err == nil {
		if record != nil {
			// Remove file or directory.
			if err = table.Del(database.Id(fileId)); err == nil {
				// Remove all children.
				if err = table.Del(database.EqUInt64(FileParentId, fileId)); err == nil {
					path := record.GetString(FilePath)
					err = os.RemoveAll(vl.ResolveOsPath(path))
				}
			}
		} else {
			return ErrFileNotFound
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
	// The functions creates a volume.
	Create(api.VolumeInfo) (Volume, error)
	// The function returns the volume by id.
	Volume(uint64) (Volume, error)
	// The function returns all volumes.
	Volumes() ([]Volume, error)
}

// Returns a new instance of the volume manager.
func NewVolumeManager(db database.Database) (VolumeManager, error) {
	table := db.Table(VolumeTable)
	records, err := table.Get()
	if err == nil {
		volumes := make([]Volume, len(records))
		for idx, record := range records {
			volumes[idx] = &volume{
				info: api.VolumeInfo{
					Id:        record.GetRecordId(),
					Name:      record.GetString(VolumeName),
					OsPath:    record.GetString(VolumeOsPath),
					LocalPath: record.GetString(VolumeLocalPath),
				},
				db: db,
			}
		}

		return &volumeManager{db: db}, nil
	}
	return nil, err
}

type volumeManager struct {
	db database.Database
}

func (mng *volumeManager) Create(info api.VolumeInfo) (Volume, error) {
	table := mng.db.Table(VolumeTable)

	osPath := NormalizePath(info.OsPath, true)
	localPath := NormalizePath(info.LocalPath, true)

	record := table.New()
	record.SetString(VolumeName, info.Name)
	record.SetString(VolumeOsPath, osPath)
	record.SetString(VolumeLocalPath, localPath)

	err := table.Set(record)
	if err == nil {
		if err = os.MkdirAll(info.OsPath, 0755); err == nil { // TODO. 0755?
			return &volume{
				info: api.VolumeInfo{
					Id:        record.GetRecordId(),
					Name:      info.Name,
					OsPath:    osPath,
					LocalPath: localPath,
				},
				db: mng.db,
			}, nil
		}
	}
	return nil, err
}

func (mng *volumeManager) Volume(volumeId uint64) (Volume, error) {
	table := mng.db.Table(VolumeTable)
	record, err := table.Any(database.Id(volumeId))
	if err == nil {
		if record != nil {
			return &volume{
				info: api.VolumeInfo{
					Id:        record.GetRecordId(),
					Name:      record.GetString(VolumeName),
					LocalPath: record.GetString(VolumeLocalPath),
					OsPath:    record.GetString(VolumeOsPath),
				},
				db: mng.db,
			}, nil
		} else {
			err = ErrVolumeNotFound
		}
	}
	return nil, err
}

func (mng *volumeManager) Volumes() ([]Volume, error) {
	table := mng.db.Table(VolumeTable)
	records, err := table.Get()
	if err == nil {
		volumes := make([]Volume, len(records))
		for index, record := range records {
			volumes[index] = &volume{
				info: api.VolumeInfo{
					Id:        record.GetRecordId(),
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
