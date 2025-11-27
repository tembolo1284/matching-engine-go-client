package meclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
	order   NewOrder
	cancel  CancelOrder
}

// Client is a TCP client for the matching engine.
type Client struct {
	cfg Config

	// Connection
	conn   net.Conn
	connMu sync.RWMutex

	// Write path
	writeCh chan writeRequest
	writer  *bufio.Writer
	encoder *encoder

	// Output channels
	ackCh        chan Ack
	tradeCh      chan Trade
	bookUpdateCh chan BookUpdate
	cancelAckCh  chan CancelAck
	errorCh      chan error
	reconnectCh  chan ReconnectEvent

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// State
	connected   bool
	connectedMu sync.RWMutex

	// Metrics (cache-line padded)
	stats ClientStats
}

// New creates a new client with the given configuration.
func New(cfg Config) (*Client, error) {
	cfg = applyDefaults(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		cfg:          cfg,
		writeCh:      make(chan writeRequest, cfg.ChannelBuffer),
		ackCh:        make(chan Ack, cfg.ChannelBuffer),
		tradeCh:      make(chan Trade, cfg.ChannelBuffer),
		bookUpdateCh: make(chan BookUpdate, cfg.ChannelBuffer),
		cancelAckCh:  make(chan CancelAck, cfg.ChannelBuffer),
		errorCh:      make(chan error, cfg.ChannelBuffer),
		reconnectCh:  make(chan ReconnectEvent, 16),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Connect establishes a connection to the server and starts background goroutines.
func (c *Client) Connect() error {
	if c.ctx.Err() != nil {
		return ErrClientClosed
	}

	if err := c.dial(); err != nil {
		return err
	}

	c.wg.Add(2)
	go c.readLoop()
	go c.writeLoop()

	return nil
}

func (c *Client) dial() error {
	dialer := net.Dialer{Timeout: c.cfg.ConnectTimeout}

	conn, err := dialer.DialContext(c.ctx, "tcp", c.cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.cfg.Address, err)
	}

	c.connMu.Lock()
	c.conn = conn
	c.writer = bufio.NewWriterSize(conn, DefaultWriteBuffer)
	c.encoder = newEncoder(c.writer)
	c.connMu.Unlock()

	c.setConnected(true)
	return nil
}

// Close gracefully shuts down the client.
func (c *Client) Close() error {
	c.cancel()

	c.connMu.Lock()
	conn := c.conn
	c.conn = nil
	c.connMu.Unlock()

	c.setConnected(false)

	if conn != nil {
		_ = conn.Close()
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
func (c *Client) SendOrder(order NewOrder) error {
	if err := validateOrder(&order); err != nil {
		return err
	}

	return c.enqueueWrite(writeRequest{reqType: writeRequestOrder, order: order})
}

// SendCancel sends a cancel request to the matching engine.
func (c *Client) SendCancel(cancel CancelOrder) error {
	if err := validateCancel(&cancel); err != nil {
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
		c.stats.incMessagesSent()
		return nil
	default:
		c.stats.incDroppedMessages()
		return ErrWriteQueueFull
	}
}

// Channel accessors
func (c *Client) Acks() <-chan Ack                  { return c.ackCh }
func (c *Client) Trades() <-chan Trade              { return c.tradeCh }
func (c *Client) BookUpdates() <-chan BookUpdate    { return c.bookUpdateCh }
func (c *Client) CancelAcks() <-chan CancelAck      { return c.cancelAckCh }
func (c *Client) Errors() <-chan error              { return c.errorCh }
func (c *Client) Reconnects() <-chan ReconnectEvent { return c.reconnectCh }

// IsConnected returns true if the client is currently connected.
func (c *Client) IsConnected() bool {
	c.connectedMu.RLock()
	defer c.connectedMu.RUnlock()
	return c.connected
}

// Stats returns a snapshot of the current client statistics.
func (c *Client) Stats() StatsSnapshot {
	return c.stats.Snapshot()
}

func (c *Client) setConnected(v bool) {
	c.connectedMu.Lock()
	c.connected = v
	c.connectedMu.Unlock()
}

func (c *Client) getConn() net.Conn {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.conn
}

// readLoop continuously reads messages from the server.
func (c *Client) readLoop() {
	defer c.wg.Done()

	consecutiveErrors := 0

	for {
		if consecutiveErrors >= MaxConsecutiveErrors {
			c.sendError(fmt.Errorf("max consecutive errors (%d) exceeded", MaxConsecutiveErrors))
			return
		}

		if c.ctx.Err() != nil {
			return
		}

		conn := c.getConn()
		if conn == nil {
			if !c.waitForReconnect() {
				return
			}
			continue
		}

		err := c.processInboundMessages(conn)
		if err != nil {
			consecutiveErrors++
			if !c.handleReadError(err) {
				return
			}
			consecutiveErrors = 0
		}
	}
}

func (c *Client) processInboundMessages(conn net.Conn) error {
	decoder := newDecoder(bufio.NewReaderSize(conn, DefaultReadBuffer))
	batchCount := 0

	for {
		if batchCount >= MaxMessageBatchSize {
			if c.ctx.Err() != nil {
				return c.ctx.Err()
			}
			batchCount = 0
		}

		msg, err := decoder.decode()
		if err != nil {
			if err == io.EOF {
				return errors.New("connection closed by server")
			}
			return err
		}

		c.dispatchMessage(msg)
		c.stats.incMessagesReceived()
		batchCount++
	}
}

func (c *Client) handleReadError(err error) bool {
	if c.ctx.Err() != nil {
		return false
	}

	c.setConnected(false)
	c.sendError(fmt.Errorf("read error: %w", err))
	c.stats.incErrorCount()

	if c.cfg.AutoReconnect {
		return c.reconnect()
	}
	return false
}

func (c *Client) dispatchMessage(msg *message) {
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

func (c *Client) trySendAck(v Ack) {
	select {
	case c.ackCh <- v:
	default:
		c.stats.incDroppedMessages()
		c.sendError(ErrChannelFull)
	}
}

func (c *Client) trySendTrade(v Trade) {
	select {
	case c.tradeCh <- v:
	default:
		c.stats.incDroppedMessages()
		c.sendError(ErrChannelFull)
	}
}

func (c *Client) trySendBookUpdate(v BookUpdate) {
	select {
	case c.bookUpdateCh <- v:
	default:
		c.stats.incDroppedMessages()
		c.sendError(ErrChannelFull)
	}
}

func (c *Client) trySendCancelAck(v CancelAck) {
	select {
	case c.cancelAckCh <- v:
	default:
		c.stats.incDroppedMessages()
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
				c.stats.incErrorCount()
			}
		}
	}
}

func (c *Client) processWrite(req writeRequest) error {
	c.connMu.RLock()
	encoder := c.encoder
	writer := c.writer
	c.connMu.RUnlock()

	if encoder == nil || writer == nil {
		return ErrNotConnected
	}

	var err error

	switch req.reqType {
	case writeRequestOrder:
		err = encoder.encodeNewOrder(&req.order)
	case writeRequestCancel:
		err = encoder.encodeCancel(&req.cancel)
	case writeRequestFlush:
		err = encoder.encodeFlush()
	}

	if err != nil {
		return err
	}

	return writer.Flush()
}

// reconnect attempts to reconnect with exponential backoff.
func (c *Client) reconnect() bool {
	delay := c.cfg.ReconnectMinDelay

	for attempt := 1; attempt <= MaxReconnectAttempts; attempt++ {
		select {
		case <-c.ctx.Done():
			return false
		case <-time.After(delay):
		}

		if err := c.dial(); err != nil {
			c.sendError(fmt.Errorf("reconnect attempt %d failed: %w", attempt, err))

			delay *= 2
			if delay > c.cfg.ReconnectMaxDelay {
				delay = c.cfg.ReconnectMaxDelay
			}
			continue
		}

		c.stats.incReconnectCount()

		select {
		case c.reconnectCh <- ReconnectEvent{Attempt: attempt}:
		default:
		}

		return true
	}

	c.sendError(ErrMaxReconnects)
	return false
}

// waitForReconnect waits until a connection is available or the client closes.
func (c *Client) waitForReconnect() bool {
	ticker := time.NewTicker(ReconnectCheckInterval)
	defer ticker.Stop()

	for iterations := 0; iterations < MaxReconnectAttempts; iterations++ {
		select {
		case <-c.ctx.Done():
			return false
		case <-ticker.C:
			if c.getConn() != nil {
				return true
			}
		}
	}

	return false
}
