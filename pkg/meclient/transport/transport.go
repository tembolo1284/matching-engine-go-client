// Full path: pkg/meclient/transport/transport.go

// Package transport provides network transport implementations for the matching engine client.
package transport

import (
	"io"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/config"
)

// Transport is the interface for network communication.
type Transport interface {
	// Connect establishes a connection to the server.
	Connect() error

	// Close closes the connection.
	Close() error

	// Reader returns the underlying reader for decoding messages.
	Reader() io.Reader

	// Writer returns the underlying writer for encoding messages.
	Writer() io.Writer

	// IsConnected returns true if currently connected.
	IsConnected() bool

	// SetDeadline sets read/write deadline.
	SetDeadline(t time.Time) error

	// RemoteAddr returns the remote address string.
	RemoteAddr() string
}

// New creates a new transport based on the config.
func New(cfg *config.Config) Transport {
	switch cfg.Transport {
	case config.TransportUDP:
		return NewUDP(cfg)
	default:
		return NewTCP(cfg)
	}
}
