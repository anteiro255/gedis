package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config
type Config struct {
	receiveTimeout time.Duration // RECEIVE_TIMEOUT
	sendTimeout    time.Duration // SEND_TIMEOUT
}

func Load() (cfg *Config) {
	cfg = &Config{}
	cfg.loadReceiveTimeout()
	cfg.loadSendTimeout()
	return
}

// ReceiveTimeout
func (c *Config) loadReceiveTimeout() {
	if s := os.Getenv("RECEIVE_TIMEOUT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.receiveTimeout = time.Duration(n) * time.Millisecond
				return
			}
		}
	}
	slog.Warn("Failed to load the RECEIVE_TIMEOUT environment variable", "current_receive_timeout", defaultReceiveTimeout)
	c.receiveTimeout = time.Duration(defaultReceiveTimeout) * time.Millisecond
}

func (c *Config) ReceiveTimeout() time.Duration {
	return c.receiveTimeout
}

// SendTimeout
func (c *Config) loadSendTimeout() {
	if s := os.Getenv("SEND_TIMEOUT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.sendTimeout = time.Duration(n) * time.Millisecond
				return
			}
		}
	}
	slog.Warn("Failed to load the SEND_TIMEOUT environment variable", "current_send_timeout", defaultSendTimeout)
	c.sendTimeout = time.Duration(defaultSendTimeout) * time.Millisecond
}

func (c *Config) SendTimeout() time.Duration {
	return c.sendTimeout
}
