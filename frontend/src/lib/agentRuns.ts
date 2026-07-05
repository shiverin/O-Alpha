import type { AgentRunSummary } from "@/lib/api";

const activeStatuses = new Set(["starting", "running"]);

export function hasActivePortfolioAgent(
  agents?: AgentRunSummary[] | null,
): boolean {
  return (
    agents?.some(
      (agent) =>
        agent.strategy_type === "PORTFOLIO_CATALOG" &&
        activeStatuses.has(agent.status),
    ) ?? false
  );
}
