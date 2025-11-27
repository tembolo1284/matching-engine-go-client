# Quick Start Guide

Get up and running in 5 minutes.

## Prerequisites

- Go 1.21+
- Running matching engine server

## Step 1: Start the Server

```bash
cd matching-engine-c/build
./matching_engine --mode tcp --port 12345
```

## Step 2: Build the Client

```bash
cd matching-engine-go
go build ./...
```

## Step 3: Run the Example

```bash
cd cmd/example
go build
./example -addr localhost:12345 -demo
```

You should see:

```
Matching Engine Go Client
=========================

Connecting to localhost:12345...
Connected!

Running demo sequence...

=== Placing Buy Orders ===
-> BUY IBM 100 @ 150 (oid=1)
[ACK] user=1 order=1
...
```

## Step 4: Interactive Mode

```bash
./example -addr localhost:12345 -interactive
```

Commands:

```
> buy IBM 100 150
> sell IBM 50 150
> cancel IBM 1001
> status
> quit
```

## Step 5: Write Your Own Client

```go
package main

import (
    "fmt"
    "github.com/pauljunsukhan/matching-engine-go/pkg/meclient"
)

func main() {
    cfg := meclient.DefaultConfig("localhost:12345")
    client, _ := meclient.New(cfg)
    client.Connect()
    defer client.Close()

    go func() {
        for trade := range client.Trades() {
            fmt.Printf("Trade: %d @ %d\n", trade.Qty, trade.Price)
        }
    }()

    client.SendOrder(meclient.NewOrder{
        Symbol:  "TEST",
        Price:   100,
        Qty:     10,
        Side:    meclient.SideBuy,
        OrderID: 1,
    })

    select {}
}
```

## Connecting from Another Machine

1. Ensure server binds to `0.0.0.0` (not `127.0.0.1`)
2. Find server IP: `hostname -I` (Linux) or `ipconfig getifaddr en0` (macOS)
3. Open firewall port: `sudo ufw allow 12345/tcp`
4. Connect: `./example -addr 192.168.1.100:12345`

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Connection refused" | Start the server, check port |
| "Timeout" | Check IP address, firewall |
| No messages received | Drain all channels |
