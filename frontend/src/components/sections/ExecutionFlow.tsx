"use client";

import { useState } from "react";
import { Icon } from "@/components/ui/Icon";
import { Container } from "@/components/ui/Container";

type ProfileId = "conservative" | "opportunist" | "quant";

type ProfileCard = {
  id: ProfileId;
  label: string;
  title: string;
  summary: string;
  accent: "primary" | "secondary" | "neutral";
  items: string[];
  badge?: string;
};

const profileDescriptions: Record<ProfileId, string> = {
  conservative:
    "Prioritizes capital preservation with lower exposure, fewer active positions, and slower rebalance cadence.",
  opportunist:
    "Balances active sleeve discovery with moderate exposure, daily cadence, and standard exit controls.",
  quant:
    "Allows wider paper-risk limits for comparison runs while keeping catalog strategy recipes unchanged with statistical arbitrage.",
};

const profiles: ProfileCard[] = [
  {
    id: "conservative",
    label: "The",
    title: "Conservative",
    summary: "Capital preservation focus.",
    accent: "neutral",
    items: [
      "Low volatility targeting",
      "Blue-chip preference",
      "Strict trailing stop-loss",
    ],
  },
  {
    id: "opportunist",
    label: "The",
    title: "Opportunist",
    summary: "Agile momentum capture.",
    accent: "primary",
    badge: "POPULAR",
    items: [
      "Trend identification algorithms",
      "Mid-cap sector rotation",
      "Dynamic position sizing",
    ],
  },
  {
    id: "quant",
    label: "The",
    title: "Quant",
    summary: "Statistical arbitrage.",
    accent: "secondary",
    items: [
      "Mean reversion modeling",
      "High-frequency capability",
      "Complex multi-leg options",
    ],
  },
];

export function ExecutionFlow() {
  const [flippedCards, setFlippedCards] = useState<Record<ProfileId, boolean>>({
    conservative: false,
    opportunist: false,
    quant: false,
  });

  const toggleCardFlip = (profileId: ProfileId, e: React.MouseEvent) => {
    e.stopPropagation();
    setFlippedCards((prev) => ({ ...prev, [profileId]: !prev[profileId] }));
  };

  return (
    <Container>
      <div className="mb-16 max-w-xl">
        <span className="text-[10px] uppercase tracking-[0.3em] text-on-surface-variant font-medium mb-4 block">
          Agent Personalities
        </span>
        <h2 className="text-3xl md:text-5xl font-light tracking-tight text-on-background mb-4">
          Customisable Profiles
        </h2>
        <p className="text-sm md:text-base font-light leading-relaxed text-on-surface-variant">
          Deploy specialized agents tailored to distinct market environments.
          <br />
          Switch profiles with a single command.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 md:gap-8">
        {profiles.map((profile, index) => {
          const isPrimary = profile.accent === "primary";
          const isSecondary = profile.accent === "secondary";
          const isFlipped = flippedCards[profile.id];

          const accentGlow = isPrimary
            ? "via-primary-container/60"
            : isSecondary
              ? "via-secondary-fixed/60"
              : "via-on-surface-variant/40";

          const ambientGlow = isPrimary
            ? "bg-primary-container/10"
            : isSecondary
              ? "bg-secondary-fixed/10"
              : "bg-on-surface-variant/5";

          const textAccent = isPrimary
            ? "text-primary-container"
            : isSecondary
              ? "text-secondary-fixed"
              : "text-on-surface-variant";

          return (
            <div
              key={profile.id}
              className="group [perspective:1000px] min-h-[380px] w-full select-none"
              style={{ transitionDelay: `${index * 50}ms` }}
            >
              <div
                className={`relative h-full min-h-[380px] w-full transition-transform duration-500 [transform-style:preserve-3d] ${
                  isFlipped ? "[transform:rotateY(180deg)]" : ""
                }`}
              >
                <div className="absolute inset-0 [backface-visibility:hidden] flex flex-col h-full bg-surface-container-low border border-outline-variant/30 rounded-[32px] overflow-hidden transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] group-hover:bg-surface-container group-hover:-translate-y-2 group-hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
                  <div
                    className={`absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent ${accentGlow} to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700`}
                  />

                  <div
                    className={`absolute -top-32 -right-32 w-64 h-64 rounded-full blur-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none ${ambientGlow}`}
                  />

                  {profile.badge && (
                    <div className="absolute top-6 left-6 bg-primary-container/10 border border-primary-container/20 text-primary-container text-[10px] font-medium tracking-widest px-3 py-1 rounded-full backdrop-blur-md z-20">
                      {profile.badge}
                    </div>
                  )}

                  <button
                    type="button"
                    onClick={(e) => toggleCardFlip(profile.id, e)}
                    className="absolute right-6 top-6 z-20 text-on-surface-variant/30 hover:text-primary-fixed-dim transition-colors h-7 w-7 rounded-full flex items-center justify-center hover:bg-white/5"
                    aria-label={`Learn more about ${profile.title}`}
                  >
                    <span className="material-symbols-outlined text-[18px]">
                      help
                    </span>
                  </button>

                  <div className="p-8 pb-6 flex flex-col relative z-10">
                    <span
                      className={`text-[10px] uppercase tracking-[0.2em] font-medium mb-3 ${textAccent}`}
                    >
                      {profile.label}
                    </span>
                    <h3 className="text-2xl font-light tracking-wide text-on-surface mb-2">
                      {profile.title}
                    </h3>
                    <p className="text-sm font-light text-on-surface-variant">
                      {profile.summary}
                    </p>
                  </div>

                  <div className="h-px w-full bg-gradient-to-r from-transparent via-outline-variant/20 to-transparent relative z-10" />

                  <ul className="p-8 pt-6 space-y-4 relative z-10 flex-1">
                    {profile.items.map((item) => (
                      <li
                        key={item}
                        className="flex items-start gap-3 group/item cursor-default"
                      >
                        <div className="mt-0.5 opacity-60 group-hover/item:opacity-100 transition-opacity duration-300">
                          <Icon
                            name="check"
                            size="small"
                            className={textAccent}
                          />
                        </div>
                        <span className="text-sm font-light text-on-surface-variant group-hover/item:text-on-surface transition-colors duration-300">
                          {item}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="absolute inset-0 [backface-visibility:hidden] [transform:rotateY(180deg)] flex flex-col justify-center bg-surface-container-high border border-outline-variant/40 rounded-[32px] p-8 shadow-xl">
                  <button
                    type="button"
                    onClick={(e) => toggleCardFlip(profile.id, e)}
                    className="absolute right-6 top-6 text-primary-fixed-dim/70 hover:text-primary-fixed-dim h-7 w-7 rounded-full flex items-center justify-center bg-white/5"
                    aria-label={`Back to ${profile.title} overview`}
                  >
                    <span className="material-symbols-outlined text-[16px]">
                      flip_to_front
                    </span>
                  </button>

                  <span
                    className={`text-[10px] uppercase tracking-[0.2em] font-medium mb-3 ${textAccent}`}
                  >
                    {profile.label} {profile.title}
                  </span>
                  <p className="text-sm font-light leading-relaxed text-on-surface-variant/90 pr-8 select-text">
                    {profileDescriptions[profile.id]}
                  </p>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </Container>
  );
}
