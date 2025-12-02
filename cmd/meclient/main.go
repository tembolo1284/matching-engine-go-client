// Full path: cmd/meclient/main.go

// Command meclient is a Go client for the matching engine server.
package main

import (
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
	args := os.Args[1:]

	// Check for -list or -help first
	for _, arg := range args {
		if arg == "-list" || arg == "--list" {
			scenarios.PrintList()
			return
		}
		if arg == "-h" || arg == "-help" || arg == "--help" {
			printUsage()
			return
		}
	}

	// Parse arguments
	opts := parseArgs(args)

	// Validate: need host, port, and either scenario or interactive mode
	if opts.host == "" || opts.port == "" {
		printUsage()
		os.Exit(1)
	}

	// If no scenario and not interactive, show usage and exit
	if opts.scenarioID == 0 && !opts.interactive {
		fmt.Println("Error: specify a scenario ID or use -i for interactive mode")
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%s", opts.host, opts.port)

	fmt.Printf("Matching Engine Go Client\n")
	fmt.Printf("=========================\n\n")

	// Connect
	client, err := connect(addr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connected!\n\n")

	// Setup shutdown handler
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start background receiver
	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveMessages(client, done)
	}()

	// Run the appropriate mode
	runMode(client, opts, shutdown)

	// Cleanup
	fmt.Println("\nShutting down...")
	close(done)
	client.Close()
	wg.Wait()

	printStats(client)
	fmt.Println("Goodbye!")
}

// options holds parsed command-line options.
type options struct {
	host        string
	port        string
	scenarioID  int
	interactive bool
	verbose     bool
	dangerBurst bool
	useUDP      bool
	useTCP      bool
	useBinary   bool
	userID      uint32
}

func parseArgs(args []string) options {
	opts := options{userID: 1}

	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flag := strings.TrimPrefix(strings.TrimPrefix(arg, "-"), "-")

			switch flag {
			case "i", "interactive":
				opts.interactive = true
			case "v", "verbose":
				opts.verbose = true
			case "danger-burst":
				opts.dangerBurst = true
			case "udp":
				opts.useUDP = true
			case "tcp":
				opts.useTCP = true
			case "binary":
				opts.useBinary = true
			case "user":
				if i+1 < len(args) {
					i++
					if u, err := strconv.ParseUint(args[i], 10, 32); err == nil {
						opts.userID = uint32(u)
					}
				}
			}
		} else {
			positional = append(positional, arg)
		}
	}

	if len(positional) >= 1 {
		opts.host = positional[0]
	}
	if len(positional) >= 2 {
		opts.port = positional[1]
	}
	if len(positional) >= 3 {
		if id, err := strconv.Atoi(positional[2]); err == nil {
			opts.scenarioID = id
		}
	}

	return opts
}

func connect(addr string, opts options) (*meclient.Client, error) {
	if opts.useUDP {
		return connectWithTransport(addr, meclient.TransportUDP, opts.useBinary)
	}
	if opts.useTCP {
		return connectWithTransport(addr, meclient.TransportTCP, opts.useBinary)
	}
	return connectWithFallback(addr, opts.useBinary)
}

func runMode(client *meclient.Client, opts options, shutdown <-chan os.Signal) {
	if opts.scenarioID > 0 {
		runner := scenarios.NewRunner(client, opts.userID, opts.verbose)
		result, err := runner.Run(opts.scenarioID, opts.dangerBurst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scenario error: %v\n", err)
			if !scenarios.IsValid(opts.scenarioID) {
				fmt.Println()
				scenarios.PrintList()
			}
			os.Exit(1)
		}
		if result != nil && opts.scenarioID < 10 {
			result.Print()
		}
	} else if opts.interactive {
		runInteractive(client, opts.userID, shutdown)
	}
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

func printStats(client *meclient.Client) {
	stats := client.Stats()
	fmt.Printf("\nSession Stats:\n")
	fmt.Printf("  Messages Sent:     %d\n", stats.MessagesSent)
	fmt.Printf("  Messages Received: %d\n", stats.MessagesReceived)
	fmt.Printf("  Errors:            %d\n", stats.ErrorCount)
	fmt.Printf("  Reconnects:        %d\n", stats.ReconnectCount)
	fmt.Printf("  Dropped Messages:  %d\n", stats.DroppedMessages)
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  meclient HOST PORT SCENARIO [OPTIONS]   Run a scenario")
	fmt.Println("  meclient HOST PORT -i [OPTIONS]         Interactive mode")
	fmt.Println("  meclient -list                          List scenarios")
	fmt.Println("  meclient -help                          Show this help")
	fmt.Println()
	fmt.Println("Transport Options:")
	fmt.Println("  -tcp                TCP only (no UDP fallback)")
	fmt.Println("  -udp                UDP only")
	fmt.Println("  (default)           Auto-detect: try TCP, fall back to UDP")
	fmt.Println()
	fmt.Println("Protocol Options:")
	fmt.Println("  -binary             Use binary protocol (default: CSV)")
	fmt.Println()
	fmt.Println("Other Options:")
	fmt.Println("  -v                  Verbose output")
	fmt.Println("  -user N             Set user ID (default: 1)")
	fmt.Println("  -danger-burst       Allow unthrottled burst scenarios")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  meclient localhost 1234 1            # Run scenario 1")
	fmt.Println("  meclient localhost 1234 1 -udp       # Scenario 1 via UDP")
	fmt.Println("  meclient localhost 1234 -i           # Interactive mode")
	fmt.Println("  meclient localhost 1234 -i -udp      # Interactive via UDP")
	fmt.Println("  meclient -list                       # List all scenarios")
}
