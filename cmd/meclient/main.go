// Full path: cmd/meclient/main.go

// Command meclient is a Go client for the matching engine server.
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

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient"
	"github.com/tembolo1284/matching-engine-go-client/pkg/scenarios"
)

func main() {
	// Flags
	interactive := flag.Bool("i", false, "Run in interactive mode")
	list := flag.Bool("list", false, "List available scenarios")
	userID := flag.Uint("user", 1, "Default user ID for orders")
	dangerBurst := flag.Bool("danger-burst", false, "Allow unthrottled burst mode scenarios")
	verbose := flag.Bool("v", false, "Verbose output for scenarios")

	// Transport/Protocol flags
	useUDP := flag.Bool("udp", false, "Use UDP transport (default: TCP)")
	useBinary := flag.Bool("binary", false, "Force binary protocol (default: auto-detect for TCP, CSV for UDP)")

	flag.Parse()

	// List scenarios and exit
	if *list {
		scenarios.PrintList()
		return
	}

	// Parse positional args: HOST PORT [SCENARIO]
	args := flag.Args()
	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	addr := fmt.Sprintf("%s:%s", host, port)

	// Check for scenario ID
	var scenarioID int
	if len(args) >= 3 {
		id, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid scenario ID: %s\n", args[2])
			os.Exit(1)
		}
		scenarioID = id
	}

	fmt.Printf("Matching Engine Go Client\n")
	fmt.Printf("=========================\n\n")

	// Build config
	cfg := meclient.DefaultConfig(addr)

	// Set transport
	if *useUDP {
		cfg.Transport = meclient.TransportUDP
		cfg.AutoReconnect = false // No reconnect for UDP
	}

	// Set protocol
	if *useBinary {
		cfg.Protocol = meclient.ProtocolBinary
	} else if *useUDP {
		// UDP defaults to CSV (no auto-detect possible)
		cfg.Protocol = meclient.ProtocolCSV
	}
	// TCP defaults to ProtocolAuto (will probe)

	client, err := meclient.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Print connection info
	transportStr := "TCP"
	if *useUDP {
		transportStr = "UDP"
	}
	protocolStr := "auto-detect"
	if cfg.Protocol == meclient.ProtocolCSV {
		protocolStr = "CSV"
	} else if cfg.Protocol == meclient.ProtocolBinary {
		protocolStr = "binary"
	}
	fmt.Printf("Connecting to %s via %s (protocol: %s)...\n", addr, transportStr, protocolStr)

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Connected!\n\n")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start background receiver for interactive/wait modes
	if scenarioID == 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			receiveMessages(client, done)
		}()
	}

	// Run mode
	if scenarioID > 0 {
		// Run specific scenario
		runner := scenarios.NewRunner(client, uint32(*userID), *verbose)
		result, err := runner.Run(scenarioID, *dangerBurst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scenario error: %v\n", err)
			if !scenarios.IsValid(scenarioID) {
				fmt.Println()
				scenarios.PrintList()
			}
			os.Exit(1)
		}
		if result != nil && scenarioID < 10 { // Only print for basic scenarios
			result.Print()
		}
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
			fmt.Printf("Sending CANCEL user=%d oid=%d\n", cancel.UserID, cancel.OrderID)
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
	if len(parts) < 2 {
		return meclient.CancelOrder{}, fmt.Errorf("usage: cancel ORDER_ID [USER_ID]")
	}

	orderID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return meclient.CancelOrder{}, fmt.Errorf("invalid order_id: %s", parts[1])
	}

	userID := defaultUserID
	if len(parts) >= 3 {
		u, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			return meclient.CancelOrder{}, fmt.Errorf("invalid user_id: %s", parts[2])
		}
		userID = uint32(u)
	}

	return meclient.CancelOrder{
		UserID:  userID,
		OrderID: uint32(orderID),
	}, nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  meclient HOST PORT [SCENARIO]        Run a scenario")
	fmt.Println("  meclient HOST PORT -i                Interactive mode")
	fmt.Println("  meclient -list                       List available scenarios")
	fmt.Println()
	fmt.Println("Transport Options:")
	fmt.Println("  -udp                Use UDP transport (default: TCP)")
	fmt.Println()
	fmt.Println("Protocol Options:")
	fmt.Println("  -binary             Force binary protocol")
	fmt.Println("                      Default: TCP = auto-detect, UDP = CSV")
	fmt.Println()
	fmt.Println("Other Options:")
	fmt.Println("  -user uint          Default user ID (default: 1)")
	fmt.Println("  -v                  Verbose output")
	fmt.Println("  -danger-burst       Allow unthrottled burst scenarios")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  meclient localhost 1234              Connect via TCP")
	fmt.Println("  meclient localhost 1234 1            Run scenario 1 (simple orders)")
	fmt.Println("  meclient localhost 1234 -udp         Connect via UDP (CSV)")
	fmt.Println("  meclient localhost 1234 -udp -binary Connect via UDP with binary")
	fmt.Println("  meclient localhost 1234 -i           Interactive mode")
	fmt.Println()
	fmt.Println("Run 'meclient -list' to see all available scenarios.")
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  buy SYMBOL QTY PRICE [USER_ID]    Place buy order")
	fmt.Println("  sell SYMBOL QTY PRICE [USER_ID]   Place sell order")
	fmt.Println("  cancel ORDER_ID [USER_ID]         Cancel order")
	fmt.Println("  flush                             Flush all order books")
	fmt.Println("  status                            Show connection status")
	fmt.Println("  help                              Show this help")
	fmt.Println("  quit                              Exit")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  buy IBM 100 150      Buy 100 IBM @ 150")
	fmt.Println("  sell AAPL 50 175     Sell 50 AAPL @ 175")
	fmt.Println("  cancel 1001          Cancel order 1001")
	fmt.Println()
	fmt.Println("Shortcuts: b=buy, s=sell, c=cancel, f=flush, h=help, q=quit")
}
