package config

type StorageConfig struct {
	// Quantity of random entries in the database that are checked for liveness every second
	ttlEntryCheckPerSecond uint // GEDIS_ENTRY_CHECKS_PER_SECOND
}

func LoadStorageConfig() (c *StorageConfig) {
	c = DefaultStorageConfig()
	c.loadTTLEntryCheckPerSecond()
	return
}

// ttlEntryCheckPerSecond
func (c *StorageConfig) loadTTLEntryCheckPerSecond() {
	envKey := envPrefix + "ENTRY_CHECKS_PER_SECOND"
	if n, ok := loadUInt(envKey); ok {
		c.ttlEntryCheckPerSecond = uint(n)
	}
}

func (c *StorageConfig) TTLEntryCheckPerSecond() uint {
	return c.ttlEntryCheckPerSecond
}
