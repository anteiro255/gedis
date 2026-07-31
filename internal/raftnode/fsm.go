package raftnode

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/anteiro255/gedis/internal/db"
	dbaction "github.com/anteiro255/gedis/internal/db/action"
	protocolaction "github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
	hraft "github.com/hashicorp/raft"
)

// command is the deterministic mutation stored in the replicated Raft log.
// Reads are intentionally not commands: every node can serve them locally.
type command struct {
	Operation protocolaction.Action
	Key       db.Key
	Body      []byte
}

// applyResult is returned by the FSM after a committed mutation has changed
// the database. Raft delivers this value back to the leader's Apply caller.
type applyResult struct {
	Status status.Status
}

// Finite state machine
type fsm struct{ db *db.DB }

// Apply replays one committed log entry. It must be deterministic because the
// same entry is executed independently by every Raft member.
func (f *fsm) Apply(log *hraft.Log) any {
	var cmd command
	if err := gob.NewDecoder(bytes.NewReader(log.Data)).Decode(&cmd); err != nil {
		return err
	}
	action := dbaction.Action{DB: f.db, ActionType: cmd.Operation, Key: cmd.Key, Body: cmd.Body}
	_, sts := action.Perform()
	return applyResult{Status: sts}
}

// Snapshot captures the current database state so Raft can discard old log
// entries. The returned object is persisted asynchronously by Raft.
func (f *fsm) Snapshot() (hraft.FSMSnapshot, error) {
	var buffer bytes.Buffer
	if err := f.db.WriteSnapshot(&buffer); err != nil {
		return nil, err
	}
	return snapshot{data: buffer.Bytes()}, nil
}

// Restore replaces the local database with a snapshot received from Raft.
// This is used when a node joins after compaction or restarts from a snapshot.
func (f *fsm) Restore(reader io.ReadCloser) error {
	defer reader.Close()
	return f.db.ReadSnapshot(reader)
}
