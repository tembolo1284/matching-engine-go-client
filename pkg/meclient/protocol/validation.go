// Full path: pkg/meclient/protocol/validation.go

package protocol

import "errors"

// Validation constants
const (
	MaxSymbolLength = 16
)

// Validation errors
var (
	ErrEmptySymbol   = errors.New("symbol cannot be empty")
	ErrSymbolTooLong = errors.New("symbol exceeds maximum length")
	ErrZeroQuantity  = errors.New("quantity must be greater than zero")
	ErrInvalidSide   = errors.New("invalid order side")
)

// ValidateOrder validates a new order.
func ValidateOrder(order *NewOrder) error {
	if order.Symbol == "" {
		return ErrEmptySymbol
	}
	if len(order.Symbol) > MaxSymbolLength {
		return ErrSymbolTooLong
	}
	if order.Qty == 0 {
		return ErrZeroQuantity
	}
	if order.Side != SideBuy && order.Side != SideSell {
		return ErrInvalidSide
	}
	return nil
}

// ValidateCancel validates a cancel order.
func ValidateCancel(cancel *CancelOrder) error {
	// Cancel doesn't need symbol validation since it's not sent on wire
	return nil
}
