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

// HTTP server.
type Server struct {
	DB      *badger.DB
	Config  Config
	Context context.Context
}

// Start http requests listening.
func (serv *Server) Listen() {
	http.HandleFunc(_API.Host, serv._HostHandle)

	http.HandleFunc(_API.FileInfo.URL, serv._FileInfoHandle)
	http.HandleFunc(_API.FileCreate.URL, serv._FileCreateHandle)
	http.HandleFunc(_API.FileWrite.URL, serv._FileWriteHandle)

	http.HandleFunc(_API.FileCopyStart.URL, serv._FileCopyStartHandle)

	// Execute async tasks
	serv._ExecuteCopyTask()

	// Run HTTP server
	http.ListenAndServe(":8080", nil)
}

// Gets local IP address or error.
func GetLocalIP() (net.IP, error) {
	connection, err := net.Dial(_UDP, _UDP_HOST)
	if connection != nil {
		defer connection.Close()
		return connection.LocalAddr().(*net.UDPAddr).IP, nil
	}
	return nil, err
}

// Gets all IPs of local network or error.
func GetLocalNetworkIPs() ([]net.IP, error) {
	ips := []net.IP{}

	local, err := GetLocalIP()
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

// Gets information about host by IP.
func GetHost(ip net.IP) (*RemoteHost, error) {
	res, err := http.Get("http://" + ip.String() + ":8080/do-sync/api/host") // TODO. Fix it
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

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

const _CIDR_END = "1.0/24"
const _IP_SEPARATOR = "."
const _UDP = "udp"
const _UDP_HOST = "1.1.1.1:80"
const _PARAM_START = "?"
const _PARAM_VALUE = "="
const _PARAM_NEXT = "&"

var _LOCAL_HOST = _GetLocalHost()

// Gets information about localhost.
func _GetLocalHost() *RemoteHost {
	var ip net.IP
	var hostname string
	var err error

	if ip, err = GetLocalIP(); err == nil {
		if hostname, err = os.Hostname(); err == nil {
			return &RemoteHost{Name: hostname, IP: ip}
		}
	}

	panic(err.Error())
}

// Returns information about the current host.
func (serv *Server) _HostHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		data, err := json.Marshal(_LOCAL_HOST)
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
