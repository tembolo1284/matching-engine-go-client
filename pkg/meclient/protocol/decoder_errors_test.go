// Full path: pkg/meclient/protocol/decoder_errors_test.go

package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestDecodeUnknownMessageType(t *testing.T) {
	input := frameMessage("X, IBM, 1, 1001")
	dec := NewDecoder(bytes.NewReader(input))

	_, err := dec.Decode()
	if err == nil {
		t.Error("expected error for unknown message type")
	}
}

func TestDecodeInvalidAckFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too few fields", "A, IBM, 1"},
		{"invalid user_id", "A, IBM, notanumber, 1001"},
		{"invalid order_id", "A, IBM, 1, notanumber"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := frameMessage(tt.input)
			dec := NewDecoder(bytes.NewReader(input))

			_, err := dec.Decode()
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestDecodeInvalidTradeFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too few fields", "T, IBM, 1, 1001, 2, 2001, 100"},
		{"invalid buy_user_id", "T, IBM, x, 1001, 2, 2001, 100, 50"},
		{"invalid buy_order_id", "T, IBM, 1, x, 2, 2001, 100, 50"},
		{"invalid sell_user_id", "T, IBM, 1, 1001, x, 2001, 100, 50"},
		{"invalid sell_order_id", "T, IBM, 1, 1001, 2, x, 100, 50"},
		{"invalid price", "T, IBM, 1, 1001, 2, 2001, x, 50"},
		{"invalid qty", "T, IBM, 1, 1001, 2, 2001, 100, x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := frameMessage(tt.input)
			dec := NewDecoder(bytes.NewReader(input))

			_, err := dec.Decode()
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestDecodeInvalidBookUpdateFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too few fields", "B, IBM, B, 100"},
		{"invalid price", "B, IBM, B, x, 50"},
		{"invalid qty", "B, IBM, B, 100, x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := frameMessage(tt.input)
			dec := NewDecoder(bytes.NewReader(input))

			_, err := dec.Decode()
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestDecodeEmptyMessage(t *testing.T) {
	input := frameMessage("")
	dec := NewDecoder(bytes.NewReader(input))

	_, err := dec.Decode()
	if err == nil {
		t.Error("expected error for empty message")
	}
}

func TestDecodeEOF(t *testing.T) {
	dec := NewDecoder(bytes.NewReader([]byte{}))

	_, err := dec.Decode()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestDecodeInvalidFrameLength(t *testing.T) {
	// Frame length of 0
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, 0)

	dec := NewDecoder(bytes.NewReader(buf))
	_, err := dec.Decode()
	if err == nil {
		t.Error("expected error for zero frame length")
	}
}

func TestDecodeFrameTooLarge(t *testing.T) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, MaxFrameSize+1)

	dec := NewDecoder(bytes.NewReader(buf))
	_, err := dec.Decode()
	if err == nil {
		t.Error("expected error for frame too large")
	}
}
