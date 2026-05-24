type ProfileCard = {
  label: string;
  title: string;
  summary: string;
  accent: "primary" | "secondary" | "neutral";
  items: string[];
  badge?: string;
};

const profiles: ProfileCard[] = [
  {
    label: "PROFILE 01",
    title: "Conservative",
    summary: "Capital preservation focus.",
    accent: "neutral",
    items: ["Low volatility targeting", "Blue-chip preference", "Strict trailing stop-loss"],
  },
  {
    label: "PROFILE 02",
    title: "Opportunist",
    summary: "Agile momentum capture.",
    accent: "primary",
    badge: "POPULAR",
    items: ["Trend identification algorithms", "Mid-cap sector rotation", "Dynamic position sizing"],
  },
  {
    label: "PROFILE 03",
    title: "Quant",
    summary: "Statistical arbitrage.",
    accent: "secondary",
    items: ["Mean reversion modeling", "High-frequency capability", "Complex multi-leg options"],
  },
];

import { Panel } from '@/components/ui/Panel';
import { Icon } from '@/components/ui/Icon';
import { getAccentStyle, getBorderStyle } from '@/lib/ui';

export function ExecutionFlow() {
  return (
    <section className="px-margin-desktop max-w-[1440px] mx-auto w-full">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 mb-12">
        <div>
          <h2 className="font-headline-xl text-headline-xl text-on-background mb-4">
            Customisable Profiles
          </h2>
          <p className="font-body-md text-body-md text-on-surface-variant max-w-xl">
            Deploy specialized agents tailored to distinct market environments.
            Switch profiles with a single command.
          </p>
        </div>
        <button className="px-6 py-2 rounded-full border border-outline-variant text-on-background font-body-md hover:bg-surface-container-high transition-colors">
          Compare Strategies
        </button>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-gutter">
        {profiles.map((profile) => {
          const isPrimary = profile.accent === "primary";
          const isSecondary = profile.accent === "secondary";

          return (
            <Panel
              key={profile.title}
              className={
                `overflow-hidden relative transform lg:scale-105 shadow-[0_16px_32px_rgba(0,0,0,0.25)] transition-transform duration-300 ease-out hover:-translate-y-2 ${
                  isPrimary
                    ? getBorderStyle('strong').replace('border ', 'border-primary-container/40 ') + 'bg-surface-container-high/80'
                    : getBorderStyle('medium').replace('border ', 'border-outline-variant/40 ') + 'bg-surface-container-high/70'
                }`
              }
            >
              {profile.badge && (
                <div className="absolute top-4 right-4 bg-primary-container text-on-primary-container font-data-sm text-data-sm px-3 py-1 rounded-full">
                  {profile.badge}
                </div>
              )}
              <div className="p-8 border-b border-outline-variant/30">
                <span
                  className={
                    isSecondary
                      ? "font-data-sm text-data-sm text-secondary-fixed"
                      : isPrimary
                        ? "font-data-sm text-data-sm text-primary-container"
                        : "font-data-sm text-data-sm text-on-surface-variant"
                  }
                >
                  {profile.label}
                </span>
                <h3 className="font-headline-lg text-headline-lg text-on-background mt-2 mb-2">
                  {profile.title}
                </h3>
                <p className="font-body-md text-body-md text-on-surface-variant">
                  {profile.summary}
                </p>
              </div>
              <ul className="p-8 space-y-4">
                {profile.items.map((item) => (
                  <li
                    key={item}
                    className="flex items-center gap-3 text-on-surface font-body-md border-b border-outline-variant/20 pb-4 last:border-b-0 last:pb-0"
                  >
                    <Icon
                      name="check"
                      size="small"
                      color={
                        isSecondary
                          ? "text-secondary-fixed"
                          : isPrimary
                            ? "text-primary-container"
                            : "text-on-surface-variant"
                      }
                    />
                    {item}
                  </li>
                ))}
              </ul>
            </Panel>
          );
        })}
      </div>
    </section>
  );
}
