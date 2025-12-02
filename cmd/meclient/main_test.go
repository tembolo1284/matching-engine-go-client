// Full path: cmd/meclient/main_test.go

package main

import (
	"testing"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient"
)

func TestParseOrderCommand_Valid(t *testing.T) {
	var orderID uint32 = 100

	tests := []struct {
		name     string
		parts    []string
		side     meclient.Side
		wantSym  string
		wantQty  uint32
		wantPx   uint32
	}{
		{
			name:    "basic buy",
			parts:   []string{"buy", "IBM", "100", "150"},
			side:    meclient.SideBuy,
			wantSym: "IBM",
			wantQty: 100,
			wantPx:  150,
		},
		{
			name:    "sell",
			parts:   []string{"sell", "AAPL", "50", "200"},
			side:    meclient.SideSell,
			wantSym: "AAPL",
			wantQty: 50,
			wantPx:  200,
		},
		{
			name:    "lowercase symbol uppercased",
			parts:   []string{"buy", "goog", "10", "100"},
			side:    meclient.SideBuy,
			wantSym: "GOOG",
			wantQty: 10,
			wantPx:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid := orderID
			order, err := parseOrderCommand(tt.parts, tt.side, 1, &oid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if order.Symbol != tt.wantSym {
				t.Errorf("symbol: got %s, want %s", order.Symbol, tt.wantSym)
			}
			if order.Qty != tt.wantQty {
				t.Errorf("qty: got %d, want %d", order.Qty, tt.wantQty)
			}
			if order.Price != tt.wantPx {
				t.Errorf("price: got %d, want %d", order.Price, tt.wantPx)
			}
		})
	}
}

func TestParseOrderCommand_Invalid(t *testing.T) {
	var orderID uint32 = 100

	tests := []struct {
		name  string
		parts []string
	}{
		{"too few args", []string{"buy", "IBM"}},
		{"invalid qty", []string{"buy", "IBM", "abc", "100"}},
		{"invalid price", []string{"buy", "IBM", "100", "abc"}},
		{"invalid user", []string{"buy", "IBM", "100", "150", "abc"}},
		{"symbol too long", []string{"buy", "VERYLONGSYMBOLNAME123", "100", "150"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid := orderID
			_, err := parseOrderCommand(tt.parts, meclient.SideBuy, 1, &oid)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestParseCancelCommand_Valid(t *testing.T) {
	cancel, err := parseCancelCommand([]string{"cancel", "1001"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancel.OrderID != 1001 {
		t.Errorf("order_id: got %d, want 1001", cancel.OrderID)
	}
	if cancel.UserID != 1 {
		t.Errorf("user_id: got %d, want 1", cancel.UserID)
	}
}

func TestParseCancelCommand_WithUser(t *testing.T) {
	cancel, err := parseCancelCommand([]string{"cancel", "1001", "5"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancel.UserID != 5 {
		t.Errorf("user_id: got %d, want 5", cancel.UserID)
	}
}

func TestParseCancelCommand_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
	}{
		{"too few args", []string{"cancel"}},
		{"invalid order_id", []string{"cancel", "abc"}},
		{"invalid user_id", []string{"cancel", "1001", "abc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCancelCommand(tt.parts, 1)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}
