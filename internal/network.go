package netfs

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
)

// Network operations.
type Network struct {
	_Config *Config
	_Client *http.Client
}

// Get information about available hosts.
func (network *Network) GetHosts() ([]RemoteHost, error) {
	var hosts []RemoteHost

	ips, err := _GetLocalNetworkIPs()
	if err == nil {
		callback := make(chan *RemoteHost)
		for _, ip := range ips {
			go func(ip net.IP, callback chan *RemoteHost) {
				host, _ := network.GetHost(ip)
				callback <- host
			}(ip, callback)
		}

		for range ips {
			if host := <-callback; host != nil {
				hosts = append(hosts, *host)
			}
		}
	}
	return hosts, err
}

// Gets information about host by IP.
func (network *Network) GetHost(ip net.IP) (*RemoteHost, error) {
	url := _GetURL(network._Config.Server.Protocol, ip, int(network._Config.Server.Port), _API.Host)
	res, err := network._Client.Get(url)
	if err == nil {
		defer res.Body.Close()

		var data []byte
		if data, err = io.ReadAll(res.Body); err == nil {
			var host *RemoteHost = &RemoteHost{_Client: network._Client}

			if err = json.Unmarshal(data, host); err == nil {
				return host, nil
			}
		}
	}
	return nil, err
}

// Creates a new instance of Network, returns an error if creation failed.
func NewNetwork(config *Config) (*Network, error) {
	return &Network{_Config: config, _Client: &http.Client{Timeout: config.Client.Timeout}}, nil
}
