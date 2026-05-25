package alpaca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/pkg/models"
)

// Client fetches market data from Alpaca.
type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// NewClient creates an Alpaca data API client.
func NewClient(baseURL, apiKey, apiSecret string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// BaseURL returns the base URL for the API client
func (c *Client) BaseURL() string {
	return c.baseURL
}

// APIKey returns the API key for the client
func (c *Client) APIKey() string {
	return c.apiKey
}

// APISecret returns the API secret for the client
func (c *Client) APISecret() string {
	return c.apiSecret
}

type barsResponse struct {
	Bars          []alpacaBar `json:"bars"`
	NextPageToken string      `json:"next_page_token"`
}

type alpacaBar struct {
	T  string  `json:"t"`
	O  float64 `json:"o"`
	H  float64 `json:"h"`
	L  float64 `json:"l"`
	C  float64 `json:"c"`
	V  uint64  `json:"v"`
}

// GetBars fetches historical bars with pagination.
func (c *Client) GetBars(ctx context.Context, symbol, timeframe string, start, end time.Time, limit int) ([]models.Bar, error) {
	if limit <= 0 {
		limit = 10000
	}

	var all []models.Bar
	pageToken := ""

	for {
		u, err := url.Parse(fmt.Sprintf("%s/v2/stocks/%s/bars", c.baseURL, url.PathEscape(strings.ToUpper(symbol))))
		if err != nil {
			return nil, err
		}

		q := u.Query()
		q.Set("timeframe", timeframe)
		q.Set("start", start.UTC().Format(time.RFC3339))
		q.Set("end", end.UTC().Format(time.RFC3339))
		q.Set("limit", strconv.Itoa(limit))
		q.Set("adjustment", "raw")
		q.Set("feed", "iex")
		if pageToken != "" {
			q.Set("page_token", pageToken)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("APCA-API-KEY-ID", c.apiKey)
		req.Header.Set("APCA-API-SECRET-KEY", c.apiSecret)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("alpaca request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("alpaca status %d: %s", resp.StatusCode, string(body))
		}

		var payload barsResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode bars: %w", err)
		}

		for _, ab := range payload.Bars {
			t, err := time.Parse(time.RFC3339Nano, ab.T)
			if err != nil {
				t, err = time.Parse(time.RFC3339, ab.T)
				if err != nil {
					return nil, fmt.Errorf("parse time %q: %w", ab.T, err)
				}
			}
			b := models.Bar{
				Time:   t.UTC(),
				Symbol: strings.ToUpper(symbol),
				Open:   ab.O,
				High:   ab.H,
				Low:    ab.L,
				Close:  ab.C,
				Volume: int64(ab.V),
			}
			if err := ValidateBar(b); err != nil {
				return nil, fmt.Errorf("invalid bar at %s: %w", b.Time, err)
			}
			all = append(all, b)
		}

		if payload.NextPageToken == "" {
			break
		}
		pageToken = payload.NextPageToken
	}

	return all, nil
}

// ValidateBar checks OHLCV consistency.
func ValidateBar(b models.Bar) error {
	if b.Open < 0 || b.High < 0 || b.Low < 0 || b.Close < 0 {
		return fmt.Errorf("negative price")
	}
	if b.Volume < 0 {
		return fmt.Errorf("negative volume")
	}
	if b.High < b.Low {
		return fmt.Errorf("high < low")
	}
	if b.High < b.Open || b.High < b.Close {
		return fmt.Errorf("high below open/close")
	}
	if b.Low > b.Open || b.Low > b.Close {
		return fmt.Errorf("low above open/close")
	}
	return nil
}

// OrderRequest represents a request to place an order with Alpaca.
type OrderRequest struct {
	Symbol string `json:"symbol"`
	Qty    int    `json:"qty"`
	Side   string `json:"side"` // "buy" or "sell"
	Type   string `json:"type"` // "market", "limit", etc.
}

// OrderResponse represents the response from placing an order with Alpaca.
type OrderResponse struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Qty    string `json:"qty"`
	Side   string `json:"side"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// PlaceOrder places an order with the Alpaca Trading API.
func (c *Client) PlaceOrder(ctx context.Context, req *OrderRequest) (*OrderResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if req.Qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if req.Side != "buy" && req.Side != "sell" {
		return nil, fmt.Errorf("side must be 'buy' or 'sell'")
	}

	// In a real implementation, this would make an HTTP request to the Alpaca Trading API
	// For now, we'll return a mock response for testing purposes
	return &OrderResponse{
		ID:     "order123",
		Symbol: req.Symbol,
		Qty:    fmt.Sprintf("%d", req.Qty),
		Side:   req.Side,
		Type:   req.Type,
		Status: "new",
	}, nil
}
