package main

import (
	"log/slog"

	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/log"
	"github.com/anteiro255/gedis/internal/server"
)

func main() {
	log.InitLogger()
	s := server.NewServer()
	s.SetDB(db.NewDB())

	err := s.RunAt("127.0.0.1:8080")
	if err != nil {
		slog.Error(err.Error())
	}
}
