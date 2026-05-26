"use client";

import { Icon } from "@/components/ui/Icon";
import { useEffect, useRef, useState } from "react";

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
    // Uses your surface-container tokens for a clean, theme-aware background
    <div className="group relative flex flex-col h-full bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-8 overflow-hidden hover:bg-surface-container transition-colors duration-500 ease-out">
      {/* Premium subtle top-edge glow on hover using your theme accents */}
      <div
        className={`absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent ${isPrimary ? "via-primary-container/60" : "via-secondary-fixed/60"} to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700`}
      />

      {/* Ambient background glow mapped to primary/secondary tokens */}
      <div
        className={`absolute -top-24 -right-24 w-48 h-48 rounded-full blur-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none ${isPrimary ? "bg-primary-container/10" : "bg-secondary-fixed/10"}`}
      />

      {/* Icon Wrapper */}
      <div className="relative z-10 flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl border border-outline-variant/40 bg-surface-container-highest mb-8 group-hover:scale-110 transition-transform duration-500 ease-out">
        <Icon
          name={card.icon}
          size="medium"
          className={`transition-colors duration-500 ${isPrimary ? "text-primary-container/80 group-hover:text-primary-container" : "text-secondary-fixed/80 group-hover:text-secondary-fixed"}`}
        />
      </div>

      {/* Typography mapped to your 'on-surface' tokens */}
      <div className="relative z-10 flex flex-col flex-1">
        <h3 className="text-lg font-medium tracking-wide text-on-surface mb-4">
          {card.title}
        </h3>
        <p className="text-sm font-light leading-relaxed text-on-surface-variant">
          {card.description}
        </p>
      </div>
    </div>
  );
}

export function FeatureGrid() {
  const cardRefs = useRef<Array<HTMLDivElement | null>>([]);
  const [isVisible, setIsVisible] = useState<boolean[]>(() =>
    features.map(() => false),
  );

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const index = Number((entry.target as HTMLElement).dataset.index);
            setIsVisible((current) => {
              const next = [...current];
              next[index] = true;
              return next;
            });
          }
        });
      },
      {
        threshold: 0.2,
        rootMargin: "0px 0px -5% 0px",
      },
    );

    const currentRefs = cardRefs.current;
    currentRefs.forEach((element) => {
      if (element) observer.observe(element);
    });

    return () => {
      currentRefs.forEach((element) => {
        if (element) observer.unobserve(element);
      });
    };
  }, []);

  return (
    <section className="px-6 md:px-12 max-w-[1200px] mx-auto w-full py-24">
      {/* Redesigned Header using theme typography tokens */}
      <div className="text-center mb-20 flex flex-col items-center">
        <span className="text-[10px] uppercase tracking-[0.3em] text-on-surface-variant font-medium mb-4 block">
          The Product
        </span>
        <h2 className="text-3xl md:text-5xl font-light tracking-tight text-on-background mb-6 max-w-2xl">
          Human intuition, <br className="hidden md:block" />
          <span className="text-on-surface-variant/60">
            optmized to perfection.
          </span>
        </h2>
      </div>

      {/* The Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 md:gap-8">
        {features.map((card, index) => (
          <div
            key={card.title}
            ref={(element) => {
              cardRefs.current[index] = element;
            }}
            data-index={index}
            // Buttery smooth cascading slide-up reveal remains identical
            className={`h-full transform transition-all duration-1000 ease-[cubic-bezier(0.16,1,0.3,1)] ${
              isVisible[index]
                ? "translate-y-0 opacity-100 scale-100"
                : "translate-y-12 opacity-0 scale-[0.98]"
            }`}
            style={{ transitionDelay: `${index * 150}ms` }}
          >
            <FeatureCardView card={card} />
          </div>
        ))}
      </div>
    </section>
  );
}
