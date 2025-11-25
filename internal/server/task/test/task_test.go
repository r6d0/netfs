package task

import (
	"bytes"
	"errors"
	"net"
	"netfs/internal/api"
	"netfs/internal/api/transport"
	"netfs/internal/logger"
	"netfs/internal/server/database"
	"netfs/internal/server/task"
	"os"
	"testing"
	"time"
)

func TestSubmitSuccess(t *testing.T) {
	db := database.NewDatabase(database.DatabaseConfig{})
	client := &transport.CallbackTransport{}
	log := logger.NewLogger(logger.LoggerConfig{})
	exec, _ := task.NewTaskExecutor(task.TaskExecuteConfig{}, db, client, log)

	copyTask, _ := task.NewCopyTask(api.RemoteFile{}, api.RemoteFile{})
	taskId, err := exec.Submit(copyTask)
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	table := db.Table(task.TaskTable)
	records, _ := table.Get(database.Id(taskId))
	if len(records) != 1 {
		t.Fatal("database should contains only one record")
	}
}

func TestStartSuccess(t *testing.T) {
	generated := generate(100) // 100 bytes
	os.WriteFile("./TestStartSuccess", generated, os.ModeAppend)
	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
		if tp[0] == api.API.FileWrite("")[0] {
			if !bytes.Equal(generated, data) {
				return nil, errors.New("data not equals")
			}
		}
		return nil, nil
	}}

	config := task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}} // 1024 bytes
	db := database.NewDatabase(database.DatabaseConfig{})
	log := logger.NewLogger(logger.LoggerConfig{})
	exec, _ := task.NewTaskExecutor(config, db, &client, log)

	copyTask := &task.CopyTask{Status: task.Completed, Type: task.Copy, Source: api.RemoteFile{Info: api.FileInfo{FilePath: "./TestStartSuccess"}}, Target: api.RemoteFile{}}
	taskId, _ := exec.Submit(copyTask)

	err := exec.Start()
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}

	time.Sleep(5 * time.Second)
	exec.Stop()

	table := db.Table(task.TaskTable)
	records, _ := table.Get(database.Id(taskId))
	status := records[0].GetUint8(task.Status)
	if status != uint8(task.Completed) {
		t.Fatalf("status should be Completed, but status is [%d]", status)
	}

	os.RemoveAll("./TestStartSuccess")
}

func TestStopSuccess(t *testing.T) {
	client := transport.CallbackTransport{Callback: func(ip net.IP, tp transport.TransportPoint, data []byte, result any) (any, error) {
		return nil, nil
	}}

	config := task.TaskExecuteConfig{MaxAvailableTasks: 100, Copy: task.TaskCopyConfig{BufferSize: 1024}} // 1024 bytes
	db := database.NewDatabase(database.DatabaseConfig{})
	log := logger.NewLogger(logger.LoggerConfig{})
	exec, _ := task.NewTaskExecutor(config, db, &client, log)

	exec.Start()
	time.Sleep(2 * time.Second)

	err := exec.Stop()
	if err != nil {
		t.Fatalf("error should be nil, but err is [%s]", err)
	}
}

func generate(size int) []byte {
	result := make([]byte, size)
	for i := range size {
		result[i] = byte(1)
	}
	return result
}
