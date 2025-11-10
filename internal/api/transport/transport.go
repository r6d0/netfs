package transport

import (
	"errors"
	"net"
	"net/http"
	"time"
)

// Returns if the protocol is not supporting.
var ErrUnsupportedProtocol = errors.New("unsupported protocol")

// eturns if the response was unexpected.
var ErrUnexpectedAnswer = errors.New("unexpected answer")

type TransportPoint []string

// Available protocols.
type TransportProtocol uint16

const (
	HTTP TransportProtocol = iota
	CALL
)

// Abstraction of the transport layer.
type Transport interface {
	// Sends request.
	Send(net.IP, TransportPoint) error
	// Sends request with body.
	SendBody(net.IP, TransportPoint, any) error
	// Sends request with raw body.
	SendRawBody(net.IP, TransportPoint, []byte) error
	// Sends request and receives response.
	SendAndReceive(net.IP, TransportPoint, any) (any, error)
	// Sends request with body and receives response.
	SendBodyAndReceive(net.IP, TransportPoint, any, any) (any, error)
	// Sends request with raw body and receives response.
	SendRawBodyAndReceive(net.IP, TransportPoint, []byte, any) (any, error)
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
	return nil, ErrUnsupportedProtocol
}
