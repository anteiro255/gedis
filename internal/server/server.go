package server

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/anteiro255/gedis/internal/action"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type Server struct {
	Listener net.Listener
	Db       *db.DB
}

func NewServer() Server {
	return Server{}
}

func (s *Server) SetDB(d *db.DB) {
	s.Db = d
}

func (s *Server) RunAt(address string) error {
	slog.Info("Starting listening...")

	var err error
	s.Listener, err = net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer s.Listener.Close()

	slog.Info("Starting accepting connections...")
	for {
		var conn Connection
		conn.Conn, err = s.Listener.Accept()
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		conn.Db = s.Db
		go conn.Serve()
	}
}

type Connection struct {
	Conn net.Conn
	Db   *db.DB
}

func (conn *Connection) read(timeout time.Duration) (*protocol.Request, error) {
	// Set an absolute deadline for when this read operation must complete
	if err := conn.Conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	// Ensure the deadline is cleared afterwards so it doesn't affect future reads
	defer conn.Conn.SetReadDeadline(time.Time{})

	// io.ReadFull will block until 'headerAsBytes' is completely filled OR the deadline is hit
	var headerAsBytes [protocol.RequestHeaderSize]byte
	_, err := io.ReadFull(conn.Conn, headerAsBytes[:])
	if err != nil {
		// If a timeout occurred, err will be a network error where Timeout() == true
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, err // custom timeout error or handle accordingly
		}
		return nil, err // connection closed, EOF, etc.
	}

	var req protocol.Request
	req.Header = protocol.NewRequestHeaderFromBytes(headerAsBytes)

	body := make([]byte, req.Header.BodySize)
	if req.Header.BodySize > 0 {
		_, err = io.ReadFull(conn.Conn, body)
		if err != nil {
			return nil, err
		}
	}
	req.Body = body

	return &req, nil
}

func (conn *Connection) writeResponse(resp *protocol.Response) error {
	respBytes := resp.ToBytes()
	totalWritten := 0
	for totalWritten < len(respBytes) {
		n, err := conn.Conn.Write(respBytes[totalWritten:])
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

func (conn *Connection) writeError(code status.Status) {
	addr := conn.Conn.RemoteAddr().String()
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

func (conn *Connection) Serve() {
	defer conn.Conn.Close()
	addr := conn.Conn.RemoteAddr().String()

	slog.Info("A TCP connection was accepted", "addr", addr)

	for {
		req, err := conn.read(3 * time.Second)
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
			DB:         conn.Db,
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
