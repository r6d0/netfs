package netfs

import (
	"encoding/json"
	"io"
	"os"
)

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

// Server configuration.
type ServerConfig struct {
	// The server port.
	Port uint16 `json:"port"`
	// The server protocol. HTTP only.
	Protocol string `json:"protocol"`
}

// Database configuration.
type DatabaseConfig struct {
	Path string `json:"path"`
}

// netfs configuration.
type Config struct {
	// Server configuration.
	Server ServerConfig `json:"server"`
	// Database configuration.
	Database DatabaseConfig `json:"database"`
	// The size of the data transfer buffer.
	BufferSize uint64
	// The size of the asynchronous task pool.
	TaskCount uint64
}

// Creates a new instance of Config, returns an error if creation failed.
func NewConfig(paths ...string) (*Config, error) {
	config := Config{
		Server:     ServerConfig{Port: _DEFAULT_PORT, Protocol: _DEFAULT_PROTOCOL},
		Database:   DatabaseConfig{Path: _DEFAULT_DATABASE_PATH},
		BufferSize: _DEFAULT_BUFFER_SIZE,
		TaskCount:  _DEFAULT_TASKS_COUNT,
	}

	path := _DEFAULT_PATH
	if len(paths) > 0 {
		path = paths[0]
	}

	file, err := os.Open(path)
	if err == nil {
		defer file.Close()

		var data []byte
		if data, err = io.ReadAll(file); err == nil {
			err = json.Unmarshal(data, &config)
		}
	}

	if err == nil || path == _DEFAULT_PATH {
		return &config, nil
	}
	return nil, err
}

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

const _DEFAULT_PATH = "./netfs-config.json"
const _DEFAULT_PORT = 49153
const _DEFAULT_PROTOCOL = "http"
const _DEFAULT_DATABASE_PATH = "./data"
const _DEFAULT_BUFFER_SIZE = 1024 * 1024 // 1MB
const _DEFAULT_TASKS_COUNT = 10
