package alpaca

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
		Open:  100,
		High:  110,
		Low:   90,
		Close: 105,
		Volume: 1000,
	}))

	// Negative price
	assert.Error(t, ValidateBar(models.Bar{
		Open:  -100,
		High:  110,
		Low:   90,
		Close: 105,
		Volume: 1000,
	}))

	// Negative volume
	assert.Error(t, ValidateBar(models.Bar{
		Open:  100,
		High:  110,
		Low:   90,
		Close: 105,
		Volume: -1000,
	}))

	// High < low
	assert.Error(t, ValidateBar(models.Bar{
		Open:  100,
		High:  90,
		Low:   110,
		Close: 105,
		Volume: 1000,
	}))

	// High below open/close
	assert.Error(t, ValidateBar(models.Bar{
		Open:  100,
		High:  95,
		Low:   90,
		Close: 105,
		Volume: 1000,
	}))

	// Low above open/close
	assert.Error(t, ValidateBar(models.Bar{
		Open:  100,
		High:  110,
		Low:   115,
		Close: 105,
		Volume: 1000,
	}))
}

func TestPlaceOrderValidation(t *testing.T) {
	// Create a test server to mock the Alpaca API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/orders", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("APCA-API-KEY-ID"))
		assert.Equal(t, "test-secret", r.Header.Get("APCA-API-SECRET-KEY"))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"order123","symbol":"AAPL","qty":"10","side":"buy","status":"new"}`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-key", "test-secret")

	// Test nil order
	_, err := c.PlaceOrder(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order cannot be nil")

	// Test empty symbol
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Qty:   10,
		Side:  "buy",
		Type:  "market",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symbol is required")

	// Test non-positive quantity
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol: "AAPL",
		Qty:    0,
		Side:   "buy",
		Type:   "market",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")

	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol: "AAPL",
		Qty:    -5,
		Side:   "buy",
		Type:   "market",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")

	// Test invalid side
	_, err = c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol: "AAPL",
		Qty:    10,
		Side:   "invalid",
		Type:   "market",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "side must be 'buy' or 'sell'")

	// Test valid order (should succeed)
	orderResp, err := c.PlaceOrder(context.Background(), &OrderRequest{
		Symbol: "AAPL",
		Qty:    10,
		Side:   "buy",
		Type:   "market",
	})
	assert.NoError(t, err)
	assert.NotNil(t, orderResp)
	assert.Equal(t, "order123", orderResp.ID)
	assert.Equal(t, "AAPL", orderResp.Symbol)
	assert.Equal(t, "new", orderResp.Status)
}
