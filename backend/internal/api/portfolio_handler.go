package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetPortfolioSummary fetches the latest portfolio snapshot for dashboard views.
func (h *Handler) GetPortfolioSummary(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	h.refreshPortfolioMarks(c.Request.Context(), userID)
	summary, err := h.PortfolioRepo.GetLatestSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if summary == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no portfolio history snapshots discovered"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetActivePositions returns currently open portfolio positions.
func (h *Handler) GetActivePositions(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	h.refreshPortfolioMarks(c.Request.Context(), userID)
	positions, err := h.PortfolioRepo.GetActivePositions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// GetExecutionStream returns the user's latest persisted trade records.
func (h *Handler) GetExecutionStream(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	limit := 50

	if queryLimit := c.Query("limit"); queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}
	limit = clampLimit(limit, 50, 500)

	trades, err := h.PortfolioRepo.GetExecutionStream(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// GetSystemAlerts returns recent user-scoped alerts.
func (h *Handler) GetSystemAlerts(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	limit := 10
	if queryLimit := c.Query("limit"); queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}
	limit = clampLimit(limit, 10, 100)

	alerts, err := h.PortfolioRepo.GetSystemAlerts(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetPortfolioHistory returns snapshots sorted chronologically for charting.
func (h *Handler) GetPortfolioHistory(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	limit := 30
	if queryLimit := c.Query("limit"); queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}
	limit = clampLimit(limit, 30, 365)

	h.refreshPortfolioMarks(c.Request.Context(), userID)
	history, err := h.PortfolioRepo.GetPortfolioHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *Handler) refreshPortfolioMarks(ctx context.Context, userID int64) {
	if h == nil || h.PortfolioRepo == nil || h.Alpaca == nil || h.Alpaca.APIKey() == "" || h.Alpaca.APISecret() == "" {
		return
	}

	positions, err := h.PortfolioRepo.GetActivePositions(ctx, userID)
	if err != nil || len(positions) == 0 {
		return
	}

	symbols := make([]string, 0, len(positions))
	for _, position := range positions {
		symbols = append(symbols, position.Symbol)
	}

	priceCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	prices, _, err := h.Alpaca.GetLatestPrices(priceCtx, symbols)
	if err != nil || len(prices) == 0 {
		return
	}

	for _, position := range positions {
		price := prices[position.Symbol]
		if price <= 0 {
			continue
		}
		if err := h.PortfolioRepo.MarkPositionPrice(ctx, userID, position.Symbol, price); err != nil {
			return
		}
	}
	_ = h.PortfolioRepo.SavePortfolioSnapshot(ctx, userID, 0, 0)
}

type portfolioLiveEvent struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol,omitempty"`
	Price     float64     `json:"price,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Summary   interface{} `json:"summary,omitempty"`
	Positions interface{} `json:"positions,omitempty"`
	History   interface{} `json:"history,omitempty"`
}

func (h *Handler) StreamPortfolioLive(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming is not supported"})
		return
	}

	var writeMu sync.Mutex
	writeEvent := func(event portfolioLiveEvent) bool {
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now().UTC()
		}
		payload, err := json.Marshal(event)
		if err != nil {
			return false
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		if _, err := c.Writer.Write(append(payload, '\n')); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	writeSnapshot := func() bool {
		summary, _ := h.PortfolioRepo.GetLatestSummary(c.Request.Context(), userID)
		positions, _ := h.PortfolioRepo.GetActivePositions(c.Request.Context(), userID)
		history, _ := h.PortfolioRepo.GetPortfolioHistory(c.Request.Context(), userID, 30)
		return writeEvent(portfolioLiveEvent{
			Type:      "snapshot",
			Timestamp: time.Now().UTC(),
			Summary:   summary,
			Positions: positions,
			History:   history,
		})
	}

	h.refreshPortfolioMarks(c.Request.Context(), userID)
	if !writeSnapshot() {
		return
	}

	positions, err := h.PortfolioRepo.GetActivePositions(c.Request.Context(), userID)
	if err != nil || len(positions) == 0 || h.Alpaca == nil || h.Alpaca.APIKey() == "" || h.Alpaca.APISecret() == "" {
		h.streamPortfolioSnapshotFallback(c.Request.Context(), writeSnapshot)
		return
	}

	symbols := make([]string, 0, len(positions))
	for _, position := range positions {
		symbols = append(symbols, strings.ToUpper(strings.TrimSpace(position.Symbol)))
	}
	if err := h.streamAlpacaTrades(c.Request.Context(), userID, symbols, writeEvent, writeSnapshot); err != nil {
		log.Printf("[PortfolioLive] stream fallback for user %d: %v", userID, err)
		h.streamPortfolioSnapshotFallback(c.Request.Context(), writeSnapshot)
	}
}

func (h *Handler) streamPortfolioSnapshotFallback(ctx context.Context, writeSnapshot func() bool) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !writeSnapshot() {
				return
			}
		}
	}
}

func (h *Handler) streamAlpacaTrades(ctx context.Context, userID int64, symbols []string, writeEvent func(portfolioLiveEvent) bool, writeSnapshot func() bool) error {
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	wsURL := h.Alpaca.RealtimeStockStreamURL(time.Now().UTC())

	header := http.Header{}
	header.Set("APCA-API-KEY-ID", h.Alpaca.APIKey())
	header.Set("APCA-API-SECRET-KEY", h.Alpaca.APISecret())

	conn, resp, err := websocket.DefaultDialer.DialContext(streamCtx, wsURL, header)
	if err != nil {
		return fmt.Errorf("alpaca websocket dial: %w", err)
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	defer func() { _ = conn.Close() }()

	if err := conn.WriteJSON(map[string]interface{}{
		"action": "subscribe",
		"trades": symbols,
		"quotes": symbols,
	}); err != nil {
		return fmt.Errorf("alpaca quote/trade subscribe: %w", err)
	}

	snapshotTicker := time.NewTicker(time.Second)
	defer snapshotTicker.Stop()
	go func() {
		for {
			select {
			case <-streamCtx.Done():
				_ = conn.Close()
				return
			case <-snapshotTicker.C:
				_ = h.PortfolioRepo.SavePortfolioSnapshot(streamCtx, userID, 0, 0)
				if !writeSnapshot() {
					_ = conn.Close()
					return
				}
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("alpaca websocket read: %w", err)
		}
		var payload []map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}
		for _, item := range payload {
			kind, _ := item["T"].(string)
			if kind != "t" && kind != "q" {
				continue
			}
			symbol, _ := item["S"].(string)
			price, ok := livePriceFromAlpacaMessage(kind, item)
			if !ok || symbol == "" || price <= 0 {
				continue
			}
			timestamp := time.Now().UTC()
			if rawTime, ok := item["t"].(string); ok {
				if parsed, err := time.Parse(time.RFC3339Nano, rawTime); err == nil {
					timestamp = parsed.UTC()
				}
			}
			if err := h.PortfolioRepo.MarkPositionPrice(streamCtx, userID, symbol, price); err != nil {
				continue
			}
			if !writeEvent(portfolioLiveEvent{
				Type:      "price",
				Symbol:    strings.ToUpper(symbol),
				Price:     price,
				Timestamp: timestamp,
			}) {
				return nil
			}
		}
	}
}

func livePriceFromAlpacaMessage(kind string, item map[string]interface{}) (float64, bool) {
	if kind == "t" {
		return numericField(item["p"])
	}
	if kind != "q" {
		return 0, false
	}
	bid, hasBid := numericField(item["bp"])
	ask, hasAsk := numericField(item["ap"])
	if hasBid && hasAsk && bid > 0 && ask > 0 {
		return (bid + ask) / 2, true
	}
	if hasAsk && ask > 0 {
		return ask, true
	}
	if hasBid && bid > 0 {
		return bid, true
	}
	return 0, false
}

func numericField(value interface{}) (float64, bool) {
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

func clampLimit(value, fallback, max int) int {
	if value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}
