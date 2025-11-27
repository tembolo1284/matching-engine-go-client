package meclient

import "testing"

func TestSide_String(t *testing.T) {
	tests := []struct {
		side     Side
		expected string
	}{
		{SideBuy, "BUY"},
		{SideSell, "SELL"},
		{Side('X'), "UNKNOWN"},
		{Side(0), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.side.String(); got != tt.expected {
			t.Errorf("Side(%c).String() = %s, want %s", tt.side, got, tt.expected)
		}
	}
}

func TestSide_Constants(t *testing.T) {
	// Verify the byte values match protocol
	if SideBuy != Side('B') {
		t.Errorf("SideBuy should be 'B', got %c", SideBuy)
	}

	if SideSell != Side('S') {
		t.Errorf("SideSell should be 'S', got %c", SideSell)
	}
}

func TestNewOrder_Fields(t *testing.T) {
	order := NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	if order.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", order.UserID)
	}

	if order.Symbol != "IBM" {
		t.Errorf("expected Symbol IBM, got %s", order.Symbol)
	}

	if order.Price != 150 {
		t.Errorf("expected Price 150, got %d", order.Price)
	}

	if order.Qty != 100 {
		t.Errorf("expected Qty 100, got %d", order.Qty)
	}

	if order.Side != SideBuy {
		t.Errorf("expected Side BUY, got %c", order.Side)
	}

	if order.OrderID != 1001 {
		t.Errorf("expected OrderID 1001, got %d", order.OrderID)
	}
}

func TestTrade_Fields(t *testing.T) {
	trade := Trade{
		Symbol:      "IBM",
		BuyUserID:   1,
		BuyOrderID:  1001,
		SellUserID:  2,
		SellOrderID: 2001,
		Price:       150,
		Qty:         100,
	}

	if trade.Symbol != "IBM" {
		t.Errorf("expected Symbol IBM, got %s", trade.Symbol)
	}

	if trade.BuyUserID != 1 {
		t.Errorf("expected BuyUserID 1, got %d", trade.BuyUserID)
	}

	if trade.SellUserID != 2 {
		t.Errorf("expected SellUserID 2, got %d", trade.SellUserID)
	}

	if trade.Price != 150 {
		t.Errorf("expected Price 150, got %d", trade.Price)
	}

	if trade.Qty != 100 {
		t.Errorf("expected Qty 100, got %d", trade.Qty)
	}
}

func TestBookUpdate_Fields(t *testing.T) {
	update := BookUpdate{
		Symbol: "AAPL",
		Side:   SideSell,
		Price:  175,
		Qty:    500,
	}

	if update.Symbol != "AAPL" {
		t.Errorf("expected Symbol AAPL, got %s", update.Symbol)
	}

	if update.Side != SideSell {
		t.Errorf("expected Side SELL, got %c", update.Side)
	}

	if update.Price != 175 {
		t.Errorf("expected Price 175, got %d", update.Price)
	}

	if update.Qty != 500 {
		t.Errorf("expected Qty 500, got %d", update.Qty)
	}
}

func TestReconnectEvent_Fields(t *testing.T) {
	event := ReconnectEvent{Attempt: 3}

	if event.Attempt != 3 {
		t.Errorf("expected Attempt 3, got %d", event.Attempt)
	}
}
