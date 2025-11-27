# Architecture Guide

## Design Principles

1. **Idiomatic Go**: Channels for async, goroutines for concurrency
2. **Zero dependencies**: Standard library only
3. **Stateless client**: Server is source of truth
4. **Non-blocking sends**: Never block the caller
5. **Cache-friendly**: Padded stats prevent false sharing

## Component Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                           Client                                 │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                     Public API                            │   │
│  │  New() Connect() Close() SendOrder() SendCancel()        │   │
│  │  Acks() Trades() BookUpdates() CancelAcks() Stats()      │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│              ┌───────────────┴───────────────┐                  │
│              ▼                               ▼                  │
│  ┌─────────────────────┐        ┌─────────────────────┐        │
│  │     Write Path      │        │      Read Path      │        │
│  │                     │        │                     │        │
│  │  writeCh (buffered) │        │  decoder            │        │
│  │       │             │        │       │             │        │
│  │       ▼             │        │       ▼             │        │
│  │  writeLoop()        │        │  readLoop()         │        │
│  │       │             │        │       │             │        │
│  │       ▼             │        │       ▼             │        │
│  │  encoder (reuses    │        │  dispatchMessage()  │        │
│  │   buffer)           │        │       │             │        │
│  └──────────┬──────────┘        └───────┼─────────────┘        │
│             │                           │                       │
│             └───────────┬───────────────┘                       │
│                         ▼                                       │
│               ┌─────────────────┐                               │
│               │  net.Conn (TCP) │                               │
│               └─────────────────┘                               │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ClientStats (cache-line padded)                          │   │
│  │ messagesSent [pad] messagesReceived [pad] errors [pad]   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Cache-Line Optimization

From the LinkedIn posts about false sharing - we pad each stat counter to its own cache line:

```go
type ClientStats struct {
    messagesSent     uint64
    _pad1            [56]byte  // Pad to 64 bytes
    messagesReceived uint64
    _pad2            [56]byte
    // ...
}
```

This prevents cache line bouncing when multiple goroutines update different counters.

## Memory Efficiency

**Pre-allocated encoder buffer:**
```go
type encoder struct {
    w   io.Writer
    buf []byte  // Reused across encodes
}
```

**Atomic stats (no locks):**
```go
func (s *ClientStats) incMessagesSent() {
    atomic.AddUint64(&s.messagesSent, 1)
}
```

## Bounded Operations

All loops have fixed upper bounds:

| Constant | Value | Purpose |
|----------|-------|---------|
| MaxReconnectAttempts | 1000 | Prevent infinite reconnect |
| MaxConsecutiveErrors | 100 | Force disconnect on repeated failures |
| MaxMessageBatchSize | 1000 | Periodic shutdown check |

## Wire Protocol

**Outbound (CSV):**
```
N, user_id, symbol, price, qty, side, order_id
C, symbol, user_id, order_id
F
```

**Inbound (CSV):**
```
A, user_id, order_id
T, buy_user, buy_oid, sell_user, sell_oid, price, qty
B, symbol, side, price, qty
X, user_id, order_id
```

## Reconnection Strategy

```
Disconnect → Wait(delay) → Dial → Success? → Resume
                ↑                    │
                └── delay *= 2 ──────┘ (on failure)
```

Backoff: 100ms → 200ms → 400ms → ... → 30s max

## Thread Safety

| Method | Thread-Safe |
|--------|-------------|
| SendOrder/Cancel/Flush | Yes |
| Connect | No (call once) |
| Close | No (call once) |
| IsConnected | Yes |
| Stats | Yes |
| Channel reads | Yes |

## Input Validation

Orders are validated before queuing:

- Symbol: non-empty, ≤ 16 chars
- Quantity: > 0
- Side: B or S

Invalid orders return immediately with descriptive errors.
