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
	"github.com/anteiro255/gedis/internal/raftnode"
	"github.com/anteiro255/gedis/internal/server"
)

func main() {
	log.InitLogger(config.LoadLogConfig())

	cfg := config.Load()

	database := db.NewDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	if !cfg.Raft.Enabled() {
		if err := database.LoadSnapshot(cfg.Raft.SnapshotPath()); err != nil {
			slog.Error("Failed to load snapshot", "path", cfg.Raft.SnapshotPath(), "error", err)
			os.Exit(1)
		}
		wg.Go(func() { database.RunTTLManager(ctx, cfg.Storage.TTLEntryCheckPerSecond()) })
		wg.Go(func() { database.RunSnapshotter(ctx, cfg.Raft.SnapshotPath(), cfg.Raft.SnapshotInterval()) })
	}

	s := server.NewServer(database)
	s.SetConfig(cfg.Server)

	if cfg.Raft.Enabled() {
		raftNode, err := raftnode.NewNode(cfg.Raft, database)
		if err != nil {
			slog.Error("Failed to start Raft", "error", err)
			os.Exit(1)
		}
		s.SetRaftNode(raftNode)
		wg.Go(func() { raftNode.RunTTLManager(ctx, cfg.Storage.TTLEntryCheckPerSecond()) })
	}

	wg.Go(func() { interceptor.SetInterceptorOn(cancel) }) // cancel the context on shutdowng

	if err := s.RunAt(ctx, cfg.Server.Address()); err != nil {
		slog.Error("Error on server starting", "error", err.Error())
		os.Exit(1)
	}
	wg.Wait()
}
