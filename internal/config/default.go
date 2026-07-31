package config

import "time"

const (
	defaultAddress                string        = ":8080"
	defaultVerbosity              uint8         = 2
	defaultTCPPingInterval        time.Duration = 30 * time.Second
	defaultReceiveTimeout         time.Duration = 3 * time.Second
	defaultSendTimeout            time.Duration = 3 * time.Second
	defaultTTLEntryCheckPerSecond uint          = 200
	defaultSnapshotPath           string        = "./gedis.snap"
	defaultSnapshotInterval       time.Duration = 300 * time.Second
)

func Default() (cfg *Config) {
	return &Config{
		Server: &ServerConfig{
			address:         defaultAddress,
			tcpPingInterval: defaultTCPPingInterval,
			receiveTimeout:  defaultReceiveTimeout,
			sendTimeout:     defaultSendTimeout,
		},
		Storage: &StorageConfig{
			ttlEntryCheckPerSecond: defaultTTLEntryCheckPerSecond,
			snapshotPath:           defaultSnapshotPath,
			snapshotInterval:       defaultSnapshotInterval,
		},
	}
}
