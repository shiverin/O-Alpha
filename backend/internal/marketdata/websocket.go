package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/pkg/models"
)

// WebSocketConnector handles Alpaca market data streams
type WebSocketConnector struct {
	alpacaClient *alpaca.Client
	symbols      []string
	timeframe    string
	dataChan     chan models.Bar
	doneChan     chan struct{}
	errChan      chan error
	wsConn       *websocket.Conn
	mu           sync.Mutex
	connected    bool
	reconnectTicker *time.Ticker
}

// NewWebSocketConnector creates a new market data WebSocket connector
func NewWebSocketConnector(alpacaClient *alpaca.Client, symbols []string, timeframe string) *WebSocketConnector {
	return &WebSocketConnector{
		alpacaClient: alpacaClient,
		symbols:      symbols,
		timeframe:    timeframe,
		dataChan:     make(chan models.Bar, 100),
		doneChan:     make(chan struct{}),
		errChan:      make(chan error, 1),
		reconnectTicker: time.NewTicker(30 * time.Second),
	}
}

// Start begins the WebSocket connection
func (w *WebSocketConnector) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.connected {
		w.mu.Unlock()
		return fmt.Errorf("websocket already connected")
	}
	w.mu.Unlock()

	// Determine WebSocket URL based on environment
	var wsURL string
	if w.alpacaClient.BaseURL() == "https://api.alpaca.markets" {
		// Use SIP for paid data
		wsURL = "wss://stream.data.alpaca.markets/v2/sip"
	} else {
		// Use IEX for free/paper data
		wsURL = "wss://stream.data.alpaca.markets/v2/iex"
	}

	// Dial WebSocket connection
	header := http.Header{}
	header.Add("APCA-API-KEY-ID", w.alpacaClient.APIKey())
	header.Add("APCA-API-SECRET-KEY", w.alpacaClient.APISecret())

	c, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}
	if resp.StatusCode != 101 {
		resp.Body.Close()
		return fmt.Errorf("unexpected WS response: %d", resp.StatusCode)
	}
	resp.Body.Close()
	w.wsConn = c

	// Subscribe to bars for symbols
	subMsg := map[string]interface{}{
		"action": "subscribe",
		"bars":   w.symbols,
	}
	if err := w.wsConn.WriteJSON(subMsg); err != nil {
		w.wsConn.Close()
		return fmt.Errorf("subscription failed: %w", err)
	}

	w.mu.Lock()
	w.connected = true
	w.mu.Unlock()

	// Start reading messages
	go w.readMessages(ctx)
	go w.handleReconnect(ctx)

	return nil
}

// readMessages continuously reads from WebSocket
func (w *WebSocketConnector) readMessages(ctx context.Context) {
	defer func() {
		if w.wsConn != nil {
			w.wsConn.Close()
		}
		close(w.dataChan)
		close(w.errChan)
		w.mu.Lock()
		w.connected = false
		w.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.doneChan:
			return
		default:
			_, msg, err := w.wsConn.ReadMessage()
			if err != nil {
				select {
				case w.errChan <- fmt.Errorf("websocket read error: %w", err):
				case <-ctx.Done():
				case <-w.doneChan:
				}
				return
			}

			// Parse Alpaca WebSocket message
			var wsMsg []map[string]interface{}
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				log.Printf("Failed to parse WS message: %v", err)
				continue
			}

			// Handle bar data
			for _, barData := range wsMsg {
				if barData["T"] == nil || barData["S"] == nil {
					// Not a bar message, could be trade, quote, or status
					continue
				}

				if bar := w.parseAlpacaBar(barData); bar != nil {
					select {
					case w.dataChan <- *bar:
					case <-ctx.Done():
						return
					case <-w.doneChan:
						return
					}
				}
			}
		}
	}
}

// handleReconnect manages reconnection logic
func (w *WebSocketConnector) handleReconnect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.doneChan:
			return
		case <-w.reconnectTicker.C:
			w.mu.Lock()
			if !w.connected {
				w.mu.Unlock()
				// Attempt reconnection
				log.Println("Attempting WebSocket reconnection...")
				if err := w.Start(ctx); err != nil {
					log.Printf("Reconnection failed: %v", err)
					// Will try again on next tick
				} else {
					log.Println("WebSocket reconnected successfully")
				}
			} else {
				w.mu.Unlock()
			}
		}
	}
}

// parseAlpacaBar converts Alpaca WebSocket bar to internal model
func (w *WebSocketConnector) parseAlpacaBar(data map[string]interface{}) *models.Bar {
	t := data["t"].(string)
	timestamp, err := time.Parse(time.RFC3339Nano, t)
	if err != nil {
		// Try alternative format
		timestamp, err = time.Parse(time.RFC3339, t)
		if err != nil {
			log.Printf("Failed to parse time %s: %v", t, err)
			return nil
		}
	}

	return &models.Bar{
		Time:   timestamp.UTC(),
		Symbol: strings.ToUpper(data["S"].(string)),
		Open:   data["o"].(float64),
		High:   data["h"].(float64),
		Low:    data["l"].(float64),
		Close:  data["c"].(float64),
		Volume: int64(data["v"].(float64)),
	}
}

// Data returns the market data channel
func (w *WebSocketConnector) Data() <-chan models.Bar {
	return w.dataChan
}

// Errors returns the error channel
func (w *WebSocketConnector) Errors() <-chan error {
	return w.errChan
}

// IsConnected returns connection status
func (w *WebSocketConnector) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.connected
}

// Stop closes the WebSocket connection
func (w *WebSocketConnector) Stop() {
	w.mu.Lock()
	if !w.connected {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	close(w.doneChan)
	w.reconnectTicker.Stop()

	if w.wsConn != nil {
		w.wsConn.Close()
	}

	// Wait for goroutines to cleanup
	time.Sleep(100 * time.Millisecond)
}