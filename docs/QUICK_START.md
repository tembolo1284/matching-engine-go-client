# Quick Start Guide

Get up and running in 5 minutes.

## Prerequisites

- Go 1.21+
- Running matching engine server (C implementation)

## Step 1: Start the Matching Engine Server
```bash
cd matching-engine/build
./matching_engine 1234
```

The server listens on port 1234 by default.

## Step 2: Build the Go Client
```bash
cd matching-engine-go-client
make build
```

This creates the binary at `./bin/meclient`.

## Step 3: Run a Basic Scenario

Using make:
```bash
make scenario S=1
```

Or directly:
```bash
./bin/meclient localhost 1234 1
```

You should see:
```
Matching Engine Go Client
=========================

Connecting to localhost:1234 via TCP (protocol: auto-detect)...
Connected!

=== Scenario 1: Simple Orders ===

Sending: BUY IBM 50@100
[ACK] IBM user=1 order=1

Sending: SELL IBM 50@105
[ACK] IBM user=1 order=2

Sending: FLUSH
...
```

## Step 4: Interactive Mode

Using make:
```bash
make run-i
```

Or directly:
```bash
./bin/meclient localhost 1234 -i
```

Commands:
```
> buy IBM 100 150       # Buy 100 IBM @ 150
> sell AAPL 50 175      # Sell 50 AAPL @ 175  
> cancel 1001           # Cancel order 1001
> flush                 # Flush all order books
> status                # Show connection status
> help                  # Show all commands
> quit                  # Exit
```

Shortcuts: `b`=buy, `s`=sell, `c`=cancel, `f`=flush, `h`=help, `q`=quit

## Step 5: List Available Scenarios
```bash
make list
```

Or:
```bash
./bin/meclient -list
```

Output:
```
Available scenarios:

Basic:
  1   - Simple orders (no match)
  2   - Matching trade execution
  3   - Cancel order

Stress Tests:
  10  - Stress: 1K orders
  11  - Stress: 10K orders

Matching Stress (generates trades):
  20  - Matching: 1K pairs (2K orders)
  21  - Matching: 10K pairs
```

## Step 6: Run Stress Tests

1K orders:
```bash
make stress-1k
# or
./bin/meclient localhost 1234 10
```

10K orders:
```bash
make stress-10k
# or
./bin/meclient localhost 1234 11
```

## Step 7: Write Your Own Client
```go
package main

import (
    "fmt"
    "github.com/tembolo1284/matching-engine-go-client/pkg/meclient"
)

func main() {
    // Create client
    cfg := meclient.DefaultConfig("localhost:1234")
    client, err := meclient.New(cfg)
    if err != nil {
        panic(err)
    }
    
    // Connect
    if err := client.Connect(); err != nil {
        panic(err)
    }
    defer client.Close()

    // Handle responses in background
    go func() {
        for {
            select {
            case ack := <-client.Acks():
                fmt.Printf("ACK: %s order=%d\n", ack.Symbol, ack.OrderID)
            case trade := <-client.Trades():
                fmt.Printf("TRADE: %s %d @ %d\n", trade.Symbol, trade.Qty, trade.Price)
            case err := <-client.Errors():
                fmt.Printf("ERROR: %v\n", err)
            }
        }
    }()

    // Send an order
    err = client.SendOrder(meclient.NewOrder{
        UserID:  1,
        Symbol:  "IBM",
        Price:   100,
        Qty:     50,
        Side:    meclient.SideBuy,
        OrderID: 1,
    })
    if err != nil {
        panic(err)
    }

    // Wait for response
    select {}
}
```

## Command Reference

### Make Targets

| Command | Description |
|---------|-------------|
| `make build` | Build the client binary |
| `make run` | Connect and listen for messages |
| `make run-i` | Interactive mode |
| `make list` | List available scenarios |
| `make scenario S=N` | Run scenario N |
| `make stress-1k` | Run 1K order stress test |
| `make stress-10k` | Run 10K order stress test |
| `make test` | Run all tests |
| `make clean` | Remove build artifacts |

### Direct Commands
```bash
# Basic connection
./bin/meclient localhost 1234

# Run scenario
./bin/meclient localhost 1234 <scenario_id>

# Interactive mode
./bin/meclient localhost 1234 -i

# With options
./bin/meclient localhost 1234 -v           # Verbose output
./bin/meclient localhost 1234 -user 5      # Set user ID
./bin/meclient localhost 1234 -udp         # Use UDP transport
./bin/meclient localhost 1234 -binary      # Force binary protocol

# List scenarios
./bin/meclient -list
```

## Connecting from Another Machine

1. Ensure server binds to `0.0.0.0`:
```bash
   ./matching_engine 1234  # Should bind to all interfaces
```

2. Find server IP:
```bash
   # Linux
   hostname -I
   
   # macOS
   ipconfig getifaddr en0
```

3. Open firewall port (if needed):
```bash
   sudo ufw allow 1234/tcp
```

4. Connect from client machine:
```bash
   ./bin/meclient 192.168.1.100 1234 -i
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Connection refused" | Start the server, verify port 1234 |
| "Timeout" | Check IP address and firewall settings |
| No messages received | Ensure you're draining all response channels |
| "Invalid scenario" | Run `./bin/meclient -list` to see valid IDs |
| Build errors | Run `go mod tidy` then `make build` |

## Next Steps

- Read [ARCHITECTURE.md](ARCHITECTURE.md) for design details
- Read [BUILD.md](BUILD.md) for build options
- Check `pkg/meclient/` for the client library API
- Check `pkg/scenarios/` for scenario implementations
