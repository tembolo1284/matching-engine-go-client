package meclient

import (
	"errors"
	"testing"
	"time"
)

func TestClient_New_Valid(t *testing.T) {
	cfg := DefaultConfig("localhost:12345")
	client, err := New(cfg)

	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClient_New_InvalidConfig(t *testing.T) {
	cfg := Config{Address: ""} // Invalid - empty address
	_, err := New(cfg)

	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got: %v", err)
	}
}

func TestClient_ConnectAndClose(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	cfg.AutoReconnect = false
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("expected client to be disconnected after Close")
	}
}

func TestClient_SendOrder(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	cfg.AutoReconnect = false
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	order := NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	if err := client.SendOrder(order); err != nil {
		t.Fatalf("SendOrder failed: %v", err)
	}

	if !server.waitForReceived(1, time.Second) {
		t.Fatal("server did not receive message")
	}

	received := server.getReceived()
	expected := "N, 1, IBM, 150, 100, B, 1001"
	if received[0] != expected {
		t.Errorf("expected %q, got %q", expected, received[0])
	}
}

func TestClient_SendOrder_Invalid(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	// Empty symbol
	order := NewOrder{Symbol: "", Qty: 100, Side: SideBuy}
	err = client.SendOrder(order)
	if !errors.Is(err, ErrEmptySymbol) {
		t.Errorf("expected ErrEmptySymbol, got: %v", err)
	}

	// Zero quantity
	order = NewOrder{Symbol: "IBM", Qty: 0, Side: SideBuy}
	err = client.SendOrder(order)
	if !errors.Is(err, ErrZeroQuantity) {
		t.Errorf("expected ErrZeroQuantity, got: %v", err)
	}
}

func TestClient_SendCancel(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	cfg.AutoReconnect = false
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	cancel := CancelOrder{
		Symbol:  "IBM",
		UserID:  1,
		OrderID: 1001,
	}

	if err := client.SendCancel(cancel); err != nil {
		t.Fatalf("SendCancel failed: %v", err)
	}

	if !server.waitForReceived(1, time.Second) {
		t.Fatal("server did not receive message")
	}

	received := server.getReceived()
	expected := "C, IBM, 1, 1001"
	if received[0] != expected {
		t.Errorf("expected %q, got %q", expected, received[0])
	}
}

func TestClient_SendFlush(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	cfg.AutoReconnect = false
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	if err := client.SendFlush(); err != nil {
		t.Fatalf("SendFlush failed: %v", err)
	}

	if !server.waitForReceived(1, time.Second) {
		t.Fatal("server did not receive message")
	}

	received := server.getReceived()
	if received[0] != "F" {
		t.Errorf("expected F, got %q", received[0])
	}
}

func TestClient_MultipleMessages(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	cfg.AutoReconnect = false
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	for i := uint32(1); i <= 5; i++ {
		order := NewOrder{
			UserID:  1,
			Symbol:  "TEST",
			Price:   100 + i,
			Qty:     10 * i,
			Side:    SideBuy,
			OrderID: 1000 + i,
		}
		if err := client.SendOrder(order); err != nil {
			t.Fatalf("SendOrder %d failed: %v", i, err)
		}
	}

	if !server.waitForReceived(5, time.Second) {
		t.Fatal("server did not receive all messages")
	}

	received := server.getReceived()
	if len(received) != 5 {
		t.Errorf("expected 5 messages, got %d", len(received))
	}
}

func TestClient_SendAfterClose(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	cfg.AutoReconnect = false
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	client.Close()

	err = client.SendOrder(NewOrder{
		UserID:  1,
		Symbol:  "TEST",
		Price:   100,
		Qty:     10,
		Side:    SideBuy,
		OrderID: 1,
	})

	if !errors.Is(err, ErrClientClosed) {
		t.Errorf("expected ErrClientClosed, got: %v", err)
	}
}

func TestClient_ConnectFailure(t *testing.T) {
	cfg := DefaultConfig("127.0.0.1:1") // Port 1 should be unavailable
	cfg.ConnectTimeout = 100 * time.Millisecond
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = client.Connect()
	if err == nil {
		client.Close()
		t.Fatal("expected connection to fail")
	}
}

func TestClient_Stats(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	cfg := DefaultConfig(server.addr())
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	// Send some messages
	for i := uint32(0); i < 3; i++ {
		client.SendOrder(NewOrder{
			Symbol:  "TEST",
			Qty:     10,
			Side:    SideBuy,
			OrderID: i,
		})
	}

	time.Sleep(50 * time.Millisecond)

	stats := client.Stats()
	if stats.MessagesSent != 3 {
		t.Errorf("expected 3 messages sent, got %d", stats.MessagesSent)
	}
}
