package server

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
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
		}
		conn.Db = s.Db
		go conn.Serve()
	}
}

type Connection struct {
	Conn net.Conn
	Db   *db.DB
}

func (conn *Connection) Read(timeout time.Duration) (*protocol.Request, error) {
	// Set an absolute deadline for when this read operation must complete
	if err := conn.Conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	// Ensure the deadline is cleared afterwards so it doesn't affect future reads
	defer conn.Conn.SetReadDeadline(time.Time{})

	// io.ReadFull will block until 'headerAsBytes' is completely filled OR the deadline is hit
	var headerAsBytes [protocol.RequestHeaderSize]byte
	n, err := io.ReadFull(conn.Conn, headerAsBytes[:])
	if err != nil {
		// If a timeout occurred, err will be a network error where Timeout() == true
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, os.ErrDeadlineExceeded // custom timeout error or handle accordingly
		}
		return nil, err // connection closed, EOF, etc.
	}
	if n != protocol.RequestHeaderSize {
		return nil, status.WrongInput
	}

	var req protocol.Request
	req.Header = protocol.NewRequestHeaderFromBytes(headerAsBytes)

	body := make([]byte, req.Header.BodySize)
	n, err = io.ReadFull(conn.Conn, body)
	req.Body = body

	return &req, nil
}

func (conn Connection) write(code status.Status, body []byte) error {
	_, err := conn.Conn.Write(binary.BigEndian.AppendUint32(nil, uint32(code)))
	if err != nil {
		return err
	}
	_, err = conn.Conn.Write(body)
	return err
}

func (conn Connection) writeError(code status.Status) {
	ip := conn.Conn.RemoteAddr().String()
	err := conn.write(code, nil)
	if err != nil {
		slog.Error("Failed to write to the connection with ip = " + ip + ". Error: " + err.Error())
	}
	switch code {
	case status.NoSuchKey:
		slog.Info("The user at " + ip + " submitted not existing key")
	case status.WrongInput:
		slog.Info("The user at " + ip + " submitted wrong input")
	case status.InternalError:
		slog.Error("Internal error occured during the serving the user at " + ip)
	default:
		slog.Error("Wrong error code value was passed to the handlers.error() function, the code value: " + strconv.Itoa(int(code)))
	}
}

func (conn Connection) Serve() {
	defer conn.Conn.Close()
	ip := conn.Conn.RemoteAddr().String()

	slog.Info("A connection with ip = " + ip + " was accepted")

	req, err := conn.Read(3 * time.Second)

	switch {
	case errors.Is(err, os.ErrDeadlineExceeded):
		slog.Info("The user exceeded timeout.", "ip", ip)
		return
	case errors.Is(err, status.WrongInput):
		conn.writeError(status.WrongInput)
		slog.Error("Failed to read from the connection with ip = " + ip + ". Error: " + err.Error())
		return
	}

	action := action.Action{
		DB:         conn.Db,
		ActionType: protocol.ActionType(req.Header.Operation),
		Key:        req.Header.Key,
		Body:       req.Body,
	}
	err = action.Perform()
	if err != nil {
		stat, ok := err.(status.Status)
		if !ok {
			stat = status.InternalError
		}
		conn.writeError(stat)
		slog.Error(stat.Error())
		return
	}

	err = conn.Conn.Close()
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("The connection was successfully served and closed", "ip", ip)
}
