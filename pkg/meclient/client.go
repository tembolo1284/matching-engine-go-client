// Full path: pkg/meclient/client.go

// Package meclient provides a Go client for the matching engine server.
package meclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/config"
	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/internal/stats"
	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/protocol"
	"github.com/tembolo1284/matching-engine-go-client/pkg/meclient/transport"
)

// Re-export types from subpackages for convenient access
type (
	Config         = config.Config
	Transport      = config.Transport
	Protocol       = config.Protocol
	Side           = protocol.Side
	NewOrder       = protocol.NewOrder
	CancelOrder    = protocol.CancelOrder
	Ack            = protocol.Ack
	Trade          = protocol.Trade
	BookUpdate     = protocol.BookUpdate
	CancelAck      = protocol.CancelAck
	ReconnectEvent = protocol.ReconnectEvent
	StatsSnapshot  = stats.Snapshot
)

// Re-export constants
const (
	SideBuy  = protocol.SideBuy
	SideSell = protocol.SideSell

	TransportTCP = config.TransportTCP
	TransportUDP = config.TransportUDP

	ProtocolAuto   = config.ProtocolAuto
	ProtocolCSV    = config.ProtocolCSV
	ProtocolBinary = config.ProtocolBinary

	DefaultPort = config.DefaultPort
)

// Re-export config functions
var (
	DefaultConfig = config.Default
)

// Re-export errors
var (
	ErrInvalidConfig = config.ErrInvalidConfig
	ErrEmptySymbol   = protocol.ErrEmptySymbol
	ErrSymbolTooLong = protocol.ErrSymbolTooLong
	ErrZeroQuantity  = protocol.ErrZeroQuantity
	ErrInvalidSide   = protocol.ErrInvalidSide
)

// Client-specific errors
var (
	ErrClientClosed   = errors.New("client closed")
	ErrNotConnected   = errors.New("not connected")
	ErrWriteQueueFull = errors.New("write queue full")
	ErrChannelFull    = errors.New("channel full, message dropped")
	ErrMaxReconnects  = errors.New("maximum reconnection attempts exceeded")
)

// Internal write request types
type writeRequestType uint8

const (
	writeRequestOrder writeRequestType = iota
	writeRequestCancel
	writeRequestFlush
)

type writeRequest struct {
	reqType writeRequestType
	order   protocol.NewOrder
	cancel  protocol.CancelOrder
}

// FlushableTransport extends transport with Flush capability
type FlushableTransport interface {
	transport.Transport
	Flush() error
}

// Client is a client for the matching engine.
type Client struct {
	cfg config.Config

	// Transport
	transport transport.Transport

	// Protocol
	encoder *protocol.Encoder

	// Write path
	writeCh chan writeRequest

	// Output channels
	ackCh        chan protocol.Ack
	tradeCh      chan protocol.Trade
	bookUpdateCh chan protocol.BookUpdate
	cancelAckCh  chan protocol.CancelAck
	errorCh      chan error
	reconnectCh  chan protocol.ReconnectEvent

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	stats stats.Stats
}

// New creates a new client with the given configuration.
func New(cfg config.Config) (*Client, error) {
	cfg = config.ApplyDefaults(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		cfg:          cfg,
		writeCh:      make(chan writeRequest, cfg.ChannelBuffer),
		ackCh:        make(chan protocol.Ack, cfg.ChannelBuffer),
		tradeCh:      make(chan protocol.Trade, cfg.ChannelBuffer),
		bookUpdateCh: make(chan protocol.BookUpdate, cfg.ChannelBuffer),
		cancelAckCh:  make(chan protocol.CancelAck, cfg.ChannelBuffer),
		errorCh:      make(chan error, cfg.ChannelBuffer),
		reconnectCh:  make(chan protocol.ReconnectEvent, 16),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Connect establishes a connection to the server and starts background goroutines.
func (c *Client) Connect() error {
	if c.ctx.Err() != nil {
		return ErrClientClosed
	}

	// Create transport
	c.transport = transport.New(&c.cfg)

	if err := c.transport.Connect(); err != nil {
		return err
	}

	// Create encoder
	c.encoder = protocol.NewEncoder(c.transport.Writer())

	c.wg.Add(2)
	go c.readLoop()
	go c.writeLoop()

	return nil
}

// Close gracefully shuts down the client.
func (c *Client) Close() error {
	c.cancel()

	if c.transport != nil {
		_ = c.transport.Close()
	}

	c.wg.Wait()
	c.closeChannels()

	return nil
}

func (c *Client) closeChannels() {
	close(c.ackCh)
	close(c.tradeCh)
	close(c.bookUpdateCh)
	close(c.cancelAckCh)
	close(c.errorCh)
	close(c.reconnectCh)
}

// SendOrder sends a new order to the matching engine.
func (c *Client) SendOrder(order protocol.NewOrder) error {
	if err := protocol.ValidateOrder(&order); err != nil {
		return err
	}

	return c.enqueueWrite(writeRequest{reqType: writeRequestOrder, order: order})
}

// SendCancel sends a cancel request to the matching engine.
func (c *Client) SendCancel(cancel protocol.CancelOrder) error {
	if err := protocol.ValidateCancel(&cancel); err != nil {
		return err
	}

	return c.enqueueWrite(writeRequest{reqType: writeRequestCancel, cancel: cancel})
}

// SendFlush sends a flush command to clear all order books.
func (c *Client) SendFlush() error {
	return c.enqueueWrite(writeRequest{reqType: writeRequestFlush})
}

func (c *Client) enqueueWrite(req writeRequest) error {
	if c.ctx.Err() != nil {
		return ErrClientClosed
	}

	select {
	case <-c.ctx.Done():
		return ErrClientClosed
	case c.writeCh <- req:
		c.stats.IncMessagesSent()
		return nil
	default:
		c.stats.IncDroppedMessages()
		return ErrWriteQueueFull
	}
}

// Channel accessors
func (c *Client) Acks() <-chan protocol.Ack               { return c.ackCh }
func (c *Client) Trades() <-chan protocol.Trade           { return c.tradeCh }
func (c *Client) BookUpdates() <-chan protocol.BookUpdate { return c.bookUpdateCh }
func (c *Client) CancelAcks() <-chan protocol.CancelAck   { return c.cancelAckCh }
func (c *Client) Errors() <-chan error                    { return c.errorCh }
func (c *Client) Reconnects() <-chan protocol.ReconnectEvent { return c.reconnectCh }

// IsConnected returns true if the client is currently connected.
func (c *Client) IsConnected() bool {
	if c.transport == nil {
		return false
	}
	return c.transport.IsConnected()
}

// Stats returns a snapshot of the current client statistics.
func (c *Client) Stats() stats.Snapshot {
	return c.stats.GetSnapshot()
}

// readLoop continuously reads messages from the server.
func (c *Client) readLoop() {
	defer c.wg.Done()

	consecutiveErrors := 0

	for {
		if consecutiveErrors >= config.MaxConsecutiveErrors {
			c.sendError(fmt.Errorf("max consecutive errors (%d) exceeded", config.MaxConsecutiveErrors))
			return
		}

		if c.ctx.Err() != nil {
			return
		}

		if c.transport == nil || !c.transport.IsConnected() {
			if !c.waitForReconnect() {
				return
			}
			continue
		}

		err := c.processInboundMessages()
		if err != nil {
			consecutiveErrors++
			if !c.handleReadError(err) {
				return
			}
			consecutiveErrors = 0
		}
	}
}

func (c *Client) processInboundMessages() error {
	reader := c.transport.Reader()
	if reader == nil {
		return errors.New("no reader available")
	}

	decoder := protocol.NewDecoder(reader)
	batchCount := 0

	for {
		if batchCount >= config.MaxMessageBatchSize {
			if c.ctx.Err() != nil {
				return c.ctx.Err()
			}
			batchCount = 0
		}

		msg, err := decoder.Decode()
		if err != nil {
			if err == io.EOF {
				return errors.New("connection closed by server")
			}
			c.sendError(fmt.Errorf("decode error: %w", err))
			return err
		}

		c.dispatchMessage(msg)
		c.stats.IncMessagesReceived()
		batchCount++
	}
}

func (c *Client) handleReadError(err error) bool {
	if c.ctx.Err() != nil {
		return false
	}

	c.sendError(fmt.Errorf("read error: %w", err))
	c.stats.IncErrorCount()

	if c.cfg.AutoReconnect {
		return c.reconnect()
	}
	return false
}

func (c *Client) dispatchMessage(msg *protocol.Message) {
	switch {
	case msg.Ack != nil:
		c.trySendAck(*msg.Ack)
	case msg.Trade != nil:
		c.trySendTrade(*msg.Trade)
	case msg.BookUpdate != nil:
		c.trySendBookUpdate(*msg.BookUpdate)
	case msg.CancelAck != nil:
		c.trySendCancelAck(*msg.CancelAck)
	}
}

func (c *Client) trySendAck(v protocol.Ack) {
	select {
	case c.ackCh <- v:
	default:
		c.stats.IncDroppedMessages()
		c.sendError(ErrChannelFull)
	}
}

func (c *Client) trySendTrade(v protocol.Trade) {
	select {
	case c.tradeCh <- v:
	default:
		c.stats.IncDroppedMessages()
		c.sendError(ErrChannelFull)
	}
}

func (c *Client) trySendBookUpdate(v protocol.BookUpdate) {
	select {
	case c.bookUpdateCh <- v:
	default:
		c.stats.IncDroppedMessages()
		c.sendError(ErrChannelFull)
	}
}

func (c *Client) trySendCancelAck(v protocol.CancelAck) {
	select {
	case c.cancelAckCh <- v:
	default:
		c.stats.IncDroppedMessages()
		c.sendError(ErrChannelFull)
	}
}

func (c *Client) sendError(err error) {
	select {
	case c.errorCh <- err:
	default:
	}
}

// writeLoop processes outbound messages.
func (c *Client) writeLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case req := <-c.writeCh:
			if err := c.processWrite(req); err != nil {
				c.sendError(fmt.Errorf("write error: %w", err))
				c.stats.IncErrorCount()
			}
		}
	}
}

func (c *Client) processWrite(req writeRequest) error {
	if c.encoder == nil {
		return ErrNotConnected
	}

	var err error

	switch req.reqType {
	case writeRequestOrder:
		err = c.encoder.EncodeNewOrder(&req.order)
	case writeRequestCancel:
		err = c.encoder.EncodeCancel(&req.cancel)
	case writeRequestFlush:
		err = c.encoder.EncodeFlush()
	}

	if err != nil {
		return err
	}

	// Flush the transport if it supports it
	if ft, ok := c.transport.(FlushableTransport); ok {
		return ft.Flush()
	}

	return nil
}

// reconnect attempts to reconnect with exponential backoff.
func (c *Client) reconnect() bool {
	delay := c.cfg.ReconnectMinDelay

	for attempt := 1; attempt <= config.MaxReconnectAttempts; attempt++ {
		select {
		case <-c.ctx.Done():
			return false
		case <-time.After(delay):
		}

		// Close existing transport
		if c.transport != nil {
			_ = c.transport.Close()
		}

		// Create new transport
		c.transport = transport.New(&c.cfg)

		if err := c.transport.Connect(); err != nil {
			c.sendError(fmt.Errorf("reconnect attempt %d failed: %w", attempt, err))

			delay *= 2
			if delay > c.cfg.ReconnectMaxDelay {
				delay = c.cfg.ReconnectMaxDelay
			}
			continue
		}

		// Recreate encoder
		c.encoder = protocol.NewEncoder(c.transport.Writer())

		c.stats.IncReconnectCount()

		select {
		case c.reconnectCh <- protocol.ReconnectEvent{Attempt: attempt}:
		default:
		}

		return true
	}

	c.sendError(ErrMaxReconnects)
	return false
}

// waitForReconnect waits until a connection is available or the client closes.
func (c *Client) waitForReconnect() bool {
	ticker := time.NewTicker(config.ReconnectCheckInterval)
	defer ticker.Stop()

	for iterations := 0; iterations < config.MaxReconnectAttempts; iterations++ {
		select {
		case <-c.ctx.Done():
			return false
		case <-ticker.C:
			if c.transport != nil && c.transport.IsConnected() {
				return true
			}
		}
	}

	return false
}
