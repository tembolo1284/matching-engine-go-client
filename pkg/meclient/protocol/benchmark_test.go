// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}// Full path: pkg/meclient/protocol/benchmark_test.go

package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeNewOrder(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeNewOrder(order)
	}
}

func BenchmarkEncodeCancel(b *testing.B) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	cancel := &CancelOrder{
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.EncodeCancel(cancel)
	}
}

func BenchmarkDecodeAck(b *testing.B) {
	input := frameMessage("A, IBM, 1, 1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeTrade(b *testing.B) {
	input := frameMessage("T, IBM, 1, 1001, 2, 2001, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkDecodeBookUpdate(b *testing.B) {
	input := frameMessage("B, IBM, B, 100, 50")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(input))
		_, _ = dec.Decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   100,
		Qty:     50,
		Side:    SideBuy,
		OrderID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrder(order)
	}
}
