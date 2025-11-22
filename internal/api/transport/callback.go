package transport

import (
	"encoding/json"
	"net"
)

// The transport layer with a callback.
type CallbackTransport struct {
	Callback func(net.IP, TransportPoint, []byte, any) (any, error)
}

// Sends request.
func (tr *CallbackTransport) Send(ip net.IP, point TransportPoint) error {
	_, err := tr.SendAndReceive(ip, point, nil)
	return err
}

// Sends request with body.
func (tr *CallbackTransport) SendBody(ip net.IP, point TransportPoint, body any) error {
	_, err := tr.SendBodyAndReceive(ip, point, body, nil)
	return err
}

// Sends request with raw body.
func (tr *CallbackTransport) SendRawBody(ip net.IP, point TransportPoint, body []byte) error {
	_, err := tr.SendRawBodyAndReceive(ip, point, body, nil)
	return err
}

// Sends request and receives response.
func (tr *CallbackTransport) SendAndReceive(ip net.IP, point TransportPoint, result any) (any, error) {
	return tr.SendRawBodyAndReceive(ip, point, nil, result)
}

// Sends request with body and receives response.
func (tr *CallbackTransport) SendBodyAndReceive(ip net.IP, point TransportPoint, body any, result any) (any, error) {
	data, err := json.Marshal(body)
	if err == nil {
		return tr.SendRawBodyAndReceive(ip, point, data, result)
	}
	return nil, err
}

// Sends request with raw body and receives response.
func (tr *CallbackTransport) SendRawBodyAndReceive(ip net.IP, point TransportPoint, body []byte, result any) (any, error) {
	return tr.Callback(ip, point, body, result)
}

// Returns protocol.
func (tr *CallbackTransport) Protocol() TransportProtocol {
	return CALL
}

// Returns port.
func (tr *CallbackTransport) Port() uint16 {
	return 0
}
