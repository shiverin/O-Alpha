package portfolio

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/oalpha/internal/alpaca"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePaperBroker struct {
	placed       []*alpaca.OrderRequest
	byID         map[string][]*alpaca.OrderResponse
	byClientID   map[string]*alpaca.OrderResponse
	lookupClient error
}

func (f *fakePaperBroker) PlaceOrder(_ context.Context, req *alpaca.OrderRequest) (*alpaca.OrderResponse, error) {
	f.placed = append(f.placed, req)
	return &alpaca.OrderResponse{
		ID:            "order-" + req.ClientOrderID,
		ClientOrderID: req.ClientOrderID,
		Symbol:        req.Symbol,
		Qty:           fmt.Sprintf("%.8f", *req.Qty),
		Side:          req.Side,
		Status:        "accepted",
	}, nil
}

func (f *fakePaperBroker) GetOrder(_ context.Context, orderID string) (*alpaca.OrderResponse, error) {
	queue := f.byID[orderID]
	if len(queue) == 0 {
		return nil, fmt.Errorf("missing order %s", orderID)
	}
	order := queue[0]
	if len(queue) > 1 {
		f.byID[orderID] = queue[1:]
	}
	return order, nil
}

func (f *fakePaperBroker) GetOrderByClientOrderID(_ context.Context, clientOrderID string) (*alpaca.OrderResponse, error) {
	if f.lookupClient != nil {
		return nil, f.lookupClient
	}
	order := f.byClientID[clientOrderID]
	if order == nil {
		return nil, fmt.Errorf("missing client order %s", clientOrderID)
	}
	return order, nil
}

func TestAlpacaPaperPollOrderRecordsFilledStatus(t *testing.T) {
	broker := &fakePaperBroker{
		byID: map[string][]*alpaca.OrderResponse{
			"order-1": {
				{ID: "order-1", Status: "accepted"},
				{ID: "order-1", Status: "filled", FilledQty: "2.5", FilledAvgPrice: "101.25"},
			},
		},
	}
	router := NewAlpacaPaperExecutionRouter(nil, broker, 1, 2, 100000)
	router.pollInterval = time.Millisecond
	router.pollTimeout = time.Second

	order, err := router.pollOrder(context.Background(), &alpaca.OrderResponse{ID: "order-1", Status: "accepted"})
	require.NoError(t, err)
	assert.Equal(t, "filled", order.Status)
	assert.Equal(t, 2.5, filledQty(order))
	assert.Equal(t, 101.25, filledAvgPrice(order))
}

func TestAlpacaPaperPollOrderReturnsRejectedWithoutFill(t *testing.T) {
	router := NewAlpacaPaperExecutionRouter(nil, &fakePaperBroker{}, 1, 2, 100000)
	router.pollInterval = time.Millisecond
	router.pollTimeout = time.Second

	order, err := router.pollOrder(context.Background(), &alpaca.OrderResponse{
		ID:             "order-1",
		Status:         "rejected",
		RejectedReason: "insufficient buying power",
	})
	require.NoError(t, err)
	assert.Equal(t, "rejected", order.Status)

	err = router.recordConfirmedFill(context.Background(), "client-1", "BUY_LONG", "AAPL", order)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "produced no fill")
	assert.Contains(t, err.Error(), "insufficient buying power")
}

func TestPortfolioExecutionModeSelectsRouter(t *testing.T) {
	internal := NewPortfolioOrchestrator(nil, nil, nil, nil, nil, nil, "internal", StrategyCatalogConfig{})
	assert.Equal(t, PortfolioExecutionInternal, internal.executionMode)
	assert.IsType(t, &DBExecutionRouter{}, internal.executionRouterFor(1, 2, 100000))

	alpacaPaper := NewPortfolioOrchestrator(nil, nil, nil, nil, nil, &fakePaperBroker{}, "alpaca_paper", StrategyCatalogConfig{})
	assert.Equal(t, PortfolioExecutionAlpacaPaper, alpacaPaper.executionMode)
	assert.IsType(t, &AlpacaPaperExecutionRouter{}, alpacaPaper.executionRouterFor(1, 2, 100000))

	fallback := NewPortfolioOrchestrator(nil, nil, nil, nil, nil, nil, "alpaca_paper", StrategyCatalogConfig{})
	assert.Equal(t, PortfolioExecutionAlpacaPaper, fallback.executionMode)
	assert.IsType(t, &DBExecutionRouter{}, fallback.executionRouterFor(1, 2, 100000))
}

func TestNormalizePortfolioExecutionMode(t *testing.T) {
	assert.Equal(t, PortfolioExecutionAlpacaPaper, normalizePortfolioExecutionMode(" ALPACA_PAPER "))
	assert.Equal(t, PortfolioExecutionInternal, normalizePortfolioExecutionMode(""))
	assert.Equal(t, PortfolioExecutionInternal, normalizePortfolioExecutionMode("something_else"))
}
