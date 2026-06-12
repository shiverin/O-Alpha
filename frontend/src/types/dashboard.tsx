export interface ServerPortfolioSummary {
  total_asset_value: number;
  change_percent_24h: number;
  change_dollar_24h: number;
  estimated_annual_yield: number;
  target_progress_percent: number;
  timestamp: string;
}

export interface ServerTradeLog {
  id: number;
  timestamp: string;
  action: string;
  symbol: string;
  price: number;
  qty: number;
  slippage: number;
  status: string;
}

export interface MockLogItem {
  time: string;
  asset: string;
  side: string;
  price: string;
  primary?: boolean;
  highlight?: boolean;
}

export interface SnapshotPoint {
  total_asset_value: number;
  timestamp: string;
}

export interface ServerPositionMetrics {
  symbol: string;
  qty: number;
  avg_entry_price: number;
  current_price: number;
  unrealized_pnl: number;
  exposure: number;
}
