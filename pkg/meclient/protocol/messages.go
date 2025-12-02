// Full path: pkg/meclient/protocol/messages.go

// Package protocol defines message types and wire formats for the matching engine.
package protocol

// Side represents buy or sell side of an order.
type Side byte

const (
	SideBuy  Side = 'B'
	SideSell Side = 'S'
)

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

// NewOrder represents a new order request.
type NewOrder struct {
	UserID  uint32
	Symbol  string
	Price   uint32
	Qty     uint32
	Side    Side
	OrderID uint32
}

// CancelOrder represents a cancel order request.
type CancelOrder struct {
	Symbol  string // Not sent in wire format, kept for client tracking
	UserID  uint32
	OrderID uint32
}

// Ack represents an order acknowledgment from the server.
type Ack struct {
	Symbol  string
	UserID  uint32
	OrderID uint32
}

// Trade represents a trade execution notification.
type Trade struct {
	Symbol      string
	BuyUserID   uint32
	BuyOrderID  uint32
	SellUserID  uint32
	SellOrderID uint32
	Price       uint32
	Qty         uint32
}

// BookUpdate represents a top-of-book update.
type BookUpdate struct {
	Symbol string
	Side   Side
	Price  uint32
	Qty    uint32
}

// CancelAck represents a cancel acknowledgment.
type CancelAck struct {
	Symbol  string
	UserID  uint32
	OrderID uint32
}

// ReconnectEvent is sent when the client reconnects.
type ReconnectEvent struct {
	Attempt int
}

// Message is a union type for all possible server responses.
type Message struct {
	Ack        *Ack
	Trade      *Trade
	BookUpdate *BookUpdate
	CancelAck  *CancelAck
}
