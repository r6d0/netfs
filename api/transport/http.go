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
	"net/url"
	"strconv"
	"strings"
)

const portSeparator = ":"
const paramsSeparator = "?"
const httpProtocol = "http://"

type httpRequest struct {
	ip       net.IP
	endpoint string
	params   []string
	rawBody  []byte
}

func (req *httpRequest) IP() net.IP {
	return req.ip
}

func (req *httpRequest) Endpoint() string {
	return req.endpoint
}

func (req *httpRequest) Param(name string) string {
	for index, param := range req.params {
		if param == name && index < len(req.params)-1 {
			return req.params[index+1]
		}
	}
	return ""
}

func (req *httpRequest) Params() []string {
	return req.params
}

func (req *httpRequest) RawBody() []byte {
	return req.rawBody
}

func (req *httpRequest) Body(target any) (any, error) {
	return target, json.Unmarshal(req.rawBody, target)
}

type httpResponse struct {
	ip       net.IP
	endpoint string
	rawBody  []byte
}

func (res *httpResponse) IP() net.IP {
	return res.ip
}

func (res *httpResponse) Endpoint() string {
	return res.endpoint
}

func (res *httpResponse) RawBody() []byte {
	return res.rawBody
}

func (res *httpResponse) Body(target any) (any, error) {
	return target, json.Unmarshal(res.rawBody, target)
}

// Sending data via the HTTP protocol.
type HttpTransportSender struct {
	client *http.Client
	port   uint16
}

// Creates new request instance by parameters.
func (tr *HttpTransportSender) NewRequest(ip net.IP, endpoint string, parameters []string, rawBody []byte, body any) (Request, error) {
	var err error
	req := &httpRequest{ip: ip, endpoint: endpoint, params: parameters, rawBody: rawBody}
	if body != nil {
		req.rawBody, err = json.Marshal(body)
	}
	return req, err
}

// Sends request.
func (tr *HttpTransportSender) Send(req Request) (Response, error) {
	var reader io.Reader
	if body := req.RawBody(); body != nil {
		reader = bytes.NewReader(body)
	}

	endpoint, err := url.JoinPath(httpProtocol, req.IP().String())
	if err == nil {
		endpoint = strings.Join([]string{endpoint, strconv.Itoa(int(tr.Port()))}, portSeparator)
		endpoint, err = url.JoinPath(endpoint, req.Endpoint())
		if params := req.Params(); err == nil && len(params) > 0 {
			urlParams := url.Values{}
			for index := range params {
				if index%2 == 0 {
					urlParams.Add(params[index], params[index+1])
				}
			}
			endpoint = strings.Join([]string{endpoint, urlParams.Encode()}, paramsSeparator)
		}
	}

	if err == nil {
		var httpReq *http.Request
		if httpReq, err = http.NewRequest(http.MethodPost, endpoint, reader); err == nil {
			var httpRes *http.Response
			if httpRes, err = tr.client.Do(httpReq); err == nil {
				defer httpRes.Body.Close()

				message, _ := io.ReadAll(httpRes.Body)
				if httpRes.StatusCode == http.StatusOK {
					return &httpResponse{ip: req.IP(), endpoint: req.Endpoint(), rawBody: message}, nil
				} else {
					if len(message) > 0 {
						err = errors.Join(ErrUnexpectedAnswer, fmt.Errorf("status code is [%d], message is [%s]", httpRes.StatusCode, string(message)))
					} else {
						err = errors.Join(ErrUnexpectedAnswer, fmt.Errorf("status code is [%d]", httpRes.StatusCode))
					}
				}
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

// Receiving data via the HTTP protocol.
type HttpTransportReceiver struct {
	port   uint16
	mux    *http.ServeMux
	server *http.Server
}

// Creates new request instance by parameters.
func (tr *HttpTransportReceiver) NewRequest(ip net.IP, endpoint string, parameters []string, rawBody []byte, body any) (Request, error) {
	var err error
	req := &httpRequest{ip: ip, endpoint: endpoint, params: parameters, rawBody: rawBody}
	if body != nil {
		req.rawBody, err = json.Marshal(body)
	}
	return req, err
}

// Receives request.
func (tr *HttpTransportReceiver) Receive(endpoint string, handle func(Request) ([]byte, any, error)) {
	tr.mux.HandleFunc(endpoint, func(httpRes http.ResponseWriter, httpReq *http.Request) {
		defer httpReq.Body.Close()

		var rawResBody []byte
		body, err := io.ReadAll(httpReq.Body)
		if err == nil {
			ip := net.ParseIP(httpReq.RemoteAddr)

			query := httpReq.URL.Query()
			parameters := []string{}
			for key := range query {
				parameters = append(parameters, key)
				parameters = append(parameters, query.Get(key))
			}

			var req Request
			if req, err = tr.NewRequest(ip, endpoint, parameters, body, nil); err == nil {
				var resBody any
				if rawResBody, resBody, err = handle(req); err == nil {
					if resBody != nil {
						rawResBody, err = json.Marshal(resBody)
					}
				}
			}
		}

		if err != nil {
			httpRes.WriteHeader(http.StatusInternalServerError)
			httpRes.Write([]byte(err.Error()))
		} else {
			httpRes.Write(rawResBody)
		}
	})
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

// Returns protocol.
func (tr *HttpTransportReceiver) Protocol() TransportProtocol {
	return HTTP
}

// Returns port.
func (tr *HttpTransportReceiver) Port() uint16 {
	return tr.port
}
