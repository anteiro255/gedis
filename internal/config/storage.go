package config

import (
	"time"
)

type StorageConfig struct {

	// Quantity of random entries in the database that are checked for liveness every second
	ttlEntryCheckPerSecond uint // GEDIS_ENTRY_CHECKS_PER_SECOND

	// Path to the file in which snapsots will be saved (./gedis.snap by default)
	snapshotPath string // GEDIS_SNAPSHOT_PATH

	// time of the gaps between snapsot savings
	snapshotInterval time.Duration // GEDIS_SNAPSHOT_INTERVAL
}

func LoadStorageCfg() (c *StorageConfig) {
	c.loadTTLEntryCheckPerSecond()
	c.loadSnapshotPath()
	c.loadSnapshotInterval()
	return
}

// ttlEntryCheckPerSecond
func (c *StorageConfig) loadTTLEntryCheckPerSecond() {
	envKey := "GEDIS_ENTRY_CHECKS_PER_SECOND"
	if n, ok := loadUInt(envKey); ok {
		c.ttlEntryCheckPerSecond = uint(n)
		return
	}
	c.ttlEntryCheckPerSecond = defaultTTLEntryCheckPerSecond
}

func (c *StorageConfig) TTLEntryCheckPerSecond() uint {
	return c.ttlEntryCheckPerSecond
}

// snapshotPath
func (c *StorageConfig) loadSnapshotPath() {
	envKey := "GEDIS_SNAPSHOT_PATH"
	if s, ok := loadString(envKey); ok {
		c.snapshotPath = s
		return
	}
	c.snapshotPath = defaultSnapshotPath
}

func (c *StorageConfig) SnapshotPath() string {
	return c.snapshotPath
}

// snapshotInterval
func (c *StorageConfig) loadSnapshotInterval() {
	envKey := "GEDIS_SNAPSHOT_INTERVAL"
	if dur, ok := loadTime(envKey); ok {
		c.snapshotInterval = dur
		return
	}
	c.snapshotInterval = defaultSnapshotInterval
}

func (c *StorageConfig) SnapshotInterval() time.Duration {
	return c.snapshotInterval
}
