package meclient

import (
	"strings"
	"testing"
)

func TestEncoder_NewOrder(t *testing.T) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	if err := enc.encodeNewOrder(order); err != nil {
		t.Fatalf("encodeNewOrder failed: %v", err)
	}

	expected := "N, 1, IBM, 150, 100, B, 1001\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestEncoder_NewOrder_Sell(t *testing.T) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	order := &NewOrder{
		UserID:  2,
		Symbol:  "AAPL",
		Price:   175,
		Qty:     50,
		Side:    SideSell,
		OrderID: 2001,
	}

	if err := enc.encodeNewOrder(order); err != nil {
		t.Fatalf("encodeNewOrder failed: %v", err)
	}

	expected := "N, 2, AAPL, 175, 50, S, 2001\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestEncoder_MarketOrder(t *testing.T) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	order := &NewOrder{
		UserID:  2,
		Symbol:  "AAPL",
		Price:   0, // Market order
		Qty:     50,
		Side:    SideSell,
		OrderID: 2002,
	}

	if err := enc.encodeNewOrder(order); err != nil {
		t.Fatalf("encodeNewOrder failed: %v", err)
	}

	expected := "N, 2, AAPL, 0, 50, S, 2002\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestEncoder_Cancel(t *testing.T) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	cancel := &CancelOrder{
		Symbol:  "IBM",
		UserID:  1,
		OrderID: 1001,
	}

	if err := enc.encodeCancel(cancel); err != nil {
		t.Fatalf("encodeCancel failed: %v", err)
	}

	expected := "C, IBM, 1, 1001\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestEncoder_Flush(t *testing.T) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	if err := enc.encodeFlush(); err != nil {
		t.Fatalf("encodeFlush failed: %v", err)
	}

	expected := "F\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestEncoder_LargeValues(t *testing.T) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	order := &NewOrder{
		UserID:  4294967295, // Max uint32
		Symbol:  "VERYLONGSYMBOL",
		Price:   4294967295,
		Qty:     4294967295,
		Side:    SideBuy,
		OrderID: 4294967295,
	}

	if err := enc.encodeNewOrder(order); err != nil {
		t.Fatalf("encodeNewOrder failed: %v", err)
	}

	expected := "N, 4294967295, VERYLONGSYMBOL, 4294967295, 4294967295, B, 4294967295\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestEncoder_BufferReuse(t *testing.T) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	// Send multiple messages to verify buffer reuse
	for i := uint32(1); i <= 3; i++ {
		buf.Reset()
		order := &NewOrder{
			UserID:  i,
			Symbol:  "TEST",
			Price:   100 + i,
			Qty:     10,
			Side:    SideBuy,
			OrderID: 1000 + i,
		}
		if err := enc.encodeNewOrder(order); err != nil {
			t.Fatalf("encodeNewOrder %d failed: %v", i, err)
		}
	}

	// Last message should be correct
	expected := "N, 3, TEST, 103, 10, B, 1003\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}
