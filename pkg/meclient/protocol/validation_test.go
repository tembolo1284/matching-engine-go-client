// Full path: pkg/meclient/protocol/validation_test.go

package protocol

import (
	"testing"
)

func TestValidateOrder_ValidBuy(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	if err := ValidateOrder(order); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOrder_ValidSell(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "AAPL",
		Price:   150,
		Qty:     25,
		Side:    SideSell,
		OrderID: 2,
	}

	if err := ValidateOrder(order); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOrder_EmptySymbol(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	err := ValidateOrder(order)
	if err == nil {
		t.Error("expected error for empty symbol")
	}
	if err != ErrEmptySymbol {
		t.Errorf("expected ErrEmptySymbol, got %v", err)
	}
}

func TestValidateOrder_SymbolTooLong(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "VERYLONGSYMBOLNAME",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	err := ValidateOrder(order)
	if err == nil {
		t.Error("expected error for symbol too long")
	}
	if err != ErrSymbolTooLong {
		t.Errorf("expected ErrSymbolTooLong, got %v", err)
	}
}

func TestValidateOrder_ZeroQuantity(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     0,
		Side:    SideBuy,
		OrderID: 1,
	}

	err := ValidateOrder(order)
	if err == nil {
		t.Error("expected error for zero quantity")
	}
	if err != ErrZeroQuantity {
		t.Errorf("expected ErrZeroQuantity, got %v", err)
	}
}

func TestValidateOrder_InvalidSide(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    Side('X'),
		OrderID: 1,
	}

	err := ValidateOrder(order)
	if err == nil {
		t.Error("expected error for invalid side")
	}
	if err != ErrInvalidSide {
		t.Errorf("expected ErrInvalidSide, got %v", err)
	}
}

func TestValidateOrder_MaxSymbolLength(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "1234567890123456", // exactly 16 chars
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	err := ValidateOrder(order)
	if err != nil {
		t.Errorf("unexpected error for max length symbol: %v", err)
	}
}

func TestValidateCancel_Valid(t *testing.T) {
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	if err := ValidateCancel(cancel); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateCancel_ZeroOrderID(t *testing.T) {
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 0,
	}

	err := ValidateCancel(cancel)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateCancel_ZeroUserID(t *testing.T) {
	cancel := &CancelOrder{
		UserID:  0,
		OrderID: 1001,
	}

	err := ValidateCancel(cancel)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
