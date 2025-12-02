// Full path: pkg/meclient/internal/stats/stats_test.go

package stats

import (
	"sync"
	"testing"
)

func TestStatsIncrement(t *testing.T) {
	s := &Stats{}

	s.IncMessagesSent()
	s.IncMessagesSent()
	s.IncMessagesReceived()
	s.IncErrorCount()
	s.IncReconnectCount()
	s.IncDroppedMessages()

	snap := s.GetSnapshot()

	if snap.MessagesSent != 2 {
		t.Errorf("expected MessagesSent=2, got %d", snap.MessagesSent)
	}
	if snap.MessagesReceived != 1 {
		t.Errorf("expected MessagesReceived=1, got %d", snap.MessagesReceived)
	}
	if snap.ErrorCount != 1 {
		t.Errorf("expected ErrorCount=1, got %d", snap.ErrorCount)
	}
	if snap.ReconnectCount != 1 {
		t.Errorf("expected ReconnectCount=1, got %d", snap.ReconnectCount)
	}
	if snap.DroppedMessages != 1 {
		t.Errorf("expected DroppedMessages=1, got %d", snap.DroppedMessages)
	}
}

func TestStatsReset(t *testing.T) {
	s := &Stats{}

	s.IncMessagesSent()
	s.IncMessagesReceived()
	s.IncErrorCount()

	s.Reset()

	snap := s.GetSnapshot()

	if snap.MessagesSent != 0 {
		t.Errorf("expected MessagesSent=0 after reset, got %d", snap.MessagesSent)
	}
	if snap.MessagesReceived != 0 {
		t.Errorf("expected MessagesReceived=0 after reset, got %d", snap.MessagesReceived)
	}
	if snap.ErrorCount != 0 {
		t.Errorf("expected ErrorCount=0 after reset, got %d", snap.ErrorCount)
	}
}

func TestStatsConcurrency(t *testing.T) {
	s := &Stats{}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				s.IncMessagesSent()
			}
		}()
	}

	wg.Wait()

	snap := s.GetSnapshot()
	expected := uint64(100 * 1000)

	if snap.MessagesSent != expected {
		t.Errorf("expected MessagesSent=%d, got %d", expected, snap.MessagesSent)
	}
}

func TestSnapshotIsImmutable(t *testing.T) {
	s := &Stats{}
	s.IncMessagesSent()

	snap1 := s.GetSnapshot()

	s.IncMessagesSent()

	snap2 := s.GetSnapshot()

	if snap1.MessagesSent != 1 {
		t.Error("snap1 should not be affected by later increments")
	}
	if snap2.MessagesSent != 2 {
		t.Errorf("snap2 should reflect new value, got %d", snap2.MessagesSent)
	}
}

func TestStatsZeroInitialization(t *testing.T) {
	s := &Stats{}
	snap := s.GetSnapshot()

	if snap.MessagesSent != 0 {
		t.Errorf("expected MessagesSent=0, got %d", snap.MessagesSent)
	}
	if snap.MessagesReceived != 0 {
		t.Errorf("expected MessagesReceived=0, got %d", snap.MessagesReceived)
	}
	if snap.ErrorCount != 0 {
		t.Errorf("expected ErrorCount=0, got %d", snap.ErrorCount)
	}
	if snap.ReconnectCount != 0 {
		t.Errorf("expected ReconnectCount=0, got %d", snap.ReconnectCount)
	}
	if snap.DroppedMessages != 0 {
		t.Errorf("expected DroppedMessages=0, got %d", snap.DroppedMessages)
	}
}
