package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"netfs/internal/api"
	"netfs/internal/transport"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestGetIPs(t *testing.T) {
	network, _ := api.NewNetwork(api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 1 * time.Second})
	ips, err := network.GetIPs()
	if err != nil {
		t.Fatal("error should be nil")
	}

	if len(ips) == 0 {
		t.Fatal("ips should be not empty")
	}
}

func TestGetLocalIP(t *testing.T) {
	network, _ := api.NewNetwork(api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 1 * time.Second})
	ip, err := network.GetLocalIP()
	if err != nil {
		t.Fatal("error should be nil")
	}

	if ip == nil {
		t.Fatal("IP should be not nil")
	}
}

func TestGetLocalHost(t *testing.T) {
	network, _ := api.NewNetwork(api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 1 * time.Second})
	ip, _ := network.GetLocalIP()
	hostname, _ := os.Hostname()

	host, err := network.GetLocalHost()
	if err != nil {
		t.Fatal("error should be nil")
	}

	if host.IP.String() != ip.String() {
		t.Fatalf("[%s] and [%s] should be equals", host.IP.String(), ip.String())
	}

	if host.Name != hostname {
		t.Fatalf("[%s] and [%s] should be equals", host.Name, hostname)
	}
}

func TestGetHostSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local, _ := network.GetLocalHost()

	go func() {
		http.HandleFunc(api.API.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(*local)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), nil)
	}()
	time.Sleep(2 * time.Second)

	host, err := network.GetHost(local.IP)
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
	config := api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local, _ := network.GetLocalHost()

	_, err := network.GetHost(local.IP)
	if err == nil {
		t.Fatalf("error should be not nil")
	}
}

func TestGetHostResponseError(t *testing.T) {
	config := api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local, _ := network.GetLocalHost()

	go func() {
		http.HandleFunc(api.API.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), nil)
	}()
	time.Sleep(2 * time.Second)

	_, err := network.GetHost(local.IP)
	if !errors.Is(err, transport.UnexpectedAnswer) {
		t.Fatalf("error should be [transport.UnexpectedAnswer], but error is [%s]", err)
	}
}

func TestGetHostsSuccess(t *testing.T) {
	config := api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)
	local, _ := network.GetLocalHost()

	go func() {
		http.HandleFunc(api.API.ServerHost, func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(*local)

			w.Write(data)
		})
		http.ListenAndServe(":"+strconv.Itoa(int(config.Port)), nil)
	}()
	time.Sleep(2 * time.Second)

	hosts, err := network.GetHosts()
	if err != nil {
		t.Fatal("error should be nil")
	}

	if len(hosts) == 0 {
		t.Fatal("hosts should be not empty")
	}
}

func TestGetHostsNoAvailable(t *testing.T) {
	config := api.NetworkConfig{Port: 80, Protocol: transport.HTTP, Timeout: 5 * time.Second}
	network, _ := api.NewNetwork(config)

	hosts, err := network.GetHosts()
	if err != nil {
		t.Fatal("error should be nil")
	}

	if len(hosts) > 0 {
		t.Fatal("hosts should be empty")
	}
}
