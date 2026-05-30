"use client";

import { Container } from "@/components/ui/Container";
import { cn } from "@/lib/utils";

type TickerTrend = "positive" | "negative" | "neutral";

type TickerItem = {
  symbol: string;
  change: string;
  trend: TickerTrend;
};

const tickerRows: {
  items: TickerItem[];
  rowClassName: string;
  reverse?: boolean;
}[] = [
  {
    rowClassName: "top-[28%] opacity-80",
    items: [
      { symbol: "NVDA", change: "+2.8%", trend: "positive" },
      { symbol: "AMD", change: "+1.4%", trend: "positive" },
      { symbol: "SPY", change: "+0.6%", trend: "positive" },
      { symbol: "QQQ", change: "+1.1%", trend: "positive" },
      { symbol: "TSLA", change: "-0.7%", trend: "negative" },
    ],
  },
  {
    rowClassName: "top-[45%] opacity-70",
    reverse: true,
    items: [
      { symbol: "AAPL", change: "+0.9%", trend: "positive" },
      { symbol: "MSFT", change: "+1.0%", trend: "positive" },
      { symbol: "META", change: "+1.7%", trend: "positive" },
      { symbol: "AMZN", change: "-0.5%", trend: "negative" },
      { symbol: "GOOGL", change: "-0.3%", trend: "negative" },
    ],
  },
  {
    rowClassName: "top-[62%] opacity-60",
    items: [
      { symbol: "AVGO", change: "+1.8%", trend: "positive" },
      { symbol: "NFLX", change: "+1.2%", trend: "positive" },
      { symbol: "SHOP", change: "-1.1%", trend: "negative" },
      { symbol: "COIN", change: "+1.9%", trend: "positive" },
      { symbol: "INTC", change: "0.0%", trend: "neutral" },
    ],
  },
];

const tickerRepeatGroups = Array.from({ length: 8 }, (_, index) => index);

const tickerChipClasses: Record<TickerTrend, string> = {
  positive:
    "border-emerald-300/35 bg-emerald-400/15 shadow-[0_0_18px_rgba(52,211,153,0.16)]",
  negative:
    "border-rose-300/35 bg-rose-400/15 shadow-[0_0_18px_rgba(251,113,133,0.16)]",
  neutral:
    "border-slate-300/25 bg-slate-200/10 shadow-[0_0_16px_rgba(203,213,225,0.08)]",
};

const tickerValueClasses: Record<TickerTrend, string> = {
  positive: "text-emerald-200",
  negative: "text-rose-200",
  neutral: "text-slate-200",
};

function TickerChip({ ticker }: { ticker: TickerItem }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-2 rounded-full border px-3 py-1.5 leading-none ring-1 ring-white/[0.04] backdrop-blur-[1px]",
        tickerChipClasses[ticker.trend],
      )}
    >
      <span className="text-on-background/90">{ticker.symbol}</span>
      <span className={cn("font-medium", tickerValueClasses[ticker.trend])}>
        {ticker.change}
      </span>
    </span>
  );
}

export function Hero() {
  return (
    <section className="relative min-h-screen flex items-center overflow-hidden">
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_72%_38%,rgba(125,170,255,0.14),transparent_48%),radial-gradient(circle_at_22%_70%,rgba(117,210,185,0.10),transparent_50%)]" />
        {tickerRows.map((row, rowIndex) => (
          <div
            key={rowIndex}
            className={cn(
              "ticker-row absolute inset-x-0 overflow-hidden py-3",
              row.rowClassName,
            )}
          >
            <div
              className={cn(
                "ticker-track w-max font-data-sm text-data-sm uppercase tracking-[0.16em]",
                row.reverse && "ticker-track-reverse",
              )}
            >
              {tickerRepeatGroups.map((groupIndex) => (
                <div
                  className="ticker-group"
                  key={`${rowIndex}-${groupIndex}`}
                  aria-hidden={groupIndex > 0}
                >
                  {row.items.map((ticker) => (
                    <TickerChip
                      key={`${ticker.symbol}-${groupIndex}`}
                      ticker={ticker}
                    />
                  ))}
                </div>
              ))}
            </div>
          </div>
        ))}
        <div className="absolute inset-0 bg-gradient-to-r from-background from-[0%] via-background/80 via-[38%] to-background/15" />
      </div>

      <Container>
        <div className="max-w-3xl">
          <h1 className="font-headline-xl text-headline-xl text-on-background mb-6">
            <span className="text-primary-container">Build</span> Your Own
            Trading Agent.
          </h1>
          <p className="font-body-md text-body-md text-on-surface-variant text-lg max-w-2xl">
            O(Alpha) translates your strategic intent into autonomous execution.
            High-frequency capability meets institutional-grade intelligence,
            now in your control.
          </p>
        </div>
      </Container>

      <style jsx>{`
        .ticker-row {
          mask-image: linear-gradient(
            90deg,
            transparent,
            #000 7%,
            #000 93%,
            transparent
          );
          -webkit-mask-image: linear-gradient(
            90deg,
            transparent,
            #000 7%,
            #000 93%,
            transparent
          );
        }

        .ticker-track {
          display: flex;
          white-space: nowrap;
          animation: tickerScroll 22s linear infinite;
          text-shadow: 0 0 12px rgba(185, 241, 255, 0.22);
          will-change: transform;
        }

        .ticker-group {
          display: inline-flex;
          gap: 1rem;
          padding-right: 1rem;
        }

        .ticker-track-reverse {
          animation-direction: reverse;
          animation-duration: 26s;
        }

        @keyframes tickerScroll {
          0% {
            transform: translateX(0);
          }
          100% {
            transform: translateX(-12.5%);
          }
        }
      `}</style>
    </section>
  );
}
