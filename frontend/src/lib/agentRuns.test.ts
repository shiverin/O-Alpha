import { describe, expect, it } from "vitest";
import { hasActivePortfolioAgent } from "@/lib/agentRuns";
import type { AgentRunSummary } from "@/lib/api";

const run = (strategyType: string, status: string): AgentRunSummary =>
  ({
    id: 1,
    symbol: "VOO",
    strategy_type: strategyType,
    timeframe: "1Day",
    mode: "paper",
    status,
    initial_cash: 100000,
    started_at: "2026-07-05T00:00:00Z",
  }) as AgentRunSummary;

describe("hasActivePortfolioAgent", () => {
  it("detects starting and running portfolio catalog runs", () => {
    expect(
      hasActivePortfolioAgent([run("PORTFOLIO_CATALOG", "starting")]),
    ).toBe(true);
    expect(hasActivePortfolioAgent([run("PORTFOLIO_CATALOG", "running")])).toBe(
      true,
    );
  });

  it("ignores stopped portfolio and non-portfolio runs", () => {
    expect(hasActivePortfolioAgent([run("PORTFOLIO_CATALOG", "stopped")])).toBe(
      false,
    );
    expect(hasActivePortfolioAgent([run("MA_CROSSOVER", "running")])).toBe(
      false,
    );
  });
});
