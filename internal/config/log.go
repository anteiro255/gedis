package config

import (
	"log/slog"
	"strconv"
)

type LogConfig struct {

	// verbosity is the level of logs:
	//     0: Error
	//     1: Warn, Error
	//     2: Info, Warn, Error
	//     3: Debug, Info, Warn, Error
	verbosity uint8 // GEDIS_VERBOSITY
}

func LoadLogConfig() (c *LogConfig) {
	c = DefaultLogConfig()
	c.loadVerbosity()
	return
}

// verbosity
func (c *LogConfig) loadVerbosity() {
	envKey := envPrefix + "VERBOSITY"

	if n, ok := loadInt(envKey); ok {
		if n >= int(minVerbosity) && n <= int(maxVerbosity) {
			c.verbosity = uint8(n)
			return
		}
		slog.Error(envKey + " should be not less than " + strconv.Itoa(int(minVerbosity)) + " and not greater than " + strconv.Itoa(int(maxVerbosity)))
	}
}

func (c *LogConfig) Verbosity() uint8 {
	return c.verbosity
}
