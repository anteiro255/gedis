package server

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
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

func (s *Server) Addr() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *Server) RunAt(ctx context.Context, address string) error {
	slog.Info("Starting listening...")

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

	slog.Info("Starting accepting connections...")
	for {
		netConn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				slog.Info("Listener closed, stopping accept loop")
				return nil
			}
			slog.Error(err.Error())
			continue
		}

		c := connection{
			conn:   netConn,
			server: s,
		}
		go c.Serve()
	}
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
