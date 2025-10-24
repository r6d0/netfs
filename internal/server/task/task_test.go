package task_test

// import (
// 	"bytes"
// 	netfs "netfs/internal"
// 	"netfs/internal/logger"
// 	"netfs/internal/server/task"
// 	"netfs/internal/store"
// 	"os"
// 	"testing"
// )

// const dbPath = "./data"

// func TestSubmitSuccess(t *testing.T) {
// 	log := logger.NewLogger(logger.LoggerConfig{Level: logger.Info})
// 	db := store.NewStore(store.StoreConfig{Path: dbPath})
// 	db.Start()

// 	source := netfs.RemoteFile{}
// 	target := netfs.RemoteFile{}
// 	cp, _ := task.NewCopyTask(source, target)

// 	exec, _ := task.NewTaskExecutor(task.TaskConfig{}, db, log)
// 	err := exec.Submit(cp)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err.Error())
// 	}

// 	item, _ := cp.ToWaiting()
// 	saved, _ := db.Get(item.Key)
// 	if !bytes.Equal(item.Value, saved.Value) {
// 		t.Fatalf("value should be equals to expected value")
// 	}

// 	db.Stop()
// 	os.RemoveAll(dbPath)
// }

// func TestSubmitFailure(t *testing.T) {
// 	log := logger.NewLogger(logger.LoggerConfig{Level: logger.Info})
// 	db := store.NewStore(store.StoreConfig{Path: dbPath})

// 	source := netfs.RemoteFile{}
// 	target := netfs.RemoteFile{}
// 	cp, _ := task.NewCopyTask(source, target)

// 	exec, _ := task.NewTaskExecutor(task.TaskConfig{}, db, log)
// 	err := exec.Submit(cp)
// 	if err != store.NotStarted {
// 		t.Fatalf("error should be [%s], but error is [%s]", store.NotStarted, err)
// 	}

// 	db.Stop()
// 	os.RemoveAll(dbPath)
// }

// func TestCopyDirectory(t *testing.T) {
// 	log := logger.NewLogger(logger.LoggerConfig{Level: logger.Info})
// 	db := store.NewStore(store.StoreConfig{Path: dbPath})
// 	db.Start()

// 	source := netfs.RemoteFile{Path: "./test_directory"}
// 	target := netfs.RemoteFile{Path: "./test_directory_copy"}
// 	cp, _ := task.NewCopyTask(source, target)

// 	exec, _ := task.NewTaskExecutor(task.TaskConfig{}, db, log)
// 	err := exec.Submit(cp)
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err.Error())
// 	}

// 	item, _ := cp.ToWaiting()
// 	saved, _ := db.Get(item.Key)
// 	if !bytes.Equal(item.Value, saved.Value) {
// 		t.Fatalf("value should be equals to expected value")
// 	}

// 	db.Stop()
// 	os.RemoveAll(dbPath)
// }
