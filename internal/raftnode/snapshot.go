package raftnode

import (
	hraft "github.com/hashicorp/raft"
)

// snapshot is the in-memory snapshot handed from the FSM to Raft's storage
// layer. Persist and Release are part of the hashicorp/raft FSM contract.
type snapshot struct{ data []byte }

// Persist writes the snapshot bytes to Raft's temporary snapshot sink. Raft
// only makes the snapshot visible after Close succeeds.
func (s snapshot) Persist(sink hraft.SnapshotSink) error {
	if _, err := sink.Write(s.data); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

// Release releases resources held by a snapshot. This snapshot only owns a
// byte slice managed by the Go runtime, so there is nothing to release.
func (snapshot) Release() {}
