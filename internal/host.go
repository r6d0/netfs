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

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

// Information about host.
type RemoteHost struct {
	Name    string
	IP      net.IP
	URL     string
	_Client *http.Client
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
	res, err := host._Client.Get(host.GetURL(_API.FileInfo.URL, _API.FileInfo.Path, path))
	if err == nil {
		defer res.Body.Close()

		var data []byte
		if data, err = io.ReadAll(res.Body); err == nil {
			var file *RemoteFile = &RemoteFile{_Client: host._Client}

			if err = json.Unmarshal(data, file); err == nil {
				return file, nil
			}
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
