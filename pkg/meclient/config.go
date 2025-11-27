package meclient

import (
	"errors"
	"fmt"
	"time"
)

// Configuration constants
const (
	DefaultChannelBuffer = 1024
	MaxChannelBuffer     = 65536

	DefaultWriteBuffer = 64 * 1024
	DefaultReadBuffer  = 64 * 1024

	DefaultReconnectMinDelay = 100 * time.Millisecond
	DefaultReconnectMaxDelay = 30 * time.Second
	DefaultConnectTimeout    = 5 * time.Second

	// Safety bounds
	MaxReconnectAttempts   = 1000
	MaxMessageBatchSize    = 1000
	MaxConsecutiveErrors   = 100
	ReconnectCheckInterval = 50 * time.Millisecond

	// Protocol constraints
	MaxSymbolLength = 16
)

// Sentinel errors
var (
	ErrClientClosed   = errors.New("client closed")
	ErrNotConnected   = errors.New("not connected")
	ErrWriteQueueFull = errors.New("write queue full")
	ErrChannelFull    = errors.New("channel full, message dropped")
	ErrInvalidConfig  = errors.New("invalid configuration")
	ErrEmptySymbol    = errors.New("symbol cannot be empty")
	ErrSymbolTooLong  = errors.New("symbol exceeds maximum length")
	ErrZeroQuantity   = errors.New("quantity must be greater than zero")
	ErrInvalidSide    = errors.New("invalid order side")
	ErrMaxReconnects  = errors.New("maximum reconnection attempts exceeded")
)

// Config holds client configuration options.
type Config struct {
	Address           string
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

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(address string) Config {
	return Config{
		Address:           address,
		ChannelBuffer:     DefaultChannelBuffer,
		ReconnectMinDelay: DefaultReconnectMinDelay,
		ReconnectMaxDelay: DefaultReconnectMaxDelay,
		ConnectTimeout:    DefaultConnectTimeout,
		AutoReconnect:     true,
	}
}

func applyDefaults(cfg Config) Config {
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
