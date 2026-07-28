package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config
type Config struct {

	// address at which the server will be run
	address string // GEDIS_ADDRESS

	// OS-level probing on idle TCP sockets
	// time of the gaps between OS-level TCP pings to a client
	tcpPingInterval time.Duration // GEDIS_TCP_PING_INTERVAL

	// Max time allowed to read a complete request frame AFTER the first byte arrives
	receiveTimeout time.Duration // RECEIVE_TIMEOUT

	// Max time allowed to write a complete response frame AFTER the first byte is written
	sendTimeout time.Duration // SEND_TIMEOUT

	// Quantity of random entries in the database that are checked for liveness every second
	ttlEntryCheckPerSecond uint // GEDIS_ENTRY_CHECKS_PER_SECOND

	// Path to the file in which snapsots will be saved (./gedis.snap by default)
	snapshotPath string // GEDIS_SNAPSHOT_PATH

	// time of the gaps between snapsot savings
	snapshotInterval time.Duration // GEDIS_SNAPSHOT_INTERVAL

}

func Load() (cfg *Config) {
	cfg = &Config{}
	cfg.loadAddress()
	cfg.loadTCPPingInterval()
	cfg.loadReceiveTimeout()
	cfg.loadSendTimeout()
	cfg.loadTTLEntryCheckPerSecond()
	cfg.loadSnapshotPath()
	cfg.loadSnapshotInterval()
	return
}

// address
func (c *Config) loadAddress() {
	envKey := "GEDIS_ADDRESS"
	if s := os.Getenv(envKey); s != "" {
		c.address = s
		return
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_address", defaultAddress)
	c.address = defaultAddress
}

func (c *Config) Address() string {
	return c.address
}

// tcpPingInterval
func (c *Config) loadTCPPingInterval() {
	envKey := "GEDIS_TCP_PING_INTERVAL"
	if s := os.Getenv(envKey); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.tcpPingInterval = time.Duration(n) * time.Millisecond
				return
			}
		}
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_tcp_keep_alive_period", defaultTCPPingIntervalInMs)
	c.tcpPingInterval = time.Duration(defaultTCPPingIntervalInMs) * time.Millisecond
}

func (c *Config) TCPPingInterval() time.Duration {
	return c.tcpPingInterval
}

// receiveTimeout
func (c *Config) loadReceiveTimeout() {
	envKey := "GEDIS_RECEIVE_TIMEOUT"
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

// sendTimeout
func (c *Config) loadSendTimeout() {
	envKey := "GEDIS_SEND_TIMEOUT"
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

// ttlEntryCheckPerSecond
func (c *Config) loadTTLEntryCheckPerSecond() {
	envKey := "GEDIS_ENTRY_CHECKS_PER_SECOND"
	if s := os.Getenv(envKey); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 {
				c.ttlEntryCheckPerSecond = uint(n)
				return
			}
		}
	}
	slog.Warn("Failed to load the "+envKey+" environment variable", "current_ttl_entries_checking_per_second", defaultTTLEntryCheckPerSecond)
	c.ttlEntryCheckPerSecond = defaultTTLEntryCheckPerSecond
}

func (c *Config) TTLEntryCheckPerSecond() uint {
	return c.ttlEntryCheckPerSecond
}

// snapshotPath
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

// snapshotInterval
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
