// Full path: pkg/meclient/config/config_test.go

package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default("localhost:12345")

	if cfg.Address != "localhost:12345" {
		t.Errorf("expected address localhost:12345, got %s", cfg.Address)
	}
	if cfg.ChannelBuffer != DefaultChannelBuffer {
		t.Errorf("expected channel buffer %d, got %d", DefaultChannelBuffer, cfg.ChannelBuffer)
	}
}

func TestConfigValidation_EmptyAddress(t *testing.T) {
	cfg := Default("")

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for empty address")
	}
}

func TestConfigValidation_InvalidChannelBuffer(t *testing.T) {
	tests := []struct {
		name   string
		buffer int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too large", MaxChannelBuffer + 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default("localhost:12345")
			cfg.ChannelBuffer = tt.buffer
			if err := cfg.Validate(); err == nil {
				t.Errorf("expected error for channel buffer %d", tt.buffer)
			}
		})
	}
}

func TestConfigValidation_ValidConfig(t *testing.T) {
	cfg := Default("localhost:12345")

	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConfigValidation_InvalidTimeouts(t *testing.T) {
	tests := []struct {
		name           string
		connectTimeout time.Duration
		minDelay       time.Duration
		maxDelay       time.Duration
	}{
		{"zero connect timeout", 0, DefaultReconnectMinDelay, DefaultReconnectMaxDelay},
		{"negative connect timeout", -1, DefaultReconnectMinDelay, DefaultReconnectMaxDelay},
		{"zero min delay", DefaultConnectTimeout, 0, DefaultReconnectMaxDelay},
		{"zero max delay", DefaultConnectTimeout, DefaultReconnectMinDelay, 0},
		{"min > max delay", DefaultConnectTimeout, time.Second, time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default("localhost:12345")
			cfg.ConnectTimeout = tt.connectTimeout
			cfg.ReconnectMinDelay = tt.minDelay
			cfg.ReconnectMaxDelay = tt.maxDelay
			if err := cfg.Validate(); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestConfigValidation_EdgeCases(t *testing.T) {
	cfg := Default("localhost:12345")
	cfg.ChannelBuffer = 1 // minimum valid
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error for minimum buffer: %v", err)
	}
}

func TestConfigValidation_MaxChannelBuffer(t *testing.T) {
	cfg := Default("localhost:12345")
	cfg.ChannelBuffer = MaxChannelBuffer // maximum valid
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error for maximum buffer: %v", err)
	}
}

func TestApplyDefaults_ZeroValues(t *testing.T) {
	cfg := Config{
		Address:           "localhost:12345",
		ChannelBuffer:     0,
		ReconnectMinDelay: 0,
		ReconnectMaxDelay: 0,
		ConnectTimeout:    0,
	}

	cfg = ApplyDefaults(cfg)

	if cfg.ChannelBuffer != DefaultChannelBuffer {
		t.Errorf("expected channel buffer %d, got %d", DefaultChannelBuffer, cfg.ChannelBuffer)
	}
	if cfg.ReconnectMinDelay != DefaultReconnectMinDelay {
		t.Errorf("expected min delay %v, got %v", DefaultReconnectMinDelay, cfg.ReconnectMinDelay)
	}
	if cfg.ReconnectMaxDelay != DefaultReconnectMaxDelay {
		t.Errorf("expected max delay %v, got %v", DefaultReconnectMaxDelay, cfg.ReconnectMaxDelay)
	}
	if cfg.ConnectTimeout != DefaultConnectTimeout {
		t.Errorf("expected connect timeout %v, got %v", DefaultConnectTimeout, cfg.ConnectTimeout)
	}
}

func TestApplyDefaults_PreservesNonZeroValues(t *testing.T) {
	cfg := Config{
		Address:           "localhost:12345",
		ChannelBuffer:     500,
		ReconnectMinDelay: time.Second,
		ReconnectMaxDelay: time.Minute,
		ConnectTimeout:    10 * time.Second,
	}

	cfg = ApplyDefaults(cfg)

	if cfg.ChannelBuffer != 500 {
		t.Errorf("expected channel buffer 500, got %d", cfg.ChannelBuffer)
	}
	if cfg.ReconnectMinDelay != time.Second {
		t.Errorf("expected min delay 1s, got %v", cfg.ReconnectMinDelay)
	}
	if cfg.ReconnectMaxDelay != time.Minute {
		t.Errorf("expected max delay 1m, got %v", cfg.ReconnectMaxDelay)
	}
	if cfg.ConnectTimeout != 10*time.Second {
		t.Errorf("expected connect timeout 10s, got %v", cfg.ConnectTimeout)
	}
}

func TestTransportString(t *testing.T) {
	if TransportTCP.String() != "tcp" {
		t.Errorf("expected 'tcp', got %s", TransportTCP.String())
	}
	if TransportUDP.String() != "udp" {
		t.Errorf("expected 'udp', got %s", TransportUDP.String())
	}
}

func TestProtocolString(t *testing.T) {
	if ProtocolAuto.String() != "auto" {
		t.Errorf("expected 'auto', got %s", ProtocolAuto.String())
	}
	if ProtocolCSV.String() != "csv" {
		t.Errorf("expected 'csv', got %s", ProtocolCSV.String())
	}
	if ProtocolBinary.String() != "binary" {
		t.Errorf("expected 'binary', got %s", ProtocolBinary.String())
	}
}

func TestConfigHelpers(t *testing.T) {
	cfg := Default("localhost:1234")

	if !cfg.IsTCP() {
		t.Error("default should be TCP")
	}
	if cfg.IsUDP() {
		t.Error("default should not be UDP")
	}

	cfg.Transport = TransportUDP
	if cfg.IsTCP() {
		t.Error("should not be TCP after setting UDP")
	}
	if !cfg.IsUDP() {
		t.Error("should be UDP after setting UDP")
	}
}
