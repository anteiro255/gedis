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
	snapshotPath                string        // GEDIS_SNAPSHOT_PATH
	snapshotInterval            time.Duration // GEDIS_SNAPSHOT_INTERVAL
}

func Load() (cfg *Config) {
	cfg = &Config{}
	cfg.loadReceiveTimeout()
	cfg.loadSendTimeout()
	cfg.loadTTLEntriesCheckingPerSecond()
	cfg.loadSnapshotPath()
	cfg.loadSnapshotInterval()
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
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_receive_timeout", defaultReceiveTimeoutInMs)
	c.receiveTimeout = time.Duration(defaultReceiveTimeoutInMs) * time.Millisecond
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
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_send_timeout", defaultSendTimeoutInMs)
	c.sendTimeout = time.Duration(defaultSendTimeoutInMs) * time.Millisecond
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

// SnapshotPath
func (c *Config) loadSnapshotPath() {
	envKey := "GEDIS_SNAPSHOT_PATH"
	if s := os.Getenv(envKey); s != "" {
		c.snapshotPath = s
		return
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_snapshot_path", defaultSnapshotPath)
	c.snapshotPath = defaultSnapshotPath
}

func (c *Config) SnapshotPath() string {
	return c.snapshotPath
}

// SnapshotInterval
func (c *Config) loadSnapshotInterval() {
	envKey := "GEDIS_SNAPSHOT_INTERVAL"
	if s := os.Getenv(envKey); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.snapshotInterval = time.Duration(n) * time.Second
				return
			}
		}
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_snapshot_interval_seconds", defaultSnapshotIntervalInSec)
	c.snapshotInterval = time.Duration(defaultSnapshotIntervalInSec) * time.Second
}

func (c *Config) SnapshotInterval() time.Duration {
	return c.snapshotInterval
}
