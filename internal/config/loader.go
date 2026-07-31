package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

func loadString(envKey string) (string, bool) {
	if val := os.Getenv(envKey); val != "" {
		slog.Debug(envKey + "=" + val + " was loaded")
		return val, true
	}
	slog.Debug("There's no " + envKey + " in environment")
	return "", false
}

func loadInt(envKey string) (int, bool) {
	if val := os.Getenv(envKey); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n, true
		}
		slog.Error("Failed to convert "+envKey+"="+val+" to int", "error", "the value is not a number")
	}
	return 0, false
}

func loadUInt(envKey string) (int, bool) {
	if val := os.Getenv(envKey); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			if n >= 0 {
				slog.Debug(envKey + "=" + val + "was loaded")
				return n, true
			}
			slog.Error("Failed to convert "+envKey+"="+val+" to uint", "error", "the value is less than 0")
			return 0, false
		}
		slog.Error("Failed to convert "+envKey+"="+val+" to uint", "error", "the value is not a number")
		return 0, false
	}
	slog.Debug("There's no " + envKey + " in environment")
	return 0, false
}

func loadTime(envKey string) (time.Duration, bool) {
	if val := os.Getenv(envKey); val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			slog.Debug(envKey + "=" + val + " was loaded")
			return d, true
		}
		slog.Error("Failed to convert "+envKey+"="+val+" to time.Duration", "error", err.Error())
	}
	slog.Debug("There's no " + envKey + " in environment")
	return time.Duration(0), false
}
