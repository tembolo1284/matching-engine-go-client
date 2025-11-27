package meclient

// validateOrder checks that an order has valid fields.
func validateOrder(order *NewOrder) error {
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

// validateCancel checks that a cancel request has valid fields.
func validateCancel(cancel *CancelOrder) error {
	if cancel.Symbol == "" {
		return ErrEmptySymbol
	}

	if len(cancel.Symbol) > MaxSymbolLength {
		return ErrSymbolTooLong
	}

	return nil
}
