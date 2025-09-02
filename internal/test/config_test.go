package netfs_test

import (
	"encoding/json"
	netfs "netfs/internal"
	"os"
	"testing"
)

const EXPECTED_USER_PATH = "./user-netfs-config.json"
const EXPECTED_PATH = "./netfs-config.json"
const EXPECTED_PORT = 49153
const EXPECTED_PROTOCOL = "http"
const EXPECTED_DATABASE_PATH = "./data"
const EXPECTED_BUFFER_SIZE = 1024 * 1024 // 1MB
const EXPECTED_TASKS_COUNT = 10

func TestNewConfigDefaultValues(t *testing.T) {
	config, err := netfs.NewConfig()
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if config.BufferSize != EXPECTED_BUFFER_SIZE {
		t.Fatalf("[config.BufferSize] should be [%d], but error is [%d]", EXPECTED_BUFFER_SIZE, config.BufferSize)
	}

	if config.TaskCount != EXPECTED_TASKS_COUNT {
		t.Fatalf("[config.TaskCount] should be [%d], but value is [%d]", EXPECTED_TASKS_COUNT, config.TaskCount)
	}

	if config.Server.Port != EXPECTED_PORT {
		t.Fatalf("[config.Server.Port] should be [%d], but value is [%d]", EXPECTED_PORT, config.Server.Port)
	}

	if config.Server.Protocol != EXPECTED_PROTOCOL {
		t.Fatalf("[config.Server.Protocol] should be [%s], but value is [%s]", EXPECTED_PROTOCOL, config.Server.Protocol)
	}

	if config.Database.Path != EXPECTED_DATABASE_PATH {
		t.Fatalf("[config.Database.Path] should be [%s], but value is [%s]", EXPECTED_DATABASE_PATH, config.Database.Path)
	}
}

func TestNewConfigWithDefaultPath(t *testing.T) {
	config, _ := netfs.NewConfig()
	config.BufferSize = 0

	data, _ := json.Marshal(config)
	os.WriteFile(EXPECTED_PATH, data, 0644)

	var err error
	config, err = netfs.NewConfig()
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if config.BufferSize != 0 {
		t.Fatalf("[config.BufferSize] should be [%d], but error is [%d]", 0, config.BufferSize)
	}

	os.Remove(EXPECTED_PATH)
}

func TestNewConfigWithUserPathSuccess(t *testing.T) {
	config, _ := netfs.NewConfig()
	data, _ := json.Marshal(config)
	os.WriteFile(EXPECTED_USER_PATH, data, 0644)

	var err error
	config, err = netfs.NewConfig(EXPECTED_USER_PATH)
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	os.Remove(EXPECTED_USER_PATH)
}

func TestNewConfigWithUserPathNotFound(t *testing.T) {
	config, err := netfs.NewConfig(EXPECTED_USER_PATH)
	if err == nil {
		t.Fatalf("error should be not nil, but error is nil")
	}

	if config != nil {
		t.Fatalf("config should be nil, but config is not nil")
	}
}
