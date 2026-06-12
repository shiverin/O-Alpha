package alpaca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/pkg/models"
)

// Client fetches market data from Alpaca.
type Client struct {
	baseURL       string
	apiKey        string
	apiSecret     string
	regularFeed   string
	overnightFeed string
	httpClient    *http.Client
}

// NewClient creates an Alpaca data API client.
func NewClient(baseURL, apiKey, apiSecret string) *Client {
	return &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		regularFeed:   normalizeStockFeed(os.Getenv("ALPACA_REALTIME_FEED"), "iex"),
		overnightFeed: normalizeStockFeed(os.Getenv("ALPACA_OVERNIGHT_FEED"), "overnight"),
		httpClient:    &http.Client{Timeout: 60 * time.Second},
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

// RealtimeStockFeed returns the Alpaca stock feed that should be used right now.
//
// Alpaca's overnight feed lives on a different endpoint family than the normal
// IEX/SIP stream. The regular feed can be overridden with ALPACA_REALTIME_FEED
// (iex, sip, delayed_sip). The overnight feed can be overridden with
// ALPACA_OVERNIGHT_FEED (overnight, boats) for paid BOATS access.
func (c *Client) RealtimeStockFeed(now time.Time) string {
	if IsUSEquityOvernightSession(now) {
		if c != nil && c.overnightFeed != "" {
			return c.overnightFeed
		}
		return "overnight"
	}
	if c != nil && c.regularFeed != "" {
		return c.regularFeed
	}
	return "iex"
}

// RealtimeStockStreamURL returns the websocket URL for the current stock session.
func (c *Client) RealtimeStockStreamURL(now time.Time) string {
	feed := c.RealtimeStockFeed(now)
	version := "v2"
	if feed == "overnight" || feed == "boats" {
		version = "v1beta1"
	}
	return fmt.Sprintf("wss://stream.data.alpaca.markets/%s/%s", version, feed)
}

// IsUSEquityOvernightSession reports whether now is inside Alpaca's overnight
// stock session: Sunday-Friday, 8:00 PM-4:00 AM America/New_York time.
func IsUSEquityOvernightSession(now time.Time) bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.FixedZone("America/New_York", -5*60*60)
	}
	ny := now.In(loc)
	hour := ny.Hour()
	switch ny.Weekday() {
	case time.Sunday:
		return hour >= 20
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday:
		return hour < 4 || hour >= 20
	case time.Friday:
		return hour < 4
	default:
		return false
	}
}

type barsResponse struct {
	Bars          []alpacaBar `json:"bars"`
	NextPageToken string      `json:"next_page_token"`
}

type multiBarsResponse struct {
	Bars          map[string][]alpacaBar `json:"bars"`
	NextPageToken string                 `json:"next_page_token"`
}

type latestBarsResponse struct {
	Bars map[string]alpacaBar `json:"bars"`
}

type alpacaBar struct {
	T string  `json:"t"`
	O float64 `json:"o"`
	H float64 `json:"h"`
	L float64 `json:"l"`
	C float64 `json:"c"`
	V uint64  `json:"v"`
}

// GetBars fetches historical bars with pagination.
func (c *Client) GetBars(ctx context.Context, symbol, timeframe string, start, end time.Time, limit int) ([]models.Bar, error) {
	return c.GetBarsWithDataset(ctx, symbol, timeframe, start, end, limit, "iex", "raw")
}

// GetBarsWithDataset fetches historical bars for a single symbol with explicit
// feed and adjustment parameters.
func (c *Client) GetBarsWithDataset(ctx context.Context, symbol, timeframe string, start, end time.Time, limit int, feed string, adjustment string) ([]models.Bar, error) {
	if limit <= 0 {
		limit = 10000
	}
	feed, adjustment = normalizeMarketDataOptions(feed, adjustment)

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
		q.Set("adjustment", adjustment)
		q.Set("feed", feed)
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

func (c *Client) GetBarsMulti(
	ctx context.Context,
	symbols []string,
	timeframe string,
	start time.Time,
	end time.Time,
	limit int,
	feed string,
	adjustment string,
) (map[string][]models.Bar, error) {
	if limit <= 0 {
		limit = 10000
	}
	feed, adjustment = normalizeMarketDataOptions(feed, adjustment)
	normalizedSymbols := normalizeSymbols(symbols)
	if len(normalizedSymbols) == 0 {
		return nil, fmt.Errorf("at least one symbol is required")
	}

	all := make(map[string][]models.Bar, len(normalizedSymbols))
	pageToken := ""
	for {
		u, err := url.Parse(fmt.Sprintf("%s/v2/stocks/bars", c.baseURL))
		if err != nil {
			return nil, err
		}

		q := u.Query()
		q.Set("symbols", strings.Join(normalizedSymbols, ","))
		q.Set("timeframe", timeframe)
		q.Set("start", start.UTC().Format(time.RFC3339))
		q.Set("end", end.UTC().Format(time.RFC3339))
		q.Set("limit", strconv.Itoa(limit))
		q.Set("adjustment", adjustment)
		q.Set("feed", feed)
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
			return nil, fmt.Errorf("alpaca multi-bars request: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read multi-bars body: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("alpaca multi-bars status %d: %s", resp.StatusCode, string(body))
		}

		var payload multiBarsResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode multi-bars: %w", err)
		}
		for symbol, symbolBars := range payload.Bars {
			normalized := strings.ToUpper(symbol)
			for _, ab := range symbolBars {
				bar, err := alpacaBarToModel(normalized, ab)
				if err != nil {
					return nil, err
				}
				all[normalized] = append(all[normalized], bar)
			}
		}
		if payload.NextPageToken == "" {
			break
		}
		pageToken = payload.NextPageToken
	}

	for _, symbol := range normalizedSymbols {
		if _, ok := all[symbol]; !ok {
			all[symbol] = nil
		}
	}
	return all, nil
}

func (c *Client) GetLatestBars(ctx context.Context, symbols []string, feed string, adjustment string) (map[string]models.Bar, error) {
	normalizedSymbols := normalizeSymbols(symbols)
	if len(normalizedSymbols) == 0 {
		return nil, fmt.Errorf("at least one symbol is required")
	}
	feed, _ = normalizeMarketDataOptions(feed, adjustment)

	u, err := url.Parse(fmt.Sprintf("%s/v2/stocks/bars/latest", c.baseURL))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("symbols", strings.Join(normalizedSymbols, ","))
	q.Set("feed", feed)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("APCA-API-KEY-ID", c.apiKey)
	req.Header.Set("APCA-API-SECRET-KEY", c.apiSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("alpaca latest-bars request: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read latest-bars body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("alpaca latest-bars status %d: %s", resp.StatusCode, string(body))
	}

	var payload latestBarsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode latest bars: %w", err)
	}

	out := make(map[string]models.Bar, len(normalizedSymbols))
	for symbol, ab := range payload.Bars {
		normalized := strings.ToUpper(symbol)
		bar, err := alpacaBarToModel(normalized, ab)
		if err != nil {
			return nil, err
		}
		out[normalized] = bar
	}
	return out, nil
}

func (c *Client) GetLatestPrices(ctx context.Context, symbols []string) (map[string]float64, time.Time, error) {
	if c == nil || c.apiKey == "" || c.apiSecret == "" {
		return nil, time.Time{}, nil
	}
	normalizedSymbols := normalizeSymbols(symbols)
	if len(normalizedSymbols) == 0 {
		return nil, time.Time{}, nil
	}

	now := time.Now().UTC()
	barsBySymbol, err := c.GetLatestBars(ctx, normalizedSymbols, c.RealtimeStockFeed(now), "raw")
	if err != nil {
		return nil, time.Time{}, err
	}

	prices := make(map[string]float64, len(normalizedSymbols))
	var latestTime time.Time
	for _, symbol := range normalizedSymbols {
		latest, ok := barsBySymbol[symbol]
		if !ok {
			continue
		}
		if latest.Close <= 0 {
			continue
		}
		prices[symbol] = latest.Close
		if latest.Time.After(latestTime) {
			latestTime = latest.Time
		}
	}
	return prices, latestTime, nil
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
	Symbol      string   `json:"symbol"`
	Qty         *float64 `json:"qty,omitempty"`
	Notional    *float64 `json:"notional,omitempty"`
	Side        string   `json:"side"` // "buy" or "sell"
	Type        string   `json:"type"` // initially "market"
	TimeInForce string   `json:"time_in_force"`
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

type Asset struct {
	Symbol       string `json:"symbol"`
	Tradable     bool   `json:"tradable"`
	Marginable   bool   `json:"marginable"`
	Shortable    bool   `json:"shortable"`
	EasyToBorrow bool   `json:"easy_to_borrow"`
	Fractionable bool   `json:"fractionable"`
}

// PlaceOrder places an order with the Alpaca Trading API.
func (c *Client) PlaceOrder(ctx context.Context, req *OrderRequest) (*OrderResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if (req.Qty == nil && req.Notional == nil) || (req.Qty != nil && req.Notional != nil) {
		return nil, fmt.Errorf("exactly one of qty or notional is required")
	}
	if req.Qty != nil && *req.Qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if req.Notional != nil && *req.Notional <= 0 {
		return nil, fmt.Errorf("notional must be positive")
	}
	if req.Side != "buy" && req.Side != "sell" {
		return nil, fmt.Errorf("side must be 'buy' or 'sell'")
	}
	if req.Type == "" {
		req.Type = "market"
	}
	if req.Type != "market" {
		return nil, fmt.Errorf("order type must be market")
	}
	if req.TimeInForce == "" {
		req.TimeInForce = "day"
	}
	if req.TimeInForce != "day" {
		return nil, fmt.Errorf("time_in_force must be day")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode order: %w", err)
	}

	u, err := url.Parse(c.baseURL + "/v2/orders")
	if err != nil {
		return nil, fmt.Errorf("parse order url: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create order request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("APCA-API-KEY-ID", c.apiKey)
	httpReq.Header.Set("APCA-API-SECRET-KEY", c.apiSecret)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("alpaca order request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read order response: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("alpaca order status %d: %s", resp.StatusCode, string(respBody))
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, fmt.Errorf("decode order response: %w", err)
	}
	return &orderResp, nil
}

func (c *Client) GetAsset(ctx context.Context, symbol string) (*Asset, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	u, err := url.Parse(fmt.Sprintf("%s/v2/assets/%s", c.baseURL, url.PathEscape(symbol)))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("APCA-API-KEY-ID", c.apiKey)
	req.Header.Set("APCA-API-SECRET-KEY", c.apiSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("alpaca asset request: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read asset body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("alpaca asset status %d: %s", resp.StatusCode, string(body))
	}
	var asset Asset
	if err := json.Unmarshal(body, &asset); err != nil {
		return nil, fmt.Errorf("decode asset: %w", err)
	}
	return &asset, nil
}

func alpacaBarToModel(symbol string, ab alpacaBar) (models.Bar, error) {
	t, err := time.Parse(time.RFC3339Nano, ab.T)
	if err != nil {
		t, err = time.Parse(time.RFC3339, ab.T)
		if err != nil {
			return models.Bar{}, fmt.Errorf("parse time %q: %w", ab.T, err)
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
		return models.Bar{}, fmt.Errorf("invalid bar at %s: %w", b.Time, err)
	}
	return b, nil
}

func normalizeMarketDataOptions(feed string, adjustment string) (string, string) {
	feed = normalizeStockFeed(feed, "")
	adjustment = strings.ToLower(strings.TrimSpace(adjustment))
	if feed == "" {
		feed = "iex"
	}
	if adjustment == "" {
		adjustment = "raw"
	}
	return feed, adjustment
}

func normalizeStockFeed(feed string, fallback string) string {
	feed = strings.ToLower(strings.TrimSpace(feed))
	switch feed {
	case "iex", "sip", "delayed_sip", "overnight", "boats":
		return feed
	case "":
		return fallback
	default:
		return fallback
	}
}

func normalizeSymbols(symbols []string) []string {
	out := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}
