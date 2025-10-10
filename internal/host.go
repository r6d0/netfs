package netfs

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

const paramStart = "?"
const paramValue = "="
const paramNext = "&"
const protocolSeparator = "://"
const portSeparator = ":"

// Information about host.
type RemoteHost struct {
	Name   string
	IP     net.IP
	URL    string
	client *http.Client
}

// Gets host url in format - [protocol]://[ip]:[port]/
func (host *RemoteHost) GetURL(path string, params ...any) string {
	builder := strings.Builder{}
	builder.WriteString(host.URL)
	builder.WriteString(path)

	if len(params) > 0 {
		builder.WriteString(paramStart)

		index := 0
		for index < len(params)-1 {
			builder.WriteString(fmt.Sprint(params[index]))
			builder.WriteString(paramValue)
			builder.WriteString(fmt.Sprint(params[index+1]))

			index++
			if index < len(params)-1 {
				builder.WriteString(paramNext)
			}
		}
	}
	return builder.String()
}

// Gets information about file by path.
func (host *RemoteHost) GetFileInfo(path string) (*RemoteFile, error) {
	res, err := host.client.Get(host.GetURL(API.FileInfo.URL, API.FileInfo.Path, path))
	if err == nil {
		defer res.Body.Close()

		var data []byte
		if data, err = io.ReadAll(res.Body); err == nil {
			var file *RemoteFile = &RemoteFile{client: host.client}

			if err = json.Unmarshal(data, file); err == nil {
				return file, nil
			}
		}
	}
	return nil, err
}
