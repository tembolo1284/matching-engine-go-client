// Full path: pkg/meclient/config/config.go

// Package config provides configuration types for the matching engine client.
package config

import (
	"errors"
	"fmt"
	"time"
)

// Default configuration values
const (
	DefaultPort              = 1234
	DefaultChannelBuffer     = 1024
	MaxChannelBuffer         = 65536
	DefaultWriteBuffer       = 64 * 1024
	DefaultReadBuffer        = 64 * 1024
	DefaultReconnectMinDelay = 100 * time.Millisecond
	DefaultReconnectMaxDelay = 30 * time.Second
	DefaultConnectTimeout    = 5 * time.Second
)

// Safety bounds
const (
	MaxReconnectAttempts   = 1000
	MaxMessageBatchSize    = 1000
	MaxConsecutiveErrors   = 100
	ReconnectCheckInterval = 50 * time.Millisecond
	MaxSymbolLength        = 16
)

// Transport mode
type Transport int

const (
	TransportTCP Transport = iota // TCP (default)
	TransportUDP                  // UDP
)

func (t Transport) String() string {
	switch t {
	case TransportTCP:
		return "tcp"
	case TransportUDP:
		return "udp"
	default:
		return "unknown"
	}
}

// Protocol mode for message encoding
type Protocol int

const (
	ProtocolAuto   Protocol = iota // Auto-detect (TCP: probe, UDP: CSV)
	ProtocolCSV                    // CSV format
	ProtocolBinary                 // Binary format
)

func (p Protocol) String() string {
	switch p {
	case ProtocolAuto:
		return "auto"
	case ProtocolCSV:
		return "csv"
	case ProtocolBinary:
		return "binary"
	default:
		return "unknown"
	}
}

// Sentinel errors
var (
	ErrInvalidConfig = errors.New("invalid configuration")
)

// Config holds client configuration options.
type Config struct {
	Address           string
	Transport         Transport // TCP or UDP
	Protocol          Protocol  // CSV, Binary, or Auto
	ChannelBuffer     int
	ReconnectMinDelay time.Duration
	ReconnectMaxDelay time.Duration
	ConnectTimeout    time.Duration
	AutoReconnect     bool
}

// Validate checks configuration for validity.
func (c *Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("%w: address is empty", ErrInvalidConfig)
	}

	if c.ChannelBuffer <= 0 || c.ChannelBuffer > MaxChannelBuffer {
		return fmt.Errorf("%w: channel buffer must be 1-%d", ErrInvalidConfig, MaxChannelBuffer)
	}

	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("%w: connect timeout must be positive", ErrInvalidConfig)
	}

	if c.ReconnectMinDelay <= 0 || c.ReconnectMaxDelay <= 0 {
		return fmt.Errorf("%w: reconnect delays must be positive", ErrInvalidConfig)
	}

	if c.ReconnectMinDelay > c.ReconnectMaxDelay {
		return fmt.Errorf("%w: min delay cannot exceed max delay", ErrInvalidConfig)
	}

	return nil
}

// IsTCP returns true if using TCP transport.
func (c *Config) IsTCP() bool {
	return c.Transport == TransportTCP
}

// IsUDP returns true if using UDP transport.
func (c *Config) IsUDP() bool {
	return c.Transport == TransportUDP
}

// IsBinary returns true if using binary protocol.
func (c *Config) IsBinary() bool {
	return c.Protocol == ProtocolBinary
}

// IsCSV returns true if using CSV protocol.
func (c *Config) IsCSV() bool {
	return c.Protocol == ProtocolCSV
}

// Default returns a Config with sensible defaults.
func Default(address string) Config {
	return Config{
		Address:           address,
		Transport:         TransportTCP,
		Protocol:          ProtocolAuto,
		ChannelBuffer:     DefaultChannelBuffer,
		ReconnectMinDelay: DefaultReconnectMinDelay,
		ReconnectMaxDelay: DefaultReconnectMaxDelay,
		ConnectTimeout:    DefaultConnectTimeout,
		AutoReconnect:     true,
	}
}

// ApplyDefaults fills in zero values with defaults.
func ApplyDefaults(cfg Config) Config {
	if cfg.ChannelBuffer <= 0 {
		cfg.ChannelBuffer = DefaultChannelBuffer
	}
	if cfg.ReconnectMinDelay <= 0 {
		cfg.ReconnectMinDelay = DefaultReconnectMinDelay
	}
	if cfg.ReconnectMaxDelay <= 0 {
		cfg.ReconnectMaxDelay = DefaultReconnectMaxDelay
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = DefaultConnectTimeout
	}
	return cfg
}
