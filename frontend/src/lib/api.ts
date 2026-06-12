import { getToken } from "@/lib/auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ??
  "http://localhost:8080";

export interface BacktestRequest {
  symbol?: string;
  symbols?: string[];
  start: string;
  end: string;
  strategy_type: string; // "MA_CROSSOVER", "KALMAN", or portfolio strategy types
  timeframe?: string; // e.g., "1Day"
  // Kalman-specific parameters
  q_noise?: number;
  r_noise?: number;
  z_threshold?: number;
  // MA-specific parameters
  fast_period?: number;
  slow_period?: number;
  initial_cash?: number;
  parameters?: Record<string, unknown>;
}

export interface EquityPoint {
  time: string;
  equity: number;
}

export interface BacktestResult {
  symbol?: string;
  symbols?: string[];
  equity_curve: EquityPoint[];
  metrics?: {
    total_return: number;
    annual_return: number;
    sharpe: number;
    sortino: number;
    max_drawdown: number;
    num_trades: number;
    turnover: number;
  };
  final_equity?: number;
  sharpe?: number;
  sortino?: number;
  max_drawdown?: number;
  total_return?: number;
  annual_return?: number;
  num_trades?: number;
}

export interface PortfolioLiveSummary {
  total_asset_value: number;
  change_percent_24h: number;
  change_dollar_24h: number;
  estimated_annual_yield: number;
  target_progress_percent: number;
  timestamp: string;
}

export interface PortfolioLivePosition {
  symbol: string;
  qty: number;
  avg_entry_price: number;
  current_price: number;
  unrealized_pnl: number;
  exposure: number;
}

export type PortfolioLiveEvent =
  | {
      type: "snapshot";
      timestamp: string;
      summary?: PortfolioLiveSummary | null;
      positions?: PortfolioLivePosition[] | null;
      history?: PortfolioLiveSummary[] | null;
    }
  | {
      type: "price";
      timestamp: string;
      symbol: string;
      price: number;
    };

export interface BacktestProgress {
  index: number;
  total: number;
  point: EquityPoint;
  percent: number;
}

export type BacktestStreamEvent =
  | { type: "started" }
  | { type: "progress"; progress: BacktestProgress }
  | { type: "completed"; result: BacktestResult }
  | { type: "error"; error: string };

const nextAnimationFrame = () =>
  new Promise<void>((resolve) => {
    if (typeof window !== "undefined" && window.requestAnimationFrame) {
      window.requestAnimationFrame(() => resolve());
      return;
    }
    setTimeout(resolve, 0);
  });

const getAuthHeaders = (): HeadersInit => {
  const token = typeof window !== "undefined" ? getToken() : null;
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return headers;
};

export async function runBacktest(
  payload: BacktestRequest,
): Promise<BacktestResult> {
  const res = await fetch(`${API_BASE}/api/v1/backtest`, {
    method: "POST",
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Backtest failed (${res.status})`);
  }

  return res.json();
}

export async function runBacktestStream(
  payload: BacktestRequest,
  onEvent: (event: BacktestStreamEvent) => void,
): Promise<BacktestResult> {
  const res = await fetch(`${API_BASE}/api/v1/backtest/stream`, {
    method: "POST",
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    if (res.status === 404 || res.status === 405) {
      return runBacktest(payload);
    }
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Backtest failed (${res.status})`);
  }
  if (!res.body) {
    throw new Error("Backtest stream did not return a response body");
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let finalResult: BacktestResult | null = null;
  let progressEventsSincePaint = 0;

  const processLine = async (line: string) => {
    const trimmed = line.trim();
    if (!trimmed) return;
    const event = JSON.parse(trimmed) as BacktestStreamEvent;
    onEvent(event);
    if (event.type === "completed") {
      finalResult = event.result;
    }
    if (event.type === "error") {
      throw new Error(event.error || "Backtest stream failed");
    }
    if (event.type === "progress") {
      progressEventsSincePaint += 1;
      if (progressEventsSincePaint >= 8) {
        progressEventsSincePaint = 0;
        await nextAnimationFrame();
      }
    }
  };

  while (true) {
    const { value, done } = await reader.read();
    buffer += decoder.decode(value, { stream: !done });
    const lines = buffer.split("\n");
    buffer = lines.pop() ?? "";
    for (const line of lines) {
      await processLine(line);
    }
    if (done) break;
  }
  await processLine(buffer);

  if (!finalResult) {
    throw new Error("Backtest stream ended before returning a final result");
  }
  return finalResult;
}

export async function streamPortfolioLive(
  onEvent: (event: PortfolioLiveEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const res = await fetch(`${API_BASE}/api/v1/user/portfolio/live`, {
    headers: getAuthHeaders(),
    signal,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Live stream failed (${res.status})`);
  }
  if (!res.body) {
    throw new Error("Live stream did not return a response body");
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { value, done } = await reader.read();
    buffer += decoder.decode(value, { stream: !done });
    const lines = buffer.split("\n");
    buffer = lines.pop() ?? "";
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed) continue;
      onEvent(JSON.parse(trimmed) as PortfolioLiveEvent);
    }
    if (done) break;
  }

  const trimmed = buffer.trim();
  if (trimmed) {
    onEvent(JSON.parse(trimmed) as PortfolioLiveEvent);
  }
}

export const api = {
  get: async <R>(endpoint: string): Promise<R> => {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      headers: getAuthHeaders(),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error ?? `Request failed (${res.status})`);
    }

    return res.json();
  },

  post: async <R, T = unknown>(endpoint: string, data: T): Promise<R> => {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error ?? `Request failed (${res.status})`);
    }

    return res.json();
  },

  put: async <R, T = unknown>(endpoint: string, data: T): Promise<R> => {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      method: "PUT",
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error ?? `Request failed (${res.status})`);
    }

    return res.json();
  },

  delete: async <R>(endpoint: string): Promise<R> => {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      method: "DELETE",
      headers: getAuthHeaders(),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error ?? `Request failed (${res.status})`);
    }

    return res.json();
  },
};

export interface ServerAgentSettings {
  risk_profile: string;
  leverage: number;
  max_positions: number;
  stop_loss_pct: number;
  take_profit_pct: number;
  rebalance_freq: string;
}

export interface SettingsCheckResponse {
  found: boolean;
  settings?: ServerAgentSettings;
}

export const settingsApi = {
  check: async (): Promise<SettingsCheckResponse> => {
    return api.get<SettingsCheckResponse>("/api/v1/user/settings");
  },
  save: async (payload: {
    risk_profile: string;
    leverage: number;
    max_positions: number;
    stop_loss_pct: number;
    take_profit_pct: number;
    rebalance_freq: string;
    strategy_key?: string;
    backtest_accepted?: boolean;
  }): Promise<{ status: string }> => {
    return api.post<{ status: string }, typeof payload>(
      "/api/v1/user/settings",
      payload,
    );
  },
};

export const userApi = {
  completeOnboarding: async (payload: {
    risk_profile: string;
    strategy_key: string;
    backtest_accepted: boolean;
  }): Promise<{
    status: string;
    risk_profile: string;
    strategy_key: string;
  }> => {
    return api.post<
      { status: string; risk_profile: string; strategy_key: string },
      typeof payload
    >("/api/v1/user/onboarding/complete", payload);
  },
};

export interface AgentControlPayload {
  symbol: string;
  strategy_type: "KALMAN" | "MA_CROSSOVER" | "HMM_ENSEMBLE";
  timeframe?: string;
  initial_cash?: number;
  use_websocket?: boolean;
  q_noise?: number;
  r_noise?: number;
  z_threshold?: number;
  fast_period?: number;
  slow_period?: number;
  risk_profile?: "conservative" | "moderate" | "aggressive";
}

export const agentApi = {
  start: async (
    payload: AgentControlPayload,
  ): Promise<{ status: string; symbol: string; run_id: number }> => {
    return api.post<
      { status: string; symbol: string; run_id: number },
      AgentControlPayload
    >("/api/v1/agent/start", payload);
  },
  stop: async (symbol: string): Promise<{ status: string; symbol: string }> => {
    return api.post<{ status: string; symbol: string }, { symbol: string }>(
      "/api/v1/agent/stop",
      { symbol },
    );
  },
};

export interface StrategySpec {
  key: string;
  display_name: string;
  family: string;
  risk_profile: "low" | "medium" | "high";
  deployment_status: string;
  promoted_checkpoint: boolean;
  requires_model_artifacts: boolean;
  paper_only: boolean;
  benchmark_symbol: string;
  description: string;
  evidence_paths?: string[];
  notes?: string[];
}

export interface StrategyCatalogResponse {
  strategies: StrategySpec[];
  recommended: {
    conservative: string;
    moderate: string;
    aggressive: string;
  };
  default_universe: string[];
}

export interface PortfolioAgentStartPayload {
  strategy_key?: string;
  risk_profile?: "conservative" | "moderate" | "aggressive";
  symbols?: string[];
  timeframe?: string;
  initial_cash?: number;
}

export interface PortfolioAgentStartResult {
  status: string;
  run_id: number;
  strategy_key: string;
  display_name: string;
  deployment_status: string;
  paper_only: boolean;
  symbols: string[];
}

export interface AgentRunSummary {
  id: number;
  symbol: string;
  strategy_type: string;
  strategy_key?: string;
  timeframe: string;
  mode: string;
  status: string;
  initial_cash: number;
  parameters?: Record<string, unknown>;
  runtime_state?: {
    source?: string;
    benchmark_symbol?: string;
    model_healthy?: boolean;
    regime_label?: string;
    confidence?: number;
    probability_low?: number;
    probability_medium?: number;
    probability_high?: number;
    overlay_role?: string;
    overlay_multiplier?: number;
    overlay_confidence?: number;
    overlay_vetoed?: boolean;
    bar_time?: string;
    updated_at?: string;
  };
  started_at: string;
  last_heartbeat_at?: string;
}

export const strategyCatalogApi = {
  list: async (): Promise<StrategyCatalogResponse> => {
    return api.get<StrategyCatalogResponse>("/api/v1/strategies/catalog");
  },
};

export const portfolioAgentApi = {
  start: async (
    payload: PortfolioAgentStartPayload,
  ): Promise<PortfolioAgentStartResult> => {
    return api.post<PortfolioAgentStartResult, PortfolioAgentStartPayload>(
      "/api/v1/agent/portfolio/start",
      payload,
    );
  },
  stop: async (): Promise<{ status: string }> => {
    return api.post<{ status: string }, Record<string, never>>(
      "/api/v1/agent/portfolio/stop",
      {},
    );
  },
};

export const agentStatusApi = {
  list: async (): Promise<{ agents: AgentRunSummary[] }> => {
    return api.get<{ agents: AgentRunSummary[] }>("/api/v1/agent/list");
  },
};
