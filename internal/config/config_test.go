package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/config"
)

func TestDefaultValues(t *testing.T) {
	cfg := config.Default()

	if addr := cfg.Address(); addr != ":8080" {
		t.Errorf("expected default address :8080, got %q", addr)
	}
	if ping := cfg.TCPPingInterval(); ping != 30*time.Second {
		t.Errorf("expected default TCP ping interval 30s, got %v", ping)
	}
	if rcv := cfg.ReceiveTimeout(); rcv != 3*time.Second {
		t.Errorf("expected default receive timeout 3s, got %v", rcv)
	}
	if snd := cfg.SendTimeout(); snd != 3*time.Second {
		t.Errorf("expected default send timeout 3s, got %v", snd)
	}
	if checks := cfg.TTLEntryCheckPerSecond(); checks != 200 {
		t.Errorf("expected default TTL entry checks per second 200, got %d", checks)
	}
	if path := cfg.SnapshotPath(); path != "./gedis.snap" {
		t.Errorf("expected default snapshot path ./gedis.snap, got %q", path)
	}
	if interval := cfg.SnapshotInterval(); interval != 5*time.Minute {
		t.Errorf("expected default snapshot interval 5m, got %v", interval)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	cleanup := setEnv(t, map[string]string{
		"GEDIS_ADDRESS":                  "127.0.0.1:9999",
		"GEDIS_TCP_PING_INTERVAL":        "15000",
		"GEDIS_RECEIVE_TIMEOUT":          "5000",
		"GEDIS_SEND_TIMEOUT":             "7000",
		"GEDIS_ENTRY_CHECKS_PER_SECOND":  "500",
		"GEDIS_SNAPSHOT_PATH":            "/tmp/test.snap",
		"GEDIS_SNAPSHOT_INTERVAL":        "60",
	})
	defer cleanup()

	cfg := config.Load()

	if addr := cfg.Address(); addr != "127.0.0.1:9999" {
		t.Errorf("expected address 127.0.0.1:9999, got %q", addr)
	}
	if ping := cfg.TCPPingInterval(); ping != 15*time.Second {
		t.Errorf("expected TCP ping interval 15s, got %v", ping)
	}
	if rcv := cfg.ReceiveTimeout(); rcv != 5*time.Second {
		t.Errorf("expected receive timeout 5s, got %v", rcv)
	}
	if snd := cfg.SendTimeout(); snd != 7*time.Second {
		t.Errorf("expected send timeout 7s, got %v", snd)
	}
	if checks := cfg.TTLEntryCheckPerSecond(); checks != 500 {
		t.Errorf("expected TTL entry checks per second 500, got %d", checks)
	}
	if path := cfg.SnapshotPath(); path != "/tmp/test.snap" {
		t.Errorf("expected snapshot path /tmp/test.snap, got %q", path)
	}
	if interval := cfg.SnapshotInterval(); interval != 60*time.Second {
		t.Errorf("expected snapshot interval 60s, got %v", interval)
	}
}

func TestLoadEnvironment_InvalidValuesFallbackToDefaults(t *testing.T) {
	cleanup := setEnv(t, map[string]string{
		"GEDIS_TCP_PING_INTERVAL":        "invalid",
		"GEDIS_RECEIVE_TIMEOUT":          "-1",
		"GEDIS_SEND_TIMEOUT":             "0",
		"GEDIS_ENTRY_CHECKS_PER_SECOND":  "0",
		"GEDIS_SNAPSHOT_INTERVAL":        "invalid",
	})
	defer cleanup()

	cfg := config.Load()

	if ping := cfg.TCPPingInterval(); ping != 30*time.Second {
		t.Errorf("expected fallback TCP ping interval 30s, got %v", ping)
	}
	if rcv := cfg.ReceiveTimeout(); rcv != 3*time.Second {
		t.Errorf("expected fallback receive timeout 3s, got %v", rcv)
	}
	if snd := cfg.SendTimeout(); snd != 3*time.Second {
		t.Errorf("expected fallback send timeout 3s, got %v", snd)
	}
	if checks := cfg.TTLEntryCheckPerSecond(); checks != 200 {
		t.Errorf("expected fallback TTL entry checks per second 200, got %d", checks)
	}
	if interval := cfg.SnapshotInterval(); interval != 5*time.Minute {
		t.Errorf("expected fallback snapshot interval 5m, got %v", interval)
	}
}

func TestLoadEnvironment_EmptyEnvVarsFallbackToDefaults(t *testing.T) {
	cleanup := setEnv(t, map[string]string{
		"GEDIS_ADDRESS":                  "",
		"GEDIS_TCP_PING_INTERVAL":        "",
		"GEDIS_RECEIVE_TIMEOUT":          "",
		"GEDIS_SEND_TIMEOUT":             "",
		"GEDIS_ENTRY_CHECKS_PER_SECOND":  "",
		"GEDIS_SNAPSHOT_PATH":            "",
		"GEDIS_SNAPSHOT_INTERVAL":        "",
	})
	defer cleanup()

	cfg := config.Load()
	if addr := cfg.Address(); addr != ":8080" {
		t.Errorf("expected default address :8080 for empty env, got %q", addr)
	}
	if ping := cfg.TCPPingInterval(); ping != 30*time.Second {
		t.Errorf("expected default TCP ping interval 30s, got %v", ping)
	}
}

func TestLoadEnvironment_PartialOverride(t *testing.T) {
	cleanup := setEnv(t, map[string]string{
		"GEDIS_ADDRESS":   "0.0.0.0:3000",
		"GEDIS_SNAPSHOT_PATH": "/custom/path.snap",
	})
	defer cleanup()

	cfg := config.Load()

	if addr := cfg.Address(); addr != "0.0.0.0:3000" {
		t.Errorf("expected address 0.0.0.0:3000, got %q", addr)
	}
	if path := cfg.SnapshotPath(); path != "/custom/path.snap" {
		t.Errorf("expected snapshot path /custom/path.snap, got %q", path)
	}
	if ping := cfg.TCPPingInterval(); ping != 30*time.Second {
		t.Errorf("expected unchanged TCP ping interval 30s, got %v", ping)
	}
}

func setEnv(t *testing.T, vars map[string]string) func() {
	t.Helper()
	prev := make(map[string]string)
	for k, v := range vars {
		prev[k] = os.Getenv(k)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	return func() {
		for k, v := range prev {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}
}
