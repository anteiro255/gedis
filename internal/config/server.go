package config

import (
	"log/slog"
	"os"
	"time"
)

type ServerConfig struct {

	// address at which the server will be run
	address string // GEDIS_ADDRESS

	// OS-level probing on idle TCP sockets
	// time of the gaps between OS-level TCP pings to a client
	tcpPingInterval time.Duration // GEDIS_TCP_PING_INTERVAL

	// Max time allowed to read a complete request frame AFTER the first byte arrives
	receiveTimeout time.Duration // RECEIVE_TIMEOUT

	// Max time allowed to write a complete response frame AFTER the first byte is written
	sendTimeout time.Duration // SEND_TIMEOUT
}

func LoadServerCfg() (c *ServerConfig) {
	c.loadAddress()
	c.loadTCPPingInterval()
	c.loadReceiveTimeout()
	c.loadSendTimeout()
	return
}

// address
func (c *ServerConfig) loadAddress() {
	envKey := "GEDIS_ADDRESS"
	if s := os.Getenv(envKey); s != "" {
		c.address = s
		return
	}
	slog.Debug("Failed to load the "+envKey+" environment variable", "current_address", defaultAddress)
	c.address = defaultAddress
}

func (c *ServerConfig) Address() string {
	return c.address
}

// tcpPingInterval
func (c *ServerConfig) loadTCPPingInterval() {
	envKey := "GEDIS_TCP_PING_INTERVAL"
	if dur, ok := loadTime(envKey); ok {
		c.tcpPingInterval = dur
		return
	}
	c.tcpPingInterval = defaultTCPPingInterval
}

func (c *ServerConfig) TCPPingInterval() time.Duration {
	return c.tcpPingInterval
}

// receiveTimeout
func (c *ServerConfig) loadReceiveTimeout() {
	envKey := "GEDIS_RECEIVE_TIMEOUT"
	if t, ok := loadTime(envKey); ok {
		c.receiveTimeout = t
		return
	}
	c.receiveTimeout = defaultReceiveTimeout
}

func (c *ServerConfig) ReceiveTimeout() time.Duration {
	return c.receiveTimeout
}

// sendTimeout
func (c *ServerConfig) loadSendTimeout() {
	envKey := "GEDIS_SEND_TIMEOUT"
	if t, ok := loadTime(envKey); ok {
		c.sendTimeout = t
		return
	}
	c.sendTimeout = defaultSendTimeout
}

func (c *ServerConfig) SendTimeout() time.Duration {
	return c.sendTimeout
}
