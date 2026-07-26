package config

import "time"

const (
	defaultReceiveTimeout int64 = 3000 // RECEIVE_TIMEOUT
	defaultSendTimeout    int64 = 3000 // SEND_TIMEOUT
)

func Default() (cfg *Config) {
	return &Config{
		receiveTimeout: time.Duration(defaultReceiveTimeout) * time.Millisecond,
		sendTimeout:    time.Duration(defaultSendTimeout) * time.Millisecond,
	}
}
