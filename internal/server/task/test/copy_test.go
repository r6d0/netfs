package task_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"netfs/internal/api"
	"netfs/internal/api/transport"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"os"
	"testing"
)

func TestCopyTaskToRecord(t *testing.T) {
	db := database.NewDatabase(database.DatabaseConfig{})
	table := db.Table("task")
	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{}, Target: api.RemoteFile{}}
	record, err := task.CopyTaskToRecord(table, copyTask)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if record.GetUint8(task.Status) != uint8(copyTask.Status) {
		t.Fatalf("status should be [%d], but status is [%d]", record.GetUint8(task.Status), copyTask.Status)
	}

	if record.GetUint8(task.Type) != uint8(copyTask.Type) {
		t.Fatalf("type should be [%d], but type is [%d]", record.GetUint8(task.Type), copyTask.Type)
	}

	payload := record.GetField(task.Payload)
	data, _ := json.Marshal(copyTask)
	if !bytes.Equal(data, payload) {
		t.Fatalf("payload should be [%s], but payload is [%s]", string(data), string(payload))
	}
}

func TestCopyTaskFromRecord(t *testing.T) {
	db := database.NewDatabase(database.DatabaseConfig{})
	table := db.Table("task")
	orig := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{}, Target: api.RemoteFile{}}
	record, _ := task.CopyTaskToRecord(table, orig)
	copyTask, err := task.CopyTaskFromRecord(record)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	if orig.Status != copyTask.Status {
		t.Fatalf("status should be [%d], but status is [%d]", orig.Status, copyTask.Status)
	}

	if orig.Type != copyTask.Type {
		t.Fatalf("type should be [%d], but type is [%d]", orig.Type, copyTask.Type)
	}
}

func TestExecuteSuccess(t *testing.T) {
	generated := generate(100) // 100 bytes
	os.WriteFile("./TestExecuteSuccess", generated, os.ModeAppend)

	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
		if tp[0] == api.API.FileWrite("")[0] {
			if !bytes.Equal(generated, data) {
				return nil, errors.New("data not equals")
			}
		}
		return nil, nil
	}}
	config := task.TaskExecuteConfig{Copy: task.TaskCopyConfig{BufferSize: 1024}} // 1024 bytes

	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{Info: api.FileInfo{FilePath: "./TestExecuteSuccess"}}, Target: api.RemoteFile{}}
	err := copyTask.Execute(task.TaskExecuteContext{Transport: &client, Config: config})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	os.RemoveAll("./TestExecuteSuccess")
}

func TestExecuteChunkSuccess(t *testing.T) {
	generated := generate(100) // 100 bytes
	os.WriteFile("./TestExecuteChunkSuccess", generated, os.ModeAppend)

	config := task.TaskExecuteConfig{Copy: task.TaskCopyConfig{BufferSize: 10}} // 10 bytes
	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
		if tp[0] == api.API.FileWrite("")[0] {
			if !bytes.Equal(generated[0:config.Copy.BufferSize], data) {
				return nil, errors.New("data not equals")
			}
		}
		return nil, nil
	}}

	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{Info: api.FileInfo{FilePath: "./TestExecuteChunkSuccess"}}, Target: api.RemoteFile{}}
	err := copyTask.Execute(task.TaskExecuteContext{Transport: &client, Config: config})
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	os.RemoveAll("./TestExecuteChunkSuccess")
}

func TestExecuteFailure(t *testing.T) {
	config := task.TaskExecuteConfig{Copy: task.TaskCopyConfig{BufferSize: 10}} // 10 bytes
	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
		if tp[0] == api.API.FileWrite("")[0] {
			return nil, errors.New("")
		}
		return nil, nil
	}}

	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Offset: 555, Source: api.RemoteFile{}, Target: api.RemoteFile{}}
	err := copyTask.Execute(task.TaskExecuteContext{Transport: &client, Config: config})
	if err == nil {
		t.Fatalf("error should be not nil")
	}
}

func generate(size int) []byte {
	result := make([]byte, size)
	for i := range size {
		result[i] = byte(1)
	}
	return result
}
