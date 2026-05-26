"use client";

import { Container } from "@/components/ui/Container";

export function Hero() {
  return (
    <section className="relative min-h-screen flex items-center overflow-hidden">
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_72%_38%,rgba(125,170,255,0.14),transparent_48%),radial-gradient(circle_at_22%_70%,rgba(117,210,185,0.10),transparent_50%)]" />
        <div className="absolute inset-x-0 top-[28%] overflow-hidden opacity-45">
          <div className="ticker-track w-max rounded-full border border-outline-variant/20 bg-surface-container-high/45 px-6 py-3 font-data-sm text-data-sm uppercase tracking-[0.22em] text-on-surface-variant/80 backdrop-blur-[2px]">
            <span>NVDA +2.8%</span>
            <span>AMD +1.4%</span>
            <span>SPY +0.6%</span>
            <span>QQQ +1.1%</span>
            <span>TSLA +3.2%</span>
            <span>NVDA +2.8%</span>
            <span>AMD +1.4%</span>
            <span>SPY +0.6%</span>
            <span>QQQ +1.1%</span>
            <span>TSLA +3.2%</span>
          </div>
        </div>
        <div className="absolute inset-x-0 top-[45%] overflow-hidden opacity-35">
          <div className="ticker-track ticker-track-reverse w-max rounded-full border border-outline-variant/15 bg-surface-container-high/35 px-6 py-3 font-data-sm text-data-sm uppercase tracking-[0.22em] text-on-surface-variant/75 backdrop-blur-[2px]">
            <span>AAPL +0.9%</span>
            <span>MSFT +1.0%</span>
            <span>META +1.7%</span>
            <span>AMZN +1.3%</span>
            <span>GOOGL +0.8%</span>
            <span>AAPL +0.9%</span>
            <span>MSFT +1.0%</span>
            <span>META +1.7%</span>
            <span>AMZN +1.3%</span>
            <span>GOOGL +0.8%</span>
          </div>
        </div>
        <div className="absolute inset-x-0 top-[62%] overflow-hidden opacity-28">
          <div className="ticker-track w-max rounded-full border border-outline-variant/10 bg-surface-container-high/25 px-6 py-3 font-data-sm text-data-sm uppercase tracking-[0.22em] text-on-surface-variant/65 backdrop-blur-[2px]">
            <span>AVGO +1.8%</span>
            <span>NFLX +1.2%</span>
            <span>SHOP +2.1%</span>
            <span>COIN +1.9%</span>
            <span>INTC +0.4%</span>
            <span>AVGO +1.8%</span>
            <span>NFLX +1.2%</span>
            <span>SHOP +2.1%</span>
            <span>COIN +1.9%</span>
            <span>INTC +0.4%</span>
          </div>
        </div>
        <div className="absolute inset-0 bg-gradient-to-r from-background via-background/70 to-background/45" />
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
        .ticker-track {
          display: inline-flex;
          gap: 1.25rem;
          white-space: nowrap;
          animation: tickerScroll 22s linear infinite;
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
            transform: translateX(-50%);
          }
        }
      `}</style>
    </section>
  );
}
