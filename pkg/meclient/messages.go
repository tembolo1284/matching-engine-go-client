// Package meclient provides a Go client for the matching engine TCP server.
package meclient

// Side represents the order side (buy or sell).
type Side byte

const (
	SideBuy  Side = 'B'
	SideSell Side = 'S'
)

// String returns the human-readable side name.
func (s Side) String() string {
	switch s {
	case SideBuy:
		return "BUY"
	case SideSell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// NewOrder represents a new order to be sent to the matching engine.
// Fields ordered by size (largest first) to minimize padding.
type NewOrder struct {
	Symbol  string // 16 bytes (string header)
	Price   uint32 // 4 bytes
	Qty     uint32 // 4 bytes
	UserID  uint32 // 4 bytes
	OrderID uint32 // 4 bytes
	Side    Side   // 1 byte
	// 3 bytes implicit padding
}

// CancelOrder represents a cancel request.
type CancelOrder struct {
	Symbol  string // 16 bytes
	UserID  uint32 // 4 bytes
	OrderID uint32 // 4 bytes
}

// Ack represents an order acknowledgment from the server.
type Ack struct {
	UserID  uint32
	OrderID uint32
}

// Trade represents an execution report from the server.
// Fields ordered to pack tightly.
type Trade struct {
	Price       uint32
	Qty         uint32
	BuyUserID   uint32
	BuyOrderID  uint32
	SellUserID  uint32
	SellOrderID uint32
}

// BookUpdate represents a top-of-book update from the server.
type BookUpdate struct {
	Symbol string
	Price  uint32
	Qty    uint32
	Side   Side
}

// CancelAck represents a cancel confirmation from the server.
type CancelAck struct {
	UserID  uint32
	OrderID uint32
}

// ReconnectEvent is sent when the client reconnects to the server.
type ReconnectEvent struct {
	Attempt int
}
