// Full path: pkg/meclient/protocol/decoder_test.go

package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// Helper to create a framed message
func frameMessage(msg string) []byte {
	payload := []byte(msg)
	buf := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(payload)))
	copy(buf[4:], payload)
	return buf
}

func TestDecodeAck(t *testing.T) {
	input := frameMessage("A, IBM, 1, 1001")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.Ack == nil {
		t.Fatal("expected Ack message")
	}

	ack := msg.Ack
	if ack.Symbol != "IBM" {
		t.Errorf("expected symbol IBM, got %s", ack.Symbol)
	}
	if ack.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", ack.UserID)
	}
	if ack.OrderID != 1001 {
		t.Errorf("expected order_id 1001, got %d", ack.OrderID)
	}
}

func TestDecodeTrade(t *testing.T) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.Trade == nil {
		t.Fatal("expected Trade message")
	}

	trade := msg.Trade
	if trade.Symbol != "IBM" {
		t.Errorf("expected symbol IBM, got %s", trade.Symbol)
	}
	if trade.BuyUserID != 1 {
		t.Errorf("expected buy_user_id 1, got %d", trade.BuyUserID)
	}
	if trade.BuyOrderID != 1001 {
		t.Errorf("expected buy_order_id 1001, got %d", trade.BuyOrderID)
	}
	if trade.SellUserID != 2 {
		t.Errorf("expected sell_user_id 2, got %d", trade.SellUserID)
	}
	if trade.SellOrderID != 2001 {
		t.Errorf("expected sell_order_id 2001, got %d", trade.SellOrderID)
	}
	if trade.Price != 100 {
		t.Errorf("expected price 100, got %d", trade.Price)
	}
	if trade.Qty != 50 {
		t.Errorf("expected qty 50, got %d", trade.Qty)
	}
}

func TestDecodeBookUpdate(t *testing.T) {
	input := frameMessage("B, IBM, B, 100, 50")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.BookUpdate == nil {
		t.Fatal("expected BookUpdate message")
	}

	update := msg.BookUpdate
	if update.Symbol != "IBM" {
		t.Errorf("expected symbol IBM, got %s", update.Symbol)
	}
	if update.Side != SideBuy {
		t.Errorf("expected side BUY, got %v", update.Side)
	}
	if update.Price != 100 {
		t.Errorf("expected price 100, got %d", update.Price)
	}
	if update.Qty != 50 {
		t.Errorf("expected qty 50, got %d", update.Qty)
	}
}

func TestDecodeBookUpdateSell(t *testing.T) {
	input := frameMessage("B, IBM, S, 105, 25")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.BookUpdate == nil {
		t.Fatal("expected BookUpdate message")
	}

	if msg.BookUpdate.Side != SideSell {
		t.Errorf("expected side SELL, got %v", msg.BookUpdate.Side)
	}
}

func TestDecodeBookUpdateEmpty(t *testing.T) {
	input := frameMessage("B, IBM, B, -, -")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.BookUpdate == nil {
		t.Fatal("expected BookUpdate message")
	}

	update := msg.BookUpdate
	if update.Price != 0 {
		t.Errorf("expected price 0 for empty book, got %d", update.Price)
	}
	if update.Qty != 0 {
		t.Errorf("expected qty 0 for empty book, got %d", update.Qty)
	}
}

func TestDecodeCancelAck(t *testing.T) {
	input := frameMessage("C, IBM, 1, 1001")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.CancelAck == nil {
		t.Fatal("expected CancelAck message")
	}

	cancelAck := msg.CancelAck
	if cancelAck.Symbol != "IBM" {
		t.Errorf("expected symbol IBM, got %s", cancelAck.Symbol)
	}
	if cancelAck.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", cancelAck.UserID)
	}
	if cancelAck.OrderID != 1001 {
		t.Errorf("expected order_id 1001, got %d", cancelAck.OrderID)
	}
}

func TestDecodeMultipleMessages(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(frameMessage("A, IBM, 1, 1001"))
	buf.Write(frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50"))

	dec := NewDecoder(&buf)

	// First message - Ack
	msg1, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode msg1 error: %v", err)
	}
	if msg1.Ack == nil {
		t.Error("expected first message to be Ack")
	}

	// Second message - Trade
	msg2, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode msg2 error: %v", err)
	}
	if msg2.Trade == nil {
		t.Error("expected second message to be Trade")
	}
}

func TestDecodeNoSpaces(t *testing.T) {
	input := frameMessage("A,IBM,1,1001")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.Ack == nil {
		t.Fatal("expected Ack message")
	}
	if msg.Ack.Symbol != "IBM" {
		t.Errorf("expected symbol IBM, got %s", msg.Ack.Symbol)
	}
}

func TestDecodeLargeValues(t *testing.T) {
	input := frameMessage("A, TEST, 4294967295, 4294967295")
	dec := NewDecoder(bytes.NewReader(input))

	msg, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if msg.Ack == nil {
		t.Fatal("expected Ack message")
	}
	if msg.Ack.UserID != 4294967295 {
		t.Errorf("expected user_id max uint32, got %d", msg.Ack.UserID)
	}
	if msg.Ack.OrderID != 4294967295 {
		t.Errorf("expected order_id max uint32, got %d", msg.Ack.OrderID)
	}
}
