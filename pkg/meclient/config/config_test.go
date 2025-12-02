package meclient

import (
	"errors"
	"testing"
	"time"
)

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := DefaultConfig("localhost:12345")

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestConfig_Validate_EmptyAddress(t *testing.T) {
	cfg := DefaultConfig("")

	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got: %v", err)
	}
}

func TestConfig_Validate_InvalidChannelBuffer(t *testing.T) {
	tests := []struct {
		name   string
		buffer int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too_large", MaxChannelBuffer + 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig("localhost:12345")
			cfg.ChannelBuffer = tt.buffer

			err := cfg.Validate()
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("expected ErrInvalidConfig, got: %v", err)
			}
		})
	}
}

func TestConfig_Validate_InvalidTimeout(t *testing.T) {
	cfg := DefaultConfig("localhost:12345")
	cfg.ConnectTimeout = 0

	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got: %v", err)
	}
}

func TestConfig_Validate_InvalidReconnectDelays(t *testing.T) {
	tests := []struct {
		name     string
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{"zero_min", 0, time.Second},
		{"zero_max", time.Second, 0},
		{"negative_min", -time.Second, time.Second},
		{"negative_max", time.Second, -time.Second},
		{"min_greater_than_max", 10 * time.Second, time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig("localhost:12345")
			cfg.ReconnectMinDelay = tt.minDelay
			cfg.ReconnectMaxDelay = tt.maxDelay

			err := cfg.Validate()
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("expected ErrInvalidConfig, got: %v", err)
			}
		})
	}
}

func TestConfig_Validate_MaxChannelBuffer(t *testing.T) {
	cfg := DefaultConfig("localhost:12345")
	cfg.ChannelBuffer = MaxChannelBuffer

	if err := cfg.Validate(); err != nil {
		t.Errorf("max channel buffer should be valid, got: %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("localhost:12345")

	if cfg.Address != "localhost:12345" {
		t.Errorf("expected address localhost:12345, got: %s", cfg.Address)
	}

	if cfg.ChannelBuffer != DefaultChannelBuffer {
		t.Errorf("expected default channel buffer %d, got: %d", DefaultChannelBuffer, cfg.ChannelBuffer)
	}

	if cfg.ConnectTimeout != DefaultConnectTimeout {
		t.Errorf("expected default connect timeout %v, got: %v", DefaultConnectTimeout, cfg.ConnectTimeout)
	}

	if !cfg.AutoReconnect {
		t.Error("expected auto reconnect to be true by default")
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := Config{Address: "localhost:12345"}
	cfg = applyDefaults(cfg)

	if cfg.ChannelBuffer != DefaultChannelBuffer {
		t.Errorf("expected default channel buffer")
	}

	if cfg.ReconnectMinDelay != DefaultReconnectMinDelay {
		t.Errorf("expected default reconnect min delay")
	}

	if cfg.ReconnectMaxDelay != DefaultReconnectMaxDelay {
		t.Errorf("expected default reconnect max delay")
	}

	if cfg.ConnectTimeout != DefaultConnectTimeout {
		t.Errorf("expected default connect timeout")
	}
}

func TestApplyDefaults_PreservesExisting(t *testing.T) {
	cfg := Config{
		Address:           "localhost:12345",
		ChannelBuffer:     2048,
		ReconnectMinDelay: 500 * time.Millisecond,
		ReconnectMaxDelay: 60 * time.Second,
		ConnectTimeout:    10 * time.Second,
	}

	cfg = applyDefaults(cfg)

	if cfg.ChannelBuffer != 2048 {
		t.Errorf("should preserve existing channel buffer")
	}

	if cfg.ReconnectMinDelay != 500*time.Millisecond {
		t.Errorf("should preserve existing reconnect min delay")
	}
}
