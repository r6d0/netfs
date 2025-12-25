package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"netfs/api"
	"netfs/api/transport"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestGetIPs(t *testing.T) {
	network, _ := api.NewNetwork(api.NetworkConfig{Port: 1, Protocol: transport.HTTP, Timeout: 1 * time.Second})
	ips, err := network.IPs()
	if err != nil {
		t.Fatal("error should be nil")
	}

	if len(ips) == 0 {
		t.Fatal("ips should be not empty")
	}
}

func TestLocalIP(t *testing.T) {
	network, _ := api.NewNetwork(api.NetworkConfig{Port: 1, Protocol: transport.HTTP, Timeout: 1 * time.Second})
	ip := network.LocalIP()
	if ip == nil {
		t.Fatal("IP should be not nil")
	}
}

func TestGetLocalHost(t *testing.T) {
	network, _ := api.NewNetwork(api.NetworkConfig{Port: 1, Protocol: transport.HTTP, Timeout: 1 * time.Second})
	ip := network.LocalIP()
	hostname, _ := os.Hostname()

	host := network.LocalHost()
	if host.IP.String() != ip.String() {
		t.Fatalf("[%s] and [%s] should be equals", host.IP.String(), ip.String())
	}

	if host.Name != hostname {
		t.Fatalf("[%s] and [%s] should be equals", host.Name, hostname)
	}
}

func TestGetHostSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 2, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	host, err := network.Host(local.IP)
	if err != nil {
		t.Fatal("error should be nil")
	}

	if host.IP.String() != local.IP.String() {
		t.Fatalf("[%s] and [%s] should be equals", host.IP.String(), local.IP.String())
	}

	if host.Name != local.Name {
		t.Fatalf("[%s] and [%s] should be equals", host.Name, local.Name)
	}
}

func TestGetHostUnavailable(t *testing.T) {
	config := api.NetworkConfig{Port: 1, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	_, err := network.Host(local.IP)
	if err == nil {
		t.Fatalf("error should be not nil")
	}
}

func TestGetHostResponseError(t *testing.T) {
	config := api.NetworkConfig{Port: 3, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	_, err := network.Host(local.IP)
	if !errors.Is(err, transport.ErrUnexpectedAnswer) {
		t.Fatalf("error should be [transport.UnexpectedAnswer], but error is [%s]", err)
	}
}

func TestGetHostsSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 4, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local := network.LocalHost()

	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc(api.Endpoints.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(local)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), mux)
	}()
	time.Sleep(2 * time.Second)

	hosts, err := network.Hosts()
	if err != nil {
		t.Fatal("error should be nil")
	}

	if len(hosts) == 0 {
		t.Fatal("hosts should be not empty")
	}
}
