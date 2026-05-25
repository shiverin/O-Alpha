"use client";

import { Panel } from '@/components/ui/Panel';
import { Icon } from '@/components/ui/Icon';
import { useEffect, useRef, useState } from 'react';
//import { getAccentStyle } from '@/lib/ui';

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

function FeatureCardView({ card, expanded }: { card: FeatureCard; expanded: boolean }) {
  const isPrimary = card.accent === "primary";

  return (
    <Panel
      className={`relative flex h-[350px] sm:h-[350px] flex-col overflow-hidden p-6 sm:p-7 transition-all duration-500 ease-out ${
        expanded
          ? 'bg-surface-container-highest/92 shadow-[0_18px_50px_rgba(0,0,0,0.14)] border-outline-variant/70'
          : 'bg-surface-container-high/85'
      }`}
    >
      <div className={`absolute inset-x-6 top-6 h-24 rounded-[28px] blur-2xl opacity-60 transition-opacity duration-500 ${
        isPrimary
          ? expanded
            ? 'bg-primary-container/28'
            : 'bg-primary-container/20'
          : expanded
            ? 'bg-secondary-fixed/28'
            : 'bg-secondary-fixed/20'
      }`} />
      <div className="relative z-10 flex h-full flex-col">
        <div className="flex items-start gap-4">
          <div
            className={`relative mt-0.5 flex h-12 w-12 shrink-0 items-center justify-center rounded-full border bg-surface-container/80 transition-all duration-500 ${
              expanded ? 'scale-110' : 'scale-100'
            } ${
              isPrimary ? 'border-primary-container/50' : 'border-secondary-fixed/50'
            }`}
          >
            <Icon
              name={card.icon}
              size="medium"
              color={isPrimary ? "text-primary-container" : "text-secondary-fixed"}
            />
          </div>

          <div className="min-w-0 flex-1 pt-1">
            <h3 className="font-headline-lg text-headline-lg text-on-background leading-tight">
              {card.title}
            </h3>
          </div>
        </div>

        <div className={`mt-auto overflow-hidden pt-5 transition-all duration-700 ease-out ${expanded ? 'max-h-40 opacity-100' : 'max-h-0 opacity-0'}`}>
          <p className="font-body-md text-body-md leading-7 text-on-surface-variant">
            {card.description}
          </p>
        </div>
      </div>
    </Panel>
  );
}

export function FeatureGrid() {
  const cardRefs = useRef<Array<HTMLDivElement | null>>([]);
  const [expandedCards, setExpandedCards] = useState<boolean[]>(() => features.map(() => false));

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (!entry.isIntersecting) {
            return;
          }

          const index = Number((entry.target as HTMLElement).dataset.index);

          setExpandedCards((current) => {
            if (current[index]) {
              return current;
            }

            const next = [...current];
            next[index] = true;
            return next;
          });
        });
      },
      {
        threshold: 0.45,
        rootMargin: '0px 0px -10% 0px',
      }
    );

    cardRefs.current.forEach((element) => {
      if (element) {
        observer.observe(element);
      }
    });

    return () => observer.disconnect();
  }, []);

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
        {features.map((card, index) => (
          <div
            key={card.title}
            ref={(element) => {
              cardRefs.current[index] = element;
            }}
            data-index={index}
            className={`h-full transition-all duration-700 ease-out ${
              expandedCards[index] ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-70'
            }`}
            style={{ transitionDelay: `${index * 120}ms` }}
          >
            <FeatureCardView card={card} expanded={expandedCards[index]} />
          </div>
        ))}
      </div>
    </section>
  );
}
