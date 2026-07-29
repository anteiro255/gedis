package connection

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	dbaction "github.com/anteiro255/gedis/internal/db/action"
	"github.com/anteiro255/gedis/pkg/protocol"
	protocolaction "github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

// A connection with a client
// Every connection is performed in a separate goroutine
type Conn struct {

	// Actual connection
	// Use only for deadlines
	conn net.Conn

	writer *bufio.Writer
	reader *bufio.Reader
	db     *db.DB
	cfg    *config.Config
}

func New(conn net.Conn, db *db.DB, cfg *config.Config) *Conn {
	return &Conn{
		conn:   conn,
		writer: bufio.NewWriter(conn),
		reader: bufio.NewReader(conn),
		db:     db,
		cfg:    cfg,
	}
}

// This function is made for keep alive TCP connections
// The functions first tries to read the first byte of a request(hangs), when the request arrives
// the function reads the first byte, set a deadline and read all the other bytes of the request
func (conn *Conn) read() (*protocol.Request, error) {
	receiveTimeout := conn.cfg.ReceiveTimeout()

	// Read the first byte
	var firstByte [1]byte
	_, err := io.ReadFull(conn.reader, firstByte[:])
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

	// Read all the other bytes of the header after the first
	_, err = io.ReadFull(conn.reader, headerAsBytes[1:])
	if err != nil {
		return nil, err
	}

	var req protocol.Request
	req.Header = protocol.NewRequestHeaderFromBytes(headerAsBytes)

	// read the body
	req.Body = make([]byte, req.Header.BodySize)
	if req.Header.BodySize > 0 {
		_, err = io.ReadFull(conn.reader, req.Body)
		if err != nil {
			return nil, err
		}
	}

	return &req, nil
}

func (conn *Conn) writeResponse(resp *protocol.Response) error {
	sendTimeout := conn.cfg.SendTimeout()

	respBytes := resp.ToBytes()

	// set a deadline before the reading
	if sendTimeout != 0 {
		if err := conn.conn.SetWriteDeadline(time.Now().Add(sendTimeout)); err != nil {
			return err
		}
	}

	_, err := conn.writer.Write(respBytes)
	if err != nil {
		return err
	}

	return conn.writer.Flush()
}

func (conn *Conn) writeError(code status.Status) {
	addr := conn.conn.RemoteAddr().String()

	err := conn.writeResponse(protocol.NewResponse(code, nil))
	if err != nil {
		slog.Info("Failed to write to the connection", "addr", addr, "Error", err.Error())
	}
	slog.Debug(code.Error())
}

func (conn *Conn) Serve() {
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

		action := dbaction.Action{
			DB:         conn.db,
			ActionType: protocolaction.Action(req.Header.Operation),
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
