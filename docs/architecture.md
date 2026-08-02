# Architecture

This document describes the server components relevant to operators and integrators.

## Request Path

```text
client
  -> TCP listener
  -> connection framing
  -> local read or mutation dispatch
  -> database or Raft FSM
  -> response frame
```

Reads and existence checks are served locally. In standalone mode, mutations are applied directly to the local sharded database. In Raft mode, mutations are accepted by the leader, committed to the Raft log, and applied by the FSM on every member.

## Database

The database is divided into independent shards protected by read/write locks. Values and TTL metadata are stored separately. Snapshot serialization takes a consistent view of all shards and preserves both maps.

## Standalone Persistence

Standalone mode loads the snapshot from `GEDIS_SNAPSHOT_PATH` at startup. Snapshot writes use a temporary file followed by an atomic rename. The snapshot worker runs at `GEDIS_RAFT_SNAPSHOT_INTERVAL` and saves one final snapshot during shutdown.

Standalone mode intentionally avoids Raft serialization and durable log commits so direct writes remain low overhead.

## Raft Persistence

Raft mode uses three persistence locations:

- `GEDIS_RAFT_LOG_PATH`: durable Bolt-backed Raft log
- `GEDIS_RAFT_STABLE_STORE_PATH`: durable term, vote, and configuration state
- `GEDIS_RAFT_SNAPSHOTS_DIR_PATH`: retained FSM snapshots

The FSM command contains an operation, a key, and an optional body. The FSM applies committed commands deterministically. HashiCorp Raft restores the latest snapshot and replays retained log entries during startup.

## TTL Processing

Standalone mode expires keys through the local database TTL worker. In Raft mode only the leader scans for expired keys. It submits an expiration command containing the expected deadline, and every member applies that command through the FSM. A renewed key therefore cannot be removed by an earlier expiration scan.

## Failure and Routing Behavior

When a mutation reaches a follower, the server returns status `NotLeader`. The response body may contain the leader's Raft transport address. Clients should retry through the leader rather than mutate a follower directly.

## Shutdown

Process cancellation stops background workers and closes the server lifecycle. Raft shutdown is idempotent. Standalone snapshotting saves a final snapshot when its context is canceled.
