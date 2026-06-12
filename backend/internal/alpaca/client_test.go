package alpaca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://api.alpaca.markets", "test-key", "test-secret")
	assert.NotNil(t, c)
	assert.Equal(t, "https://api.alpaca.markets", c.baseURL)
	assert.Equal(t, "test-key", c.apiKey)
	assert.Equal(t, "test-secret", c.apiSecret)
	assert.NotNil(t, c.httpClient)
}

func TestValidateBar(t *testing.T) {
	// Valid bar
	assert.NoError(t, ValidateBar(models.Bar{
		Open:   100,
		High:   110,
		Low:    90,
		Close:  105,
		Volume: 1000,
	}))

	// Negative price
	assert.Error(t, ValidateBar(models.Bar{
		Open:   -100,
		High:   110,
		Low:    90,
		Close:  105,
		Volume: 1000,
	}))

	// Negative volume
	assert.Error(t, ValidateBar(models.Bar{
		Open:   100,
		High:   110,
		Low:    90,
		Close:  105,
		Volume: -1000,
	}))

	// High < low
	assert.Error(t, ValidateBar(models.Bar{
		Open:   100,
		High:   90,
		Low:    110,
		Close:  105,
		Volume: 1000,
	}))

	// High below open/close
	assert.Error(t, ValidateBar(models.Bar{
		Open:   100,
		High:   95,
		Low:    90,
		Close:  105,
		Volume: 1000,
	}))

	// Low above open/close
	assert.Error(t, ValidateBar(models.Bar{
		Open:   100,
		High:   110,
		Low:    115,
		Close:  105,
		Volume: 1000,
	}))
}

func TestRealtimeStockFeedSelectsOvernightSession(t *testing.T) {
	c := NewClient("https://data.alpaca.markets", "test-key", "test-secret")

	cases := []struct {
		name string
		at   string
		want string
	}{
		{
			name: "sunday overnight opens at eight pm",
			at:   "2026-06-14T20:30:00-04:00",
			want: "overnight",
		},
		{
			name: "monday before four am is overnight",
			at:   "2026-06-15T03:59:00-04:00",
			want: "overnight",
		},
		{
			name: "monday premarket uses regular data feed",
			at:   "2026-06-15T04:00:00-04:00",
			want: "iex",
		},
		{
			name: "friday after four am exits overnight",
			at:   "2026-06-12T04:00:00-04:00",
			want: "iex",
		},
		{
			name: "saturday is closed",
			at:   "2026-06-13T21:00:00-04:00",
			want: "iex",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			at, err := time.Parse(time.RFC3339, tc.at)
			if err != nil {
				t.Fatalf("parse time: %v", err)
			}
			assert.Equal(t, tc.want, c.RealtimeStockFeed(at))
		})
	}
}

func TestRealtimeStockStreamURLUsesOvernightEndpointFamily(t *testing.T) {
	c := NewClient("https://data.alpaca.markets", "test-key", "test-secret")
	overnight, err := time.Parse(time.RFC3339, "2026-06-15T01:00:00-04:00")
	if err != nil {
		t.Fatalf("parse overnight time: %v", err)
	}
	regular, err := time.Parse(time.RFC3339, "2026-06-15T11:00:00-04:00")
	if err != nil {
		t.Fatalf("parse regular time: %v", err)
	}

	assert.Equal(t, "wss://stream.data.alpaca.markets/v1beta1/overnight", c.RealtimeStockStreamURL(overnight))
	assert.Equal(t, "wss://stream.data.alpaca.markets/v2/iex", c.RealtimeStockStreamURL(regular))
}

func TestRealtimeStockFeedHonorsEnvironmentOverrides(t *testing.T) {
	t.Setenv("ALPACA_REALTIME_FEED", "sip")
	t.Setenv("ALPACA_OVERNIGHT_FEED", "boats")
	c := NewClient("https://data.alpaca.markets", "test-key", "test-secret")
	overnight, err := time.Parse(time.RFC3339, "2026-06-15T01:00:00-04:00")
	if err != nil {
		t.Fatalf("parse overnight time: %v", err)
	}
	regular, err := time.Parse(time.RFC3339, "2026-06-15T11:00:00-04:00")
	if err != nil {
		t.Fatalf("parse regular time: %v", err)
	}

	assert.Equal(t, "boats", c.RealtimeStockFeed(overnight))
	assert.Equal(t, "sip", c.RealtimeStockFeed(regular))
	assert.Equal(t, "wss://stream.data.alpaca.markets/v1beta1/boats", c.RealtimeStockStreamURL(overnight))
	assert.Equal(t, "wss://stream.data.alpaca.markets/v2/sip", c.RealtimeStockStreamURL(regular))
}

func TestGetLatestBarsUsesLatestEndpointAndFeed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/stocks/bars/latest", r.URL.Path)
		assert.Equal(t, "AAPL,MSFT", r.URL.Query().Get("symbols"))
		assert.Equal(t, "overnight", r.URL.Query().Get("feed"))
		assert.Empty(t, r.URL.Query().Get("adjustment"))
		assert.Equal(t, "test-key", r.Header.Get("APCA-API-KEY-ID"))
		assert.Equal(t, "test-secret", r.Header.Get("APCA-API-SECRET-KEY"))
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"bars": map[string]interface{}{
				"AAPL": map[string]interface{}{
					"t": "2026-06-12T07:56:00Z",
					"o": 295.9,
					"h": 296.3,
					"l": 295.9,
					"c": 296.3,
					"v": 1047,
				},
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-key", "test-secret")
	bars, err := c.GetLatestBars(context.Background(), []string{"aapl", "MSFT"}, "overnight", "raw")
	assert.NoError(t, err)
	assert.Equal(t, 296.3, bars["AAPL"].Close)
}

func TestPlaceOrderValidation(t *testing.T) {
	// Create a test server to mock the Alpaca API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/orders", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("APCA-API-KEY-ID"))
		assert.Equal(t, "test-secret", r.Header.Get("APCA-API-SECRET-KEY"))
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte(`{"id":"order123","symbol":"AAPL","qty":"10","side":"buy","status":"new"}`))
		if err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-key", "test-secret")

	// Test nil order
	_, err := c.PlaceOrder(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order cannot be nil")

	// Test empty symbol
	qty := 10.0
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Qty:         &qty,
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symbol is required")

	// Test non-positive quantity
	zeroQty := 0.0
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol:      "AAPL",
		Qty:         &zeroQty,
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")

	negativeQty := -5.0
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol:      "AAPL",
		Qty:         &negativeQty,
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")

	// Test invalid side
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol:      "AAPL",
		Qty:         &qty,
		Side:        "invalid",
		Type:        "market",
		TimeInForce: "day",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "side must be 'buy' or 'sell'")

	notional := 100.0
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol:      "AAPL",
		Qty:         &qty,
		Notional:    &notional,
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of qty or notional is required")

	// Test valid order (should succeed)
	orderResp, err := c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol:      "AAPL",
		Qty:         &qty,
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
	})
	assert.NoError(t, err)
	assert.NotNil(t, orderResp)
	assert.Equal(t, "order123", orderResp.ID)
	assert.Equal(t, "AAPL", orderResp.Symbol)
	assert.Equal(t, "new", orderResp.Status)
}
