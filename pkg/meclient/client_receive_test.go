package meclient

import (
	"testing"
	"time"
)

func TestClient_ReceiveAck(t *testing.T) {
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

	time.Sleep(50 * time.Millisecond)

	server.sendToConn(0, "A, 1, 1001")

	select {
	case ack := <-client.Acks():
		if ack.UserID != 1 || ack.OrderID != 1001 {
			t.Errorf("unexpected ack: %+v", ack)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for ack")
	}
}

func TestClient_ReceiveTrade(t *testing.T) {
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

	time.Sleep(50 * time.Millisecond)

	server.sendToConn(0, "T, 1, 1001, 2, 2001, 150, 100")

	select {
	case trade := <-client.Trades():
		if trade.BuyUserID != 1 || trade.BuyOrderID != 1001 {
			t.Errorf("unexpected trade buy side: %+v", trade)
		}
		if trade.SellUserID != 2 || trade.SellOrderID != 2001 {
			t.Errorf("unexpected trade sell side: %+v", trade)
		}
		if trade.Price != 150 || trade.Qty != 100 {
			t.Errorf("unexpected trade price/qty: %+v", trade)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for trade")
	}
}

func TestClient_ReceiveBookUpdate(t *testing.T) {
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

	time.Sleep(50 * time.Millisecond)

	server.sendToConn(0, "B, AAPL, S, 175, 1000")

	select {
	case update := <-client.BookUpdates():
		if update.Symbol != "AAPL" {
			t.Errorf("unexpected symbol: %s", update.Symbol)
		}
		if update.Side != SideSell {
			t.Errorf("unexpected side: %c", update.Side)
		}
		if update.Price != 175 || update.Qty != 1000 {
			t.Errorf("unexpected price/qty: %+v", update)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for book update")
	}
}

func TestClient_ReceiveCancelAck(t *testing.T) {
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

	time.Sleep(50 * time.Millisecond)

	server.sendToConn(0, "X, 3, 3001")

	select {
	case cancelAck := <-client.CancelAcks():
		if cancelAck.UserID != 3 || cancelAck.OrderID != 3001 {
			t.Errorf("unexpected cancel ack: %+v", cancelAck)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for cancel ack")
	}
}

func TestClient_ReceiveMultipleMessages(t *testing.T) {
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

	time.Sleep(50 * time.Millisecond)

	// Send multiple messages
	server.sendToConn(0, "A, 1, 1001")
	server.sendToConn(0, "A, 1, 1002")
	server.sendToConn(0, "A, 1, 1003")

	received := 0
	timeout := time.After(time.Second)

	for received < 3 {
		select {
		case <-client.Acks():
			received++
		case <-timeout:
			t.Fatalf("timeout: only received %d of 3 acks", received)
		}
	}
}

func TestClient_ReceiveMixedMessages(t *testing.T) {
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

	time.Sleep(50 * time.Millisecond)

	server.sendToConn(0, "A, 1, 1001")
	server.sendToConn(0, "T, 1, 1001, 2, 2001, 150, 100")
	server.sendToConn(0, "B, IBM, B, 149, 500")
	server.sendToConn(0, "X, 1, 1002")

	timeout := time.After(time.Second)

	// Receive ack
	select {
	case <-client.Acks():
	case <-timeout:
		t.Fatal("timeout waiting for ack")
	}

	// Receive trade
	select {
	case <-client.Trades():
	case <-timeout:
		t.Fatal("timeout waiting for trade")
	}

	// Receive book update
	select {
	case <-client.BookUpdates():
	case <-timeout:
		t.Fatal("timeout waiting for book update")
	}

	// Receive cancel ack
	select {
	case <-client.CancelAcks():
	case <-timeout:
		t.Fatal("timeout waiting for cancel ack")
	}
}
