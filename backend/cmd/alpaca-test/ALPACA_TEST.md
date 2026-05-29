# Alpaca API Dataflow Test Guide

## Quick Start

Run this test to verify your Alpaca API integration works:

```bash
# From repo root
./backend/cmd/alpaca-test/test-alpaca.sh

# Or manually from backend dir
export $(cat .env | grep -v '^#' | xargs)
go run ./cmd/alpaca-test
```

## What Gets Tested

1. **Client Creation** - Verify credentials are loaded correctly
2. **GetBars API** - Fetch real market data from Alpaca
3. **Data Validation** - Check OHLCV bars are valid
4. **Order Validation** - Verify order request validation

## Expected Output

```
✓ Alpaca client created
  Base URL: https://data.alpaca.markets
  API Key: PKU6RVJF***

📊 Testing GetBars API...
  Symbol: AAPL
  Timeframe: 1Day
  Range: 2026-05-18 to 2026-05-25
✓ GetBars succeeded! Fetched X bars

  Latest bar:
    Time:   2026-05-22 04:00:00 +0000 UTC
    Symbol: AAPL
    OHLC:   306.06 / 311.39 / 306.06 / 308.81
    Volume: 1275219

✅ All tests passed!
```

## Troubleshooting

| Issue                       | Solution                                               |
| --------------------------- | ------------------------------------------------------ |
| `Missing required env vars` | Run: `export $(cat ../.env \| grep -v '^#' \| xargs)`  |
| `404 Not Found`             | Check ALPACA_BASE_URL is `https://data.alpaca.markets` |
| `Unauthorized`              | Verify ALPACA_API_KEY and ALPACA_API_SECRET in .env    |
| Connection timeout          | Check internet connection and firewall                 |

## Running Individual Tests

```bash
# Just unit tests
go test ./internal/alpaca -v

# Just live API test
go run ./cmd/alpaca-test
```
