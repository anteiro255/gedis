package config

import "time"

const (
	defaultReceiveTimeoutInMs          int64  = 3000
	defaultSendTimeoutInMs             int64  = 3000
	defaultTTLEntriesCheckingPerSecond uint   = 200
	defaultSnapshotPath                string = "gedis.snap"
	defaultSnapshotIntervalInSec       int64  = 300
)

func Default() (cfg *Config) {
	return &Config{
		receiveTimeout:              time.Duration(defaultReceiveTimeoutInMs) * time.Millisecond,
		sendTimeout:                 time.Duration(defaultSendTimeoutInMs) * time.Millisecond,
		ttlEntriesCheckingPerSecond: defaultTTLEntriesCheckingPerSecond,
		snapshotPath:                defaultSnapshotPath,
		snapshotInterval:            time.Duration(defaultSnapshotIntervalInSec) * time.Second,
	}
}
