package meclient

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

// mockServer is a simple TCP server for testing.
type mockServer struct {
	listener net.Listener
	conns    []net.Conn
	connsMu  sync.Mutex
	received []string
	recvMu   sync.Mutex
	wg       sync.WaitGroup
}

func newMockServer(t *testing.T) *mockServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}

	s := &mockServer{
		listener: listener,
		received: make([]string, 0),
	}

	s.wg.Add(1)
	go s.acceptLoop()

	return s
}

func (s *mockServer) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		s.connsMu.Lock()
		s.conns = append(s.conns, conn)
		s.connsMu.Unlock()

		go s.handleConn(conn)
	}
}

func (s *mockServer) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		s.recvMu.Lock()
		s.received = append(s.received, line)
		s.recvMu.Unlock()
	}
}

func (s *mockServer) addr() string {
	return s.listener.Addr().String()
}

func (s *mockServer) close() {
	s.listener.Close()

	s.connsMu.Lock()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.connsMu.Unlock()

	s.wg.Wait()
}

func (s *mockServer) getReceived() []string {
	s.recvMu.Lock()
	defer s.recvMu.Unlock()
	result := make([]string, len(s.received))
	copy(result, s.received)
	return result
}

func (s *mockServer) clearReceived() {
	s.recvMu.Lock()
	s.received = s.received[:0]
	s.recvMu.Unlock()
}

func (s *mockServer) sendToAll(msg string) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()

	for _, conn := range s.conns {
		fmt.Fprintf(conn, "%s\n", msg)
	}
}

func (s *mockServer) sendToConn(idx int, msg string) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()

	if idx < len(s.conns) {
		fmt.Fprintf(s.conns[idx], "%s\n", msg)
	}
}

func (s *mockServer) waitForReceived(n int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s.recvMu.Lock()
		count := len(s.received)
		s.recvMu.Unlock()
		if count >= n {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func (s *mockServer) connCount() int {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	return len(s.conns)
}
