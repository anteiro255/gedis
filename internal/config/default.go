package config

import "time"

const (
	defaultReceiveTimeout              int64 = 3000 // GEDIS_RECEIVE_TIMEOUT
	defaultSendTimeout                 int64 = 3000 // GEDIS_SEND_TIMEOUT
	defaultTTLEntriesCheckingPerSecond uint  = 200  // GEDIS_ENTRIES_CHECKING_PER_SECOND
)

func Default() (cfg *Config) {
	return &Config{
		receiveTimeout:              time.Duration(defaultReceiveTimeout) * time.Millisecond,
		sendTimeout:                 time.Duration(defaultSendTimeout) * time.Millisecond,
		ttlEntriesCheckingPerSecond: defaultTTLEntriesCheckingPerSecond,
	}
}
