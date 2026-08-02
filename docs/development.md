# Server Development

This document is for contributors to the Gedis server. Operational users should start with [`../README.md`](../README.md).

## Module Layout

- `cmd/server`: process entrypoint
- `internal/config`: environment-backed configuration
- `internal/db`: sharded database, TTLs, and snapshot serialization
- `internal/raftnode`: HashiCorp Raft integration and FSM
- `internal/server`: listener and connection lifecycle
- `pkg/protocol`: public wire types, action codes, and status codes

## Build and Test

From the server module directory:

```bash
make format
make test
make check
```

Equivalent Go commands:

```bash
go test ./...
go test -race ./internal/...
go build ./cmd/server
```

## Benchmarks

```bash
go test ./internal/db -run '^$' -bench . -benchmem
go test ./internal/server -run '^$' -bench . -benchmem
```

Network benchmarks include TCP and client round-trip costs; compare them on the same machine and with the same client version.

## Protocol Changes

The client and server share the protocol package. Changes to header sizes, action codes, status codes, or key representation require coordinated changes in both modules and updates to protocol tests.

## Persistence Changes

When changing snapshot data:

1. Update `WriteSnapshot` and `ReadSnapshot` together.
2. Test both standalone snapshot restoration and Raft FSM restoration.
3. Consider existing persisted data and migration behavior.

When changing Raft paths, verify that log and stable store variables are full file paths and that `GEDIS_RAFT_SNAPSHOTS_DIR_PATH` remains a directory.
