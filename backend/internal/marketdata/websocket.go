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

// WebSocketConnector handles Alpaca market data streams.
type WebSocketConnector struct {
	alpacaClient    *alpaca.Client
	symbols         []string
	timeframe       string
	dataChan        chan models.Bar
	doneChan        chan struct{}
	errChan         chan error
	wsConn          *websocket.Conn
	mu              sync.Mutex
	stopOnce        sync.Once
	reconnectOnce   sync.Once
	connected       bool
	reconnectTicker *time.Ticker
}

// NewWebSocketConnector creates a market data WebSocket connector.
func NewWebSocketConnector(alpacaClient *alpaca.Client, symbols []string, timeframe string) *WebSocketConnector {
	return &WebSocketConnector{
		alpacaClient:    alpacaClient,
		symbols:         symbols,
		timeframe:       timeframe,
		dataChan:        make(chan models.Bar, 100),
		doneChan:        make(chan struct{}),
		errChan:         make(chan error, 1),
		reconnectTicker: time.NewTicker(30 * time.Second),
	}
}

// Start opens the stream and starts the reader and reconnect loops.
func (w *WebSocketConnector) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.connected {
		w.mu.Unlock()
		return fmt.Errorf("websocket already connected")
	}
	w.mu.Unlock()

	var wsURL string
	if w.alpacaClient.BaseURL() == "https://api.alpaca.markets" {
		wsURL = "wss://stream.data.alpaca.markets/v2/sip"
	} else {
		wsURL = "wss://stream.data.alpaca.markets/v2/iex"
	}

	header := http.Header{}
	header.Add("APCA-API-KEY-ID", w.alpacaClient.APIKey())
	header.Add("APCA-API-SECRET-KEY", w.alpacaClient.APISecret())

	c, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}
	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		if resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				log.Printf("close body: %v", err)
			}
		}
		if c != nil {
			_ = c.Close()
		}
		return fmt.Errorf("unexpected WS response: %d", resp.StatusCode)
	}
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Printf("close body: %v", err)
		}
	}
	w.wsConn = c

	subMsg := map[string]interface{}{
		"action": "subscribe",
		"bars":   w.symbols,
	}
	if err := w.wsConn.WriteJSON(subMsg); err != nil {
		_ = w.wsConn.Close()
		return fmt.Errorf("subscription failed: %w", err)
	}

	w.mu.Lock()
	w.connected = true
	w.mu.Unlock()

	go w.readMessages(ctx)
	w.reconnectOnce.Do(func() {
		go w.handleReconnect(ctx)
	})

	return nil
}

// readMessages streams valid bar messages into dataChan.
func (w *WebSocketConnector) readMessages(ctx context.Context) {
	defer func() {
		if w.wsConn != nil {
			_ = w.wsConn.Close()
		}
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

			var wsMsg []map[string]interface{}
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				log.Printf("Failed to parse WS message: %v", err)
				continue
			}

			for _, barData := range wsMsg {
				if barData["T"] == nil || barData["S"] == nil {
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

// handleReconnect periodically reconnects after a dropped stream.
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
				log.Println("Attempting WebSocket reconnection...")
				if err := w.Start(ctx); err != nil {
					log.Printf("Reconnection failed: %v", err)
				} else {
					log.Println("WebSocket reconnected successfully")
				}
			} else {
				w.mu.Unlock()
			}
		}
	}
}

// parseAlpacaBar converts an Alpaca bar payload to the internal model.
func (w *WebSocketConnector) parseAlpacaBar(data map[string]interface{}) *models.Bar {
	messageType, ok := stringField(data, "T")
	if !ok || messageType != "b" {
		return nil
	}

	t, ok := stringField(data, "t")
	if !ok {
		log.Printf("Skipping bar without timestamp: %v", data)
		return nil
	}
	timestamp, err := time.Parse(time.RFC3339Nano, t)
	if err != nil {
		timestamp, err = time.Parse(time.RFC3339, t)
		if err != nil {
			log.Printf("Failed to parse time %s: %v", t, err)
			return nil
		}
	}

	symbol, ok := stringField(data, "S")
	if !ok {
		log.Printf("Skipping bar without symbol: %v", data)
		return nil
	}
	open, ok := floatField(data, "o")
	if !ok {
		log.Printf("Skipping bar without open price: %v", data)
		return nil
	}
	high, ok := floatField(data, "h")
	if !ok {
		log.Printf("Skipping bar without high price: %v", data)
		return nil
	}
	low, ok := floatField(data, "l")
	if !ok {
		log.Printf("Skipping bar without low price: %v", data)
		return nil
	}
	closePrice, ok := floatField(data, "c")
	if !ok {
		log.Printf("Skipping bar without close price: %v", data)
		return nil
	}
	volume, ok := floatField(data, "v")
	if !ok {
		log.Printf("Skipping bar without volume: %v", data)
		return nil
	}

	bar := models.Bar{
		Time:   timestamp.UTC(),
		Symbol: strings.ToUpper(symbol),
		Open:   open,
		High:   high,
		Low:    low,
		Close:  closePrice,
		Volume: int64(volume),
	}
	if err := alpaca.ValidateBar(bar); err != nil {
		log.Printf("Skipping invalid bar: %v", err)
		return nil
	}
	return &bar
}

func stringField(data map[string]interface{}, key string) (string, bool) {
	value, ok := data[key]
	if !ok {
		return "", false
	}
	s, ok := value.(string)
	return s, ok && s != ""
}

func floatField(data map[string]interface{}, key string) (float64, bool) {
	value, ok := data[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// Data returns the market data channel.
func (w *WebSocketConnector) Data() <-chan models.Bar {
	return w.dataChan
}

// Errors returns the error channel.
func (w *WebSocketConnector) Errors() <-chan error {
	return w.errChan
}

// IsConnected returns connection status.
func (w *WebSocketConnector) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.connected
}

// Stop closes the WebSocket connection.
func (w *WebSocketConnector) Stop() {
	w.stopOnce.Do(func() {
		close(w.doneChan)
		w.reconnectTicker.Stop()

		w.mu.Lock()
		conn := w.wsConn
		w.connected = false
		w.mu.Unlock()

		if conn != nil {
			_ = conn.Close()
		}

		// Wait briefly for reader goroutines to observe doneChan and exit.
		time.Sleep(100 * time.Millisecond)
	})
}
