package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/oalpha/internal/alpaca"
)

func main() {
	apiKey := os.Getenv("ALPACA_API_KEY")
	apiSecret := os.Getenv("ALPACA_API_SECRET")
	baseURL := os.Getenv("ALPACA_BASE_URL")

	if apiKey == "" || apiSecret == "" || baseURL == "" {
		fmt.Println("Missing required env vars:")
		fmt.Println("   ALPACA_API_KEY")
		fmt.Println("   ALPACA_API_SECRET")
		fmt.Println("   ALPACA_BASE_URL")
		os.Exit(1)
	}

	fmt.Println("\nTest 1: Client Creation")
	fmt.Println("-----------------------")
	client := alpaca.NewClient(baseURL, apiKey, apiSecret)
	fmt.Printf("Alpaca client created\n")
	fmt.Printf("  Base URL: %s\n", client.BaseURL())
	fmt.Printf("  API Key: %s***\n", apiKey[:8])

	fmt.Println("\nTest 2: GetBars (Market Data)")
	fmt.Println("-------------------------------")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	end := time.Now().UTC()
	start := end.Add(-7 * 24 * time.Hour)

	fmt.Printf("  Symbol: AAPL\n")
	fmt.Printf("  Timeframe: 1Day\n")
	fmt.Printf("  Range: %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))

	bars, err := client.GetBars(ctx, "AAPL", "1Day", start, end, 100)
	if err != nil {
		fmt.Printf("GetBars failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("GetBars succeeded. Fetched %d bars\n", len(bars))

	if len(bars) > 0 {
		fmt.Println("\n  Latest bar data:")
		b := bars[len(bars)-1]
		fmt.Printf("     Time:   %s\n", b.Time)
		fmt.Printf("     Symbol: %s\n", b.Symbol)
		fmt.Printf("     Open:   $%.2f\n", b.Open)
		fmt.Printf("     High:   $%.2f\n", b.High)
		fmt.Printf("     Low:    $%.2f\n", b.Low)
		fmt.Printf("     Close:  $%.2f\n", b.Close)
		fmt.Printf("     Volume: %d\n", b.Volume)
	}

	fmt.Println("\nTest 3: Bar Data Validation")
	fmt.Println("----------------------------")
	validCount := 0
	for _, b := range bars {
		if err := alpaca.ValidateBar(b); err == nil {
			validCount++
		} else {
			fmt.Printf("Invalid bar at %s: %v\n", b.Time, err)
			os.Exit(1)
		}
	}
	fmt.Printf("All %d bars are valid\n", validCount)

	fmt.Println("\nTest 4: Order Request Validation")
	fmt.Println("---------------------------------")

	orderReq := &alpaca.OrderRequest{
		Symbol: "AAPL",
		Qty:    10,
		Side:   "buy",
		Type:   "market",
	}

	orderResp, err := client.PlaceOrder(ctx, orderReq)
	if err != nil {
		fmt.Printf("PlaceOrder failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PlaceOrder validation succeeded\n")
	fmt.Printf("  Order ID: %s\n", orderResp.ID)
	fmt.Printf("  Symbol:   %s\n", orderResp.Symbol)
	fmt.Printf("  Qty:      %s\n", orderResp.Qty)
	fmt.Printf("  Side:     %s\n", orderResp.Side)
	fmt.Printf("  Status:   %s\n", orderResp.Status)

	separator := "===================================================="
	fmt.Println("\n" + separator)
	fmt.Println("ALL TESTS PASSED!")
	fmt.Println(separator)
	fmt.Println("\nYour Alpaca API dataflow is working correctly:")
	fmt.Println("  1. Can connect to Alpaca API")
	fmt.Println("  2. Can fetch market data (bars)")
	fmt.Println("  3. Data validation works")
	fmt.Println("  4. Order validation works")
	fmt.Println("\nNext steps:")
	fmt.Println("  - Set up database migrations")
	fmt.Println("  - Configure ingest service")
	fmt.Println("  - Test data storage pipeline")
}
