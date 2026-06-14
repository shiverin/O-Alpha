import { LandingShell } from "../../layout/LandingShell";

type TierFeature = {
  icon: string;
  text: string;
  filled?: boolean;
};

type TierCard = {
  tier: string;
  title: string;
  subtitle: string;
  accent: "base" | "primary" | "secondary";
  badge?: string;
  features: TierFeature[];
  cta: string;
};

type SpecRow = {
  label: string;
  basic: string;
  pro: string;
  institutional: string;
  iconOnly?: boolean;
};

const tiers: TierCard[] = [
  {
    tier: "Tier 01",
    title: "Basic",
    subtitle: "Standard Agent",
    accent: "base",
    features: [
      { icon: "settings", text: "Standard risk controls" },
      { icon: "category", text: "Major assets universe" },
      { icon: "monitoring", text: "Daily rebalancing" },
    ],
    cta: "Deploy Basic",
  },
  {
    tier: "Tier 02",
    title: "Pro",
    subtitle: "Advanced Factors",
    accent: "primary",
    badge: "POPULAR",
    features: [
      { icon: "functions", text: "Custom factor models", filled: true },
      { icon: "language", text: "Full asset universe", filled: true },
      { icon: "speed", text: "Intraday execution", filled: true },
    ],
    cta: "Deploy Pro",
  },
  {
    tier: "Tier 03",
    title: "Institutional",
    subtitle: "Full Stack Strategy",
    accent: "secondary",
    features: [
      { icon: "model_training", text: "Convex optimization" },
      { icon: "analytics", text: "HMM regime overlay" },
      { icon: "bolt", text: "Priority execution routing" },
    ],
    cta: "Contact Sales",
  },
];

const specs: SpecRow[] = [
  {
    label: "Asset Universe",
    basic: "Top 100 Equities",
    pro: "Full US Equities",
    institutional: "Global Multi-Asset",
  },
  {
    label: "Risk Model",
    basic: "Static Volatility Caps",
    pro: "Dynamic Factor Exposure",
    institutional: "Convex Optimization",
  },
  {
    label: "Regime Detection",
    basic: "close",
    pro: "Basic Trend Following",
    institutional: "Hidden Markov Models (HMM)",
    iconOnly: true,
  },
  {
    label: "Execution Frequency",
    basic: "End of Day",
    pro: "Hourly / Intraday",
    institutional: "Real-time Priority Routing",
  },
  {
    label: "API Access",
    basic: "close",
    pro: "check",
    institutional: "check",
    iconOnly: true,
  },
];

export function PricingPage() {
  return (
    <LandingShell activePath="/pricing" className="bg-scanline">
      <main className="flex-grow pt-32 pb-24">
        <section className="text-center px-margin-mobile md:px-margin-desktop max-w-[1440px] mx-auto mb-20 relative">
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[520px] h-[520px] bg-primary-container/10 rounded-full blur-3xl -z-10 pointer-events-none"></div>
          <h1 className="font-headline-xl text-headline-xl text-on-background mb-6">
            Scale Your Sophistication.
          </h1>
          <p className="font-body-md text-body-md text-on-surface-variant max-w-2xl mx-auto">
            Deploy the exact level of quantitative rigor your strategy requires.
            <br />
            From fundamental risk controls to institutional-grade convex
            optimization &
            <br />
            regime detection.
          </p>
        </section>

        <section className="grid grid-cols-1 md:grid-cols-3 gap-gutter px-margin-mobile md:px-margin-desktop max-w-[1440px] mx-auto mb-32 relative z-10">
          {tiers.map((tier) => (
            <div
              key={tier.title}
              className={
                tier.accent === "primary"
                  ? "bg-surface-container-high/90 border border-primary-container/40 rounded-lg p-8 flex flex-col relative transition-transform duration-300 ease-out hover:-translate-y-2"
                  : tier.accent === "secondary"
                    ? "bg-surface-container-high/80 border border-secondary-fixed/30 rounded-lg p-8 flex flex-col relative hover:border-secondary-fixed/50 transition-colors transition-transform duration-300 ease-out hover:-translate-y-2"
                    : "bg-surface-container-high/70 border border-outline-variant/40 rounded-lg p-8 flex flex-col hover:border-outline-variant/60 transition-colors transition-transform duration-300 ease-out hover:-translate-y-2"
              }
            >
              {tier.accent === "primary" && (
                <>
                  <div className="absolute top-0 left-0 w-full h-[2px] bg-primary-container"></div>
                  <div className="absolute top-4 right-4 bg-primary-container/10 border border-primary-container text-primary-container font-data-sm text-data-sm px-2 py-1 rounded">
                    {tier.badge}
                  </div>
                </>
              )}
              {tier.accent === "secondary" && (
                <div className="absolute top-0 left-0 w-full h-[2px] bg-secondary-fixed/50"></div>
              )}
              <div
                className={
                  tier.accent === "secondary"
                    ? "font-label-caps text-label-caps text-secondary-fixed mb-2 uppercase tracking-widest"
                    : tier.accent === "primary"
                      ? "font-label-caps text-label-caps text-primary-container mb-2 uppercase tracking-widest"
                      : "font-label-caps text-label-caps text-on-surface-variant mb-2 uppercase tracking-widest"
                }
              >
                {tier.tier}
              </div>
              <h2 className="font-headline-lg-mobile md:font-headline-lg text-headline-lg-mobile md:text-headline-lg text-on-background mb-1">
                {tier.title}
              </h2>
              <div
                className={
                  tier.accent === "secondary"
                    ? "font-data-lg text-data-lg text-secondary-fixed mb-8"
                    : tier.accent === "primary"
                      ? "font-data-lg text-data-lg text-primary-container mb-8"
                      : "font-data-lg text-data-lg text-primary-fixed mb-8"
                }
              >
                {tier.subtitle}
              </div>
              <div className="flex-grow space-y-4 mb-8">
                {tier.features.map((feature: TierFeature) => (
                  <div key={feature.text} className="flex items-start gap-3">
                    <span
                      className={
                        tier.accent === "primary"
                          ? "material-symbols-outlined text-primary-container text-[16px] mt-1"
                          : tier.accent === "secondary"
                            ? "material-symbols-outlined text-secondary-fixed text-[16px] mt-1"
                            : "material-symbols-outlined text-outline-variant text-[16px] mt-1"
                      }
                      style={
                        feature.filled
                          ? { fontVariationSettings: "'FILL' 1" }
                          : undefined
                      }
                    >
                      {feature.icon}
                    </span>
                    <span
                      className={
                        tier.accent === "primary"
                          ? "font-body-md text-body-md text-on-background"
                          : "font-body-md text-body-md text-on-surface-variant"
                      }
                    >
                      {feature.text}
                    </span>
                  </div>
                ))}
              </div>
              <button
                className={
                  tier.accent === "primary"
                    ? "w-full py-3 bg-primary-container text-[#000000] font-label-caps text-label-caps uppercase rounded font-bold hover:bg-primary-fixed transition-colors"
                    : tier.accent === "secondary"
                      ? "w-full py-3 border border-secondary-fixed text-secondary-fixed font-label-caps text-label-caps uppercase rounded hover:bg-secondary-fixed/10 transition-colors"
                      : "w-full py-3 border border-outline-variant text-on-background font-label-caps text-label-caps uppercase rounded hover:bg-surface-container-high transition-colors"
                }
              >
                {tier.cta}
              </button>
            </div>
          ))}
        </section>

        <section className="px-margin-mobile md:px-margin-desktop max-w-[1440px] mx-auto">
          <h3 className="font-headline-lg text-headline-lg text-on-background mb-8">
            Technical Specifications
          </h3>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse min-w-[800px]">
              <thead>
                <tr>
                  <th className="py-4 px-4 border-b border-outline-variant/30 font-label-caps text-label-caps text-on-surface-variant w-1/4">
                    CAPABILITY
                  </th>
                  <th className="py-4 px-4 border-b border-outline-variant/30 font-label-caps text-label-caps text-on-surface-variant w-1/4">
                    BASIC
                  </th>
                  <th className="py-4 px-4 border-b border-primary-container/50 font-label-caps text-label-caps text-primary-container w-1/4 bg-primary-container/5">
                    PRO
                  </th>
                  <th className="py-4 px-4 border-b border-secondary-fixed/50 font-label-caps text-label-caps text-secondary-fixed w-1/4">
                    INSTITUTIONAL
                  </th>
                </tr>
              </thead>
              <tbody className="font-body-md text-body-md text-on-background">
                {specs.map((row) => (
                  <tr key={row.label}>
                    <td className="py-4 px-4 border-b border-outline-variant/10">
                      {row.label}
                    </td>
                    <td className="py-4 px-4 border-b border-outline-variant/10 font-data-sm text-data-sm text-on-surface-variant">
                      {row.iconOnly && row.basic === "close" ? (
                        <span className="material-symbols-outlined text-[20px] text-outline-variant">
                          close
                        </span>
                      ) : (
                        row.basic
                      )}
                    </td>
                    <td className="py-4 px-4 border-b border-outline-variant/10 font-data-sm text-data-sm text-primary-fixed bg-primary-container/5">
                      {row.iconOnly && row.pro === "check" ? (
                        <span
                          className="material-symbols-outlined text-[20px] text-primary-container"
                          style={{ fontVariationSettings: "'FILL' 1" }}
                        >
                          check
                        </span>
                      ) : (
                        row.pro
                      )}
                    </td>
                    <td className="py-4 px-4 border-b border-outline-variant/10 font-data-sm text-data-sm text-secondary-fixed">
                      {row.iconOnly && row.institutional === "check" ? (
                        <span
                          className="material-symbols-outlined text-[20px] text-secondary-fixed"
                          style={{ fontVariationSettings: "'FILL' 1" }}
                        >
                          check
                        </span>
                      ) : (
                        row.institutional
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      </main>
    </LandingShell>
  );
}
