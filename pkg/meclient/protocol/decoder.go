// Full path: pkg/meclient/protocol/decoder.go

package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Decoder decodes messages from the wire format with length-prefix framing.
type Decoder struct {
	r      io.Reader
	lenBuf [4]byte
	buf    []byte
}

// NewDecoder creates a new decoder reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:   r,
		buf: make([]byte, MaxFrameSize),
	}
}

// Decode reads and decodes the next message.
func (d *Decoder) Decode() (*Message, error) {
	// Read 4-byte length header
	if _, err := io.ReadFull(d.r, d.lenBuf[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(d.lenBuf[:])

	// Validate length
	if length == 0 {
		return nil, fmt.Errorf("invalid frame length: 0")
	}
	if length > MaxFrameSize {
		return nil, fmt.Errorf("frame too large: %d > %d", length, MaxFrameSize)
	}

	// Read payload
	if _, err := io.ReadFull(d.r, d.buf[:length]); err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	// Parse CSV
	line := strings.TrimSpace(string(d.buf[:length]))
	return d.parseLine(line)
}

func (d *Decoder) parseLine(line string) (*Message, error) {
	if len(line) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	// Split by comma, handling optional spaces
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	msgType := parts[0]

	switch msgType {
	case "A":
		return d.parseAck(parts)
	case "T":
		return d.parseTrade(parts)
	case "B":
		return d.parseBookUpdate(parts)
	case "C":
		return d.parseCancelAck(parts)
	default:
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}
}

func (d *Decoder) parseAck(parts []string) (*Message, error) {
	// A, symbol, user_id, order_id
	if len(parts) < 4 {
		return nil, fmt.Errorf("ack: expected 4 fields, got %d", len(parts))
	}

	userID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ack: invalid user_id: %w", err)
	}

	orderID, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ack: invalid order_id: %w", err)
	}

	return &Message{
		Ack: &Ack{
			Symbol:  parts[1],
			UserID:  uint32(userID),
			OrderID: uint32(orderID),
		},
	}, nil
}

func (d *Decoder) parseTrade(parts []string) (*Message, error) {
	// T, symbol, buy_user, buy_oid, sell_user, sell_oid, price, qty
	if len(parts) < 8 {
		return nil, fmt.Errorf("trade: expected 8 fields, got %d", len(parts))
	}

	buyUserID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("trade: invalid buy_user_id: %w", err)
	}

	buyOrderID, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("trade: invalid buy_order_id: %w", err)
	}

	sellUserID, err := strconv.ParseUint(parts[4], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("trade: invalid sell_user_id: %w", err)
	}

	sellOrderID, err := strconv.ParseUint(parts[5], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("trade: invalid sell_order_id: %w", err)
	}

	price, err := strconv.ParseUint(parts[6], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("trade: invalid price: %w", err)
	}

	qty, err := strconv.ParseUint(parts[7], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("trade: invalid qty: %w", err)
	}

	return &Message{
		Trade: &Trade{
			Symbol:      parts[1],
			BuyUserID:   uint32(buyUserID),
			BuyOrderID:  uint32(buyOrderID),
			SellUserID:  uint32(sellUserID),
			SellOrderID: uint32(sellOrderID),
			Price:       uint32(price),
			Qty:         uint32(qty),
		},
	}, nil
}

func (d *Decoder) parseBookUpdate(parts []string) (*Message, error) {
	// B, symbol, side, price, qty
	// or B, symbol, side, -, - (empty book)
	if len(parts) < 5 {
		return nil, fmt.Errorf("book: expected 5 fields, got %d", len(parts))
	}

	side := SideBuy
	if len(parts[2]) > 0 && (parts[2][0] == 'S' || parts[2][0] == 's') {
		side = SideSell
	}

	var price, qty uint32

	if parts[3] != "-" {
		p, err := strconv.ParseUint(parts[3], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("book: invalid price: %w", err)
		}
		price = uint32(p)
	}

	if parts[4] != "-" {
		q, err := strconv.ParseUint(parts[4], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("book: invalid qty: %w", err)
		}
		qty = uint32(q)
	}

	return &Message{
		BookUpdate: &BookUpdate{
			Symbol: parts[1],
			Side:   side,
			Price:  price,
			Qty:    qty,
		},
	}, nil
}

func (d *Decoder) parseCancelAck(parts []string) (*Message, error) {
	// C, symbol, user_id, order_id
	if len(parts) < 4 {
		return nil, fmt.Errorf("cancel_ack: expected 4 fields, got %d", len(parts))
	}

	userID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cancel_ack: invalid user_id: %w", err)
	}

	orderID, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cancel_ack: invalid order_id: %w", err)
	}

	return &Message{
		CancelAck: &CancelAck{
			Symbol:  parts[1],
			UserID:  uint32(userID),
			OrderID: uint32(orderID),
		},
	}, nil
}
