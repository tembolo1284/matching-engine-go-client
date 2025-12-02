// Full path: pkg/meclient/transport/tcp_test.go

package transport

import (
	"net"
	"testing"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/config"
)

func TestTCPConnect_InvalidAddress(t *testing.T) {
	cfg := &config.Config{
		Address:        "invalid:99999",
		ConnectTimeout: 100 * time.Millisecond,
	}

	tcp := NewTCP(cfg)
	err := tcp.Connect()
	if err == nil {
		tcp.Close()
		t.Error("expected error for invalid address")
	}
}

func TestTCPConnect_Timeout(t *testing.T) {
	cfg := &config.Config{
		Address:        "10.255.255.1:1234",
		ConnectTimeout: 100 * time.Millisecond,
	}

	tcp := NewTCP(cfg)
	start := time.Now()
	err := tcp.Connect()
	elapsed := time.Since(start)

	if err == nil {
		tcp.Close()
		t.Error("expected timeout error")
	}

	if elapsed > 500*time.Millisecond {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

func TestTCPIsConnected_NotConnected(t *testing.T) {
	cfg := &config.Config{
		Address:        "localhost:1234",
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)

	if tcp.IsConnected() {
		t.Error("should not be connected initially")
	}
}

func TestTCPClose_NotConnected(t *testing.T) {
	cfg := &config.Config{
		Address:        "localhost:1234",
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)

	if err := tcp.Close(); err != nil {
		t.Errorf("unexpected error closing unconnected transport: %v", err)
	}
}

func TestTCPReaderWriter_NotConnected(t *testing.T) {
	cfg := &config.Config{
		Address:        "localhost:1234",
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)

	// Reader should be nil when not connected
	if tcp.Reader() != nil {
		t.Error("reader should be nil when not connected")
	}

	// Writer behavior: returns the writer field which may be nil
	// This is implementation-dependent, just verify no panic
	_ = tcp.Writer() // Just ensure it doesn't panic
}

func TestTCPRemoteAddr_NotConnected(t *testing.T) {
	cfg := &config.Config{
		Address:        "localhost:1234",
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)

	if tcp.RemoteAddr() != "" {
		t.Error("remote addr should be empty when not connected")
	}
}

func TestTCPWithMockServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(time.Second)
	}()

	cfg := &config.Config{
		Address:        addr,
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)

	if err := tcp.Connect(); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer tcp.Close()

	if !tcp.IsConnected() {
		t.Error("should be connected")
	}

	if tcp.Reader() == nil {
		t.Error("reader should not be nil when connected")
	}
	if tcp.Writer() == nil {
		t.Error("writer should not be nil when connected")
	}
	if tcp.RemoteAddr() == "" {
		t.Error("remote addr should not be empty when connected")
	}
}

func TestTCPFlush(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()

	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			defer conn.Close()
			time.Sleep(time.Second)
		}
	}()

	cfg := &config.Config{
		Address:        addr,
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)
	if err := tcp.Connect(); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer tcp.Close()

	_, err = tcp.Writer().Write([]byte("test"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	if err := tcp.Flush(); err != nil {
		t.Errorf("flush error: %v", err)
	}
}

func TestTCPFlush_NotConnected(t *testing.T) {
	cfg := &config.Config{
		Address:        "localhost:1234",
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)

	err := tcp.Flush()
	if err == nil {
		t.Error("expected error flushing when not connected")
	}
}

func TestTCPSetDeadline_NotConnected(t *testing.T) {
	cfg := &config.Config{
		Address:        "localhost:1234",
		ConnectTimeout: time.Second,
	}

	tcp := NewTCP(cfg)

	err := tcp.SetDeadline(time.Now().Add(time.Second))
	if err == nil {
		t.Error("expected error setting deadline when not connected")
	}
}
