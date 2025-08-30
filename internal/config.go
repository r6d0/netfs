package netfs

// Server configuration
type ServerConfig struct {
	// The server port
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

func NewConfig() (Config, error) {
	return Config{}, nil
}
