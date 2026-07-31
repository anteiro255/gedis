package config

import "time"

type Config struct {
	Server  *ServerConfig
	Storage *StorageConfig
}

func Load() (cfg *Config) {
	cfg = &Config{}

	cfg.Server = LoadServerCfg()
	cfg.Storage = LoadStorageCfg()
	return
}

// Server config delegation

func (c *Config) Address() string {
	return c.Server.Address()
}

func (c *Config) TCPPingInterval() time.Duration {
	return c.Server.TCPPingInterval()
}

func (c *Config) ReceiveTimeout() time.Duration {
	return c.Server.ReceiveTimeout()
}

func (c *Config) SendTimeout() time.Duration {
	return c.Server.SendTimeout()
}

// Storage config delegation

func (c *Config) TTLEntryCheckPerSecond() uint {
	return c.Storage.TTLEntryCheckPerSecond()
}

func (c *Config) SnapshotPath() string {
	return c.Storage.SnapshotPath()
}

func (c *Config) SnapshotInterval() time.Duration {
	return c.Storage.SnapshotInterval()
}
