// Full path: pkg/meclient/protocol/messages_test.go

package protocol

import "testing"

func TestSideString(t *testing.T) {
	if SideBuy.String() != "BUY" {
		t.Errorf("expected BUY, got %s", SideBuy.String())
	}
	if SideSell.String() != "SELL" {
		t.Errorf("expected SELL, got %s", SideSell.String())
	}

	unknown := Side('X')
	if unknown.String() != "UNKNOWN" {
		t.Errorf("expected UNKNOWN, got %s", unknown.String())
	}
}

func TestSideValues(t *testing.T) {
	if SideBuy != 'B' {
		t.Errorf("expected SideBuy to be 'B', got %c", SideBuy)
	}
	if SideSell != 'S' {
		t.Errorf("expected SideSell to be 'S', got %c", SideSell)
	}
}
