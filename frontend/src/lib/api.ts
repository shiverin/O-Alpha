import { getToken } from "@/lib/auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ??
  "http://localhost:8080";

export interface BacktestRequest {
  symbol: string;
  start: string;
  end: string;
  strategy_type?: string; // "MA_CROSSOVER" or "KALMAN"
  timeframe?: string; // e.g., "1Day"
  // Kalman-specific parameters
  q_noise?: number;
  r_noise?: number;
  z_threshold?: number;
  // MA-specific parameters
  fast_period?: number;
  slow_period?: number;
  initial_cash?: number;
}

export interface EquityPoint {
  time: string;
  equity: number;
}

export interface BacktestResult {
  symbol: string;
  equity_curve: EquityPoint[];
  final_equity: number;
  sharpe: number;
  sortino: number;
  max_drawdown: number;
  total_return: number;
  annual_return?: number;
  num_trades: number;
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
  completeOnboarding: async (): Promise<{ status: string }> => {
    return api.post<{ status: string }, Record<string, never>>(
      "/api/v1/user/onboarding/complete",
      {},
    );
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
