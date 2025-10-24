package transport

import (
	"errors"
	"net"
	"net/http"
	"time"
)

// Returns if the protocol is not supporting.
var UnsupportedProtocol = errors.New("unsupported protocol")

// eturns if the response was unexpected.
var UnexpectedAnswer = errors.New("unexpected answer")

// Available protocols.
type TransportProtocol uint16

const (
	HTTP TransportProtocol = iota
)

// Abstraction of the transport layer.
type Transport interface {
	// Sends request.
	Send(net.IP, string) error
	// Sends request with body.
	SendBody(net.IP, string, any) error
	// Sends request with raw body.
	SendRawBody(net.IP, string, []byte) error
	// Sends request and receives response.
	SendAndReceive(net.IP, string, any) (any, error)
	// Sends request with body and receives response.
	SendBodyAndReceive(net.IP, string, any, any) (any, error)
	// Sends request with raw body and receives response.
	SendRawBodyAndReceive(net.IP, string, []byte, any) (any, error)
	// Returns protocol.
	Protocol() TransportProtocol
	// Returns port.
	Port() uint16
}

// Creates new instance of Transport.
func NewTransport(protocol TransportProtocol, port uint16, timeout time.Duration) (Transport, error) {
	if protocol == HTTP {
		return &HttpTransport{client: &http.Client{Timeout: timeout}, port: port}, nil
	}
	return nil, UnsupportedProtocol
}
