package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/pkg/models"
)

const yahooChartBaseURL = "https://query1.finance.yahoo.com/v8/finance/chart/"

// YahooDailyClient fetches adjusted daily bars from Yahoo's chart endpoint for
// research backfills. It is not used for live execution.
type YahooDailyClient struct {
	HTTPClient *http.Client
	BaseURL    string
	UserAgent  string
}

func (c YahooDailyClient) FetchDailyBars(ctx context.Context, symbol string, start time.Time, end time.Time) ([]models.Bar, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if !end.After(start) {
		return nil, fmt.Errorf("end must be after start")
	}

	base := strings.TrimSpace(c.BaseURL)
	if base == "" {
		base = yahooChartBaseURL
	}
	endpoint, err := url.Parse(base + url.PathEscape(symbol))
	if err != nil {
		return nil, fmt.Errorf("parse yahoo chart url: %w", err)
	}
	q := endpoint.Query()
	q.Set("period1", strconv.FormatInt(start.Unix(), 10))
	q.Set("period2", strconv.FormatInt(end.Unix(), 10))
	q.Set("interval", "1d")
	q.Set("events", "history")
	q.Set("includeAdjustedClose", "true")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build yahoo request: %w", err)
	}
	userAgent := strings.TrimSpace(c.UserAgent)
	if userAgent == "" {
		userAgent = "Mozilla/5.0"
	}
	req.Header.Set("User-Agent", userAgent)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch yahoo bars for %s: %w", symbol, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo bars for %s returned status %d", symbol, resp.StatusCode)
	}

	var payload yahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode yahoo bars for %s: %w", symbol, err)
	}
	return ParseYahooDailyBars(symbol, payload)
}

func ParseYahooDailyBars(symbol string, payload yahooChartResponse) ([]models.Bar, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if payload.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo chart error: %s", payload.Chart.Error.Description)
	}
	if len(payload.Chart.Result) == 0 {
		return nil, nil
	}
	result := payload.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, nil
	}
	quote := result.Indicators.Quote[0]
	adjCloses := []float64Ptr(nil)
	if len(result.Indicators.AdjClose) > 0 {
		adjCloses = result.Indicators.AdjClose[0].AdjClose
	}

	n := minInt5(len(result.Timestamp), len(quote.Open), len(quote.High), len(quote.Low), len(quote.Close))
	if len(quote.Volume) < n {
		n = len(quote.Volume)
	}
	if len(adjCloses) > 0 && len(adjCloses) < n {
		n = len(adjCloses)
	}

	bars := make([]models.Bar, 0, n)
	for i := 0; i < n; i++ {
		open, high, low, closePrice := quote.Open[i], quote.High[i], quote.Low[i], quote.Close[i]
		volume := quote.Volume[i]
		if open == nil || high == nil || low == nil || closePrice == nil || volume == nil {
			continue
		}
		if !validPositive(*open) || !validPositive(*high) || !validPositive(*low) || !validPositive(*closePrice) || *volume < 0 {
			continue
		}

		factor := 1.0
		adjustedClose := *closePrice
		if len(adjCloses) > i && adjCloses[i] != nil && validPositive(*adjCloses[i]) {
			adjustedClose = *adjCloses[i]
			factor = adjustedClose / *closePrice
		}
		adjustedOpen := *open * factor
		adjustedHigh := *high * factor
		adjustedLow := *low * factor
		adjustedHigh = math.Max(adjustedHigh, math.Max(adjustedOpen, adjustedClose))
		adjustedLow = math.Min(adjustedLow, math.Min(adjustedOpen, adjustedClose))
		bar := models.Bar{
			Time:   time.Unix(result.Timestamp[i], 0).UTC(),
			Symbol: symbol,
			Open:   adjustedOpen,
			High:   adjustedHigh,
			Low:    adjustedLow,
			Close:  adjustedClose,
			Volume: int64(math.Round(*volume)),
		}
		if bar.High < bar.Low {
			continue
		}
		bars = append(bars, bar)
	}
	return bars, nil
}

type float64Ptr = *float64

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64Ptr `json:"open"`
					High   []float64Ptr `json:"high"`
					Low    []float64Ptr `json:"low"`
					Close  []float64Ptr `json:"close"`
					Volume []float64Ptr `json:"volume"`
				} `json:"quote"`
				AdjClose []struct {
					AdjClose []float64Ptr `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func validPositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func minInt5(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	out := values[0]
	for _, value := range values[1:] {
		if value < out {
			out = value
		}
	}
	return out
}
