import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AgentSettingsPage from "./page";
import {
  agentStatusApi,
  settingsApi,
  strategyCatalogApi,
  type AgentRunSummary,
  type StrategyCatalogResponse,
} from "@/lib/api";

const mocks = vi.hoisted(() => ({
  useAuth: vi.fn(),
}));

vi.mock("@/context/AuthContext", () => ({
  useAuth: mocks.useAuth,
}));

vi.mock("@/components/app/AppShell", () => ({
  AppShell: ({ children }: { children: React.ReactNode }) => (
    <main>{children}</main>
  ),
}));

vi.mock("@/components/EquityCurveChart", () => ({
  EquityCurveChart: () => <div data-testid="equity-chart" />,
}));

vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    ...actual,
    agentStatusApi: { list: vi.fn() },
    runBacktestStream: vi.fn(),
    settingsApi: { check: vi.fn(), save: vi.fn() },
    strategyCatalogApi: { list: vi.fn() },
  };
});

const catalog: StrategyCatalogResponse = {
  default_universe: ["VOO", "AAPL", "MSFT"],
  recommended: {
    conservative: "ranker_proxy_h63_low",
    moderate: "ranker_proxy_h63_medium",
    aggressive: "composite_momentum_high",
  },
  strategies: [
    {
      key: "ranker_proxy_h63_low",
      display_name: "Low risk proxy",
      family: "ranker",
      risk_profile: "low",
      deployment_status: "conservative_variant",
      promoted_checkpoint: false,
      requires_model_artifacts: false,
      paper_only: true,
      benchmark_symbol: "VOO",
      description: "Low-risk catalog strategy.",
    },
    {
      key: "ranker_proxy_h63_medium",
      display_name: "Medium risk proxy",
      family: "ranker",
      risk_profile: "medium",
      deployment_status: "promoted_research_checkpoint",
      promoted_checkpoint: true,
      requires_model_artifacts: false,
      paper_only: true,
      benchmark_symbol: "VOO",
      description: "Medium-risk catalog strategy.",
    },
    {
      key: "composite_momentum_high",
      display_name: "High risk momentum",
      family: "momentum",
      risk_profile: "high",
      deployment_status: "experimental_variant",
      promoted_checkpoint: false,
      requires_model_artifacts: false,
      paper_only: true,
      benchmark_symbol: "VOO",
      description: "High-risk catalog strategy.",
    },
  ],
};

const activePortfolioRun = {
  id: 99,
  symbol: "VOO",
  strategy_type: "PORTFOLIO_CATALOG",
  timeframe: "1Day",
  mode: "paper",
  status: "running",
  initial_cash: 100000,
  started_at: "2026-07-05T00:00:00Z",
} as AgentRunSummary;

describe("AgentSettingsPage active-run safety", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(window, "alert").mockImplementation(() => {});
    mocks.useAuth.mockReturnValue({
      user: { id: 7, username: "ada", is_onboarded: true },
    });
    vi.mocked(strategyCatalogApi.list).mockResolvedValue(catalog);
    vi.mocked(settingsApi.check).mockResolvedValue({
      found: true,
      settings: {
        risk_profile: "moderate",
        leverage: 2,
        max_positions: 6,
        stop_loss_pct: 2.5,
        take_profit_pct: 5,
        rebalance_freq: "daily",
      },
    });
    vi.mocked(settingsApi.save).mockResolvedValue({ status: "synchronized" });
  });

  it("disables save and all advanced controls while a portfolio agent is active", async () => {
    vi.mocked(agentStatusApi.list).mockResolvedValue({
      agents: [activePortfolioRun],
    });

    render(<AgentSettingsPage />);

    expect(
      await screen.findByText(
        /stop it from the dashboard before editing settings/i,
      ),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /save terminal configuration/i }),
    ).toBeDisabled();

    await userEvent.click(
      screen.getByRole("button", { name: /advanced tuning/i }),
    );
    const sliders = document.querySelectorAll<HTMLInputElement>(
      'input[type="range"]',
    );
    expect(sliders).toHaveLength(4);
    sliders.forEach((slider) => expect(slider).toBeDisabled());
    expect(screen.getByRole("button", { name: "hourly" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "daily" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "weekly" })).toBeDisabled();
    expect(settingsApi.save).not.toHaveBeenCalled();
  });

  it("does not reset advanced settings when the saved profile is clicked while locked", async () => {
    vi.mocked(settingsApi.check).mockResolvedValue({
      found: true,
      settings: {
        risk_profile: "moderate",
        leverage: 5,
        max_positions: 11,
        stop_loss_pct: 6,
        take_profit_pct: 13,
        rebalance_freq: "weekly",
      },
    });
    vi.mocked(agentStatusApi.list).mockResolvedValue({
      agents: [activePortfolioRun],
    });

    render(<AgentSettingsPage />);

    await screen.findByText(/before editing settings/i);
    await userEvent.click(
      screen.getByRole("button", { name: /advanced tuning/i }),
    );
    const leverageSlider = document.querySelector<HTMLInputElement>(
      'input[type="range"]',
    );
    await waitFor(() => expect(leverageSlider).toHaveValue("5"));

    await userEvent.click(screen.getByText("moderate", { selector: "h4" }));

    expect(leverageSlider).toHaveValue("5");
    expect(
      screen.getByText(
        /stop the running portfolio agent before editing settings/i,
      ),
    ).toBeInTheDocument();
  });

  it("performs a fresh active-run check before save and blocks a late-starting run", async () => {
    vi.mocked(agentStatusApi.list)
      .mockResolvedValueOnce({ agents: [] })
      .mockResolvedValueOnce({ agents: [activePortfolioRun] });

    render(<AgentSettingsPage />);

    const save = await screen.findByRole("button", {
      name: /save terminal configuration/i,
    });
    await waitFor(() => expect(save).toBeEnabled());
    await userEvent.click(save);

    expect(
      await screen.findByText(
        /stop the running portfolio agent before editing settings/i,
      ),
    ).toBeInTheDocument();
    expect(settingsApi.save).not.toHaveBeenCalled();
  });

  it("saves the expected payload while inactive", async () => {
    vi.mocked(agentStatusApi.list).mockResolvedValue({ agents: [] });

    render(<AgentSettingsPage />);

    await userEvent.click(
      await screen.findByRole("button", { name: /advanced tuning/i }),
    );
    const leverageSlider = document.querySelector<HTMLInputElement>(
      'input[type="range"]',
    );
    expect(leverageSlider).not.toBeNull();
    fireEvent.change(leverageSlider!, { target: { value: "3" } });

    await userEvent.click(
      screen.getByRole("button", { name: /save terminal configuration/i }),
    );

    await waitFor(() =>
      expect(settingsApi.save).toHaveBeenCalledWith({
        risk_profile: "moderate",
        leverage: 3,
        max_positions: 6,
        stop_loss_pct: 2.5,
        take_profit_pct: 5,
        rebalance_freq: "daily",
        strategy_key: undefined,
        backtest_accepted: undefined,
      }),
    );
  });
});
