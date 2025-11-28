// Command example demonstrates usage of the matching engine Go client.
//
// Usage:
//
//	./example -addr localhost:1234
//	./example -addr localhost:1234 -interactive
//	./example -addr localhost:1234 -demo
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient"
)

func main() {
	addr := flag.String("addr", "localhost:1234", "Server address (host:port)")
	interactive := flag.Bool("interactive", false, "Run in interactive mode")
	demo := flag.Bool("demo", false, "Run demo sequence")
	userID := flag.Uint("user", 1, "Default user ID for orders")
	flag.Parse()

	fmt.Printf("Matching Engine Go Client\n")
	fmt.Printf("=========================\n\n")

	cfg := meclient.DefaultConfig(*addr)
	client, err := meclient.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connecting to %s...\n", *addr)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connected!\n\n")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Only start background receiver for non-demo modes
	if !*demo {
		wg.Add(1)
		go func() {
			defer wg.Done()
			receiveMessages(client, done)
		}()
	}

	if *demo {
		runDemo(client, uint32(*userID))
	} else if *interactive {
		runInteractive(client, uint32(*userID), shutdown)
	} else {
		printUsage()
		fmt.Println("\nWaiting for messages (Ctrl+C to quit)...")
		<-shutdown
	}

	fmt.Println("\nShutting down...")
	close(done)
	client.Close()
	wg.Wait()

	stats := client.Stats()
	fmt.Printf("\nSession Stats:\n")
	fmt.Printf("  Messages Sent:     %d\n", stats.MessagesSent)
	fmt.Printf("  Messages Received: %d\n", stats.MessagesReceived)
	fmt.Printf("  Errors:            %d\n", stats.ErrorCount)
	fmt.Printf("  Reconnects:        %d\n", stats.ReconnectCount)
	fmt.Printf("  Dropped Messages:  %d\n", stats.DroppedMessages)

	fmt.Println("Goodbye!")
}

func receiveMessages(client *meclient.Client, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return

		case ack, ok := <-client.Acks():
			if !ok {
				return
			}
			fmt.Printf("[ACK] %s user=%d order=%d\n", ack.Symbol, ack.UserID, ack.OrderID)

		case trade, ok := <-client.Trades():
			if !ok {
				return
			}
			fmt.Printf("[TRADE] %s buy(user=%d,oid=%d) sell(user=%d,oid=%d) price=%d qty=%d\n",
				trade.Symbol,
				trade.BuyUserID, trade.BuyOrderID,
				trade.SellUserID, trade.SellOrderID,
				trade.Price, trade.Qty)

		case update, ok := <-client.BookUpdates():
			if !ok {
				return
			}
			fmt.Printf("[BOOK] %s %s price=%d qty=%d\n",
				update.Symbol, update.Side, update.Price, update.Qty)

		case cancelAck, ok := <-client.CancelAcks():
			if !ok {
				return
			}
			fmt.Printf("[CANCEL] %s user=%d order=%d\n", cancelAck.Symbol, cancelAck.UserID, cancelAck.OrderID)

		case err, ok := <-client.Errors():
			if !ok {
				return
			}
			fmt.Printf("[ERROR] %v\n", err)

		case event, ok := <-client.Reconnects():
			if !ok {
				return
			}
			fmt.Printf("[RECONNECT] Connected after %d attempts\n", event.Attempt)
		}
	}
}

func runDemo(client *meclient.Client, userID uint32) {
	fmt.Println("Running demo sequence...")
	fmt.Println()

	// Helper to wait for and print responses
	drainResponses := func(timeout time.Duration) {
		deadline := time.After(timeout)
		for {
			select {
			case ack := <-client.Acks():
				fmt.Printf("   <- ACK %s user=%d order=%d\n", ack.Symbol, ack.UserID, ack.OrderID)
			case trade := <-client.Trades():
				fmt.Printf("   <- TRADE %s buy(%d,%d) sell(%d,%d) price=%d qty=%d\n",
					trade.Symbol,
					trade.BuyUserID, trade.BuyOrderID,
					trade.SellUserID, trade.SellOrderID,
					trade.Price, trade.Qty)
			case update := <-client.BookUpdates():
				fmt.Printf("   <- BOOK %s %s price=%d qty=%d\n",
					update.Symbol, update.Side, update.Price, update.Qty)
			case cancelAck := <-client.CancelAcks():
				fmt.Printf("   <- CANCEL %s user=%d order=%d\n", cancelAck.Symbol, cancelAck.UserID, cancelAck.OrderID)
			case err := <-client.Errors():
				fmt.Printf("   <- ERROR: %v\n", err)
			case <-deadline:
				return
			}
		}
	}

	// Check for any immediate errors
	fmt.Println("Checking connection...")
	drainResponses(100 * time.Millisecond)
	fmt.Println("Connection OK, starting scenarios...")
	fmt.Println()

	// Helper to send order and wait for response
	sendOrder := func(order meclient.NewOrder) {
		side := "BUY"
		if order.Side == meclient.SideSell {
			side = "SELL"
		}
		if order.Price == 0 {
			fmt.Printf("-> %s %s %d @ MARKET (oid=%d)\n", side, order.Symbol, order.Qty, order.OrderID)
		} else {
			fmt.Printf("-> %s %s %d @ %d (oid=%d)\n", side, order.Symbol, order.Qty, order.Price, order.OrderID)
		}
		if err := client.SendOrder(order); err != nil {
			fmt.Printf("   Send error: %v\n", err)
			return
		}
		drainResponses(500 * time.Millisecond)
	}

	// Helper to send cancel and wait for response
	sendCancel := func(cancel meclient.CancelOrder) {
		fmt.Printf("-> CANCEL %s user=%d oid=%d\n", cancel.Symbol, cancel.UserID, cancel.OrderID)
		if err := client.SendCancel(cancel); err != nil {
			fmt.Printf("   Send error: %v\n", err)
			return
		}
		drainResponses(500 * time.Millisecond)
	}

	// Helper to send flush and wait for responses
	sendFlush := func() {
		fmt.Printf("-> FLUSH\n")
		if err := client.SendFlush(); err != nil {
			fmt.Printf("   Error: %v\n", err)
			return
		}
		drainResponses(500 * time.Millisecond)
	}

	fmt.Println("=== Scenario 1: Simple Orders ===")
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 100, Qty: 50, Side: meclient.SideBuy, OrderID: 1})
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 105, Qty: 50, Side: meclient.SideSell, OrderID: 2})
	sendFlush()

	fmt.Println()
	fmt.Println("=== Scenario 2: Matching Orders ===")
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 100, Qty: 50, Side: meclient.SideBuy, OrderID: 3})
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 100, Qty: 50, Side: meclient.SideSell, OrderID: 4})
	sendFlush()

	fmt.Println()
	fmt.Println("=== Scenario 3: Partial Fill ===")
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 100, Qty: 100, Side: meclient.SideBuy, OrderID: 5})
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 100, Qty: 30, Side: meclient.SideSell, OrderID: 6})
	sendFlush()

	fmt.Println()
	fmt.Println("=== Scenario 4: Cancel Order ===")
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 100, Qty: 50, Side: meclient.SideBuy, OrderID: 7})
	sendCancel(meclient.CancelOrder{Symbol: "IBM", UserID: userID, OrderID: 7})
	sendFlush()

	fmt.Println()
	fmt.Println("=== Scenario 5: Multiple Symbols ===")
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "IBM", Price: 100, Qty: 50, Side: meclient.SideBuy, OrderID: 8})
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "AAPL", Price: 150, Qty: 30, Side: meclient.SideBuy, OrderID: 9})
	sendOrder(meclient.NewOrder{UserID: userID, Symbol: "TSLA", Price: 200, Qty: 25, Side: meclient.SideSell, OrderID: 10})
	sendFlush()

	fmt.Println()
	fmt.Println("Demo complete!")
}

func runInteractive(client *meclient.Client, defaultUserID uint32, shutdown <-chan os.Signal) {
	fmt.Println("Interactive mode - enter commands:")
	fmt.Println()
	printHelp()
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	orderID := uint32(1000)

	for {
		fmt.Print("> ")

		select {
		case <-shutdown:
			return
		default:
		}

		if !scanner.Scan() {
			return
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "help", "h", "?":
			printHelp()

		case "quit", "exit", "q":
			return

		case "buy", "b":
			order, err := parseOrderCommand(parts, meclient.SideBuy, defaultUserID, &orderID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Sending BUY %s %d @ %d (oid=%d)\n",
				order.Symbol, order.Qty, order.Price, order.OrderID)
			if err := client.SendOrder(order); err != nil {
				fmt.Printf("Send error: %v\n", err)
			}

		case "sell", "s":
			order, err := parseOrderCommand(parts, meclient.SideSell, defaultUserID, &orderID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Sending SELL %s %d @ %d (oid=%d)\n",
				order.Symbol, order.Qty, order.Price, order.OrderID)
			if err := client.SendOrder(order); err != nil {
				fmt.Printf("Send error: %v\n", err)
			}

		case "cancel", "c":
			cancel, err := parseCancelCommand(parts, defaultUserID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Sending CANCEL %s user=%d oid=%d\n",
				cancel.Symbol, cancel.UserID, cancel.OrderID)
			if err := client.SendCancel(cancel); err != nil {
				fmt.Printf("Send error: %v\n", err)
			}

		case "flush", "f":
			fmt.Println("Sending FLUSH")
			if err := client.SendFlush(); err != nil {
				fmt.Printf("Send error: %v\n", err)
			}

		case "status", "stat":
			if client.IsConnected() {
				fmt.Println("Connected")
			} else {
				fmt.Println("Disconnected")
			}
			stats := client.Stats()
			fmt.Printf("  Sent: %d  Received: %d  Errors: %d  Dropped: %d\n",
				stats.MessagesSent, stats.MessagesReceived,
				stats.ErrorCount, stats.DroppedMessages)

		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmd)
		}
	}
}

func parseOrderCommand(parts []string, side meclient.Side, defaultUserID uint32, nextOrderID *uint32) (meclient.NewOrder, error) {
	if len(parts) < 4 {
		return meclient.NewOrder{}, fmt.Errorf("usage: %s SYMBOL QTY PRICE [USER_ID]", parts[0])
	}

	symbol := strings.ToUpper(parts[1])
	if len(symbol) > 16 {
		return meclient.NewOrder{}, fmt.Errorf("symbol too long (max 16 chars)")
	}

	qty, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return meclient.NewOrder{}, fmt.Errorf("invalid quantity: %s", parts[2])
	}

	price, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return meclient.NewOrder{}, fmt.Errorf("invalid price: %s", parts[3])
	}

	userID := defaultUserID
	if len(parts) >= 5 {
		u, err := strconv.ParseUint(parts[4], 10, 32)
		if err != nil {
			return meclient.NewOrder{}, fmt.Errorf("invalid user_id: %s", parts[4])
		}
		userID = uint32(u)
	}

	orderID := *nextOrderID
	*nextOrderID++

	return meclient.NewOrder{
		UserID:  userID,
		Symbol:  symbol,
		Price:   uint32(price),
		Qty:     uint32(qty),
		Side:    side,
		OrderID: orderID,
	}, nil
}

func parseCancelCommand(parts []string, defaultUserID uint32) (meclient.CancelOrder, error) {
	if len(parts) < 3 {
		return meclient.CancelOrder{}, fmt.Errorf("usage: cancel SYMBOL ORDER_ID [USER_ID]")
	}

	symbol := strings.ToUpper(parts[1])

	orderID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return meclient.CancelOrder{}, fmt.Errorf("invalid order_id: %s", parts[2])
	}

	userID := defaultUserID
	if len(parts) >= 4 {
		u, err := strconv.ParseUint(parts[3], 10, 32)
		if err != nil {
			return meclient.CancelOrder{}, fmt.Errorf("invalid user_id: %s", parts[3])
		}
		userID = uint32(u)
	}

	return meclient.CancelOrder{
		Symbol:  symbol,
		UserID:  userID,
		OrderID: uint32(orderID),
	}, nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ./example -addr HOST:PORT              Connect and listen for messages")
	fmt.Println("  ./example -addr HOST:PORT -demo        Run demo order sequence")
	fmt.Println("  ./example -addr HOST:PORT -interactive Interactive command mode")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -addr string    Server address (default: localhost:1234)")
	fmt.Println("  -user uint      Default user ID (default: 1)")
	fmt.Println("  -demo           Run demo sequence")
	fmt.Println("  -interactive    Interactive mode")
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  buy SYMBOL QTY PRICE [USER_ID]    Place buy order")
	fmt.Println("  sell SYMBOL QTY PRICE [USER_ID]   Place sell order")
	fmt.Println("  cancel SYMBOL ORDER_ID [USER_ID]  Cancel order")
	fmt.Println("  flush                             Flush all order books")
	fmt.Println("  status                            Show connection status")
	fmt.Println("  help                              Show this help")
	fmt.Println("  quit                              Exit")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  buy IBM 100 150      Buy 100 IBM @ 150")
	fmt.Println("  sell AAPL 50 175     Sell 50 AAPL @ 175")
	fmt.Println("  cancel IBM 1001      Cancel order 1001 on IBM")
	fmt.Println()
	fmt.Println("Shortcuts: b=buy, s=sell, c=cancel, f=flush, h=help, q=quit")
}
