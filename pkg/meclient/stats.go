package meclient

import (
	"sync/atomic"
)

// Cache line size on most modern CPUs
const cacheLineSize = 64

// ClientStats tracks client health metrics.
// Each counter is padded to its own cache line to prevent false sharing
// when multiple goroutines update different counters concurrently.
type ClientStats struct {
	messagesSent     uint64
	_pad1            [cacheLineSize - 8]byte
	messagesReceived uint64
	_pad2            [cacheLineSize - 8]byte
	errorCount       uint64
	_pad3            [cacheLineSize - 8]byte
	reconnectCount   uint64
	_pad4            [cacheLineSize - 8]byte
	droppedMessages  uint64
	_pad5            [cacheLineSize - 8]byte
}

// StatsSnapshot is a point-in-time copy of client statistics.
type StatsSnapshot struct {
	MessagesSent     uint64
	MessagesReceived uint64
	ErrorCount       uint64
	ReconnectCount   uint64
	DroppedMessages  uint64
}

// Snapshot returns a point-in-time copy of the statistics.
func (s *ClientStats) Snapshot() StatsSnapshot {
	return StatsSnapshot{
		MessagesSent:     atomic.LoadUint64(&s.messagesSent),
		MessagesReceived: atomic.LoadUint64(&s.messagesReceived),
		ErrorCount:       atomic.LoadUint64(&s.errorCount),
		ReconnectCount:   atomic.LoadUint64(&s.reconnectCount),
		DroppedMessages:  atomic.LoadUint64(&s.droppedMessages),
	}
}

// Atomic increment methods - no locks needed
func (s *ClientStats) incMessagesSent()     { atomic.AddUint64(&s.messagesSent, 1) }
func (s *ClientStats) incMessagesReceived() { atomic.AddUint64(&s.messagesReceived, 1) }
func (s *ClientStats) incErrorCount()       { atomic.AddUint64(&s.errorCount, 1) }
func (s *ClientStats) incReconnectCount()   { atomic.AddUint64(&s.reconnectCount, 1) }
func (s *ClientStats) incDroppedMessages()  { atomic.AddUint64(&s.droppedMessages, 1) }
