package netfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

// Information about host.
type RemoteHost struct {
	Name string
	IP   net.IP
	URL  string
}

// Gets host url in format - [protocol]://[ip]:[port]/
func (host *RemoteHost) GetURL(path string, params ...any) string {
	builder := strings.Builder{}
	builder.WriteString(host.URL)
	builder.WriteString(path)

	if len(params) > 0 {
		builder.WriteString(_PARAM_START)
		for index := range params {
			builder.WriteString(fmt.Sprint(params[index]))
			builder.WriteString(_PARAM_VALUE)
			builder.WriteString(fmt.Sprint(params[index+1]))

			if index < len(params)-1 {
				builder.WriteString(_PARAM_NEXT)
			}
		}
	}
	return builder.String()
}

// Gets information about file by path.
func (host *RemoteHost) GetFileInfo(path string) (*RemoteFile, error) {
	res, err := http.Get(host.GetURL(_API.FileInfo.URL, _API.FileInfo.Path, path))
	if err == nil {
		defer res.Body.Close()

		var data []byte
		if data, err = io.ReadAll(res.Body); err == nil {
			var file *RemoteFile = &RemoteFile{}

			if err = json.Unmarshal(data, file); err == nil {
				return file, nil
			}
		}
	}
	return nil, err
}

type Network struct {
	_Config Config
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
	res, err := http.Get(url)
	if err == nil {
		defer res.Body.Close()

		var data []byte
		if data, err = io.ReadAll(res.Body); err == nil {
			var host *RemoteHost = &RemoteHost{}

			if err = json.Unmarshal(data, host); err == nil {
				return host, nil
			}
		}
	}
	return nil, err
}

func NewNetwork(config Config) (*Network, error) {
	return nil, nil
}

// HTTP server.
type Server struct {
	_DB      *badger.DB
	_Config  Config
	_Context context.Context
	_Host    RemoteHost
}

// Start requests listening. It's blocking current thread.
func (serv *Server) Listen() {
	http.HandleFunc(_API.Host, serv._HostHandle)

	http.HandleFunc(_API.FileInfo.URL, serv._FileInfoHandle)
	http.HandleFunc(_API.FileCreate.URL, serv._FileCreateHandle)
	http.HandleFunc(_API.FileWrite.URL, serv._FileWriteHandle)

	http.HandleFunc(_API.FileCopyStart.URL, serv._FileCopyStartHandle)

	// Execute async tasks
	serv._ExecuteCopyTask()

	// Run HTTP server
	http.ListenAndServe(_PORT_SEPARATOR+strconv.Itoa(int(serv._Config.Server.Port)), nil)
}

// New instance of the netfs server.
func New(config Config) (*Server, error) {
	// Database connection
	db, err := badger.Open(badger.DefaultOptions(config.Database.Path))

	// Information about server host
	if err == nil {
		var host *RemoteHost
		if host, err = _GetLocalHost(config.Server.Protocol, int(config.Server.Port)); err == nil {
			return &Server{_Config: config, _DB: db, _Host: *host, _Context: context.Background()}, nil
		}
	}
	return nil, err
}

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

const _CIDR_END = "1.0/24"
const _IP_SEPARATOR = "."
const _UDP = "udp"
const _UDP_HOST = "1.1.1.1:80"
const _PARAM_START = "?"
const _PARAM_VALUE = "="
const _PARAM_NEXT = "&"
const _PROTOCOL_SEPARATOR = "://"
const _PORT_SEPARATOR = ":"

// Server API.
var _API = struct {
	// Information about file.
	FileInfo struct {
		URL    string
		Method string
		Path   string
	}
	// Information about host.
	Host string
	// Create directory.
	FileCreate struct {
		URL         string
		Method      string
		ContentType string
	}
	// Write data to file.
	FileWrite struct {
		URL         string
		Method      string
		ContentType string
		Path        string
	}
	// Starting a file or directory copy operation.
	FileCopyStart struct {
		URL         string
		Method      string
		ContentType string
	}
}{
	Host: "/do-sync/api/host",
	FileInfo: struct {
		URL    string
		Method string
		Path   string
	}{URL: "/do-sync/api/file/info", Method: http.MethodGet, Path: "path"},
	FileCreate: struct {
		URL         string
		Method      string
		ContentType string
	}{URL: "/do-sync/api/file/create", Method: http.MethodPost, ContentType: "application/octet-stream"},
	FileWrite: struct {
		URL         string
		Method      string
		ContentType string
		Path        string
	}{URL: "/do-sync/api/file/write", Method: http.MethodPost, Path: "path", ContentType: "application/octet-stream"},
	FileCopyStart: struct {
		URL         string
		Method      string
		ContentType string
	}{URL: "/do-sync/api/file/copy/start", Method: http.MethodPost, ContentType: "application/octet-stream"},
}

// Returns information about the current host.
func (serv *Server) _HostHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		data, err := json.Marshal(serv._Host)
		if err == nil {
			_, err = writer.Write(data)
		}

		if err != nil {
			fmt.Println(err)

			writer.Write([]byte(err.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Gets local IP address or error.
func _GetLocalIP() (net.IP, error) {
	connection, err := net.Dial(_UDP, _UDP_HOST)
	if connection != nil {
		defer connection.Close()
		return connection.LocalAddr().(*net.UDPAddr).IP, nil
	}
	return nil, err
}

// Gets information about localhost.
func _GetLocalHost(protocol string, port int) (*RemoteHost, error) {
	var ip net.IP
	var hostname string
	var err error

	if ip, err = _GetLocalIP(); err == nil {
		if hostname, err = os.Hostname(); err == nil {
			return &RemoteHost{Name: hostname, IP: ip, URL: _GetURL(strings.ToLower(protocol), ip, port, "")}, nil
		}
	}
	return nil, err
}

// Gets all IPs of local network or error.
func _GetLocalNetworkIPs() ([]net.IP, error) {
	ips := []net.IP{}

	local, err := _GetLocalIP()
	if err == nil {
		localString := local.String()
		parts := strings.Split(localString, _IP_SEPARATOR)
		cidr := strings.Join([]string{parts[0], parts[1], _CIDR_END}, _IP_SEPARATOR)

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

// Gets host url in format - [protocol]://[ip]:[port]/
func _GetURL(protocol string, ip net.IP, port int, path string, params ...any) string {
	buffer := strings.Builder{}
	buffer.WriteString(strings.ToLower(protocol))
	buffer.WriteString(_PROTOCOL_SEPARATOR)
	buffer.WriteString(ip.String())
	buffer.WriteString(_PORT_SEPARATOR)
	buffer.WriteString(strconv.Itoa(port))

	if len(path) > 0 {
		buffer.WriteString(path)
	}

	if len(params) > 0 {
		buffer.WriteString(_PARAM_START)
		for index := range params {
			buffer.WriteString(fmt.Sprint(params[index]))
			buffer.WriteString(_PARAM_VALUE)
			buffer.WriteString(fmt.Sprint(params[index+1]))

			if index < len(params)-1 {
				buffer.WriteString(_PARAM_NEXT)
			}
		}
	}
	return buffer.String()
}
