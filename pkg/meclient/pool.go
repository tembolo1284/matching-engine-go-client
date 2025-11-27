package meclient

import "sync"

// Message pools for reducing allocations in the hot path.
// These pools reuse message structs instead of allocating new ones.

var ackPool = sync.Pool{
	New: func() interface{} {
		return &Ack{}
	},
}

var tradePool = sync.Pool{
	New: func() interface{} {
		return &Trade{}
	},
}

var bookUpdatePool = sync.Pool{
	New: func() interface{} {
		return &BookUpdate{}
	},
}

var cancelAckPool = sync.Pool{
	New: func() interface{} {
		return &CancelAck{}
	},
}

// getAck retrieves an Ack from the pool.
func getAck() *Ack {
	return ackPool.Get().(*Ack)
}

// putAck returns an Ack to the pool.
func putAck(a *Ack) {
	a.UserID = 0
	a.OrderID = 0
	ackPool.Put(a)
}

// getTrade retrieves a Trade from the pool.
func getTrade() *Trade {
	return tradePool.Get().(*Trade)
}

// putTrade returns a Trade to the pool.
func putTrade(t *Trade) {
	*t = Trade{} // Zero out
	tradePool.Put(t)
}

// getBookUpdate retrieves a BookUpdate from the pool.
func getBookUpdate() *BookUpdate {
	return bookUpdatePool.Get().(*BookUpdate)
}

// putBookUpdate returns a BookUpdate to the pool.
func putBookUpdate(b *BookUpdate) {
	*b = BookUpdate{} // Zero out
	bookUpdatePool.Put(b)
}

// getCancelAck retrieves a CancelAck from the pool.
func getCancelAck() *CancelAck {
	return cancelAckPool.Get().(*CancelAck)
}

// putCancelAck returns a CancelAck to the pool.
func putCancelAck(c *CancelAck) {
	c.UserID = 0
	c.OrderID = 0
	cancelAckPool.Put(c)
}
