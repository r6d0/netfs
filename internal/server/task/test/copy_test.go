package task_test

// func TestCopyTaskToRecord(t *testing.T) {
// 	db := database.NewDatabase(database.DatabaseConfig{})
// 	table := db.Table("task")
// 	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{}, Target: api.RemoteFile{}}
// 	record, err := task.CopyTaskToRecord(table, copyTask)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	if record.GetUint8(task.Status) != uint8(copyTask.Status) {
// 		t.Fatalf("status should be [%d], but status is [%d]", record.GetUint8(task.Status), copyTask.Status)
// 	}

// 	if record.GetUint8(task.Type) != uint8(copyTask.Type) {
// 		t.Fatalf("type should be [%d], but type is [%d]", record.GetUint8(task.Type), copyTask.Type)
// 	}

// 	payload := record.GetField(task.Payload)
// 	data, _ := json.Marshal(copyTask)
// 	if !bytes.Equal(data, payload) {
// 		t.Fatalf("payload should be [%s], but payload is [%s]", string(data), string(payload))
// 	}
// }

// func TestCopyTaskFromRecord(t *testing.T) {
// 	db := database.NewDatabase(database.DatabaseConfig{})
// 	table := db.Table("task")
// 	orig := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{}, Target: api.RemoteFile{}}
// 	record, _ := task.CopyTaskToRecord(table, orig)
// 	copyTask, err := task.CopyTaskFromRecord(record)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	if orig.Status != copyTask.Status {
// 		t.Fatalf("status should be [%d], but status is [%d]", orig.Status, copyTask.Status)
// 	}

// 	if orig.Type != copyTask.Type {
// 		t.Fatalf("type should be [%d], but type is [%d]", orig.Type, copyTask.Type)
// 	}
// }

// func TestExecuteSuccess(t *testing.T) {
// 	generated := generate(100) // 100 bytes
// 	vlOsPath, _ := filepath.Abs("./")
// 	osPath, _ := filepath.Abs("./TestExecuteSuccess")
// 	os.WriteFile(osPath, generated, os.ModeAppend)

// 	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
// 		if tp[0] == api.API.FileWrite("")[0] {
// 			if !bytes.Equal(generated, data) {
// 				return nil, errors.New("data not equals")
// 			}
// 		}
// 		return nil, nil
// 	}}
// 	config := task.TaskExecuteConfig{Copy: task.TaskCopyConfig{BufferSize: 1024}} // 1024 bytes

// 	db := database.NewDatabase(database.DatabaseConfig{})
// 	vlTable := db.Table(volume.VolumeTable)
// 	vlRecord := database.NewRecord(3)
// 	vlRecord.SetRecordId(vlTable.NextId())
// 	vlRecord.SetField(volume.VolumeName, []byte("root"))
// 	vlRecord.SetField(volume.VolumePath, []byte(vlOsPath))
// 	vlRecord.SetUint8(volume.VolumePerm, uint8(volume.Read|volume.Write))
// 	vlTable.Set(vlRecord)

// 	flTable := db.Table(volume.VolumeFileTable)
// 	flRecord := database.NewRecord(5)
// 	flRecord.SetField(volume.FileName, []byte("TestExecuteSuccess"))
// 	flRecord.SetField(volume.FilePath, []byte("root:/TestExecuteSuccess"))
// 	flRecord.SetUint64(volume.FileSize, uint64(len(generated)))
// 	flRecord.SetUint8(volume.FileType, uint8(api.FILE))
// 	flTable.Set(flRecord)

// 	manager, _ := volume.NewVolumeManager(db)

// 	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 0, Source: api.RemoteFile{Info: api.FileInfo{FilePath: "root:/TestExecuteSuccess"}}, Target: api.RemoteFile{}}
// 	err := copyTask.Execute(task.TaskExecuteContext{Transport: &client, Config: config, Volumes: manager})
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	os.RemoveAll(osPath)
// }

// func TestExecuteChunkSuccess(t *testing.T) {
// 	generated := generate(100) // 100 bytes
// 	vlOsPath, _ := filepath.Abs("./")
// 	osPath, _ := filepath.Abs("./TestExecuteChunkSuccess")
// 	os.WriteFile(osPath, generated, os.ModeAppend)

// 	config := task.TaskExecuteConfig{Copy: task.TaskCopyConfig{BufferSize: 10}} // 10 bytes
// 	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
// 		if tp[0] == api.API.FileWrite("")[0] {
// 			if !bytes.Equal(generated[0:config.Copy.BufferSize], data) {
// 				return nil, errors.New("data not equals")
// 			}
// 		}
// 		return nil, nil
// 	}}

// 	db := database.NewDatabase(database.DatabaseConfig{})
// 	vlTable := db.Table(volume.VolumeTable)
// 	vlRecord := database.NewRecord(3)
// 	vlRecord.SetRecordId(vlTable.NextId())
// 	vlRecord.SetField(volume.VolumeName, []byte("root"))
// 	vlRecord.SetField(volume.VolumePath, []byte(vlOsPath))
// 	vlRecord.SetUint8(volume.VolumePerm, uint8(volume.Read|volume.Write))
// 	vlTable.Set(vlRecord)

// 	flTable := db.Table(volume.VolumeFileTable)
// 	flRecord := database.NewRecord(5)
// 	flRecord.SetField(volume.FileName, []byte("TestExecuteChunkSuccess"))
// 	flRecord.SetField(volume.FilePath, []byte("root:/TestExecuteChunkSuccess"))
// 	flRecord.SetUint64(volume.FileSize, uint64(len(generated)))
// 	flRecord.SetUint8(volume.FileType, uint8(api.FILE))
// 	flTable.Set(flRecord)

// 	manager, _ := volume.NewVolumeManager(db)

// 	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 0, Source: api.RemoteFile{Info: api.FileInfo{FilePath: "root:/TestExecuteChunkSuccess"}}, Target: api.RemoteFile{}}
// 	err := copyTask.Execute(task.TaskExecuteContext{Transport: &client, Config: config, Volumes: manager})
// 	if err != nil {
// 		t.Fatalf("error should be nil, but err is [%s]", err)
// 	}

// 	os.RemoveAll(osPath)
// }

// func TestExecuteFailure(t *testing.T) {
// 	config := task.TaskExecuteConfig{Copy: task.TaskCopyConfig{BufferSize: 10}} // 10 bytes
// 	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
// 		if tp[0] == api.API.FileWrite("")[0] {
// 			return nil, errors.New("")
// 		}
// 		return nil, nil
// 	}}

// 	db := database.NewDatabase(database.DatabaseConfig{})
// 	manager, _ := volume.NewVolumeManager(db)

// 	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{}, Target: api.RemoteFile{}}
// 	err := copyTask.Execute(task.TaskExecuteContext{Transport: &client, Config: config, Volumes: manager})
// 	if err == nil {
// 		t.Fatalf("error should be not nil")
// 	}
// }

// func generate(size int) []byte {
// 	result := make([]byte, size)
// 	for i := range size {
// 		result[i] = byte(1)
// 	}
// 	return result
// }
