package server

import (
	"encoding/binary"
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
			return nil, status.DeadLineExceeded // custom timeout error or handle accordingly
		}
		return nil, err // connection closed, EOF, etc.
	}

	var req protocol.Request
	req.Header = protocol.NewRequestHeaderFromBytes(headerAsBytes)

	body := make([]byte, req.Header.BodySize)
	_, err = io.ReadFull(conn.Conn, body)
	req.Body = body

	return &req, err
}

func (conn Connection) writeAll(code status.Status, body []byte) error {
	n, err := conn.Conn.Write(binary.BigEndian.AppendUint32(nil, uint32(code)))
	if err != nil {
		return err
	}

	if n != 4 {
		return io.ErrShortWrite
	}

	for len(body) > 0 {
		n, err := conn.Conn.Write(body)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		body = body[n:]
	}
	return nil
}

func (conn Connection) writeError(code status.Status) {
	ip := conn.Conn.RemoteAddr().String()
	err := conn.writeAll(code, nil)
	if err != nil {
		slog.Info("Failed to write to the connection", "ip", ip, "Error", err.Error())
	}
	switch code {
	case status.NoSuchKey:
		slog.Info("The user submitted not existing key", "ip", ip)
	case status.WrongInput:
		slog.Info("The user submitted wrong input", "ip", ip)
	case status.InternalError:
		slog.Info("Internal error occured during the serving the user", "ip", ip)
	default:
		slog.Error("Wrong error code value was passed to the handlers.error() function, the code value: " + strconv.Itoa(int(code)))
	}
}

func (conn Connection) Serve() {
	defer conn.Conn.Close()
	ip := conn.Conn.RemoteAddr().String()

	slog.Info("A connection was accepted", "ip", ip)

	req, err := conn.read(3 * time.Second)
	if err != nil {
		slog.Info(err.Error(), "ip", ip)
		if err, ok := err.(status.Status); ok {
			conn.writeError(err)
			return
		}
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
		return
	}

	conn.writeAll(stat, body)
	slog.Info("The connection was successfully served", "ip", ip, "status", stat.Error())
}
