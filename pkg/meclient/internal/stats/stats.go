// Full path: pkg/meclient/internal/stats/stats.go

// Package stats provides thread-safe statistics tracking.
package stats

import "sync/atomic"

// Snapshot is a point-in-time copy of statistics.
type Snapshot struct {
	MessagesSent     uint64
	MessagesReceived uint64
	ErrorCount       uint64
	ReconnectCount   uint64
	DroppedMessages  uint64
}

// Stats tracks client statistics with atomic operations.
type Stats struct {
	messagesSent     uint64
	messagesReceived uint64
	errorCount       uint64
	reconnectCount   uint64
	droppedMessages  uint64
}

// IncMessagesSent increments the sent message counter.
func (s *Stats) IncMessagesSent() {
	atomic.AddUint64(&s.messagesSent, 1)
}

// IncMessagesReceived increments the received message counter.
func (s *Stats) IncMessagesReceived() {
	atomic.AddUint64(&s.messagesReceived, 1)
}

// IncErrorCount increments the error counter.
func (s *Stats) IncErrorCount() {
	atomic.AddUint64(&s.errorCount, 1)
}

// IncReconnectCount increments the reconnect counter.
func (s *Stats) IncReconnectCount() {
	atomic.AddUint64(&s.reconnectCount, 1)
}

// IncDroppedMessages increments the dropped message counter.
func (s *Stats) IncDroppedMessages() {
	atomic.AddUint64(&s.droppedMessages, 1)
}

// GetSnapshot returns a point-in-time copy of all statistics.
func (s *Stats) GetSnapshot() Snapshot {
	return Snapshot{
		MessagesSent:     atomic.LoadUint64(&s.messagesSent),
		MessagesReceived: atomic.LoadUint64(&s.messagesReceived),
		ErrorCount:       atomic.LoadUint64(&s.errorCount),
		ReconnectCount:   atomic.LoadUint64(&s.reconnectCount),
		DroppedMessages:  atomic.LoadUint64(&s.droppedMessages),
	}
}

// Reset resets all counters to zero.
func (s *Stats) Reset() {
	atomic.StoreUint64(&s.messagesSent, 0)
	atomic.StoreUint64(&s.messagesReceived, 0)
	atomic.StoreUint64(&s.errorCount, 0)
	atomic.StoreUint64(&s.reconnectCount, 0)
	atomic.StoreUint64(&s.droppedMessages, 0)
}
