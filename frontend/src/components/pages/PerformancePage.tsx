import { LandingShell } from "../layout/LandingShell";

const metricTabs = ["1W", "1M", "YTD"] as const;

export function PerformancePage() {
  return (
    <LandingShell activePath="/performance" className="bg-performance-grid">
      <main className="pt-32 px-margin-mobile md:px-margin-desktop max-w-[1440px] mx-auto flex flex-col gap-16 md:gap-24">
        <section className="flex flex-col items-start max-w-4xl">
          <div className="flex items-center gap-3 mb-6">
            <div className="pulse-dot"></div>
            <span className="font-label-caps text-label-caps text-primary-container uppercase tracking-widest">
              Live Alpha Generation
            </span>
          </div>
          <h1 className="font-headline-xl text-headline-xl text-on-surface mb-6">
            Institutional-Grade <br className="hidden md:block" />
            <span className="text-secondary-container gold-glow">Performance.</span>
          </h1>
          <p className="font-body-md text-body-md text-on-surface-variant max-w-2xl">
            Continuous market scanning. Convex optimization. Regime-aware
            execution. O(Alpha) translates your risk appetite into systematic,
            absolute returns without emotional bias.
          </p>
        </section>

        <section>
          <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
            <span className="material-symbols-outlined text-sm">monitoring</span>
            Real-Time Metrics
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-12 gap-gutter">
            <div className="md:col-span-8 glass-card rounded p-6 flex flex-col justify-between h-[400px]">
              <div className="flex justify-between items-start mb-6">
                <div>
                  <span className="font-data-sm text-data-sm text-on-surface-variant block mb-1">
                    CUMULATIVE P&L (YTD)
                  </span>
                  <span className="font-data-lg text-[32px] font-medium text-primary-container neon-text">
                    +24.8%
                  </span>
                </div>
                <div className="flex gap-2">
                  {metricTabs.map((tab) => (
                    <span
                      key={tab}
                      className={
                        tab === "1W"
                          ? "font-data-sm text-data-sm bg-primary-container/10 border border-primary-container/30 text-primary-container px-2 py-1 rounded"
                          : "font-data-sm text-data-sm text-on-surface-variant px-2 py-1"
                      }
                    >
                      {tab}
                    </span>
                  ))}
                </div>
              </div>
              <div className="w-full flex-grow relative mt-4">
                <svg
                  className="absolute inset-0 w-full h-full"
                  preserveAspectRatio="none"
                  viewBox="0 0 100 50"
                >
                  <defs>
                    <linearGradient id="chart-gradient" x1="0%" y1="0%" x2="0%" y2="100%">
                      <stop offset="0%" stopColor="rgba(0, 229, 255, 0.2)"></stop>
                      <stop offset="100%" stopColor="rgba(0, 229, 255, 0)"></stop>
                    </linearGradient>
                  </defs>
                  <line
                    x1="0"
                    y1="25"
                    x2="100"
                    y2="25"
                    stroke="rgba(255,255,255,0.05)"
                    strokeWidth="0.5"
                    strokeDasharray="1 2"
                  ></line>
                  <path
                    className="chart-area chart-area-animate"
                    d="M0,40 C10,38 20,45 30,30 C40,15 50,25 60,10 C70,-5 80,15 90,5 L100,0 L100,50 L0,50 Z"
                  ></path>
                  <path
                    className="chart-line chart-line-animate"
                    d="M0,40 C10,38 20,45 30,30 C40,15 50,25 60,10 C70,-5 80,15 90,5 L100,0"
                  ></path>
                </svg>
              </div>
            </div>

            <div className="md:col-span-4 flex flex-col gap-gutter">
              <div className="glass-card rounded p-6 flex-1 flex flex-col justify-center">
                <span className="font-data-sm text-data-sm text-on-surface-variant mb-2">
                  SHARPE RATIO
                </span>
                <div className="flex items-baseline gap-2">
                  <span className="font-data-lg text-[40px] text-on-surface">
                    2.4
                  </span>
                  <span className="font-data-sm text-data-sm text-surface-tint">
                    Top Decile
                  </span>
                </div>
              </div>
              <div className="glass-card rounded p-6 flex-1 flex flex-col justify-center">
                <span className="font-data-sm text-data-sm text-on-surface-variant mb-2">
                  MAX DRAWDOWN (CONTROLLED)
                </span>
                <div className="flex items-baseline gap-2">
                  <span className="font-data-lg text-[40px] text-error">-4.2%</span>
                </div>
                <div className="mt-4 w-full bg-surface-container-highest h-1 rounded-full overflow-hidden">
                  <div className="bg-error h-full w-1/4"></div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="grid grid-cols-1 md:grid-cols-2 gap-gutter mb-4">
          <div>
            <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
              <span className="material-symbols-outlined text-sm">radar</span>
              Regime Detection
            </h2>
            <div className="glass-card rounded p-6 h-[300px] flex flex-col">
              <p className="font-body-md text-body-md text-on-surface-variant mb-6">
                Hidden Markov Models continuously classify market states,
                dynamically shifting agent posture.
              </p>
              <div className="flex-grow flex flex-col justify-around">
                <div className="flex items-center justify-between">
                  <span className="font-data-sm text-data-sm text-on-surface w-24">
                    BULL
                  </span>
                  <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative">
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 w-3/4 h-[2px] bg-secondary-container"></div>
                  </div>
                  <span className="font-data-sm text-data-sm text-secondary-container">
                    SCALING
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-data-sm text-data-sm text-on-surface w-24">
                    VOLATILE
                  </span>
                  <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative">
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1/2 h-[2px] bg-surface-tint"></div>
                  </div>
                  <span className="font-data-sm text-data-sm text-surface-tint">
                    ACTIVE
                  </span>
                </div>
                <div className="flex items-center justify-between opacity-50">
                  <span className="font-data-sm text-data-sm text-on-surface w-24">
                    BEAR
                  </span>
                  <div className="flex-grow mx-4 bg-surface-container-highest h-[2px] relative"></div>
                  <span className="font-data-sm text-data-sm text-on-surface-variant">
                    HEDGED
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div>
            <h2 className="font-label-caps text-label-caps text-surface-tint uppercase mb-6 flex items-center gap-2 border-b border-outline-variant/30 pb-2 inline-flex">
              <span className="material-symbols-outlined text-sm">shield</span>
              Risk Architecture
            </h2>
            <div className="glass-card rounded p-6 h-[300px] flex flex-col gap-4">
              <div className="border-b border-outline-variant/20 pb-4">
                <h3 className="font-data-lg text-data-lg text-on-surface mb-2">
                  Kalman Filtering
                </h3>
                <p className="font-body-md text-sm text-on-surface-variant">
                  Extracts true signal from market noise, enabling precise entry
                  points even in low-liquidity environments.
                </p>
              </div>
              <div className="pt-2">
                <h3 className="font-data-lg text-data-lg text-on-surface mb-2">
                  Convex Optimization
                </h3>
                <p className="font-body-md text-sm text-on-surface-variant">
                  Portfolio weights are solved continuously to maximize expected
                  return subject to strict variance constraints.
                </p>
              </div>
            </div>
          </div>
        </section>

      </main>
    </LandingShell>
  );
}
