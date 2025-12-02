// Full path: pkg/scenarios/scenarios.go

// Package scenarios provides test scenarios for the matching engine client.
// Ported from the C matching engine client.
package scenarios

import (
	"fmt"
	"time"
)

// Category represents the type of scenario
type Category int

const (
	CategoryBasic       Category = iota // Basic functional tests
	CategoryStress                      // Stress tests (throttled)
	CategoryMatching                    // Matching/trade generation
	CategoryMultiSymbol                 // Multi-symbol (dual-processor)
	CategoryBurst                       // Burst mode (no throttling)
)

// Info describes a scenario
type Info struct {
	ID            int
	Name          string
	Description   string
	Category      Category
	OrderCount    int
	RequiresBurst bool
}

// Result holds the outcome of running a scenario
type Result struct {
	OrdersSent        uint32
	OrdersFailed      uint32
	ResponsesReceived uint32
	TradesExecuted    uint32

	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	MinLatency time.Duration
	AvgLatency time.Duration
	MaxLatency time.Duration

	OrdersPerSec   float64
	MessagesPerSec float64

	// Processor distribution (for multi-symbol tests)
	Proc0Orders uint32
	Proc1Orders uint32
}

// Symbols for multi-symbol tests (spread across both processors)
var MultiSymbols = []string{
	// Processor 0 (A-M)
	"AAPL", "IBM", "GOOGL", "META", "MSFT",
	// Processor 1 (N-Z)
	"NVDA", "TSLA", "UBER", "SNAP", "ZM",
}

// Registry of all available scenarios
var Registry = []Info{
	// Basic scenarios
	{1, "simple-orders", "Simple orders (no match)", CategoryBasic, 3, false},
	{2, "matching-trade", "Matching trade execution", CategoryBasic, 2, false},
	{3, "cancel-order", "Cancel order", CategoryBasic, 2, false},

	// Stress tests (throttled)
	{10, "stress-1k", "Stress: 1K orders", CategoryStress, 1000, false},
	{11, "stress-10k", "Stress: 10K orders", CategoryStress, 10000, false},
	{12, "stress-100k", "Stress: 100K orders", CategoryStress, 100000, false},
	{13, "stress-1m", "Stress: 1M orders", CategoryStress, 1000000, false},
	{14, "stress-10m", "Stress: 10M orders ** EXTREME **", CategoryStress, 10000000, false},
	{15, "stress-100m", "Stress: 100M orders ** INSANE **", CategoryStress, 100000000, false},

	// Matching stress
	{20, "match-1k", "Matching: 1K pairs (2K orders)", CategoryMatching, 2000, false},
	{21, "match-10k", "Matching: 10K pairs", CategoryMatching, 20000, false},
	{22, "match-100k", "Matching: 100K pairs", CategoryMatching, 200000, false},
	{23, "match-1m", "Matching: 1M pairs ** EXTREME **", CategoryMatching, 2000000, false},

	// Multi-symbol stress
	{30, "multi-10k", "Multi-symbol: 10K orders", CategoryMultiSymbol, 10000, false},
	{31, "multi-100k", "Multi-symbol: 100K orders", CategoryMultiSymbol, 100000, false},
	{32, "multi-1m", "Multi-symbol: 1M orders", CategoryMultiSymbol, 1000000, false},

	// Burst mode (unthrottled - danger!)
	{40, "burst-100k", "Burst: 100K orders (raw speed)", CategoryBurst, 100000, true},
	{41, "burst-1m", "Burst: 1M orders (raw speed)", CategoryBurst, 1000000, true},
}

// GetInfo returns scenario info by ID, or nil if not found
func GetInfo(id int) *Info {
	for i := range Registry {
		if Registry[i].ID == id {
			return &Registry[i]
		}
	}
	return nil
}

// IsValid returns true if the scenario ID exists
func IsValid(id int) bool {
	return GetInfo(id) != nil
}

// RequiresBurst returns true if the scenario requires --danger-burst flag
func RequiresBurst(id int) bool {
	info := GetInfo(id)
	return info != nil && info.RequiresBurst
}

// PrintList prints all available scenarios grouped by category
func PrintList() {
	fmt.Println("Available scenarios:")
	fmt.Println()

	fmt.Println("Basic:")
	for _, s := range Registry {
		if s.Category == CategoryBasic {
			fmt.Printf("  %-3d - %s\n", s.ID, s.Description)
		}
	}

	fmt.Println("\nStress Tests (throttled):")
	for _, s := range Registry {
		if s.Category == CategoryStress {
			fmt.Printf("  %-3d - %s\n", s.ID, s.Description)
		}
	}

	fmt.Println("\nMatching Stress (generates trades):")
	for _, s := range Registry {
		if s.Category == CategoryMatching {
			fmt.Printf("  %-3d - %s\n", s.ID, s.Description)
		}
	}

	fmt.Println("\nMulti-Symbol Stress (tests dual-processor):")
	for _, s := range Registry {
		if s.Category == CategoryMultiSymbol {
			fmt.Printf("  %-3d - %s\n", s.ID, s.Description)
		}
	}

	fmt.Println("\nBurst Mode (no throttling - requires --danger-burst):")
	for _, s := range Registry {
		if s.Category == CategoryBurst {
			fmt.Printf("  %-3d - %s\n", s.ID, s.Description)
		}
	}
}

// Print prints scenario results in a formatted way
func (r *Result) Print() {
	fmt.Println()
	fmt.Println("=== Scenario Results ===")
	fmt.Println()

	fmt.Println("Orders:")
	fmt.Printf("  Sent:              %d\n", r.OrdersSent)
	fmt.Printf("  Failed:            %d\n", r.OrdersFailed)
	fmt.Printf("  Responses:         %d\n", r.ResponsesReceived)
	fmt.Printf("  Trades:            %d\n", r.TradesExecuted)
	fmt.Println()

	// Format time nicely
	if r.Duration >= time.Second {
		fmt.Printf("Time:                %.3f sec\n", r.Duration.Seconds())
	} else {
		fmt.Printf("Time:                %.3f ms\n", float64(r.Duration.Microseconds())/1000)
	}
	fmt.Println()

	fmt.Println("Throughput:")
	if r.OrdersPerSec >= 1000000 {
		fmt.Printf("  Orders/sec:        %.2fM\n", r.OrdersPerSec/1e6)
	} else if r.OrdersPerSec >= 1000 {
		fmt.Printf("  Orders/sec:        %.2fK\n", r.OrdersPerSec/1e3)
	} else {
		fmt.Printf("  Orders/sec:        %.0f\n", r.OrdersPerSec)
	}
	fmt.Println()

	if r.MinLatency > 0 {
		fmt.Println("Latency (round-trip):")
		fmt.Printf("  Min:               %.3f µs\n", float64(r.MinLatency.Nanoseconds())/1000)
		fmt.Printf("  Avg:               %.3f µs\n", float64(r.AvgLatency.Nanoseconds())/1000)
		fmt.Printf("  Max:               %.3f µs\n", float64(r.MaxLatency.Nanoseconds())/1000)
		fmt.Println()
	}

	if r.Proc0Orders > 0 || r.Proc1Orders > 0 {
		fmt.Println("Processor Distribution:")
		fmt.Printf("  Processor 0 (A-M): %d orders\n", r.Proc0Orders)
		fmt.Printf("  Processor 1 (N-Z): %d orders\n", r.Proc1Orders)
		fmt.Println()
	}
}

// Finalize calculates derived fields like throughput
func (r *Result) Finalize() {
	r.Duration = r.EndTime.Sub(r.StartTime)

	if r.Duration > 0 {
		seconds := r.Duration.Seconds()
		r.OrdersPerSec = float64(r.OrdersSent) / seconds
		r.MessagesPerSec = float64(r.OrdersSent+r.ResponsesReceived) / seconds
	}
}
