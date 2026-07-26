package config_test

import (
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/config"
)

func TestConfig_Default(t *testing.T) {
	cfg := config.Default()

	if got := cfg.ReceiveTimeout(); got != 3000*time.Millisecond {
		t.Errorf("expected default ReceiveTimeout 3000ms, got %v", got)
	}
	if got := cfg.SendTimeout(); got != 3000*time.Millisecond {
		t.Errorf("expected default SendTimeout 3000ms, got %v", got)
	}
}

func TestConfig_Load_DefaultEnv(t *testing.T) {
	t.Setenv("RECEIVE_TIMEOUT", "")
	t.Setenv("SEND_TIMEOUT", "")

	cfg := config.Load()

	if got := cfg.ReceiveTimeout(); got != 3000*time.Millisecond {
		t.Errorf("expected default ReceiveTimeout 3000ms when unset, got %v", got)
	}
	if got := cfg.SendTimeout(); got != 3000*time.Millisecond {
		t.Errorf("expected default SendTimeout 3000ms when unset, got %v", got)
	}
}

func TestConfig_Load_CustomEnv(t *testing.T) {
	t.Setenv("RECEIVE_TIMEOUT", "5000")
	t.Setenv("SEND_TIMEOUT", "1500")

	cfg := config.Load()

	if got := cfg.ReceiveTimeout(); got != 5000*time.Millisecond {
		t.Errorf("expected ReceiveTimeout 5000ms, got %v", got)
	}
	if got := cfg.SendTimeout(); got != 1500*time.Millisecond {
		t.Errorf("expected SendTimeout 1500ms, got %v", got)
	}
}

func TestConfig_Load_InvalidEnv(t *testing.T) {
	t.Run("non-integer string", func(t *testing.T) {
		t.Setenv("RECEIVE_TIMEOUT", "not_a_number")
		t.Setenv("SEND_TIMEOUT", "invalid")

		cfg := config.Load()
		if got := cfg.ReceiveTimeout(); got != 3000*time.Millisecond {
			t.Errorf("expected fallback ReceiveTimeout 3000ms, got %v", got)
		}
		if got := cfg.SendTimeout(); got != 3000*time.Millisecond {
			t.Errorf("expected fallback SendTimeout 3000ms, got %v", got)
		}
	})

	t.Run("negative and zero values", func(t *testing.T) {
		t.Setenv("RECEIVE_TIMEOUT", "-500")
		t.Setenv("SEND_TIMEOUT", "0")

		cfg := config.Load()
		if got := cfg.ReceiveTimeout(); got != 3000*time.Millisecond {
			t.Errorf("expected fallback ReceiveTimeout 3000ms for negative value, got %v", got)
		}
		if got := cfg.SendTimeout(); got != 3000*time.Millisecond {
			t.Errorf("expected fallback SendTimeout 3000ms for zero value, got %v", got)
		}
	})
}
