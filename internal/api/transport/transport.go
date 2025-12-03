package transport

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Returns if the protocol is not supporting.
var ErrUnsupportedProtocol = errors.New("unsupported protocol")

// eturns if the response was unexpected.
var ErrUnexpectedAnswer = errors.New("unexpected answer")

type TransportPoint []string

// Request data.
type Request interface {
	IP() net.IP
	Endpoint() string
	Param(string) string
	Params() []string
	RawBody() []byte
	Body(any) (any, error)
}

// Response data.
type Response interface {
	IP() net.IP
	Endpoint() string
	RawBody() []byte
	Body(any) (any, error)
}

// Available protocols.
type TransportProtocol uint16

const (
	HTTP TransportProtocol = iota
	CALL
)

// Abstraction of the data sender.
type TransportSender interface {
	// Creates new request instance by parameters.
	NewRequest(net.IP, string, []string, []byte, any) (Request, error)
	// Sends request.
	Send(Request) (Response, error)
	// Returns protocol.
	Protocol() TransportProtocol
	// Returns port.
	Port() uint16
}

// Creates new instance of TransportSender.
func NewSender(protocol TransportProtocol, port uint16, timeout time.Duration) (TransportSender, error) {
	if protocol == HTTP {
		return &HttpTransportSender{client: &http.Client{Timeout: timeout}, port: port}, nil
	}
	return nil, ErrUnsupportedProtocol
}

// Abstraction of the data receiver.
type TransportReceiver interface {
	// Creates new request instance by parameters.
	NewRequest(net.IP, string, []string, []byte, any) (Request, error)
	// Receives request.
	Receive(string, func(Request) ([]byte, any, error))
	// Starts receiver.
	Start() error
	// Stops receiver.
	Stop() error
	// Returns protocol.
	Protocol() TransportProtocol
	// Returns port.
	Port() uint16
}

// Creates new instance of TransportReceiver.
func NewReceiver(protocol TransportProtocol, port uint16) (TransportReceiver, error) {
	if protocol == HTTP {
		mux := http.NewServeMux()
		server := &http.Server{Addr: portSeparator + strconv.Itoa(int(port)), Handler: mux}

		return &HttpTransportReceiver{server: server, mux: mux, port: port}, nil
	}
	return nil, ErrUnsupportedProtocol
}
