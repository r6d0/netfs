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

func TestWriteSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 7, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileInfo.Name, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.FileInfo{FileName: "test_file.txt", FilePath: "./test_file.txt", FileType: api.FILE, FileSize: 1024},
			)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileWrite.Name, func(w http.ResponseWriter, r *http.Request) {})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	file, _ := host.FileInfo(network.Transport(), "./test_file.txt")
	err := file.Write(network.Transport(), []byte("TEST"))
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestWriteResponseError(t *testing.T) {
	config := api.NetworkConfig{Port: 8, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileInfo.Name, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.FileInfo{FileName: "test_file.txt", FilePath: "./test_file.txt", FileType: api.FILE, FileSize: 1024},
			)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	file, _ := host.FileInfo(network.Transport(), "./test_file.txt")
	err := file.Write(network.Transport(), []byte("TEST"))
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestCreateSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 9, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileInfo.Name, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.FileInfo{FileName: "test_file.txt", FilePath: "./test_file.txt", FileType: api.FILE, FileSize: 1024},
			)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileCreate, func(w http.ResponseWriter, r *http.Request) {})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	file, _ := host.FileInfo(network.Transport(), "./test_file.txt")
	err := file.Create(network.Transport())
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestCreateResponseError(t *testing.T) {
	config := api.NetworkConfig{Port: 10, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileInfo.Name, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.FileInfo{FileName: "test_file.txt", FilePath: "./test_file.txt", FileType: api.FILE, FileSize: 1024},
			)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	file, _ := host.FileInfo(network.Transport(), "./test_file.txt")
	err := file.Create(network.Transport())
	if err == nil {
		t.Fatal("error should be not nil")
	}
}

func TestCopyToSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 11, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileInfo.Name, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.FileInfo{FileName: "test_file.txt", FilePath: "./test_file.txt", FileType: api.FILE, FileSize: 1024},
			)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileCopyStart, func(w http.ResponseWriter, r *http.Request) {})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	file, _ := host.FileInfo(network.Transport(), "./test_file.txt")
	err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{FilePath: "./test_file_1.txt"}})
	if err != nil {
		t.Fatal("error should be nil")
	}
}

func TestCopyToResponseError(t *testing.T) {
	config := api.NetworkConfig{Port: 12, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileInfo.Name, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.FileInfo{FileName: "test_file.txt", FilePath: "./test_file.txt", FileType: api.FILE, FileSize: 1024},
			)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	file, _ := host.FileInfo(network.Transport(), "./test_file.txt")
	err := file.CopyTo(network.Transport(), api.RemoteFile{Info: api.FileInfo{FilePath: "./test_file_1.txt"}})
	if err == nil {
		t.Fatal("error should be not nil")
	}
}
