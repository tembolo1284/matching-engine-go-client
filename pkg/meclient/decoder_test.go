package meclient

import (
	"strings"
	"testing"
)

func TestDecoder_Ack(t *testing.T) {
	input := "A, 1, 1001\n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.Ack == nil {
		t.Fatal("expected Ack message")
	}

	if msg.Ack.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", msg.Ack.UserID)
	}

	if msg.Ack.OrderID != 1001 {
		t.Errorf("expected OrderID 1001, got %d", msg.Ack.OrderID)
	}
}

func TestDecoder_Trade(t *testing.T) {
	input := "T, 1, 1001, 2, 2001, 150, 100\n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.Trade == nil {
		t.Fatal("expected Trade message")
	}

	trade := msg.Trade
	if trade.BuyUserID != 1 {
		t.Errorf("expected BuyUserID 1, got %d", trade.BuyUserID)
	}
	if trade.BuyOrderID != 1001 {
		t.Errorf("expected BuyOrderID 1001, got %d", trade.BuyOrderID)
	}
	if trade.SellUserID != 2 {
		t.Errorf("expected SellUserID 2, got %d", trade.SellUserID)
	}
	if trade.SellOrderID != 2001 {
		t.Errorf("expected SellOrderID 2001, got %d", trade.SellOrderID)
	}
	if trade.Price != 150 {
		t.Errorf("expected Price 150, got %d", trade.Price)
	}
	if trade.Qty != 100 {
		t.Errorf("expected Qty 100, got %d", trade.Qty)
	}
}

func TestDecoder_BookUpdate_Buy(t *testing.T) {
	input := "B, IBM, B, 150, 500\n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.BookUpdate == nil {
		t.Fatal("expected BookUpdate message")
	}

	update := msg.BookUpdate
	if update.Symbol != "IBM" {
		t.Errorf("expected Symbol IBM, got %s", update.Symbol)
	}
	if update.Side != SideBuy {
		t.Errorf("expected Side BUY, got %c", update.Side)
	}
	if update.Price != 150 {
		t.Errorf("expected Price 150, got %d", update.Price)
	}
	if update.Qty != 500 {
		t.Errorf("expected Qty 500, got %d", update.Qty)
	}
}

func TestDecoder_BookUpdate_Sell(t *testing.T) {
	input := "B, AAPL, S, 175, 1000\n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.BookUpdate == nil {
		t.Fatal("expected BookUpdate message")
	}

	if msg.BookUpdate.Side != SideSell {
		t.Errorf("expected Side SELL, got %c", msg.BookUpdate.Side)
	}
}

func TestDecoder_CancelAck(t *testing.T) {
	input := "X, 1, 1001\n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.CancelAck == nil {
		t.Fatal("expected CancelAck message")
	}

	if msg.CancelAck.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", msg.CancelAck.UserID)
	}

	if msg.CancelAck.OrderID != 1001 {
		t.Errorf("expected OrderID 1001, got %d", msg.CancelAck.OrderID)
	}
}

func TestDecoder_MultipleMessages(t *testing.T) {
	input := `A, 1, 1001
T, 1, 1001, 2, 2001, 150, 100
B, IBM, S, 151, 200
X, 2, 2001
`
	dec := newDecoder(strings.NewReader(input))

	// Message 1: Ack
	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode 1 failed: %v", err)
	}
	if msg.Ack == nil {
		t.Error("expected Ack for message 1")
	}

	// Message 2: Trade
	msg, err = dec.decode()
	if err != nil {
		t.Fatalf("decode 2 failed: %v", err)
	}
	if msg.Trade == nil {
		t.Error("expected Trade for message 2")
	}

	// Message 3: BookUpdate
	msg, err = dec.decode()
	if err != nil {
		t.Fatalf("decode 3 failed: %v", err)
	}
	if msg.BookUpdate == nil {
		t.Error("expected BookUpdate for message 3")
	}

	// Message 4: CancelAck
	msg, err = dec.decode()
	if err != nil {
		t.Fatalf("decode 4 failed: %v", err)
	}
	if msg.CancelAck == nil {
		t.Error("expected CancelAck for message 4")
	}
}

func TestDecoder_SkipsEmptyLines(t *testing.T) {
	input := "\n\nA, 1, 1001\n\n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.Ack == nil {
		t.Error("expected Ack message")
	}
}

func TestDecoder_WhitespaceHandling(t *testing.T) {
	input := "A,  1  ,  1001  \n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.Ack == nil {
		t.Fatal("expected Ack message")
	}

	if msg.Ack.UserID != 1 || msg.Ack.OrderID != 1001 {
		t.Errorf("whitespace not handled correctly: got user=%d, order=%d",
			msg.Ack.UserID, msg.Ack.OrderID)
	}
}

func TestDecoder_LargeValues(t *testing.T) {
	input := "A, 4294967295, 4294967295\n"
	dec := newDecoder(strings.NewReader(input))

	msg, err := dec.decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if msg.Ack.UserID != 4294967295 {
		t.Errorf("expected max uint32, got %d", msg.Ack.UserID)
	}
}
