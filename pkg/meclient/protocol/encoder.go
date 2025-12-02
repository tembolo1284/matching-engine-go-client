// Full path: pkg/meclient/protocol/encoder.go

package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// MaxFrameSize is the maximum message size (16KB per message_framing.h)
	MaxFrameSize = 16384
)

// Encoder encodes messages to the wire format with length-prefix framing.
type Encoder struct {
	w      io.Writer
	buf    []byte
	lenBuf [4]byte
}

// NewEncoder creates a new encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:   w,
		buf: make([]byte, 0, 256),
	}
}

// writeFrame writes a length-prefixed frame.
func (e *Encoder) writeFrame(payload []byte) error {
	if len(payload) > MaxFrameSize {
		return fmt.Errorf("message too large: %d > %d", len(payload), MaxFrameSize)
	}

	// Write 4-byte big-endian length
	binary.BigEndian.PutUint32(e.lenBuf[:], uint32(len(payload)))
	if _, err := e.w.Write(e.lenBuf[:]); err != nil {
		return fmt.Errorf("write length: %w", err)
	}

	// Write payload
	if _, err := e.w.Write(payload); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}

	return nil
}

// EncodeNewOrder encodes a new order message.
// Format: N,user_id,symbol,price,qty,side,order_id\n
func (e *Encoder) EncodeNewOrder(order *NewOrder) error {
	side := byte('B')
	if order.Side == SideSell {
		side = 'S'
	}

	e.buf = e.buf[:0]
	e.buf = fmt.Appendf(e.buf, "N,%d,%s,%d,%d,%c,%d\n",
		order.UserID,
		order.Symbol,
		order.Price,
		order.Qty,
		side,
		order.OrderID)

	return e.writeFrame(e.buf)
}

// EncodeCancel encodes a cancel order message.
// Format: C,user_id,order_id\n (no symbol per message_parser.c)
func (e *Encoder) EncodeCancel(cancel *CancelOrder) error {
	e.buf = e.buf[:0]
	e.buf = fmt.Appendf(e.buf, "C,%d,%d\n",
		cancel.UserID,
		cancel.OrderID)

	return e.writeFrame(e.buf)
}

// EncodeFlush encodes a flush command.
// Format: F\n
func (e *Encoder) EncodeFlush() error {
	e.buf = e.buf[:0]
	e.buf = append(e.buf, "F\n"...)
	return e.writeFrame(e.buf)
}
