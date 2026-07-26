package config

import (
	"strconv"
	"testing"
	"time"
)

func TestConfig_Default(t *testing.T) {
	cfg := Default()

	if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeout)*time.Millisecond {
		t.Errorf("expected default ReceiveTimeout="+strconv.Itoa(int(defaultReceiveTimeout))+", got %v", got)
	}
	if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeout)*time.Millisecond {
		t.Errorf("expected default SendTimeout="+strconv.Itoa(int(defaultSendTimeout))+", got %v", got)
	}
	if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
		t.Errorf("expected default TTLEntriesCheckingPerSecond="+strconv.Itoa(int(defaultTTLEntriesCheckingPerSecond))+", got %v", got)
	}
}

func TestConfig_Load_DefaultEnv(t *testing.T) {
	t.Setenv("RECEIVE_TIMEOUT", "")
	t.Setenv("SEND_TIMEOUT", "")
	t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "")

	cfg := Load()

	if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeout)*time.Millisecond {
		t.Errorf("expected default ReceiveTimeout 3000ms when unset, got %v", got)
	}
	if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeout)*time.Millisecond {
		t.Errorf("expected default SendTimeout 3000ms when unset, got %v", got)
	}
	if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
		t.Errorf("expected default TTLEntriesCheckingPerSecond=%d when unset, got %d", defaultTTLEntriesCheckingPerSecond, got)
	}
}

func TestConfig_Load_CustomEnv(t *testing.T) {
	t.Setenv("RECEIVE_TIMEOUT", "5000")
	t.Setenv("SEND_TIMEOUT", "1500")
	t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "500")

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
}

func TestConfig_Load_InvalidEnv(t *testing.T) {
	t.Run("non-integer string", func(t *testing.T) {
		t.Setenv("RECEIVE_TIMEOUT", "not_a_number")
		t.Setenv("SEND_TIMEOUT", "invalid")
		t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "bad")

		cfg := Load()
		if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeout)*time.Millisecond {
			t.Errorf("expected fallback ReceiveTimeout 3000ms, got %v", got)
		}
		if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeout)*time.Millisecond {
			t.Errorf("expected fallback SendTimeout 3000ms, got %v", got)
		}
		if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
			t.Errorf("expected fallback TTLEntriesCheckingPerSecond=%d, got %d", defaultTTLEntriesCheckingPerSecond, got)
		}
	})

	t.Run("negative and zero values", func(t *testing.T) {
		t.Setenv("RECEIVE_TIMEOUT", "-500")
		t.Setenv("SEND_TIMEOUT", "0")
		t.Setenv("GEDIS_ENTRIES_CHECKING_PER_SECOND", "-1")

		cfg := Load()
		if got := cfg.ReceiveTimeout(); got != time.Duration(defaultReceiveTimeout)*time.Millisecond {
			t.Errorf("expected fallback ReceiveTimeout 3000ms for negative value, got %v", got)
		}
		if got := cfg.SendTimeout(); got != time.Duration(defaultSendTimeout)*time.Millisecond {
			t.Errorf("expected fallback SendTimeout 3000ms for zero value, got %v", got)
		}
		if got := cfg.TTLEntriesCheckingPerSecond(); got != defaultTTLEntriesCheckingPerSecond {
			t.Errorf("expected fallback TTLEntriesCheckingPerSecond=%d for negative value, got %d", defaultTTLEntriesCheckingPerSecond, got)
		}
	})
}
