package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const ContentType = "Content-Type"
const httpProtocol = "http"
const protocolSeparator = "://"
const portSeparator = ":"

// Transport layer via HTTP.
type HttpTransport struct {
	client *http.Client
	port   uint16
}

// Sends request.
func (tr *HttpTransport) Send(ip net.IP, point TransportPoint) error {
	_, err := tr.SendAndReceive(ip, point, nil)
	return err
}

// Sends request with body.
func (tr *HttpTransport) SendBody(ip net.IP, point TransportPoint, body any) error {
	_, err := tr.SendBodyAndReceive(ip, point, body, nil)
	return err
}

// Sends request with raw body.
func (tr *HttpTransport) SendRawBody(ip net.IP, point TransportPoint, body []byte) error {
	_, err := tr.SendRawBodyAndReceive(ip, point, body, nil)
	return err
}

// Sends request and receives response.
func (tr *HttpTransport) SendAndReceive(ip net.IP, point TransportPoint, result any) (any, error) {
	return tr.SendRawBodyAndReceive(ip, point, nil, result)
}

// Sends request with body and receives response.
func (tr *HttpTransport) SendBodyAndReceive(ip net.IP, point TransportPoint, body any, result any) (any, error) {
	data, err := json.Marshal(body)
	if err == nil {
		return tr.SendRawBodyAndReceive(ip, point, data, result)
	}
	return nil, err
}

// Sends request with raw body and receives response.
func (tr *HttpTransport) SendRawBodyAndReceive(ip net.IP, point TransportPoint, body []byte, result any) (any, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(http.MethodPost, tr.buildURL(ip, point), reader)
	if err == nil {
		var res *http.Response
		if res, err = tr.client.Do(req); err == nil {
			defer res.Body.Close()

			if res.StatusCode == http.StatusOK {
				if result != nil {
					var data []byte
					if data, err = io.ReadAll(res.Body); err == nil {
						return result, json.Unmarshal(data, result)
					}
				}
			} else {
				err = errors.Join(ErrUnexpectedAnswer, fmt.Errorf("status code is [%d]", res.StatusCode))
			}
		}
	}
	return nil, err
}

// Returns protocol.
func (tr *HttpTransport) Protocol() TransportProtocol {
	return HTTP
}

// Returns port.
func (tr *HttpTransport) Port() uint16 {
	return tr.port
}

// Returns URL.
func (tr HttpTransport) buildURL(ip net.IP, point TransportPoint) string { // TODO. Parse point
	buffer := strings.Builder{}
	buffer.WriteString(httpProtocol)
	buffer.WriteString(protocolSeparator)
	buffer.WriteString(ip.String())
	buffer.WriteString(portSeparator)
	buffer.WriteString(strconv.Itoa(int(tr.port)))
	buffer.WriteString(point[0])
	return buffer.String()
}
