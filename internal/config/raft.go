package config

import (
	"strings"
	"time"
)

const raftEnvPrefix = envPrefix + "RAFT_"

type RaftConfig struct {
	enabled              bool          // GEDIS_RAFT_ENABLED
	nodeID               string        // GEDIS_RAFT_NODE_ID
	address              string        // GEDIS_RAFT_ADDRESS
	peerAddresses        []string      // GEDIS_RAFT_PEER_ADDRESSES (peer address are separated with a comma)
	heartbeatInterval    time.Duration // GEDIS_RAFT_HEARTBEAT_INTERVAL
	electionTimeoutMin   time.Duration // GEDIS_RAFT_ELECTION_TIMEOUT_MIN
	electionTimeoutMax   time.Duration // GEDIS_RAFT_ELECTION_TIMEOUT_MAX
	replicationTimeout   time.Duration // GEDIS_RAFT_REPLICATION_TIMEOUT
	commitTimeout        time.Duration // GEDIS_RAFT_COMMIT_TIMEOUT
	snapshotInterval     time.Duration // GEDIS_RAFT_SNAPSHOT_INTERVAL
	snapshotThreshold    uint64        // GEDIS_RAFT_SNAPSHOT_THRESHOLD
	trailingLogs         uint64        // GEDIS_RAFT_TRAILING_LOGS
	maxAppendEntries     uint64        // GEDIS_RAFT_MAX_APPEND_ENTRIES
	logPath              string        // GEDIS_RAFT_LOG_PATH
	stableStorePath      string        // GEDIS_RAFT_STABLE_STORE_PATH
	raftSnapshotsDirPath string        // GEDIS_RAFT_SNAPSHOTS_DIR_PATH
}

// LoadRaftCfg loads all Raft settings. GEDIS_RAFT_ENABLED controls whether
// peer addresses are used; a disabled cluster still runs as a single node.
func LoadRaftCfg() (c *RaftConfig) {
	c = DefaultRaftConfig()

	c.loadEnabled()
	c.loadNodeID()
	c.loadAddress()
	c.loadPeerAddresses()
	c.loadHeartbeatInterval()
	c.loadElectionTimeoutMin()
	c.loadElectionTimeoutMax()
	c.loadReplicationTimeout()
	c.loadCommitTimeout()
	c.loadSnapshotInterval()
	c.loadSnapshotThreshold()
	c.loadTrailingLogs()
	c.loadMaxAppendEntries()
	c.loadLogPath()
	c.loadStableStorePath()
	c.loadRaftSnapshotsDirPath()

	return
}

func (c *RaftConfig) loadEnabled() {
	envVar := raftEnvPrefix + "ENABLED"
	enabled, ok := loadBool(envVar)
	c.enabled = ok && enabled
}
func (c *RaftConfig) loadNodeID() {
	if value, ok := loadString(raftEnvPrefix + "NODE_ID"); ok {
		c.nodeID = value
	}
}
func (c *RaftConfig) loadAddress() {
	if value, ok := loadString(raftEnvPrefix + "ADDRESS"); ok {
		c.address = value
	}
}
func (c *RaftConfig) loadPeerAddresses() {
	if value, ok := loadString(raftEnvPrefix + "PEER_ADDRESSES"); ok {
		for _, address := range strings.Split(value, ",") {
			if address = strings.TrimSpace(address); address != "" {
				c.peerAddresses = append(c.peerAddresses, address)
			}
		}
	}
}
func (c *RaftConfig) loadHeartbeatInterval() {
	if value, ok := loadTime(raftEnvPrefix + "HEARTBEAT_INTERVAL"); ok {
		c.heartbeatInterval = value
	}
}
func (c *RaftConfig) loadElectionTimeoutMin() {
	if value, ok := loadTime(raftEnvPrefix + "ELECTION_TIMEOUT_MIN"); ok {
		c.electionTimeoutMin = value
	}
}
func (c *RaftConfig) loadElectionTimeoutMax() {
	if value, ok := loadTime(raftEnvPrefix + "ELECTION_TIMEOUT_MAX"); ok {
		c.electionTimeoutMax = value
	}
}
func (c *RaftConfig) loadReplicationTimeout() {
	if value, ok := loadTime(raftEnvPrefix + "REPLICATION_TIMEOUT"); ok {
		c.replicationTimeout = value
	}
}
func (c *RaftConfig) loadCommitTimeout() {
	if value, ok := loadTime(raftEnvPrefix + "COMMIT_TIMEOUT"); ok {
		c.commitTimeout = value
	}
}
func (c *RaftConfig) loadSnapshotInterval() {
	if value, ok := loadTime(raftEnvPrefix + "SNAPSHOT_INTERVAL"); ok {
		c.snapshotInterval = value
	}
}
func (c *RaftConfig) loadSnapshotThreshold() {
	if value, ok := loadUInt(raftEnvPrefix + "SNAPSHOT_THRESHOLD"); ok {
		if value > 0 {
			c.snapshotThreshold = uint64(value)
		}
	}
}
func (c *RaftConfig) loadTrailingLogs() {
	if value, ok := loadUInt(raftEnvPrefix + "TRAILING_LOGS"); ok {
		c.trailingLogs = uint64(value)
	}
}
func (c *RaftConfig) loadMaxAppendEntries() {
	if value, ok := loadUInt(raftEnvPrefix + "MAX_APPEND_ENTRIES"); ok {
		if value > 0 {
			c.maxAppendEntries = uint64(value)
		}
	}
}
func (c *RaftConfig) loadLogPath() {
	if value, ok := loadString(raftEnvPrefix + "LOG_PATH"); ok {
		c.logPath = value
	}
}

func (c *RaftConfig) loadStableStorePath() {
	if value, ok := loadString(raftEnvPrefix + "STABLE_STORE_PATH"); ok {
		c.stableStorePath = value
	}
}

func (c *RaftConfig) loadRaftSnapshotsDirPath() {
	if value, ok := loadString(raftEnvPrefix + "SNAPSHOTS_DIR_PATH"); ok {
		c.raftSnapshotsDirPath = value
	}
}

// The accessors keep the configuration immutable to packages that consume it.
func (c *RaftConfig) Enabled() bool                     { return c.enabled }
func (c *RaftConfig) NodeID() string                    { return c.nodeID }
func (c *RaftConfig) Address() string                   { return c.address }
func (c *RaftConfig) PeerAddresses() []string           { return append([]string(nil), c.peerAddresses...) }
func (c *RaftConfig) HeartbeatInterval() time.Duration  { return c.heartbeatInterval }
func (c *RaftConfig) ElectionTimeoutMin() time.Duration { return c.electionTimeoutMin }
func (c *RaftConfig) ElectionTimeoutMax() time.Duration { return c.electionTimeoutMax }
func (c *RaftConfig) ReplicationTimeout() time.Duration { return c.replicationTimeout }
func (c *RaftConfig) CommitTimeout() time.Duration      { return c.commitTimeout }
func (c *RaftConfig) SnapshotInterval() time.Duration   { return c.snapshotInterval }
func (c *RaftConfig) SnapshotThreshold() uint64         { return c.snapshotThreshold }
func (c *RaftConfig) TrailingLogs() uint64              { return c.trailingLogs }
func (c *RaftConfig) MaxAppendEntries() uint64          { return c.maxAppendEntries }
func (c *RaftConfig) LogPath() string                   { return c.logPath }
func (c *RaftConfig) StableStorePath() string           { return c.stableStorePath }
func (c *RaftConfig) SnapshotsDirPath() string          { return c.raftSnapshotsDirPath }
