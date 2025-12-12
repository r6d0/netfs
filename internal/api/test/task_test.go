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

func TestRefreshSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 8, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileCopyStatus.Name, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(
				api.RemoteTask{Id: 1, Status: api.Completed, Host: local},
			)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	task := api.RemoteTask{Id: 1, Status: api.Waiting, Host: *host}
	err := task.Refresh(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}
	if task.Status != api.Completed {
		t.Fatalf("status should be equal [%d], but status is [%d]", api.Completed, task.Status)
	}
}

func TestCancelSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 8, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		mux.HandleFunc(api.Endpoints.FileCopyStop.Name, func(w http.ResponseWriter, r *http.Request) {})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, _ := network.GetHost(local.IP)
	task := api.RemoteTask{Id: 1, Status: api.Waiting, Host: *host}
	err := task.Cancel(network.Transport())
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}
}
