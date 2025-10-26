package api

import (
	"net"
	"net/netip"
	"netfs/internal/transport"
	"os"
	"strings"
	"time"
)

const cidrEnd = "1.0/24"
const udpProtocol = "udp"
const udpHost = "1.1.1.1:80"
const ipSeparator = "."

// Network configuration.
type NetworkConfig struct {
	Port     uint16
	Protocol transport.TransportProtocol
	Timeout  time.Duration
}

// Network operations.
type Network struct {
	host   RemoteHost
	config NetworkConfig
	client transport.Transport
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
	res, err := network.client.SendAndReceive(ip, API.ServerHost, &RemoteHost{})
	if err == nil {
		return res.(*RemoteHost), nil
	}
	return nil, err
}

// Gets all IPs of local network or error.
func (network *Network) GetIPs() ([]net.IP, error) {
	ips := []net.IP{}

	local := network.LocalIP()
	localString := local.String()
	parts := strings.Split(localString, ipSeparator)
	cidr := strings.Join([]string{parts[0], parts[1], cidrEnd}, ipSeparator)

	prefix, err := netip.ParsePrefix(cidr)
	if err == nil {
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
	return ips, err
}

// Returns local IP.
func (network *Network) LocalIP() net.IP {
	return network.host.IP
}

// Returns local host.
func (network *Network) LocalHost() RemoteHost {
	return network.host
}

// Returns the associated transport.
func (network *Network) Transport() transport.Transport {
	return network.client
}

// Creates a new instance of Network, returns an error if creation failed.
func NewNetwork(config NetworkConfig) (*Network, error) {
	connection, err := net.Dial(udpProtocol, udpHost)
	if connection != nil {
		defer connection.Close()
		ip := connection.LocalAddr().(*net.UDPAddr).IP

		var hostname string
		if hostname, err = os.Hostname(); err == nil {
			var client transport.Transport
			if client, err = transport.NewTransport(config.Protocol, config.Port, config.Timeout); err == nil {
				return &Network{config: config, client: client, host: RemoteHost{Name: hostname, IP: ip}}, nil
			}
		}
	}
	return nil, err
}
