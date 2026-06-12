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
