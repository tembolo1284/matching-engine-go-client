package meclient

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// decoder handles parsing of inbound messages from the server.
type decoder struct {
	scanner *bufio.Scanner
}

// newDecoder creates a decoder that reads from r.
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		scanner: bufio.NewScanner(r),
	}
}

// message represents a decoded message from the server.
// Exactly one of the pointer fields will be non-nil.
type message struct {
	Ack        *Ack
	Trade      *Trade
	BookUpdate *BookUpdate
	CancelAck  *CancelAck
}

// decode reads and parses the next message from the server.
// Returns io.EOF when the connection is closed.
func (d *decoder) decode() (*message, error) {
	if !d.scanner.Scan() {
		if err := d.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := strings.TrimSpace(d.scanner.Text())
	if len(line) == 0 {
		return d.decode() // Skip empty lines
	}

	return d.parseLine(line)
}

// parseLine parses a single CSV line into a message.
func (d *decoder) parseLine(line string) (*message, error) {
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	switch parts[0] {
	case "A":
		return d.parseAck(parts)
	case "T":
		return d.parseTrade(parts)
	case "B":
		return d.parseBookUpdate(parts)
	case "X":
		return d.parseCancelAck(parts)
	default:
		return nil, fmt.Errorf("unknown message type: %s", parts[0])
	}
}

// parseAck parses: A, user_id, order_id
func (d *decoder) parseAck(parts []string) (*message, error) {
	if len(parts) != 3 {
		return nil, fmt.Errorf("ack: expected 3 fields, got %d", len(parts))
	}

	userID, err := parseUint32(parts[1])
	if err != nil {
		return nil, fmt.Errorf("ack: invalid user_id: %w", err)
	}

	orderID, err := parseUint32(parts[2])
	if err != nil {
		return nil, fmt.Errorf("ack: invalid order_id: %w", err)
	}

	return &message{
		Ack: &Ack{
			UserID:  userID,
			OrderID: orderID,
		},
	}, nil
}

// parseTrade parses: T, buy_user, buy_oid, sell_user, sell_oid, price, qty
func (d *decoder) parseTrade(parts []string) (*message, error) {
	if len(parts) != 7 {
		return nil, fmt.Errorf("trade: expected 7 fields, got %d", len(parts))
	}

	buyUserID, err := parseUint32(parts[1])
	if err != nil {
		return nil, fmt.Errorf("trade: invalid buy_user: %w", err)
	}

	buyOrderID, err := parseUint32(parts[2])
	if err != nil {
		return nil, fmt.Errorf("trade: invalid buy_oid: %w", err)
	}

	sellUserID, err := parseUint32(parts[3])
	if err != nil {
		return nil, fmt.Errorf("trade: invalid sell_user: %w", err)
	}

	sellOrderID, err := parseUint32(parts[4])
	if err != nil {
		return nil, fmt.Errorf("trade: invalid sell_oid: %w", err)
	}

	price, err := parseUint32(parts[5])
	if err != nil {
		return nil, fmt.Errorf("trade: invalid price: %w", err)
	}

	qty, err := parseUint32(parts[6])
	if err != nil {
		return nil, fmt.Errorf("trade: invalid qty: %w", err)
	}

	return &message{
		Trade: &Trade{
			BuyUserID:   buyUserID,
			BuyOrderID:  buyOrderID,
			SellUserID:  sellUserID,
			SellOrderID: sellOrderID,
			Price:       price,
			Qty:         qty,
		},
	}, nil
}

// parseBookUpdate parses: B, symbol, side, price, qty
func (d *decoder) parseBookUpdate(parts []string) (*message, error) {
	if len(parts) != 5 {
		return nil, fmt.Errorf("book_update: expected 5 fields, got %d", len(parts))
	}

	symbol := parts[1]

	if len(parts[2]) != 1 {
		return nil, fmt.Errorf("book_update: invalid side: %s", parts[2])
	}
	side := Side(parts[2][0])
	if side != SideBuy && side != SideSell {
		return nil, fmt.Errorf("book_update: invalid side: %c", side)
	}

	price, err := parseUint32(parts[3])
	if err != nil {
		return nil, fmt.Errorf("book_update: invalid price: %w", err)
	}

	qty, err := parseUint32(parts[4])
	if err != nil {
		return nil, fmt.Errorf("book_update: invalid qty: %w", err)
	}

	return &message{
		BookUpdate: &BookUpdate{
			Symbol: symbol,
			Side:   side,
			Price:  price,
			Qty:    qty,
		},
	}, nil
}

// parseCancelAck parses: X, user_id, order_id
func (d *decoder) parseCancelAck(parts []string) (*message, error) {
	if len(parts) != 3 {
		return nil, fmt.Errorf("cancel_ack: expected 3 fields, got %d", len(parts))
	}

	userID, err := parseUint32(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cancel_ack: invalid user_id: %w", err)
	}

	orderID, err := parseUint32(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cancel_ack: invalid order_id: %w", err)
	}

	return &message{
		CancelAck: &CancelAck{
			UserID:  userID,
			OrderID: orderID,
		},
	}, nil
}

// parseUint32 parses a string as a uint32.
func parseUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
