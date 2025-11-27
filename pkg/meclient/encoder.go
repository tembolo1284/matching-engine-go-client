package meclient

import (
	"io"
	"strconv"
)

// encoder handles serialization of outbound messages to CSV format.
// Uses a reusable buffer to minimize allocations in the hot path.
type encoder struct {
	w   io.Writer
	buf []byte // Reusable buffer
}

// newEncoder creates an encoder that writes to w.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:   w,
		buf: make([]byte, 0, 128), // Pre-allocate typical message size
	}
}

// encodeNewOrder writes a new order message.
// Format: N, user_id, symbol, price, qty, side, order_id
func (e *encoder) encodeNewOrder(o *NewOrder) error {
	e.buf = e.buf[:0]

	e.buf = append(e.buf, "N, "...)
	e.buf = strconv.AppendUint(e.buf, uint64(o.UserID), 10)
	e.buf = append(e.buf, ", "...)
	e.buf = append(e.buf, o.Symbol...)
	e.buf = append(e.buf, ", "...)
	e.buf = strconv.AppendUint(e.buf, uint64(o.Price), 10)
	e.buf = append(e.buf, ", "...)
	e.buf = strconv.AppendUint(e.buf, uint64(o.Qty), 10)
	e.buf = append(e.buf, ", "...)
	e.buf = append(e.buf, byte(o.Side))
	e.buf = append(e.buf, ", "...)
	e.buf = strconv.AppendUint(e.buf, uint64(o.OrderID), 10)
	e.buf = append(e.buf, '\n')

	_, err := e.w.Write(e.buf)
	return err
}

// encodeCancel writes a cancel order message.
// Format: C, symbol, user_id, order_id
func (e *encoder) encodeCancel(c *CancelOrder) error {
	e.buf = e.buf[:0]

	e.buf = append(e.buf, "C, "...)
	e.buf = append(e.buf, c.Symbol...)
	e.buf = append(e.buf, ", "...)
	e.buf = strconv.AppendUint(e.buf, uint64(c.UserID), 10)
	e.buf = append(e.buf, ", "...)
	e.buf = strconv.AppendUint(e.buf, uint64(c.OrderID), 10)
	e.buf = append(e.buf, '\n')

	_, err := e.w.Write(e.buf)
	return err
}

// encodeFlush writes a flush command.
// Format: F
func (e *encoder) encodeFlush() error {
	_, err := e.w.Write([]byte("F\n"))
	return err
}
