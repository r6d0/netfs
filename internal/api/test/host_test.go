package api_test

import (
	"encoding/json"
	"net/http"
	"netfs/internal/api"
	"netfs/internal/api/transport"
	"strconv"
	"testing"
	"time"
)

func TestOpenFileSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 5, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.API.ServerHost()[0], func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.API.FileInfo()[0], func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.RemoteFile{Name: "test_file.txt", Path: "./test_file.txt", FileType: api.FILE, Size: 1024, Host: local},
			)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	file, err := host.OpenFile(network.Transport(), api.RemoteFile{Path: "./test_file.txt"})
	if err != nil {
		t.Fatal("error should be nil")
	}

	if file.FileType != api.FILE {
		t.Fatal("element should be a file")
	}

	if file.Name != "test_file.txt" {
		t.Fatal("file name should be [test_file.txt]")
	}

	if file.Path != "./test_file.txt" {
		t.Fatal("file path should be [./test_file.txt]")
	}
}

func TestOpenFileResponseError(t *testing.T) {
	config := api.NetworkConfig{Port: 6, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.API.ServerHost()[0], func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	_, err := host.OpenFile(network.Transport(), api.RemoteFile{Path: "./test_file.txt"})
	if err == nil {
		t.Fatal("error should be not nil")
	}
}
