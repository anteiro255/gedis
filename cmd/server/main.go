package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/log"
	"github.com/anteiro255/gedis/internal/server"
)

func main() {
	log.InitLogger()

	cfg := config.Load()

	database := db.NewDB()
	database.RunTTLManager(context.Background(), cfg)

	s := server.NewServer()
	s.SetDB(database)
	s.SetConfig(cfg)

	if err := s.RunAt("127.0.0.1:8080"); err != nil {
		slog.Error("Error on server starting", "Error", err.Error())
		os.Exit(1)
	}

	if err := s.Close(); err != nil {
		slog.Error("Error on closing the server", "Error", err.Error())
		os.Exit(1)
	}

}
