"use client";

import { Icon } from '@/components/ui/Icon';
import { Container } from '@/components/ui/Container';

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
    label: "The",
    title: "Conservative",
    summary: "Capital preservation focus.",
    accent: "neutral",
    items: ["Low volatility targeting", "Blue-chip preference", "Strict trailing stop-loss"],
  },
  {
    label: "The",
    title: "Opportunist",
    summary: "Agile momentum capture.",
    accent: "primary",
    badge: "POPULAR",
    items: ["Trend identification algorithms", "Mid-cap sector rotation", "Dynamic position sizing"],
  },
  {
    label: "The",
    title: "Quant",
    summary: "Statistical arbitrage.",
    accent: "secondary",
    items: ["Mean reversion modeling", "High-frequency capability", "Complex multi-leg options"],
  },
];

export function ExecutionFlow() {
  return (
    <Container>
      {/* Redesigned Minimalist Header */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 mb-16">
        <div className="max-w-xl">
          <span className="text-[10px] uppercase tracking-[0.3em] text-on-surface-variant font-medium mb-4 block">
            Agent Personalities
          </span>
          <h2 className="text-3xl md:text-5xl font-light tracking-tight text-on-background mb-4">
            Customisable Profiles
          </h2>
          <p className="text-sm md:text-base font-light leading-relaxed text-on-surface-variant">
            Deploy specialized agents tailored to distinct market environments. Switch profiles with a single command.
          </p>
        </div>
        
        {/* Sleek Glass Button */}
        <button className="px-6 py-2.5 rounded-full border border-outline-variant/40 bg-surface-container-low text-on-surface text-sm font-medium tracking-wide hover:bg-surface-container hover:border-outline-variant/60 transition-all duration-300 ease-out">
          Compare Strategies
        </button>
      </div>

      {/* The Profile Grid */}
      <div className="grid grid-cols-1 gap-6 md:gap-8">
        {profiles.map((profile, index) => {
          const isPrimary = profile.accent === "primary";
          const isSecondary = profile.accent === "secondary";
          
          // Dynamic Theme Mapping for Glows & Accents
          const accentGlow = isPrimary 
            ? 'via-primary-container/60' 
            : isSecondary 
              ? 'via-secondary-fixed/60' 
              : 'via-on-surface-variant/40';
              
          const ambientGlow = isPrimary 
            ? 'bg-primary-container/10' 
            : isSecondary 
              ? 'bg-secondary-fixed/10' 
              : 'bg-on-surface-variant/5';

          const textAccent = isPrimary 
            ? 'text-primary-container' 
            : isSecondary 
              ? 'text-secondary-fixed' 
              : 'text-on-surface-variant';

          return (
            <div
              key={profile.title}
              // Buttery smooth float interaction on card hover
              className="group relative flex flex-col h-full bg-surface-container-low border border-outline-variant/30 rounded-[32px] overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:-translate-y-2 hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)]"
              style={{ transitionDelay: `${index * 50}ms` }}
            >
              {/* Premium top-edge hover glow */}
              <div className={`absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent ${accentGlow} to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700`} />
              
              {/* Ambient radial background glow (Subtle) */}
              <div className={`absolute -top-32 -right-32 w-64 h-64 rounded-full blur-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none ${ambientGlow}`} />

              {/* Glass Pill Badge */}
              {profile.badge && (
                <div className="absolute top-6 right-6 bg-primary-container/10 border border-primary-container/20 text-primary-container text-[10px] font-medium tracking-widest px-3 py-1 rounded-full backdrop-blur-md">
                  {profile.badge}
                </div>
              )}

              {/* Card Header Section */}
              <div className="p-8 pb-6 flex flex-col relative z-10">
                <span className={`text-[10px] uppercase tracking-[0.2em] font-medium mb-3 ${textAccent}`}>
                  {profile.label}
                </span>
                <h3 className="text-2xl font-light tracking-wide text-on-surface mb-2">
                  {profile.title}
                </h3>
                <p className="text-sm font-light text-on-surface-variant">
                  {profile.summary}
                </p>
              </div>

              {/* Seamless Gradient Divider */}
              <div className="h-px w-full bg-gradient-to-r from-transparent via-outline-variant/20 to-transparent relative z-10" />

              {/* Clean Features List */}
              <ul className="p-8 pt-6 space-y-4 relative z-10 flex-1">
                {profile.items.map((item) => (
                  <li key={item} className="flex items-start gap-3 group/item cursor-default">
                    {/* Icon brightens slightly on individual item hover */}
                    <div className="mt-0.5 opacity-60 group-hover/item:opacity-100 transition-opacity duration-300">
                      <Icon
                        name="check"
                        size="small"
                        className={textAccent}
                      />
                    </div>
                    {/* Text brightens to pure on-surface when hovered */}
                    <span className="text-sm font-light text-on-surface-variant group-hover/item:text-on-surface transition-colors duration-300">
                      {item}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          );
        })}
      </div>
    </Container>
  );
}