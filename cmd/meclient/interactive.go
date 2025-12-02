// Full path: cmd/meclient/interactive.go

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient"
)

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
