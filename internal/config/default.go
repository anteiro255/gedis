package config

import "time"

func DefaultConfig() *Config {
	return &Config{
		Server:  DefaultServerConfig(),
		Storage: DefaultStorageConfig(),
		Raft:    DefaultRaftConfig(),
	}
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		address:         ":8080",
		tcpPingInterval: 30 * time.Second,
		receiveTimeout:  3 * time.Second,
		sendTimeout:     3 * time.Second,
	}
}

func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		ttlEntryCheckPerSecond: 200,
	}
}

func DefaultRaftConfig() *RaftConfig {
	return &RaftConfig{
		enabled:            false,
		nodeID:             "",
		address:            ":7000",
		peerAddresses:      nil,
		heartbeatInterval:  100 * time.Millisecond,
		electionTimeoutMin: 300 * time.Millisecond,
		electionTimeoutMax: 600 * time.Millisecond,
		replicationTimeout: time.Second,
		commitTimeout:      5 * time.Millisecond,
		snapshotInterval:   10 * time.Minute,
		snapshotThreshold:  8192,
		trailingLogs:       1024,
		maxAppendEntries:   64,
		logPath:            "./gedis-raft/aof.log",
		stableStorePath:    "./gedis-raft/stable.db",
		snapshotsPath:      "./gedis-raft/snapshots",
	}
}

func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		verbosity: 2,
	}
}
