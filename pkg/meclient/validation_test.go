package meclient

import (
	"errors"
	"testing"
)

func TestValidateOrder_Valid(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	if err := validateOrder(order); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateOrder_MarketOrder(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   0, // Market order
		Qty:     100,
		Side:    SideSell,
		OrderID: 1001,
	}

	if err := validateOrder(order); err != nil {
		t.Errorf("market order should be valid, got: %v", err)
	}
}

func TestValidateOrder_EmptySymbol(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "",
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	err := validateOrder(order)
	if !errors.Is(err, ErrEmptySymbol) {
		t.Errorf("expected ErrEmptySymbol, got: %v", err)
	}
}

func TestValidateOrder_SymbolTooLong(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "VERYLONGSYMBOLNAME", // > 16 chars
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	err := validateOrder(order)
	if !errors.Is(err, ErrSymbolTooLong) {
		t.Errorf("expected ErrSymbolTooLong, got: %v", err)
	}
}

func TestValidateOrder_MaxSymbolLength(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "1234567890123456", // Exactly 16 chars
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	if err := validateOrder(order); err != nil {
		t.Errorf("16-char symbol should be valid, got: %v", err)
	}
}

func TestValidateOrder_ZeroQuantity(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     0,
		Side:    SideBuy,
		OrderID: 1001,
	}

	err := validateOrder(order)
	if !errors.Is(err, ErrZeroQuantity) {
		t.Errorf("expected ErrZeroQuantity, got: %v", err)
	}
}

func TestValidateOrder_InvalidSide(t *testing.T) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     100,
		Side:    Side('X'),
		OrderID: 1001,
	}

	err := validateOrder(order)
	if !errors.Is(err, ErrInvalidSide) {
		t.Errorf("expected ErrInvalidSide, got: %v", err)
	}
}

func TestValidateCancel_Valid(t *testing.T) {
	cancel := &CancelOrder{
		Symbol:  "IBM",
		UserID:  1,
		OrderID: 1001,
	}

	if err := validateCancel(cancel); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateCancel_EmptySymbol(t *testing.T) {
	cancel := &CancelOrder{
		Symbol:  "",
		UserID:  1,
		OrderID: 1001,
	}

	err := validateCancel(cancel)
	if !errors.Is(err, ErrEmptySymbol) {
		t.Errorf("expected ErrEmptySymbol, got: %v", err)
	}
}

func TestValidateCancel_SymbolTooLong(t *testing.T) {
	cancel := &CancelOrder{
		Symbol:  "VERYLONGSYMBOLNAME",
		UserID:  1,
		OrderID: 1001,
	}

	err := validateCancel(cancel)
	if !errors.Is(err, ErrSymbolTooLong) {
		t.Errorf("expected ErrSymbolTooLong, got: %v", err)
	}
}
