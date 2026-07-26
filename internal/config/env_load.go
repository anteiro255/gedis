package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config
type Config struct {
	receiveTimeout              time.Duration // RECEIVE_TIMEOUT
	sendTimeout                 time.Duration // SEND_TIMEOUT
	ttlEntriesCheckingPerSecond uint          // GEDIS_ENTRIES_CHECKING_PER_SECOND
}

func Load() (cfg *Config) {
	cfg = &Config{}
	cfg.loadReceiveTimeout()
	cfg.loadSendTimeout()
	cfg.loadTTLEntriesCheckingPerSecond()
	return
}

// ReceiveTimeout
func (c *Config) loadReceiveTimeout() {
	envKey := "RECEIVE_TIMEOUT"
	if s := os.Getenv(envKey); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.receiveTimeout = time.Duration(n) * time.Millisecond
				return
			}
		}
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_receive_timeout", defaultReceiveTimeout)
	c.receiveTimeout = time.Duration(defaultReceiveTimeout) * time.Millisecond
}

func (c *Config) ReceiveTimeout() time.Duration {
	return c.receiveTimeout
}

// SendTimeout
func (c *Config) loadSendTimeout() {
	envKey := "SEND_TIMEOUT"
	if s := os.Getenv(envKey); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.sendTimeout = time.Duration(n) * time.Millisecond
				return
			}
		}
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_send_timeout", defaultSendTimeout)
	c.sendTimeout = time.Duration(defaultSendTimeout) * time.Millisecond
}

func (c *Config) SendTimeout() time.Duration {
	return c.sendTimeout
}

func (c *Config) loadTTLEntriesCheckingPerSecond() {
	envKey := "GEDIS_ENTRIES_CHECKING_PER_SECOND"
	if s := os.Getenv(envKey); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.ttlEntriesCheckingPerSecond = uint(n)
				return
			}
		}
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_ttl_entries_checking_per_second", defaultTTLEntriesCheckingPerSecond)
	c.ttlEntriesCheckingPerSecond = defaultTTLEntriesCheckingPerSecond
}

func (c *Config) TTLEntriesCheckingPerSecond() uint {
	return c.ttlEntriesCheckingPerSecond
}
