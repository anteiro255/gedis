package raftnode

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	hraft "github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb/v2"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	protocolaction "github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

var ErrNotLeader = errors.New("raft: node is not the leader")

// Node owns the Raft instance and exposes the small consensus API needed by
// the server: leadership, replicated mutation, leader discovery, and close.
type Node struct {
	raft      *hraft.Raft
	config    *config.RaftConfig
	database  *db.DB
	closeOnce sync.Once
	closeErr  error
}

// NewNode creates the durable Raft log, snapshot store, network transport, and
// FSM, then restores or bootstraps the cluster configuration.
func NewNode(cfg *config.RaftConfig, database *db.DB) (*Node, error) {
	if cfg.Address() == "" {
		return nil, errors.New("raft address is required")
	}
	// LogPath and StableStorePath are file paths. Only their parent
	// directories are created; the Bolt stores create the files themselves.
	if err := os.MkdirAll(filepath.Dir(cfg.LogPath()), 0o755); err != nil {
		return nil, fmt.Errorf("create raft log directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.StableStorePath()), 0o755); err != nil {
		return nil, fmt.Errorf("create raft stable store directory: %w", err)
	}
	if err := os.MkdirAll(cfg.SnapshotsDirPath(), 0o755); err != nil {
		return nil, fmt.Errorf("create raft snapshot directory: %w", err)
	}

	logStore, err := boltdb.NewBoltStore(cfg.LogPath())
	if err != nil {
		return nil, fmt.Errorf("open raft AOF: %w", err)
	}
	stableStore, err := boltdb.NewBoltStore(cfg.StableStorePath())
	if err != nil {
		logStore.Close()
		return nil, fmt.Errorf("open raft stable store: %w", err)
	}
	snapshots, err := hraft.NewFileSnapshotStore(cfg.SnapshotsDirPath(), 2, io.Discard)
	if err != nil {
		stableStore.Close()
		logStore.Close()
		return nil, fmt.Errorf("open raft snapshots: %w", err)
	}
	transport, err := hraft.NewTCPTransport(cfg.Address(), nil, 3, 10*time.Second, io.Discard)
	if err != nil {
		stableStore.Close()
		logStore.Close()
		return nil, fmt.Errorf("open raft transport: %w", err)
	}

	raftConfig := hraft.DefaultConfig()
	// Peer addresses are the configured identity format, so every node uses
	// its Raft transport address as its stable server ID.
	raftConfig.LocalID = hraft.ServerID(cfg.Address())
	raftConfig.HeartbeatTimeout = cfg.HeartbeatInterval()
	raftConfig.ElectionTimeout = cfg.ElectionTimeoutMin()
	raftConfig.LeaderLeaseTimeout = cfg.HeartbeatInterval()
	raftConfig.CommitTimeout = cfg.CommitTimeout()
	raftConfig.SnapshotInterval = cfg.SnapshotInterval()
	raftConfig.SnapshotThreshold = cfg.SnapshotThreshold()
	raftConfig.TrailingLogs = cfg.TrailingLogs()
	raftConfig.MaxAppendEntries = int(cfg.MaxAppendEntries())
	raftConfig.LogOutput = io.Discard

	node := &Node{config: cfg, database: database}
	node.raft, err = hraft.NewRaft(raftConfig, &fsm{db: database}, logStore, stableStore, snapshots, transport)
	if err != nil {
		stableStore.Close()
		logStore.Close()
		return nil, fmt.Errorf("create raft node: %w", err)
	}

	// Never bootstrap over an existing log. Doing so could overwrite the
	// cluster's committed history after a restart.
	existing, err := hraft.HasExistingState(logStore, stableStore, snapshots)
	if err != nil {
		_ = node.raft.Shutdown().Error()
		return nil, fmt.Errorf("check raft state: %w", err)
	}
	if !existing {
		servers := []hraft.Server{{ID: raftConfig.LocalID, Address: hraft.ServerAddress(cfg.Address())}}
		if cfg.Enabled() {
			for _, address := range cfg.PeerAddresses() {
				if address == cfg.Address() {
					continue
				}
				servers = append(servers, hraft.Server{ID: hraft.ServerID(address), Address: hraft.ServerAddress(address)})
			}
		}
		if err := node.raft.BootstrapCluster(hraft.Configuration{Servers: servers}).Error(); err != nil && !errors.Is(err, hraft.ErrCantBootstrap) {
			_ = node.raft.Shutdown().Error()
			return nil, fmt.Errorf("bootstrap raft cluster: %w", err)
		}
	}

	slog.Info("Raft node started", "id", cfg.NodeID(), "address", cfg.Address())
	return node, nil
}

// RunTTLManager expires keys through the Raft log. Only the leader scans for
// expired keys; the resulting commands are then applied on every member.
// Blocking function
func (n *Node) RunTTLManager(ctx context.Context, checks uint) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !n.IsLeader() {
				continue
			}
			n.database.TickTTLs(checks, func(s *db.Shard, k db.Key) {
				n.Apply(ctx, protocolaction.Del, k, nil)
			})
		}
	}
}

// IsLeader reports whether this member can accept replicated mutations.
func (n *Node) IsLeader() bool { return n.raft.State() == hraft.Leader }

// LeaderAddress returns the address advertised by the current leader. It is
// empty while the cluster is electing or when this node has no leader yet.
func (n *Node) LeaderAddress() string { return string(n.raft.Leader()) }

// Apply serializes a mutation, submits it to the local Raft leader, and waits
// until the entry is committed and applied to the local FSM. A follower is
// rejected before it can mutate its database independently.
func (n *Node) Apply(ctx context.Context, operation protocolaction.Action, key db.Key, body []byte) (status.Status, error) {
	if !n.IsLeader() {
		return status.NotLeader, ErrNotLeader
	}
	var encoded bytes.Buffer
	if err := gob.NewEncoder(&encoded).Encode(command{Operation: operation, Key: key, Body: body}); err != nil {
		return status.InternalError, err
	}
	if err := ctx.Err(); err != nil {
		return status.DeadlineExceeded, err
	}
	future := n.raft.Apply(encoded.Bytes(), n.config.ReplicationTimeout())
	if err := future.Error(); err != nil {
		if !n.IsLeader() {
			return status.NotLeader, ErrNotLeader
		}
		return status.DeadlineExceeded, err
	}
	response := future.Response()
	if err, ok := response.(error); ok {
		return status.InternalError, err
	}
	return response.(applyResult).Status, nil
}

// Close shuts down Raft's replication and transport goroutines.
func (n *Node) Close() error {
	n.closeOnce.Do(func() { n.closeErr = n.raft.Shutdown().Error() })
	return n.closeErr
}
