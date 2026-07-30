package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/server/connection"
)

type Server struct {
	listener net.Listener
	db       *db.DB
	config   *config.Config
}

func NewServer() Server {
	return Server{
		config: config.Default(),
	}
}

func (s *Server) SetDB(d *db.DB) {
	s.db = d
}

func (s *Server) SetConfig(cfg *config.Config) {
	s.config = cfg
}

func (s *Server) Serve(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second) // Probe every 30s
		tcpConn.SetNoDelay(true)
	}

	connection.New(conn, s.db, s.config).Serve()
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
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
