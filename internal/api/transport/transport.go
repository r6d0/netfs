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

// Available protocols.
type TransportProtocol uint16

const (
	HTTP TransportProtocol = iota
	CALL
)

// Abstraction of the data sender.
type TransportSender interface {
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

// Creates new instance of TransportSender.
func NewSender(protocol TransportProtocol, port uint16, timeout time.Duration) (TransportSender, error) {
	if protocol == HTTP {
		return &HttpTransportSender{client: &http.Client{Timeout: timeout}, port: port}, nil
	}
	return nil, ErrUnsupportedProtocol
}

// Abstraction of the data receiver.
type TransportReceiver interface {
	// Starts receiver.
	Start() error
	// Stops receiver.
	Stop() error
	// Receives request.
	Receive(TransportPoint, func() error)
	// Receives request with body.
	ReceiveBody(TransportPoint, func() any, func(any) error)
	// Receives request with raw body.
	ReceiveRawBody(TransportPoint, func([]byte) error)
	// Receives request with body and sends response.
	ReceiveBodyAndSend(TransportPoint, func() any, func(any) (any, error))
	// Receives request with raw body and sends response.
	ReceiveRawBodyAndSend(TransportPoint, func([]byte) (any, error))
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
