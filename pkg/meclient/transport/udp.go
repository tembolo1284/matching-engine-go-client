// Full path: pkg/meclient/transport/udp.go

package transport

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/config"
)

// UDP implements Transport over UDP.
// Note: UDP is connectionless, so Connect() resolves the address
// and Close() is a no-op for the underlying socket.
type UDP struct {
	cfg *config.Config

	conn   *net.UDPConn
	addr   *net.UDPAddr
	mu     sync.RWMutex
	active bool
}

// NewUDP creates a new UDP transport.
func NewUDP(cfg *config.Config) *UDP {
	return &UDP{
		cfg: cfg,
	}
}

// Connect resolves the address and creates a UDP socket.
func (u *UDP) Connect() error {
	addr, err := net.ResolveUDPAddr("udp", u.cfg.Address)
	if err != nil {
		return fmt.Errorf("udp resolve: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("udp dial: %w", err)
	}

	u.mu.Lock()
	u.conn = conn
	u.addr = addr
	u.active = true
	u.mu.Unlock()

	return nil
}

// Close closes the UDP socket.
func (u *UDP) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.active = false

	if u.conn != nil {
		err := u.conn.Close()
		u.conn = nil
		return err
	}
	return nil
}

// Reader returns the UDP connection as a reader.
func (u *UDP) Reader() io.Reader {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.conn
}

// Writer returns the UDP connection as a writer.
func (u *UDP) Writer() io.Writer {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.conn
}

// IsConnected returns true if active.
func (u *UDP) IsConnected() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.active
}

// SetDeadline sets read/write deadline.
func (u *UDP) SetDeadline(deadline time.Time) error {
	u.mu.RLock()
	conn := u.conn
	u.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}
	return conn.SetDeadline(deadline)
}

// RemoteAddr returns the remote address.
func (u *UDP) RemoteAddr() string {
	u.mu.RLock()
	addr := u.addr
	u.mu.RUnlock()

	if addr == nil {
		return ""
	}
	return addr.String()
}
