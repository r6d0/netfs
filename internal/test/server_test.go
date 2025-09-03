package netfs_test

import (
	netfs "netfs/internal"
	"os"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	config, _ := netfs.NewConfig()
	server, err := netfs.NewServer(config)
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if server == nil {
		t.Fatalf("server should be nil, but server is not nil")
	}
}

func TestStart(t *testing.T) {
	config, _ := netfs.NewConfig()
	server, _ := netfs.NewServer(config)

	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	server.Start()

	defer os.RemoveAll(config.Database.Path)
}
