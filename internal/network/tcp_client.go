package network

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"time"
)

type TCPClient struct {
	conn net.Conn

	address        string
	idleTimeout    time.Duration
	maxMessageSize int

	log *slog.Logger
}

func NewTCPClient(
	log *slog.Logger,
	options ...TCPClientOption) (*TCPClient, error) {
	const op = "network.NewTCPClient"

	if log == nil {
		return nil, fmt.Errorf("%s: failed initilize logger", op)
	}

	client := &TCPClient{
		log: log,
	}

	for _, option := range options {
		option(client)
	}

	conn, err := net.Dial("tcp", client.address)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to server: %w", op, err)
	}

	client.conn = conn

	return client, nil
}

func (tcpClient *TCPClient) SendAndReceive(request []byte) ([]byte, error) {
	const op = "network.TCPClient.SendAndReceive"

	var err error

	if tcpClient.idleTimeout != 0 {
		err = tcpClient.conn.SetWriteDeadline(time.Now().Add(tcpClient.idleTimeout))
		if err != nil {
			tcpClient.log.Debug("failed to set write deadline", slog.String("operation", op))
		}
	}
	_, err = tcpClient.conn.Write(request)
	if err != nil {

		return nil, fmt.Errorf("%s: failed to send data to server %w", op, err)
	}

	response := make([]byte, tcpClient.maxMessageSize)

	if tcpClient.idleTimeout != 0 {
		err = tcpClient.conn.SetReadDeadline(time.Now().Add(tcpClient.idleTimeout))
		if err != nil {
			tcpClient.log.Debug("failed to set read deadline", slog.String("operation", op))
		}
	}

	num, err := tcpClient.conn.Read(response)
	if err != nil {

		return nil, fmt.Errorf("%s: failed to read data from server %w", op, err)
	}

	return bytes.TrimSpace(response[:num]), nil
}

func (tcpClient *TCPClient) Close() {
	if tcpClient.conn != nil {
		_ = tcpClient.conn.Close()
	}
}
