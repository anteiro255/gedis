package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/internal/interceptor"
	"github.com/anteiro255/gedis/internal/log"
	"github.com/anteiro255/gedis/internal/server"
)

func main() {
	log.InitLogger(config.PreLoad())

	cfg := config.Load()

	database := db.NewDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := database.LoadSnapshot(cfg.SnapshotPath()); err != nil {
		slog.Error("Failed to load snapshot", "path", cfg.SnapshotPath(), "error", err)
		os.Exit(1)
	}
	slog.Info("Snapshot loaded", "path", cfg.SnapshotPath())

	var wg sync.WaitGroup

	wg.Go(func() { database.RunTTLManager(ctx, cfg) })
	wg.Go(func() { database.RunSnapshotter(ctx, cfg) })

	s := server.NewServer()
	s.SetDB(database)
	s.SetConfig(cfg)

	wg.Go(func() { interceptor.SetInterceptorOn(cancel) }) // cancel the context on shutdowng

	if err := s.RunAt(ctx, cfg.Address()); err != nil {
		slog.Error("Error on server starting", "error", err.Error())
		os.Exit(1)
	}
	wg.Wait()
}
