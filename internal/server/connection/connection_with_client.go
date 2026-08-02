package connection

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	dbaction "github.com/anteiro255/gedis/internal/db/action"
	"github.com/anteiro255/gedis/internal/raftnode"
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

	writer         *bufio.Writer
	reader         *bufio.Reader
	db             *db.DB
	cfg            *config.ServerConfig
	raft           *raftnode.Node
	requestHeader  [protocol.RequestHeaderSize]byte
	responseHeader [protocol.ResponseHeaderSize]byte
}

// SetRaftNode enables consensus-aware mutation handling for this client
// connection. Reads continue to use the local database.
func (conn *Conn) SetRaftNode(node *raftnode.Node) { conn.raft = node }

func New(conn net.Conn, db *db.DB, cfg *config.ServerConfig) *Conn {
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
func (conn *Conn) read() (protocol.Request, error) {
	receiveTimeout := conn.cfg.ReceiveTimeout()

	// Read the first byte
	_, err := io.ReadFull(conn.reader, conn.requestHeader[:1])
	if err != nil {
		return protocol.Request{}, err
	}

	// Set a deadline after the reading the first byte
	if receiveTimeout != 0 { // 0 means no timeout
		if err := conn.conn.SetReadDeadline(time.Now().Add(receiveTimeout)); err != nil {
			return protocol.Request{}, err
		}
	}

	headerAsBytes := &conn.requestHeader

	// Read all the other bytes of the header after the first
	_, err = io.ReadFull(conn.reader, headerAsBytes[1:])
	if err != nil {
		return protocol.Request{}, err
	}

	var req protocol.Request
	req.Header = protocol.RequestHeaderFromBytes(headerAsBytes)

	// read the body
	if req.Header.BodySize > 4096 {
		req.Body = make([]byte, req.Header.BodySize)
		_, err = io.ReadFull(conn.reader, req.Body)
		if err != nil {
			return protocol.Request{}, err
		}
	} else if req.Header.BodySize > 0 {
		req.Body, err = conn.reader.Peek(int(req.Header.BodySize)) // Use Peak to avoid allocations
		if err != nil {
			return protocol.Request{}, err
		}
	}

	return req, nil
}

func (conn *Conn) writeResponse(sts status.Status, body []byte) error {
	sendTimeout := conn.cfg.SendTimeout()

	// set a deadline before the reading
	if sendTimeout != 0 {
		if err := conn.conn.SetWriteDeadline(time.Now().Add(sendTimeout)); err != nil {
			return err
		}
	}

	header := protocol.ResponseHeader{Status: sts, BodySize: uint32(len(body))}
	header.MarshalTo(conn.responseHeader[:])
	_, err := conn.writer.Write(conn.responseHeader[:])
	if err != nil {
		return err
	}

	if len(body) > 0 {
		_, err = conn.writer.Write(body)
		if err != nil {
			return err
		}
	}

	return conn.writer.Flush()
}

func (conn *Conn) writeError(code status.Status) {
	addr := conn.conn.RemoteAddr().String()

	err := conn.writeResponse(code, nil)
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

		// Peek borrows the reader buffer. A SET value must be detached when it
		// is small enough to be borrowed; large bodies already have ownership of
		// their newly allocated read buffer and can be stored without another copy.
		if req.Header.Operation == uint8(protocolaction.Set) && req.Header.BodySize > 0 && req.Header.BodySize <= 4096 {
			owned := make([]byte, len(req.Body))
			copy(owned, req.Body)
			req.Body = owned
		}

		op := protocolaction.Action(req.Header.Operation)
		if conn.raft == nil {
			action := dbaction.Action{DB: conn.db, ActionType: op, Key: req.Header.Key, Body: req.Body}
			body, stat := action.Perform()
			if stat == status.InternalError {
				slog.Error(stat.Error())
				conn.writeError(stat)
				break
			}
			if req.Header.BodySize > 0 && req.Header.BodySize <= 4096 {
				_, err = conn.reader.Discard(int(req.Header.BodySize))
				if err != nil {
					slog.Error("Discarding bytes from the buffer", "error", err)
				}
			}
			if err := conn.writeResponse(stat, body); err != nil {
				slog.Debug("Failed to write response", "addr", addr, "error", err.Error())
				break
			}
			continue
		}

		var body []byte
		var stat status.Status
		var applyErr error
		if isMutation(op) {
			// Only the leader may append mutations. Followers return the known
			// leader address instead of creating divergent local state.
			if !conn.raft.IsLeader() {
				if req.Header.BodySize > 0 && req.Header.BodySize <= 4096 {
					_, _ = conn.reader.Discard(int(req.Header.BodySize))
				}
				leader := []byte(conn.raft.LeaderAddress())
				if err := conn.writeResponse(status.NotLeader, leader); err != nil {
					slog.Debug("Failed to write response", "addr", addr, "error", err.Error())
					break
				}
				continue
			}
			stat, applyErr = conn.raft.Apply(context.Background(), op, req.Header.Key, req.Body)
		} else {
			action := dbaction.Action{DB: conn.db, ActionType: op, Key: req.Header.Key, Body: req.Body}
			body, stat = action.Perform()
		}
		if applyErr != nil {
			if stat == status.NotLeader {
				body = []byte(conn.raft.LeaderAddress())
			} else {
				stat = status.InternalError
			}
		}
		if stat == status.InternalError {
			slog.Error(stat.Error())
			conn.writeError(stat)
			break
		}

		if req.Header.BodySize > 0 && req.Header.BodySize <= 4096 {
			_, err = conn.reader.Discard(int(req.Header.BodySize)) // Discard used bytes, if we used Peak during the request parsing
			if err != nil {
				slog.Error("Discarding bytes from the buffer", "error", err)
			}
		}

		if err := conn.writeResponse(stat, body); err != nil {
			slog.Debug("Failed to write response", "addr", addr, "error", err.Error())
			break
		}

	}
	slog.Info("The client was disconnected", "addr", addr)
}

func isMutation(op protocolaction.Action) bool {
	// These operations change replicated state; GET, EXISTS, and TTL_GET are
	// safe to execute locally on every member.
	return op == protocolaction.Set || op == protocolaction.Del || op == protocolaction.TTL_Set || op == protocolaction.TTL_Del
}
