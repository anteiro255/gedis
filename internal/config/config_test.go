package config

import (
	"strconv"
	"testing"
	"time"
)

func TestConfig_Default(t *testing.T) {
	cfg := Default()

	if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeoutInMs)*time.Millisecond {
		t.Errorf("expected default ReceiveTimeout="+strconv.Itoa(int(defaultReceiveTimeoutInMs))+", got %v", got)
	}
	if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeoutInMs)*time.Millisecond {
		t.Errorf("expected default SendTimeout="+strconv.Itoa(int(defaultSendTimeoutInMs))+", got %v", got)
	}
	if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
		t.Errorf("expected default TTLEntriesCheckingPerSecond="+strconv.Itoa(int(defaultTTLEntriesCheckingPerSecond))+", got %v", got)
	}
	if got := cfg.SnapshotPath(); got != defaultSnapshotPath {
		t.Errorf("expected default SnapshotPath=%s, got %s", defaultSnapshotPath, got)
	}
	if got := cfg.SnapshotInterval(); got != time.Duration(defaultSnapshotIntervalInSec)*time.Second {
		t.Errorf("expected default SnapshotInterval=%ds, got %v", defaultSnapshotIntervalInSec, got)
	}
}

func TestConfig_Load_DefaultEnv(t *testing.T) {
	t.Setenv("RECEIVE_TIMEOUT", "")
	t.Setenv("SEND_TIMEOUT", "")
	t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "")
	t.Setenv("GEDIS_SNAPSHOT_PATH", "")
	t.Setenv("GEDIS_SNAPSHOT_INTERVAL", "")

	cfg := Load()

	if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeoutInMs)*time.Millisecond {
		t.Errorf("expected default ReceiveTimeout 3000ms when unset, got %v", got)
	}
	if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeoutInMs)*time.Millisecond {
		t.Errorf("expected default SendTimeout 3000ms when unset, got %v", got)
	}
	if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
		t.Errorf("expected default TTLEntriesCheckingPerSecond=%d when unset, got %d", defaultTTLEntriesCheckingPerSecond, got)
	}
	if got := cfg.SnapshotPath(); got != defaultSnapshotPath {
		t.Errorf("expected default SnapshotPath=%s when unset, got %s", defaultSnapshotPath, got)
	}
	if got := cfg.SnapshotInterval(); got != time.Duration(defaultSnapshotIntervalInSec)*time.Second {
		t.Errorf("expected default SnapshotInterval=%ds when unset, got %v", defaultSnapshotIntervalInSec, got)
	}
}

func TestConfig_Load_CustomEnv(t *testing.T) {
	t.Setenv("RECEIVE_TIMEOUT", "5000")
	t.Setenv("SEND_TIMEOUT", "1500")
	t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "500")
	t.Setenv("GEDIS_SNAPSHOT_PATH", "/tmp/gedis.snap")
	t.Setenv("GEDIS_SNAPSHOT_INTERVAL", "120")

	cfg := Load()

	if got := cfg.ReceiveTimeout(); got != 5000*time.Millisecond {
		t.Errorf("expected ReceiveTimeout 5000ms, got %v", got)
	}
	if got := cfg.SendTimeout(); got != 1500*time.Millisecond {
		t.Errorf("expected SendTimeout 1500ms, got %v", got)
	}
	if got := cfg.TTLEntriesCheckingPerSecond(); got != 500 {
		t.Errorf("expected TTLEntriesCheckingPerSecond 500, got %d", got)
	}
	if got := cfg.SnapshotPath(); got != "/tmp/gedis.snap" {
		t.Errorf("expected SnapshotPath /tmp/gedis.snap, got %s", got)
	}
	if got := cfg.SnapshotInterval(); got != 120*time.Second {
		t.Errorf("expected SnapshotInterval 120s, got %v", got)
	}
}

func TestConfig_Load_InvalidEnv(t *testing.T) {
	t.Run("non-integer string", func(t *testing.T) {
		t.Setenv("RECEIVE_TIMEOUT", "not_a_number")
		t.Setenv("SEND_TIMEOUT", "invalid")
		t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "bad")
		t.Setenv("GEDIS_SNAPSHOT_INTERVAL", "not_a_number")

		cfg := Load()
		if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeoutInMs)*time.Millisecond {
			t.Errorf("expected fallback ReceiveTimeout 3000ms, got %v", got)
		}
		if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeoutInMs)*time.Millisecond {
			t.Errorf("expected fallback SendTimeout 3000ms, got %v", got)
		}
		if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
			t.Errorf("expected fallback TTLEntriesCheckingPerSecond=%d, got %d", defaultTTLEntriesCheckingPerSecond, got)
		}
		if got := cfg.SnapshotInterval(); got != time.Duration(defaultSnapshotIntervalInSec)*time.Second {
			t.Errorf("expected fallback SnapshotInterval=%ds for invalid value, got %v", defaultSnapshotIntervalInSec, got)
		}
	})

	t.Run("negative and zero values", func(t *testing.T) {
		t.Setenv("RECEIVE_TIMEOUT", "-500")
		t.Setenv("SEND_TIMEOUT", "0")
		t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "-1")
		t.Setenv("GEDIS_SNAPSHOT_INTERVAL", "0")

		cfg := Load()
		if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeoutInMs)*time.Millisecond {
			t.Errorf("expected fallback ReceiveTimeout 3000ms for negative value, got %v", got)
		}
		if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeoutInMs)*time.Millisecond {
			t.Errorf("expected fallback SendTimeout 3000ms for zero value, got %v", got)
		}
		if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
			t.Errorf("expected fallback TTLEntriesCheckingPerSecond=%d for negative value, got %d", defaultTTLEntriesCheckingPerSecond, got)
		}
		if got := cfg.SnapshotInterval(); got != time.Duration(defaultSnapshotIntervalInSec)*time.Second {
			t.Errorf("expected fallback SnapshotInterval=%ds for zero value, got %v", defaultSnapshotIntervalInSec, got)
		}
	})
}
