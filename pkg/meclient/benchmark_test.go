package meclient

import (
	"strings"
	"testing"
)

func BenchmarkEncoder_NewOrder(b *testing.B) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.encodeNewOrder(order)
	}
}

func BenchmarkEncoder_Cancel(b *testing.B) {
	var buf strings.Builder
	enc := newEncoder(&buf)

	cancel := &CancelOrder{
		Symbol:  "IBM",
		UserID:  1,
		OrderID: 1001,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.encodeCancel(cancel)
	}
}

func BenchmarkDecoder_Ack(b *testing.B) {
	input := "A, 1, 1001\n"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dec := newDecoder(strings.NewReader(input))
		_, _ = dec.decode()
	}
}

func BenchmarkDecoder_Trade(b *testing.B) {
	input := "T, 1, 1001, 2, 2001, 150, 100\n"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dec := newDecoder(strings.NewReader(input))
		_, _ = dec.decode()
	}
}

func BenchmarkDecoder_BookUpdate(b *testing.B) {
	input := "B, IBM, B, 150, 500\n"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dec := newDecoder(strings.NewReader(input))
		_, _ = dec.decode()
	}
}

func BenchmarkValidateOrder(b *testing.B) {
	order := &NewOrder{
		UserID:  1,
		Symbol:  "IBM",
		Price:   150,
		Qty:     100,
		Side:    SideBuy,
		OrderID: 1001,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = validateOrder(order)
	}
}

func BenchmarkStats_Increment(b *testing.B) {
	stats := &ClientStats{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		stats.incMessagesSent()
	}
}

func BenchmarkStats_Snapshot(b *testing.B) {
	stats := &ClientStats{}
	stats.incMessagesSent()
	stats.incMessagesReceived()
	stats.incErrorCount()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = stats.Snapshot()
	}
}

// Parallel benchmarks

func BenchmarkStats_IncrementParallel(b *testing.B) {
	stats := &ClientStats{}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.incMessagesSent()
		}
	})
}

func BenchmarkEncoder_NewOrder_Parallel(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		var buf strings.Builder
		enc := newEncoder(&buf)

		order := &NewOrder{
			UserID:  1,
			Symbol:  "IBM",
			Price:   150,
			Qty:     100,
			Side:    SideBuy,
			OrderID: 1001,
		}

		for pb.Next() {
			buf.Reset()
			enc.encodeNewOrder(order)
		}
	})
}
