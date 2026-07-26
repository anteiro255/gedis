package server

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/anteiro255/gedis/internal/action"
	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type connection struct {
	conn   net.Conn
	server Server
}

func (conn *connection) read() (*protocol.Request, error) {
	if err := conn.conn.SetReadDeadline(time.Now().Add(conn.server.config.ReceiveTimeout())); err != nil {
		return nil, err
	}

	var headerAsBytes [protocol.RequestHeaderSize]byte
	_, err := io.ReadFull(conn.conn, headerAsBytes[:])
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, err
		}
		return nil, err
	}

	var req protocol.Request
	req.Header = protocol.NewRequestHeaderFromBytes(headerAsBytes)

	body := make([]byte, req.Header.BodySize)
	if req.Header.BodySize > 0 {
		_, err = io.ReadFull(conn.conn, body)
		if err != nil {
			return nil, err
		}
	}
	req.Body = body

	return &req, nil
}

func (conn *connection) writeResponse(resp *protocol.Response) error {
	if err := conn.conn.SetWriteDeadline(time.Now().Add(conn.server.config.SendTimeout())); err != nil {
		return err
	}

	respBytes := resp.ToBytes()
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

func (conn *connection) writeError(code status.Status) {
	addr := conn.conn.RemoteAddr().String()
	err := conn.writeResponse(protocol.NewResponse(code, nil))
	if err != nil {
		slog.Info("Failed to write to the connection", "addr", addr, "Error", err.Error())
	}
	switch code {
	case status.NoSuchKey:
		slog.Info("The user submitted not existing key", "addr", addr)
	case status.WrongInput:
		slog.Info("The user submitted wrong input", "addr", addr)
	case status.InternalError:
		slog.Info("Internal error occured during the serving the user", "addr", addr)
	default:
		slog.Error("Wrong error code value was passed to the handlers.error() function, the code value: " + strconv.Itoa(int(code)))
	}
}

func (conn *connection) Serve() {
	defer conn.conn.Close()
	addr := conn.conn.RemoteAddr().String()

	slog.Info("A TCP connection was accepted", "addr", addr)

	for {
		req, err := conn.read()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Info("Connection closed by client", "addr", addr)
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				slog.Info("Connection read timeout", "addr", addr)
				conn.writeError(status.DeadlineExceeded)
				break
			}
			slog.Info(err.Error(), "addr", addr)
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
			slog.Error("Failed to write response", "addr", addr, "error", err.Error())
			break
		}

		slog.Info("The request was processed successfully", "addr", addr, "status", stat.Error(), "operation", req.Header.Operation, "key", req.Header.Key)
	}
	slog.Info("The TCP connection was closed", "addr", addr)
}
