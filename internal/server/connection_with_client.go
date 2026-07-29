package server

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/anteiro255/gedis/internal/action"
	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type ConnectionWithClient struct {
	conn   net.Conn
	server *Server
}

// This function is made for keep alive TCP connections
// The functions first tries to read the first byte of a request(hangs), when the request arrives
// the function reads the first byte, set a deadline and read all the other bytes of the request
func (conn *ConnectionWithClient) read() (*protocol.Request, error) {
	receiveTimeout := conn.server.config.ReceiveTimeout()

	// Read the first byte
	var firstByte [1]byte
	_, err := io.ReadFull(conn.conn, firstByte[:])
	if err != nil {
		return nil, err
	}

	// Set a deadline after the reading the first byte
	if receiveTimeout != 0 { // 0 means no timeout
		if err := conn.conn.SetReadDeadline(time.Now().Add(receiveTimeout)); err != nil {
			return nil, err
		}
	}

	var headerAsBytes [protocol.RequestHeaderSize]byte
	headerAsBytes[0] = firstByte[0]

	// Read all the other bytes after the first
	_, err = io.ReadFull(conn.conn, headerAsBytes[1:])
	if err != nil {
		return nil, err
	}

	var req protocol.Request
	req.Header = protocol.NewRequestHeaderFromBytes(headerAsBytes)

	// read the body
	req.Body = make([]byte, req.Header.BodySize)
	if req.Header.BodySize > 0 {
		_, err = io.ReadFull(conn.conn, req.Body)
		if err != nil {
			return nil, err
		}
	}

	return &req, nil
}

func (conn *ConnectionWithClient) writeResponse(resp *protocol.Response) error {
	sendTimeout := conn.server.config.SendTimeout()

	respBytes := resp.ToBytes()

	// set a deadline before the reading
	if sendTimeout != 0 {
		if err := conn.conn.SetWriteDeadline(time.Now().Add(sendTimeout)); err != nil {
			return err
		}
	}

	// write
	totalWritten := 0
	for totalWritten < len(respBytes) {
		n, err := conn.conn.Write(respBytes[totalWritten:])
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		totalWritten += n
	}
	return nil
}

func (conn *ConnectionWithClient) writeError(code status.Status) {
	addr := conn.conn.RemoteAddr().String()
	err := conn.writeResponse(protocol.NewResponse(code, nil))
	if err != nil {
		slog.Info("Failed to write to the connection", "addr", addr, "Error", err.Error())
	}
	slog.Debug(code.Error())
}

func (conn *ConnectionWithClient) Serve() {
	addr := conn.conn.RemoteAddr().String()

	slog.Info("A client was connected", "addr", addr)

	for {
		req, err := conn.read()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Debug("Connection closed by client", "addr", addr)
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				slog.Debug("Connection read timeout", "addr", addr)
				conn.writeError(status.DeadlineExceeded)
				break
			}
			slog.Debug(err.Error(), "addr", addr)
			break
		}

		action := action.Action{
			DB:         conn.server.db,
			ActionType: protocol.ActionType(req.Header.Operation),
			Key:        req.Header.Key,
			Body:       req.Body,
		}
		body, stat := action.Perform()
		if stat == status.InternalError {
			slog.Error(stat.Error())
			conn.writeError(stat)
			break
		}

		if err := conn.writeResponse(protocol.NewResponse(stat, body)); err != nil {
			slog.Debug("Failed to write response", "addr", addr, "error", err.Error())
			break
		}

		slog.Debug("The request was processed successfully", "addr", addr, "status", stat.Error(), "operation", req.Header.Operation, "key", req.Header.Key)
	}
	slog.Debug("The client was disconnected", "addr", addr)
}
