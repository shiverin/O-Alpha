import type { EquityPoint } from "@/lib/api";

export const buildFallbackEquityCurve = (): EquityPoint[] => {
  const now = Date.now();
  const points = 120;

  return Array.from({ length: points }, (_, idx) => {
    const t = idx / (points - 1);

    const trend = 10000 + t * 1650;
    const cycle = Math.sin(t * Math.PI * 3.6) * 180;
    const pullback = Math.sin(t * Math.PI * 9) * 65;

    const volatility = 180;
    const noise = (Math.random() - 0.5) * volatility;

    const equity = trend + cycle + pullback + noise;

    return {
      time: new Date(now - (points - idx) * 24 * 60 * 60 * 1000).toISOString(),
      equity: Number(equity.toFixed(2)),
    };
  });
};

export const DEFAULT_EQUITY_CURVE = buildFallbackEquityCurve();
