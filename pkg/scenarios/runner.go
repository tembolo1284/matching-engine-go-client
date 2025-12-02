// Full path: pkg/scenarios/runner.go

package scenarios

import (
	"fmt"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient"
)

// Runner executes scenarios against a matching engine client
type Runner struct {
	client      *meclient.Client
	userID      uint32
	nextOrderID uint32
	verbose     bool
	result      *Result
}

// NewRunner creates a new scenario runner
func NewRunner(client *meclient.Client, userID uint32, verbose bool) *Runner {
	return &Runner{
		client:      client,
		userID:      userID,
		nextOrderID: 1,
		verbose:     verbose,
		result:      &Result{},
	}
}

// Run executes a scenario by ID
func (r *Runner) Run(scenarioID int, dangerBurst bool) (*Result, error) {
	info := GetInfo(scenarioID)
	if info == nil {
		return nil, fmt.Errorf("unknown scenario: %d", scenarioID)
	}

	if info.RequiresBurst && !dangerBurst {
		return nil, fmt.Errorf("scenario %d requires --danger-burst flag", scenarioID)
	}

	// Reset state
	r.nextOrderID = 1
	r.result = &Result{}

	switch scenarioID {
	// Basic
	case 1:
		return r.SimpleOrders()
	case 2:
		return r.MatchingTrade()
	case 3:
		return r.CancelOrder()

	// Stress (throttled)
	case 10:
		return r.StressTest(1000, true)
	case 11:
		return r.StressTest(10000, true)
	case 12:
		return r.StressTest(100000, true)
	case 13:
		return r.StressTest(1000000, true)
	case 14:
		return r.StressTest(10000000, true)
	case 15:
		return r.StressTest(100000000, true)

	// Matching
	case 20:
		return r.MatchingStress(1000)
	case 21:
		return r.MatchingStress(10000)
	case 22:
		return r.MatchingStress(100000)
	case 23:
		return r.MatchingStress(1000000)

	// Multi-symbol
	case 30:
		return r.MultiSymbolStress(10000)
	case 31:
		return r.MultiSymbolStress(100000)
	case 32:
		return r.MultiSymbolStress(1000000)

	// Burst (unthrottled)
	case 40:
		return r.StressTest(100000, false)
	case 41:
		return r.StressTest(1000000, false)

	default:
		return nil, fmt.Errorf("scenario %d not implemented", scenarioID)
	}
}

// sendOrder sends an order and tracks the result
func (r *Runner) sendOrder(symbol string, price, qty uint32, side meclient.Side) error {
	order := meclient.NewOrder{
		UserID:  r.userID,
		Symbol:  symbol,
		Price:   price,
		Qty:     qty,
		Side:    side,
		OrderID: r.nextOrderID,
	}
	r.nextOrderID++

	if err := r.client.SendOrder(order); err != nil {
		r.result.OrdersFailed++
		return err
	}
	r.result.OrdersSent++
	return nil
}

// sendCancel sends a cancel request
func (r *Runner) sendCancel(orderID uint32) error {
	cancel := meclient.CancelOrder{
		UserID:  r.userID,
		OrderID: orderID,
	}
	return r.client.SendCancel(cancel)
}

// sendFlush sends a flush command
func (r *Runner) sendFlush() error {
	return r.client.SendFlush()
}

// drainResponses waits for and counts responses
func (r *Runner) drainResponses(timeout time.Duration) {
	deadline := time.After(timeout)
	for {
		select {
		case ack := <-r.client.Acks():
			r.result.ResponsesReceived++
			if r.verbose {
				fmt.Printf("[RECV] A, %s, %d, %d\n", ack.Symbol, ack.UserID, ack.OrderID)
			}
		case trade := <-r.client.Trades():
			r.result.ResponsesReceived++
			r.result.TradesExecuted++
			if r.verbose {
				fmt.Printf("[RECV] T, %s, %d, %d, %d, %d, %d, %d\n",
					trade.Symbol,
					trade.BuyUserID, trade.BuyOrderID,
					trade.SellUserID, trade.SellOrderID,
					trade.Price, trade.Qty)
			}
		case update := <-r.client.BookUpdates():
			r.result.ResponsesReceived++
			if r.verbose {
				if update.Price == 0 && update.Qty == 0 {
					fmt.Printf("[RECV] B, %s, %c, -, -\n", update.Symbol, update.Side)
				} else {
					fmt.Printf("[RECV] B, %s, %c, %d, %d\n",
						update.Symbol, update.Side, update.Price, update.Qty)
				}
			}
		case cancelAck := <-r.client.CancelAcks():
			r.result.ResponsesReceived++
			if r.verbose {
				fmt.Printf("[RECV] C, %s, %d, %d\n",
					cancelAck.Symbol, cancelAck.UserID, cancelAck.OrderID)
			}
		case <-r.client.Errors():
			// Count but don't print unless verbose
		case <-deadline:
			return
		}
	}
}

// SimpleOrders runs scenario 1: simple orders with no matching
func (r *Runner) SimpleOrders() (*Result, error) {
	fmt.Println("=== Scenario 1: Simple Orders ===")
	fmt.Println()

	r.result.StartTime = time.Now()

	// Buy order
	fmt.Println("Sending: BUY IBM 50@100")
	if err := r.sendOrder("IBM", 100, 50, meclient.SideBuy); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(100 * time.Millisecond)

	// Sell order at different price (no match)
	fmt.Println("\nSending: SELL IBM 50@105")
	if err := r.sendOrder("IBM", 105, 50, meclient.SideSell); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(100 * time.Millisecond)

	// Flush
	fmt.Println("\nSending: FLUSH")
	if err := r.sendFlush(); err != nil {
		return nil, err
	}
	r.result.OrdersSent++ // Count flush as an order

	// Wait for all responses
	time.Sleep(200 * time.Millisecond)
	r.drainResponses(500 * time.Millisecond)

	r.result.EndTime = time.Now()
	r.result.Finalize()
	return r.result, nil
}

// MatchingTrade runs scenario 2: orders that match
func (r *Runner) MatchingTrade() (*Result, error) {
	fmt.Println("=== Scenario 2: Matching Trade ===")
	fmt.Println()

	r.result.StartTime = time.Now()

	// Buy order
	fmt.Println("Sending: BUY IBM 50@100")
	if err := r.sendOrder("IBM", 100, 50, meclient.SideBuy); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(100 * time.Millisecond)

	// Matching sell order
	fmt.Println("\nSending: SELL IBM 50@100 (should match!)")
	if err := r.sendOrder("IBM", 100, 50, meclient.SideSell); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(300 * time.Millisecond)

	r.result.EndTime = time.Now()
	r.result.Finalize()
	return r.result, nil
}

// CancelOrder runs scenario 3: place and cancel an order
func (r *Runner) CancelOrder() (*Result, error) {
	fmt.Println("=== Scenario 3: Cancel Order ===")
	fmt.Println()

	r.result.StartTime = time.Now()

	// Buy order
	fmt.Println("Sending: BUY IBM 50@100")
	if err := r.sendOrder("IBM", 100, 50, meclient.SideBuy); err != nil {
		return nil, err
	}
	orderID := r.nextOrderID - 1
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(100 * time.Millisecond)

	// Cancel
	fmt.Printf("\nSending: CANCEL order %d\n", orderID)
	if err := r.sendCancel(orderID); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(300 * time.Millisecond)

	r.result.EndTime = time.Now()
	r.result.Finalize()
	return r.result, nil
}

// StressTest runs a stress test with the given order count
func (r *Runner) StressTest(count uint32, throttled bool) (*Result, error) {
	if throttled {
		fmt.Printf("=== Stress Test: %d Orders (throttled) ===\n\n", count)
	} else {
		fmt.Printf("=== BURST Stress Test: %d Orders (NO THROTTLING) ===\n", count)
		fmt.Println("!!! WARNING: May cause server parse errors !!!")
		fmt.Println()
	}

	r.result.StartTime = time.Now()
	r.verbose = false // Suppress output for stress tests

	// Flush first
	r.sendFlush()
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(50 * time.Millisecond)

	// Determine batch size and delay
	var batchSize uint32
	var delayMs int

	if !throttled {
		batchSize = count
		delayMs = 0
	} else if count >= 10000000 {
		batchSize = 50000
		delayMs = 50
	} else if count >= 1000000 {
		batchSize = 50000
		delayMs = 20
	} else if count >= 100000 {
		batchSize = 10000
		delayMs = 10
	} else if count >= 10000 {
		batchSize = 1000
		delayMs = 5
	} else {
		batchSize = count
		delayMs = 0
	}

	if throttled && delayMs > 0 {
		fmt.Printf("Batched mode: %d orders/batch, %d ms delay\n\n", batchSize, delayMs)
	}

	// Progress tracking
	progressInterval := count / 20
	if progressInterval == 0 {
		progressInterval = 1
	}
	var lastProgress uint32

	startTime := time.Now()

	// Send orders
	for i := uint32(0); i < count; i++ {
		price := 100 + (i % 100)
		r.sendOrder("IBM", price, 10, meclient.SideBuy)

		// Progress indicator
		if i > 0 && i/progressInterval > lastProgress {
			lastProgress = i / progressInterval
			pct := (i * 100) / count
			elapsed := time.Since(startTime)
			rate := float64(i) / elapsed.Seconds()
			fmt.Printf("  %d%% (%d orders, %.0f ms, %.0f orders/sec)\n",
				pct, i, float64(elapsed.Milliseconds()), rate)
		}

		// Batch delay
		if throttled && delayMs > 0 && i > 0 && (i%batchSize) == 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	fmt.Println("\nSending FLUSH to clear book...")
	r.sendFlush()

	// Wait based on order count
	var flushWait time.Duration
	if count >= 100000 {
		flushWait = 5 * time.Second
	} else if count >= 10000 {
		flushWait = 3 * time.Second
	} else if count >= 1000 {
		flushWait = 2 * time.Second
	} else {
		flushWait = 1 * time.Second
	}
	time.Sleep(flushWait)
	r.drainResponses(100 * time.Millisecond)

	r.result.EndTime = time.Now()
	r.result.Finalize()
	r.result.Print()
	return r.result, nil
}

// MatchingStress runs a matching stress test
func (r *Runner) MatchingStress(pairs uint32) (*Result, error) {
	fmt.Printf("=== Matching Stress Test: %d Trade Pairs ===\n\n", pairs)
	fmt.Printf("Sending %d buy/sell pairs (should generate %d trades)...\n\n", pairs, pairs)

	r.result.StartTime = time.Now()
	r.verbose = false

	// Flush first
	r.sendFlush()
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(50 * time.Millisecond)

	// Progress tracking
	progressInterval := pairs / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	// Send matching pairs
	for i := uint32(0); i < pairs; i++ {
		price := 100 + (i % 50)

		// Buy order
		r.sendOrder("IBM", price, 10, meclient.SideBuy)

		// Matching sell order
		r.sendOrder("IBM", price, 10, meclient.SideSell)

		// Progress
		if i > 0 && (i%progressInterval) == 0 {
			fmt.Printf("  Progress: %d%%\n", (i*100)/pairs)
		}
	}

	// Wait for responses
	time.Sleep(500 * time.Millisecond)
	r.drainResponses(100 * time.Millisecond)

	r.result.EndTime = time.Now()
	r.result.Finalize()
	r.result.Print()
	return r.result, nil
}

// MultiSymbolStress runs a multi-symbol stress test
func (r *Runner) MultiSymbolStress(count uint32) (*Result, error) {
	fmt.Printf("=== Multi-Symbol Stress Test: %d Orders ===\n\n", count)
	fmt.Printf("Using %d symbols across both processors...\n\n", len(MultiSymbols))

	r.result.StartTime = time.Now()
	r.verbose = false

	// Flush first
	r.sendFlush()
	time.Sleep(100 * time.Millisecond)
	r.drainResponses(50 * time.Millisecond)

	// Progress tracking
	progressInterval := count / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	// Track processor distribution
	var proc0Count, proc1Count uint32

	// Send orders across symbols
	for i := uint32(0); i < count; i++ {
		symbolIdx := int(i) % len(MultiSymbols)
		symbol := MultiSymbols[symbolIdx]
		price := 100 + (i % 100)
		side := meclient.SideBuy
		if i%2 == 1 {
			side = meclient.SideSell
		}

		if err := r.sendOrder(symbol, price, 10, side); err == nil {
			// Track processor distribution (A-M = first 5, N-Z = last 5)
			if symbolIdx < 5 {
				proc0Count++
			} else {
				proc1Count++
			}
		}

		// Progress
		if i > 0 && (i%progressInterval) == 0 {
			fmt.Printf("  Progress: %d%%\n", (i*100)/count)
		}
	}

	// Store processor distribution
	r.result.Proc0Orders = proc0Count
	r.result.Proc1Orders = proc1Count

	// Flush
	fmt.Println("\nSending FLUSH to clear all books...")
	r.sendFlush()
	time.Sleep(200 * time.Millisecond)
	r.drainResponses(100 * time.Millisecond)

	r.result.EndTime = time.Now()
	r.result.Finalize()
	r.result.Print()
	return r.result, nil
}
