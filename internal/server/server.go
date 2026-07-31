package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/raftnode"
	"github.com/anteiro255/gedis/internal/server/connection"
)

type Server struct {
	listener net.Listener
	db       *db.DB
	config   *config.ServerConfig
	raftNode *raftnode.Node
}

func NewServer(db *db.DB) Server {
	return Server{
		db:     db,
		config: config.DefaultServerConfig(),
	}
}

func (s *Server) SetConfig(cfg *config.ServerConfig) { s.config = cfg }
func (s *Server) SetRaftNode(node *raftnode.Node)    { s.raftNode = node }

func (s *Server) Serve(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second) // Probe every 30s
		tcpConn.SetNoDelay(true)
	}

	clientConn := connection.New(conn, s.db, s.config)
	clientConn.SetRaftNode(s.raftNode)
	clientConn.Serve()
}

func (s *Server) RunAt(ctx context.Context, address string) error {
	slog.Info("Run the server", "address", address)

	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return err
	}

	go func() { // close the server when the context done
		<-ctx.Done()
		if err := s.Close(); err != nil {
			slog.Error("Error on closing the server", "error", err)
		}
	}()

	for {
		netConn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				slog.Info("Close the server")
				return nil
			}
			slog.Error(err.Error())
			continue
		}

		go s.Serve(netConn)
	}
}

func (s *Server) Close() error {
	if s.raftNode != nil {
		if err := s.raftNode.Close(); err != nil {
			return err
		}
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
