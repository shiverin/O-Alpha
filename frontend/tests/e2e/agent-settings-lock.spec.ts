import { expect, test } from "@playwright/test";

test("settings page locks controls while a portfolio agent is active", async ({
  page,
}) => {
  await page.route("**/auth/me", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({ id: 7, username: "ada", is_onboarded: true }),
    });
  });
  await page.route("**/api/v1/user/settings", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        found: true,
        settings: {
          risk_profile: "moderate",
          leverage: 2,
          max_positions: 6,
          stop_loss_pct: 2.5,
          take_profit_pct: 5,
          rebalance_freq: "daily",
        },
      }),
    });
  });
  await page.route("**/api/v1/agent/list", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        agents: [
          {
            id: 99,
            symbol: "VOO",
            strategy_type: "PORTFOLIO_CATALOG",
            timeframe: "1Day",
            mode: "paper",
            status: "running",
            initial_cash: 100000,
            started_at: "2026-07-05T00:00:00Z",
          },
        ],
      }),
    });
  });
  await page.route("**/api/v1/strategies/catalog", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        default_universe: ["VOO", "AAPL", "MSFT"],
        recommended: {
          conservative: "ranker_proxy_h63_low",
          moderate: "ranker_proxy_h63_medium",
          aggressive: "composite_momentum_high",
        },
        strategies: [],
      }),
    });
  });

  await page.addInitScript((token) => {
    window.sessionStorage.setItem("token", token);
    document.cookie = `oa-auth=${token}; path=/`;
  }, fakeJwt());

  await page.goto("/app/agent-settings");

  await expect(
    page.getByRole("heading", { name: "Configuration" }),
  ).toBeVisible();
  await expect(
    page.getByText(/stop it from the dashboard before editing settings/i),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: /save terminal configuration/i }),
  ).toBeDisabled();

  await page.getByRole("button", { name: /advanced tuning/i }).click();
  await expect(page.locator('input[type="range"]').first()).toBeDisabled();
  await expect(page.getByRole("button", { name: "daily" })).toBeDisabled();
});

function fakeJwt() {
  const header = base64Url({ alg: "none", typ: "JWT" });
  const payload = base64Url({
    user_id: 7,
    username: "ada",
    is_onboarded: true,
    exp: Math.floor(Date.now() / 1000) + 3600,
  });
  return `${header}.${payload}.signature`;
}

function base64Url(value: Record<string, unknown>) {
  return Buffer.from(JSON.stringify(value)).toString("base64url");
}
