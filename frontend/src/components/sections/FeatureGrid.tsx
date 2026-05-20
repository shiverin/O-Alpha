import { SectionHeader } from "./SectionHeader";

type FeatureStat = {
  label: string;
  value: string;
  valueClassName?: string;
};

type FeatureCard = {
  title: string;
  description: string;
  iconLarge: string;
  iconSmall: string;
  accent: "primary" | "secondary";
  stats: FeatureStat[];
};

const accentStyles = {
  primary: {
    text: "text-primary-container",
    hoverBorder: "hover:border-primary-container/60",
  },
  secondary: {
    text: "text-secondary-container",
    hoverBorder: "hover:border-secondary-container/60",
  },
} as const;

const features: FeatureCard[] = [
  {
    title: "Kalman Filtering",
    description:
      "Continuous state estimation algorithm that extracts true underlying asset trends from noisy market micro-structure, enabling precise entry and exit timing.",
    iconLarge: "filter_alt",
    iconSmall: "timeline",
    accent: "primary",
    stats: [
      { label: "INPUT", value: "NOISY_TICK_DATA" },
      { label: "OUTPUT", value: "SMOOTHED_STATE_VECTOR", valueClassName: "text-secondary-container" },
    ],
  },
  {
    title: "HMM Regime Overlays",
    description:
      "Hidden Markov Models continuously classify current market states (e.g., Bull Volatile, Bear Stable) to dynamically adjust the agent's risk posture and strategy selection.",
    iconLarge: "multiline_chart",
    iconSmall: "donut_large",
    accent: "secondary",
    stats: [
      { label: "STATE_1", value: "RISK_ON_TRENDING" },
      { label: "STATE_2", value: "CAPITAL_PRESERVATION", valueClassName: "text-error" },
    ],
  },
  {
    title: "Convex Optimization",
    description:
      "Solves multi-variable portfolio construction problems in real-time, maximizing expected Sharpe ratio while strictly adhering to user-defined constraints.",
    iconLarge: "functions",
    iconSmall: "calculate",
    accent: "primary",
    stats: [
      { label: "OBJECTIVE", value: "MAX(SHARPE)" },
      { label: "CONSTRAINT", value: "USER_RISK_LIMITS", valueClassName: "text-secondary-container" },
    ],
  },
];

function FeatureCardView({ card }: { card: FeatureCard }) {
  const accent = accentStyles[card.accent];

  return (
    <div className={`bg-surface-container/60 backdrop-blur-md border border-outline-variant/40 p-8 rounded relative overflow-hidden group transition-colors duration-500 hover:bg-surface-container/80 ${accent.hoverBorder}`}>
      <div className="absolute top-0 right-0 p-4 opacity-15 group-hover:opacity-25 transition-opacity">
        <span className={`material-symbols-outlined text-[120px] ${accent.text}`}>
          {card.iconLarge}
        </span>
      </div>
      <span className={`material-symbols-outlined ${accent.text} mb-4 block`}>
        {card.iconSmall}
      </span>
      <h3 className="font-headline-lg text-headline-lg text-on-background mb-3 text-2xl">
        {card.title}
      </h3>
      <p className="font-body-md text-body-md text-on-surface-variant mb-6">
        {card.description}
      </p>
      <div className="flex flex-col gap-2 font-data-sm text-data-sm">
        {card.stats.map((stat) => (
          <div
            key={stat.label}
            className="flex justify-between border-b border-outline-variant/30 pb-1"
          >
            <span className="text-outline">{stat.label}</span>
            <span className={stat.valueClassName ?? accent.text}>
              {stat.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function FeatureGrid() {
  return (
    <section className="px-margin-desktop max-w-[1440px] mx-auto w-full">
      <SectionHeader label="AGENT BEHAVIOUR LOGIC" tone="secondary" />
      <div className="grid grid-cols-1 md:grid-cols-3 gap-gutter">
        {features.map((card) => (
          <FeatureCardView key={card.title} card={card} />
        ))}
      </div>
    </section>
  );
}
