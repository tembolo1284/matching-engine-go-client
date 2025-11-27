package meclient

import (
	"strings"
	"testing"
)

func TestDecoder_InvalidMessageType(t *testing.T) {
	input := "Z, invalid, message\n"
	dec := newDecoder(strings.NewReader(input))

	_, err := dec.decode()
	if err == nil {
		t.Fatal("expected error for invalid message type")
	}

	if !strings.Contains(err.Error(), "unknown message type") {
		t.Errorf("expected 'unknown message type' error, got: %v", err)
	}
}

func TestDecoder_InvalidFieldCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"ack_too_few", "A, IBM, 1\n"},
		{"ack_too_many", "A, IBM, 1, 2, 3\n"},
		{"trade_too_few", "T, IBM, 1, 2, 3\n"},
		{"trade_too_many", "T, IBM, 1, 2, 3, 4, 5, 6, 7, 8\n"},
		{"book_too_few", "B, IBM, B\n"},
		{"book_too_many", "B, IBM, B, 100, 200, 300\n"},
		{"cancel_too_few", "C, IBM, 1\n"},
		{"cancel_too_many", "C, IBM, 1, 2, 3\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := newDecoder(strings.NewReader(tt.input))
			_, err := dec.decode()
			if err == nil {
				t.Error("expected error for invalid field count")
			}
		})
	}
}

func TestDecoder_InvalidNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"ack_bad_user", "A, IBM, abc, 1001\n"},
		{"ack_bad_order", "A, IBM, 1, xyz\n"},
		{"ack_negative", "A, IBM, -1, 1001\n"},
		{"trade_bad_buy_user", "T, IBM, abc, 1001, 2, 2001, 150, 100\n"},
		{"trade_bad_price", "T, IBM, 1, 1001, 2, 2001, bad, 100\n"},
		{"trade_bad_qty", "T, IBM, 1, 1001, 2, 2001, 150, notanumber\n"},
		{"book_bad_price", "B, IBM, B, abc, 100\n"},
		{"book_bad_qty", "B, IBM, B, 150, notanumber\n"},
		{"cancel_bad_user", "C, IBM, abc, 1001\n"},
		{"cancel_bad_order", "C, IBM, 1, xyz\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := newDecoder(strings.NewReader(tt.input))
			_, err := dec.decode()
			if err == nil {
				t.Error("expected error for invalid number")
			}
		})
	}
}

func TestDecoder_InvalidSide(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid_char", "B, IBM, X, 150, 100\n"},
		{"lowercase_b", "B, IBM, b, 150, 100\n"},
		{"lowercase_s", "B, IBM, s, 150, 100\n"},
		{"number", "B, IBM, 1, 150, 100\n"},
		{"empty", "B, IBM, , 150, 100\n"},
		{"too_long", "B, IBM, BUY, 150, 100\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := newDecoder(strings.NewReader(tt.input))
			_, err := dec.decode()
			if err == nil {
				t.Error("expected error for invalid side")
			}
		})
	}
}

func TestDecoder_Overflow(t *testing.T) {
	// Value larger than uint32 max
	input := "A, IBM, 9999999999999, 1001\n"
	dec := newDecoder(strings.NewReader(input))

	_, err := dec.decode()
	if err == nil {
		t.Error("expected error for uint32 overflow")
	}
}

func TestDecoder_EmptyInput(t *testing.T) {
	dec := newDecoder(strings.NewReader(""))

	_, err := dec.decode()
	if err == nil {
		t.Error("expected EOF error for empty input")
	}
}
