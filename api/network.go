package api

import (
	"errors"
	"net"
	"net/netip"
	"netfs/api/transport"
	"os"
	"strings"
	"time"
)

var ErrLocalIPNotFound = errors.New("local IP address not found")
var RFC1918 = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

const cidrEnd = "1.0/24"
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
	client transport.TransportSender
}

// Get information about available hosts.
func (network *Network) Hosts() ([]RemoteHost, error) {
	var hosts []RemoteHost

	ips, err := network.IPs()
	if err == nil {
		callback := make(chan *RemoteHost)
		for _, ip := range ips {
			go func(ip net.IP, callback chan *RemoteHost) {
				host, _ := network.Host(ip)
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
func (network *Network) Host(ip net.IP) (*RemoteHost, error) {
	req, err := network.client.NewRequest(ip, Endpoints.ServerHost, nil, nil, nil)
	if err == nil {
		var res transport.Response
		if res, err = network.client.Send(req); err == nil {
			host := &RemoteHost{}
			if _, err = res.Body(host); err == nil {
				return host, nil
			}
		}
	}
	return nil, err
}

// Gets all IPs of local network or error.
func (network *Network) IPs() ([]net.IP, error) {
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
func (network *Network) Transport() transport.TransportSender {
	return network.client
}

// Creates a new instance of Network, returns an error if creation failed.
func NewNetwork(config NetworkConfig) (*Network, error) {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		var ips []net.IP
		for _, a := range addrs {
			if val, ok := a.(*net.IPAddr); ok {
				ips = append(ips, val.IP)
			} else if val, ok := a.(*net.IPNet); ok {
				ips = append(ips, val.IP)
			}
		}

		var localIP net.IP
		for _, cidr := range RFC1918 {
			_, block, _ := net.ParseCIDR(cidr)
			for _, ip := range ips {
				if block.Contains(ip) {
					localIP = ip
					break
				}
			}
		}

		if localIP != nil {
			var hostname string
			if hostname, err = os.Hostname(); err == nil {
				var client transport.TransportSender
				if client, err = transport.NewSender(config.Protocol, config.Port, config.Timeout); err == nil {
					return &Network{config: config, client: client, host: RemoteHost{Name: hostname, IP: localIP}}, nil
				}
			}
		} else {
			err = ErrLocalIPNotFound
		}
	}
	return nil, err
}
