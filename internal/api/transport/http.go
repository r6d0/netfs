package transport

import (
	"bytes"
	"context"
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

// Sending data via the HTTP protocol.
type HttpTransportSender struct {
	client *http.Client
	port   uint16
}

// Sends request.
func (tr *HttpTransportSender) Send(ip net.IP, point TransportPoint) error {
	_, err := tr.SendAndReceive(ip, point, nil)
	return err
}

// Sends request with body.
func (tr *HttpTransportSender) SendBody(ip net.IP, point TransportPoint, body any) error {
	_, err := tr.SendBodyAndReceive(ip, point, body, nil)
	return err
}

// Sends request with raw body.
func (tr *HttpTransportSender) SendRawBody(ip net.IP, point TransportPoint, body []byte) error {
	_, err := tr.SendRawBodyAndReceive(ip, point, body, nil)
	return err
}

// Sends request and receives response.
func (tr *HttpTransportSender) SendAndReceive(ip net.IP, point TransportPoint, result any) (any, error) {
	return tr.SendRawBodyAndReceive(ip, point, nil, result)
}

// Sends request with body and receives response.
func (tr *HttpTransportSender) SendBodyAndReceive(ip net.IP, point TransportPoint, body any, result any) (any, error) {
	data, err := json.Marshal(body)
	if err == nil {
		return tr.SendRawBodyAndReceive(ip, point, data, result)
	}
	return nil, err
}

// Sends request with raw body and receives response.
func (tr *HttpTransportSender) SendRawBodyAndReceive(ip net.IP, point TransportPoint, body []byte, result any) (any, error) {
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
func (tr *HttpTransportSender) Protocol() TransportProtocol {
	return HTTP
}

// Returns port.
func (tr *HttpTransportSender) Port() uint16 {
	return tr.port
}

// Returns URL.
func (tr HttpTransportSender) buildURL(ip net.IP, point TransportPoint) string { // TODO. Parse point
	buffer := strings.Builder{}
	buffer.WriteString(httpProtocol)
	buffer.WriteString(protocolSeparator)
	buffer.WriteString(ip.String())
	buffer.WriteString(portSeparator)
	buffer.WriteString(strconv.Itoa(int(tr.port)))
	buffer.WriteString(point[0])
	return buffer.String()
}

// Receiving data via the HTTP protocol.
type HttpTransportReceiver struct {
	port   uint16
	mux    *http.ServeMux
	server *http.Server
}

// Starts receiver.
func (tr *HttpTransportReceiver) Start() error {
	go func() { tr.server.ListenAndServe() }()
	return nil
}

// Stops receiver.
func (tr *HttpTransportReceiver) Stop() error {
	return tr.server.Shutdown(context.Background())
}

// Receives request.
func (tr *HttpTransportReceiver) Receive(point TransportPoint, handle func() error) {
	tr.ReceiveRawBody(point, func([]byte) error {
		return handle()
	})
}

// Receives request with body.
func (tr *HttpTransportReceiver) ReceiveBody(point TransportPoint, construct func() any, handle func(any) error) {
	tr.ReceiveRawBody(point, func(data []byte) error {
		body := construct()
		err := json.Unmarshal(data, body)
		if err == nil {
			err = handle(body)
		}
		return err
	})
}

// Receives request with raw body.
func (tr *HttpTransportReceiver) ReceiveRawBody(point TransportPoint, handle func([]byte) error) {
	tr.ReceiveRawBodyAndSend(point, func(data []byte) (any, error) {
		return nil, handle(data)
	})
}

// Receives request with body and sends response.
func (tr *HttpTransportReceiver) ReceiveBodyAndSend(point TransportPoint, construct func() any, handle func(any) (any, error)) {
	tr.ReceiveRawBodyAndSend(point, func(data []byte) (any, error) {
		var res any

		body := construct()
		err := json.Unmarshal(data, body)
		if err == nil {
			res, err = handle(body)
		}
		return res, err
	})
}

// Receives request with raw body and sends response.
func (tr *HttpTransportReceiver) ReceiveRawBodyAndSend(point TransportPoint, handle func([]byte) (any, error)) {
	tr.mux.HandleFunc(point[0], func(wrt http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var res any
		body, err := io.ReadAll(req.Body)
		if err == nil {
			if res, err = handle(body); res != nil {
				if body, err = json.Marshal(res); err == nil {
					wrt.Write(body)
				}
			}
		}

		if err != nil {
			wrt.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// Returns protocol.
func (tr *HttpTransportReceiver) Protocol() TransportProtocol {
	return HTTP
}

// Returns port.
func (tr *HttpTransportReceiver) Port() uint16 {
	return tr.port
}
