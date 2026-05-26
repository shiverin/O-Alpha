import type { EquityPoint } from "@/lib/api";

export const buildFallbackEquityCurve = (): EquityPoint[] => {
  const now = Date.now();

  // 1. Increased points for a more detailed, dense chart
  const points = 120;

  return Array.from({ length: points }, (_, idx) => {
    const t = idx / (points - 1);

    // The underlying smooth market direction
    const trend = 10000 + t * 1650;
    const cycle = Math.sin(t * Math.PI * 3.6) * 180;
    const pullback = Math.sin(t * Math.PI * 9) * 65;

    // 2. THE VARIANCE: Random jagged market noise
    // Math.random() - 0.5 generates a number between -0.5 and 0.5.
    // Multiply by a volatility factor (e.g., 180) to make the spikes larger or smaller.
    const volatility = 180;
    const noise = (Math.random() - 0.5) * volatility;

    // Combine the smooth math with the chaotic noise
    const equity = trend + cycle + pullback + noise;

    return {
      time: new Date(now - (points - idx) * 24 * 60 * 60 * 1000).toISOString(), // 1 point per day
      equity: Number(equity.toFixed(2)),
    };
  });
};

export const DEFAULT_EQUITY_CURVE = buildFallbackEquityCurve();
