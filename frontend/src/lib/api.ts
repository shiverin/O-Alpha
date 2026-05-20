const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

export interface BacktestRequest {
  symbol: string;
  start: string;
  end: string;
  fast_window?: number;
  slow_window?: number;
  initial_cash?: number;
}

export interface EquityPoint {
  time: string;
  equity: number;
}

export interface BacktestResult {
  symbol: string;
  equity_curve: EquityPoint[];
  sharpe: number;
  sortino: number;
  max_drawdown: number;
  total_return: number;
  annual_return: number;
  num_trades: number;
}

export async function runBacktest(
  payload: BacktestRequest
): Promise<BacktestResult> {
  const res = await fetch(`${API_BASE}/api/v1/backtest`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Backtest failed (${res.status})`);
  }

  return res.json();
}
