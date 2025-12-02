// Full path: pkg/meclient/protocol/encoder_test.go

package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncodeNewOrder_Buy(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1001,
	}

	if err := enc.EncodeNewOrder(order); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Check length prefix
	data := buf.Bytes()
	if len(data) < 4 {
		t.Fatalf("expected at least 4 bytes, got %d", len(data))
	}

	length := binary.BigEndian.Uint32(data[:4])
	payload := string(data[4 : 4+length])
	expected := "N,1,IBM,100,50,B,1001\n"

	if payload != expected {
		t.Errorf("expected %q, got %q", expected, payload)
	}
}

func TestEncodeNewOrder_Sell(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	order := &NewOrder{
		UserID:  2,
		Symbol:  "AAPL",
		Price:   150,
		Qty:     25,
		Side:    SideSell,
		OrderID: 2002,
	}

	if err := enc.EncodeNewOrder(order); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	data := buf.Bytes()
	length := binary.BigEndian.Uint32(data[:4])
	payload := string(data[4 : 4+length])
	expected := "N,2,AAPL,150,25,S,2002\n"

	if payload != expected {
		t.Errorf("expected %q, got %q", expected, payload)
	}
}

func TestEncodeCancel(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	if err := enc.EncodeCancel(cancel); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	data := buf.Bytes()
	length := binary.BigEndian.Uint32(data[:4])
	payload := string(data[4 : 4+length])
	expected := "C,1,1001\n"

	if payload != expected {
		t.Errorf("expected %q, got %q", expected, payload)
	}
}

func TestEncodeFlush(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	if err := enc.EncodeFlush(); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	data := buf.Bytes()
	length := binary.BigEndian.Uint32(data[:4])
	payload := string(data[4 : 4+length])
	expected := "F\n"

	if payload != expected {
		t.Errorf("expected %q, got %q", expected, payload)
	}
}

func TestEncodeMultipleMessages(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	order1 := &NewOrder{UserID: 1, Symbol: "IBM", Price: 100, Qty: 50, Side: SideBuy, OrderID: 1}
	order2 := &NewOrder{UserID: 1, Symbol: "IBM", Price: 100, Qty: 50, Side: SideSell, OrderID: 2}

	if err := enc.EncodeNewOrder(order1); err != nil {
		t.Fatalf("encode order1 error: %v", err)
	}
	if err := enc.EncodeNewOrder(order2); err != nil {
		t.Fatalf("encode order2 error: %v", err)
	}

	// Should have two framed messages
	data := buf.Bytes()
	if len(data) < 8 {
		t.Fatalf("expected at least 8 bytes for two messages")
	}

	// First message
	len1 := binary.BigEndian.Uint32(data[:4])
	msg1 := string(data[4 : 4+len1])
	if msg1 != "N,1,IBM,100,50,B,1\n" {
		t.Errorf("msg1: expected %q, got %q", "N,1,IBM,100,50,B,1\n", msg1)
	}

	// Second message
	offset := 4 + len1
	len2 := binary.BigEndian.Uint32(data[offset : offset+4])
	msg2 := string(data[offset+4 : offset+4+len2])
	if msg2 != "N,1,IBM,100,50,S,2\n" {
		t.Errorf("msg2: expected %q, got %q", "N,1,IBM,100,50,S,2\n", msg2)
	}
}

func TestEncodeLargeSymbol(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	order := &NewOrder{
		UserID:  1,
		Symbol:  "VERYLONGSYMBOL",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	if err := enc.EncodeNewOrder(order); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	data := buf.Bytes()
	length := binary.BigEndian.Uint32(data[:4])
	payload := string(data[4 : 4+length])

	if !bytes.Contains([]byte(payload), []byte("VERYLONGSYMBOL")) {
		t.Errorf("payload should contain symbol: %s", payload)
	}
}

func TestEncodeLargeValues(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	order := &NewOrder{
		UserID:  4294967295, // max uint32
		Symbol:  "TEST",
		Price:   4294967295,
		Qty:     4294967295,
		Side:    SideBuy,
		OrderID: 4294967295,
	}

	if err := enc.EncodeNewOrder(order); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Just verify it encodes without error
	if buf.Len() == 0 {
		t.Error("expected non-empty buffer")
	}
}
