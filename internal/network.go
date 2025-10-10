package netfs

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

const cidrEnd = "1.0/24"
const udpProtocol = "udp"
const udpHost = "1.1.1.1:80"
const ipSeparator = "."

// Network operations.
type Network struct {
	config *Config
	client *http.Client
}

// Get information about available hosts.
func (network *Network) GetHosts() ([]RemoteHost, error) {
	var hosts []RemoteHost

	ips, err := network.GetIPs()
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
	url := getRemoteHostURL(network.config.Server.Protocol, ip, int(network.config.Server.Port), API.Host)
	res, err := network.client.Get(url)
	if err == nil {
		defer res.Body.Close()

		var data []byte
		if data, err = io.ReadAll(res.Body); err == nil {
			var host *RemoteHost = &RemoteHost{client: network.client}

			if err = json.Unmarshal(data, host); err == nil {
				return host, nil
			}
		}
	}
	return nil, err
}

// Gets local IP address or error.
func (network *Network) GetLocalIP() (net.IP, error) {
	connection, err := net.Dial(udpProtocol, udpHost)
	if connection != nil {
		defer connection.Close()
		return connection.LocalAddr().(*net.UDPAddr).IP, nil
	}
	return nil, err
}

// Gets information about localhost.
func (network *Network) GetLocalHost() (*RemoteHost, error) {
	var ip net.IP
	var hostname string
	var err error

	if ip, err = network.GetLocalIP(); err == nil {
		if hostname, err = os.Hostname(); err == nil {
			port := network.config.Server.Port
			protocol := network.config.Server.Protocol
			return &RemoteHost{Name: hostname, IP: ip, URL: getRemoteHostURL(strings.ToLower(protocol), ip, int(port), "")}, nil
		}
	}
	return nil, err
}

// Gets all IPs of local network or error.
func (network *Network) GetIPs() ([]net.IP, error) {
	ips := []net.IP{}

	local, err := network.GetLocalIP()
	if err == nil {
		localString := local.String()
		parts := strings.Split(localString, ipSeparator)
		cidr := strings.Join([]string{parts[0], parts[1], cidrEnd}, ipSeparator)

		var prefix netip.Prefix
		if prefix, err = netip.ParsePrefix(cidr); err == nil {
			prefix = prefix.Masked()
			addr := prefix.Addr()
			for prefix.Contains(addr) {
				ip := addr.String()
				if localString != "" {
					ips = append(ips, net.ParseIP(ip))
				}
				addr = addr.Next()
			}
		}
	}
	return ips, err
}

// Creates a new instance of Network, returns an error if creation failed.
func NewNetwork(config *Config) (*Network, error) {
	return &Network{config: config, client: &http.Client{Timeout: config.Client.Timeout}}, nil
}

// Gets host url in format - [protocol]://[ip]:[port]/
func getRemoteHostURL(protocol string, ip net.IP, port int, path string, params ...any) string {
	buffer := strings.Builder{}
	buffer.WriteString(strings.ToLower(protocol))
	buffer.WriteString(protocolSeparator)
	buffer.WriteString(ip.String())
	buffer.WriteString(portSeparator)
	buffer.WriteString(strconv.Itoa(port))

	if len(path) > 0 {
		buffer.WriteString(path)
	}

	if len(params) > 0 {
		buffer.WriteString(paramStart)
		for index := range params {
			buffer.WriteString(fmt.Sprint(params[index]))
			buffer.WriteString(paramValue)
			buffer.WriteString(fmt.Sprint(params[index+1]))

			if index < len(params)-1 {
				buffer.WriteString(paramNext)
			}
		}
	}
	return buffer.String()
}
