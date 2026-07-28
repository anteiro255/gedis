package config

import "time"

const (
	defaultAddress                string = ":8080"
	defaultTCPPingIntervalInMs    int64  = 30000
	defaultReceiveTimeoutInMs     int64  = 3000
	defaultSendTimeoutInMs        int64  = 3000
	defaultTTLEntryCheckPerSecond uint   = 200
	defaultSnapshotPath           string = "./gedis.snap"
	defaultSnapshotIntervalInSec  int64  = 300
)

func Default() (cfg *Config) {
	return &Config{
		address:                defaultAddress,
		tcpPingInterval:        time.Duration(defaultTCPPingIntervalInMs) * time.Millisecond,
		receiveTimeout:         time.Duration(defaultReceiveTimeoutInMs) * time.Millisecond,
		sendTimeout:            time.Duration(defaultSendTimeoutInMs) * time.Millisecond,
		ttlEntryCheckPerSecond: defaultTTLEntryCheckPerSecond,
		snapshotPath:           defaultSnapshotPath,
		snapshotInterval:       time.Duration(defaultSnapshotIntervalInSec) * time.Second,
	}
}
