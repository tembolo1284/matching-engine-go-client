package meclient

import (
	"encoding/binary"
	"io"
	"strconv"
)

// encoder handles serialization of outbound messages to CSV format.
// Uses a reusable buffer to minimize allocations in the hot path.
// Messages are framed with a 4-byte big-endian length prefix for TCP.
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

// writeFramed writes the buffer with a 4-byte big-endian length prefix
func (e *encoder) writeFramed() error {
	// Create length prefix (4 bytes, big-endian)
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(e.buf)))

	// Write length prefix
	if _, err := e.w.Write(lenBuf); err != nil {
		return err
	}

	// Write message
	_, err := e.w.Write(e.buf)
	return err
}

// encodeNewOrder writes a new order message.
// Format: N,user_id,symbol,price,qty,side,order_id
func (e *encoder) encodeNewOrder(o *NewOrder) error {
	e.buf = e.buf[:0]

	e.buf = append(e.buf, 'N')
	e.buf = append(e.buf, ',')
	e.buf = strconv.AppendUint(e.buf, uint64(o.UserID), 10)
	e.buf = append(e.buf, ',')
	e.buf = append(e.buf, o.Symbol...)
	e.buf = append(e.buf, ',')
	e.buf = strconv.AppendUint(e.buf, uint64(o.Price), 10)
	e.buf = append(e.buf, ',')
	e.buf = strconv.AppendUint(e.buf, uint64(o.Qty), 10)
	e.buf = append(e.buf, ',')
	e.buf = append(e.buf, byte(o.Side))
	e.buf = append(e.buf, ',')
	e.buf = strconv.AppendUint(e.buf, uint64(o.OrderID), 10)
	e.buf = append(e.buf, '\n')

	return e.writeFramed()
}

// encodeCancel writes a cancel order message.
// Format: C,user_id,order_id
func (e *encoder) encodeCancel(c *CancelOrder) error {
	e.buf = e.buf[:0]

	e.buf = append(e.buf, 'C')
	e.buf = append(e.buf, ',')
	e.buf = strconv.AppendUint(e.buf, uint64(c.UserID), 10)
	e.buf = append(e.buf, ',')
	e.buf = strconv.AppendUint(e.buf, uint64(c.OrderID), 10)
	e.buf = append(e.buf, '\n')

	return e.writeFramed()
}

// encodeFlush writes a flush command.
// Format: F
func (e *encoder) encodeFlush() error {
	e.buf = e.buf[:0]
	e.buf = append(e.buf, 'F')
	e.buf = append(e.buf, '\n')

	return e.writeFramed()
}
