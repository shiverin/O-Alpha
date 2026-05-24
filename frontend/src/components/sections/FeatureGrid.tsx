import { Panel } from '@/components/ui/Panel';
import { Icon } from '@/components/ui/Icon';
import { getAccentStyle } from '@/lib/ui';

type FeatureCard = {
  title: string;
  description: string;
  icon: string;
  accent: "primary" | "secondary";
};

const features: FeatureCard[] = [
  {
    title: "Intent Engine",
    description:
      "Define your parameters in plain English or precise metrics. The agent understands risk tolerance, sector focus, and temporal horizons natively.",
    icon: "psychology",
    accent: "primary",
  },
  {
    title: "Continuous Learning",
    description:
      "Your agent evolves. It backtests theoretical strategies against live market data, suggesting optimizations to your core mandate without emotional bias.",
    icon: "model_training",
    accent: "secondary",
  },
  {
    title: "Micro-second Execution",
    description:
      "When conditions align, hesitation is eliminated. Direct exchange connectivity ensures your agent acts instantly on predefined triggers.",
    icon: "speed",
    accent: "primary",
  },
];

function FeatureCardView({ card }: { card: FeatureCard }) {
  const isPrimary = card.accent === "primary";

  return (
    <Panel
      className="p-8 relative overflow-hidden group transition-colors duration-300 hover:bg-surface-container-highest/80"
    >
      <div className={`w-12 h-12 rounded-full border border-outline-variant/40 flex items-center justify-center mb-6 transition-colors ${
        isPrimary
          ? 'group-hover:border-primary-container/60'
          : 'group-hover:border-secondary-fixed/60'
      }`}>
      <Icon
        name={card.icon}
        size="medium"
        color={isPrimary ? "text-primary-container" : "text-secondary-fixed"}
      />
      </div>
      <h3 className="font-headline-lg text-headline-lg text-on-background mb-4">
        {card.title}
      </h3>
      <p className="font-body-md text-body-md text-on-surface-variant">
        {card.description}
      </p>
    </Panel>
  );
}

export function FeatureGrid() {
  return (
    <section className="px-margin-desktop max-w-[1440px] mx-auto w-full">
      <div className="text-center mb-12">
        <h2 className="font-headline-xl text-headline-xl text-on-background mb-4">
          The Product
        </h2>
        <p className="font-body-md text-body-md text-on-surface-variant max-w-2xl mx-auto">
          Converting your unique market preferences into ruthless, continuous
          decisions. A seamless bridge between human intuition and machine
          execution.
        </p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-gutter">
        {features.map((card) => (
          <FeatureCardView key={card.title} card={card} />
        ))}
      </div>
    </section>
  );
}
