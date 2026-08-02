package raftnode

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/config"
	"github.com/anteiro255/gedis/internal/db"
	protocolaction "github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func TestNodeAppliesCommittedMutation(t *testing.T) {
	address := freeAddress(t)
	dataPath := t.TempDir()
	setEnv(t, "GEDIS_RAFT_ENABLED", "1")
	setEnv(t, "GEDIS_RAFT_ADDRESS", address)
	setEnv(t, "GEDIS_RAFT_NODE_ID", address)
	setEnv(t, "GEDIS_RAFT_LOG_PATH", filepath.Join(dataPath, "raft", "aof.log"))
	setEnv(t, "GEDIS_RAFT_STABLE_STORE_PATH", filepath.Join(dataPath, "raft", "stable.db"))
	setEnv(t, "GEDIS_RAFT_SNAPSHOTS_DIR_PATH", filepath.Join(dataPath, "snapshots"))
	setEnv(t, "GEDIS_RAFT_HEARTBEAT_INTERVAL", "20ms")
	setEnv(t, "GEDIS_RAFT_ELECTION_TIMEOUT_MIN", "50ms")
	setEnv(t, "GEDIS_RAFT_ELECTION_TIMEOUT_MAX", "100ms")
	setEnv(t, "GEDIS_RAFT_REPLICATION_TIMEOUT", "1s")

	cfg := config.LoadRaftCfg()
	if got := cfg.LogPath(); got != filepath.Join(dataPath, "raft", "aof.log") {
		t.Fatalf("log path = %q", got)
	}
	if got := cfg.StableStorePath(); got != filepath.Join(dataPath, "raft", "stable.db") {
		t.Fatalf("stable store path = %q", got)
	}
	if got := cfg.SnapshotsDirPath(); got != filepath.Join(dataPath, "snapshots") {
		t.Fatalf("snapshots dir path = %q", got)
	}
	database := db.NewDB()
	node, err := NewNode(cfg, database)
	if err != nil {
		t.Fatal(err)
	}
	defer node.Close()
	for _, path := range []string{
		cfg.LogPath(),
		cfg.StableStorePath(),
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Raft path %q was not created: %v", path, err)
		}
		if info.IsDir() {
			t.Fatalf("Raft file path %q was created as a directory", path)
		}
	}
	if info, err := os.Stat(cfg.SnapshotsDirPath()); err != nil {
		t.Fatalf("Raft snapshots directory path %q was not created: %v", cfg.SnapshotsDirPath(), err)
	} else if !info.IsDir() {
		t.Fatalf("Raft snapshots directory path %q is not a directory", cfg.SnapshotsDirPath())
	}

	deadline := time.Now().Add(2 * time.Second)
	for !node.IsLeader() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !node.IsLeader() {
		t.Fatal("single-node Raft cluster did not elect a leader")
	}

	key := db.Key{1}
	if sts, err := node.Apply(context.Background(), protocolaction.Set, key, []byte("value")); err != nil || sts != status.OK {
		t.Fatalf("apply failed: status=%v error=%v", sts, err)
	}
	got, sts := database.Get(key)
	if sts != status.OK || string(got) != "value" {
		t.Fatalf("database value mismatch: status=%v value=%q", sts, got)
	}
}

func freeAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	return address
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	previous, exists := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}
