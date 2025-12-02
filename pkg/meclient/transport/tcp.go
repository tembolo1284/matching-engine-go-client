// Full path: pkg/meclient/transport/tcp.go

package transport

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/config"
)

// TCP implements Transport over TCP.
type TCP struct {
	cfg *config.Config

	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	mu     sync.RWMutex

	connected bool
}

// NewTCP creates a new TCP transport.
func NewTCP(cfg *config.Config) *TCP {
	return &TCP{
		cfg: cfg,
	}
}

// Connect establishes a TCP connection.
func (t *TCP) Connect() error {
	dialer := net.Dialer{Timeout: t.cfg.ConnectTimeout}

	conn, err := dialer.Dial("tcp", t.cfg.Address)
	if err != nil {
		return fmt.Errorf("tcp connect: %w", err)
	}

	t.mu.Lock()
	t.conn = conn
	t.reader = bufio.NewReaderSize(conn, config.DefaultReadBuffer)
	t.writer = bufio.NewWriterSize(conn, config.DefaultWriteBuffer)
	t.connected = true
	t.mu.Unlock()

	return nil
}

// Close closes the TCP connection.
func (t *TCP) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.connected = false

	if t.conn != nil {
		err := t.conn.Close()
		t.conn = nil
		t.reader = nil
		t.writer = nil
		return err
	}
	return nil
}

// Reader returns the underlying reader.
func (t *TCP) Reader() io.Reader {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.conn == nil {
		return nil
	}
	// Return raw conn for decoder (it needs unbuffered for framing)
	return t.conn
}

// Writer returns the underlying buffered writer.
func (t *TCP) Writer() io.Writer {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.writer
}

// Flush flushes the write buffer.
func (t *TCP) Flush() error {
	t.mu.RLock()
	writer := t.writer
	t.mu.RUnlock()

	if writer == nil {
		return fmt.Errorf("not connected")
	}
	return writer.Flush()
}

// IsConnected returns true if connected.
func (t *TCP) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// SetDeadline sets read/write deadline.
func (t *TCP) SetDeadline(deadline time.Time) error {
	t.mu.RLock()
	conn := t.conn
	t.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}
	return conn.SetDeadline(deadline)
}

// RemoteAddr returns the remote address.
func (t *TCP) RemoteAddr() string {
	t.mu.RLock()
	conn := t.conn
	t.mu.RUnlock()

	if conn == nil {
		return ""
	}
	return conn.RemoteAddr().String()
}
