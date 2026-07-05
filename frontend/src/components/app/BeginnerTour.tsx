"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import type { CSSProperties } from "react";
import { Icon } from "@/components/ui/Icon";

type TourStep = {
  targetId: string;
  title: string;
  body: string;
};

type TargetBox = {
  top: number;
  left: number;
  width: number;
  height: number;
};

export const BEGINNER_TOUR_PENDING_STORAGE_KEY = "oa_beginner_tour_pending_v1";

const BEGINNER_TOUR_SEEN_STORAGE_PREFIX = "oa_beginner_tour_seen_v1";
const POPOVER_MARGIN = 16;
const POPOVER_WIDTH = 420;
const ESTIMATED_POPOVER_HEIGHT = 300;

export function beginnerTourSeenStorageKey(userID: number) {
  return `${BEGINNER_TOUR_SEEN_STORAGE_PREFIX}:${userID}`;
}

export const beginnerTourSteps: TourStep[] = [
  {
    targetId: "app-nav-overview",
    title: "Start With Overview",
    body: "Overview is your home base. It shows the paper agent status, current 24-hour P&L, selected strategy profile, recent execution events, and portfolio allocation.",
  },
  {
    targetId: "dashboard-agent-action",
    title: "Launch Or Stop The Agent",
    body: "Launch Agent starts the catalog portfolio agent with your saved risk profile, strategy, timeframe, and starting cash. When the agent is active, this button changes to Terminate Agent.",
  },
  {
    targetId: "dashboard-balance-card",
    title: "Read Agent Status First",
    body: "This card shows whether the agent is running. The P&L and regime label update from portfolio data when it is available; inactive accounts can show a flat or demo state.",
  },
  {
    targetId: "dashboard-strategy-profile",
    title: "Confirm The Active Profile",
    body: "Strategy Profile shows the risk profile chosen during onboarding or settings, plus the current catalog universe size. It is a quick check before launching the agent.",
  },
  {
    targetId: "dashboard-execution-log",
    title: "Watch Recent Actions",
    body: "Live Execution Log lists recent paper execution events. If no server trades are available yet, the app shows an empty or demo log instead of inventing activity.",
  },
  {
    targetId: "dashboard-allocation",
    title: "Check Allocation",
    body: "Portfolio Allocation summarizes how exposure is split across visible positions and cash. It uses live portfolio data when available and falls back to a demo allocation for guest data.",
  },
  {
    targetId: "app-nav-agent-settings",
    title: "Tune In Agent Settings",
    body: "Agent Settings lets you change risk profiles and, in Advanced Tuning, leverage, max positions, stop-loss, take-profit, and rebalance cadence. Settings are locked while a portfolio agent is running.",
  },
  {
    targetId: "app-nav-portfolio",
    title: "Inspect Portfolio Details",
    body: "Portfolio shows total asset value, composition, positions, and a CSV export for position data. Use it when you want more detail than the Overview allocation card.",
  },
  {
    targetId: "app-nav-activity",
    title: "Audit Activity",
    body: "Activity Console lists trade events and system alerts. You can filter by all actions, fills, errors, or asset, then export the filtered trade table as CSV.",
  },
  {
    targetId: "app-guide-button",
    title: "Replay The Guide Anytime",
    body: "Use Guide to restart this walkthrough. Skip closes it for now, Back revisits the previous step, and Next moves through the tour one point at a time.",
  },
];

export function BeginnerTour({
  open,
  onComplete,
  onSkip,
}: {
  open: boolean;
  onComplete: () => void;
  onSkip: () => void;
}) {
  const [stepIndex, setStepIndex] = useState(0);
  const [targetBox, setTargetBox] = useState<TargetBox | null>(null);
  const step = beginnerTourSteps[stepIndex];
  const isLastStep = stepIndex === beginnerTourSteps.length - 1;

  const updateTargetBox = useCallback(() => {
    if (!open || !step) return;

    const target = findVisibleTourTarget(step.targetId);
    if (!target) {
      setTargetBox(null);
      return;
    }

    target.scrollIntoView({
      block: "center",
      inline: "center",
      behavior: "smooth",
    });

    window.setTimeout(() => {
      const rect = target.getBoundingClientRect();
      setTargetBox({
        top: rect.top,
        left: rect.left,
        width: rect.width,
        height: rect.height,
      });
    }, 180);
  }, [open, step]);

  useEffect(() => {
    if (open) {
      setStepIndex(0);
    }
  }, [open]);

  useEffect(() => {
    if (!open) return;

    const animationFrame = window.requestAnimationFrame(updateTargetBox);
    window.addEventListener("resize", updateTargetBox);
    window.addEventListener("scroll", updateTargetBox, true);

    return () => {
      window.cancelAnimationFrame(animationFrame);
      window.removeEventListener("resize", updateTargetBox);
      window.removeEventListener("scroll", updateTargetBox, true);
    };
  }, [open, stepIndex, updateTargetBox]);

  useEffect(() => {
    if (!open) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onSkip();
      }
      if (event.key === "ArrowRight") {
        setStepIndex((current) =>
          Math.min(current + 1, beginnerTourSteps.length - 1),
        );
      }
      if (event.key === "ArrowLeft") {
        setStepIndex((current) => Math.max(current - 1, 0));
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onSkip, open]);

  const popoverStyle = useMemo(() => buildPopoverStyle(targetBox), [targetBox]);

  if (!open || !step) return null;

  const stepCountLabel = `${stepIndex + 1} of ${beginnerTourSteps.length}`;

  return (
    <div
      className="fixed inset-0 z-[9998] bg-background/72 backdrop-blur-[3px]"
      role="presentation"
    >
      {targetBox && (
        <div
          className="pointer-events-none absolute rounded-[24px] border border-primary-fixed-dim/80 bg-primary-fixed-dim/10 shadow-[0_0_0_9999px_rgba(11,17,23,0.72),0_0_32px_rgba(0,213,255,0.22)] transition-all duration-300"
          style={{
            top: Math.max(8, targetBox.top - 8),
            left: Math.max(8, targetBox.left - 8),
            width: targetBox.width + 16,
            height: targetBox.height + 16,
          }}
        />
      )}

      <section
        aria-labelledby="beginner-tour-title"
        aria-modal="true"
        className="absolute w-[calc(100vw-32px)] max-w-[420px] rounded-[24px] border border-outline-variant/50 bg-surface-container-high p-5 shadow-[0_24px_70px_rgba(0,0,0,0.48)] transition-all duration-300"
        role="dialog"
        style={popoverStyle}
      >
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="font-mono text-[10px] font-medium uppercase tracking-[0.2em] text-primary-fixed-dim">
              Beginner Guide
            </p>
            <h2
              id="beginner-tour-title"
              className="mt-2 text-xl font-light text-on-surface"
            >
              {step.title}
            </h2>
          </div>
          <span className="rounded-full border border-outline-variant/35 px-2.5 py-1 font-mono text-[10px] text-on-surface-variant">
            {stepCountLabel}
          </span>
        </div>

        <p className="mt-4 text-sm font-light leading-relaxed text-on-surface-variant/85">
          {step.body}
        </p>

        <div className="mt-5 h-1.5 overflow-hidden rounded-full bg-void-black/35">
          <div
            className="h-full rounded-full bg-primary-fixed-dim transition-all duration-300"
            style={{
              width: `${((stepIndex + 1) / beginnerTourSteps.length) * 100}%`,
            }}
          />
        </div>

        <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <button
            type="button"
            onClick={onSkip}
            className="inline-flex items-center justify-center gap-2 rounded-full border border-outline-variant/30 px-4 py-2 text-xs font-medium uppercase tracking-wide text-on-surface-variant transition-colors hover:bg-surface-container hover:text-on-surface"
          >
            Skip
          </button>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() =>
                setStepIndex((current) => Math.max(current - 1, 0))
              }
              disabled={stepIndex === 0}
              className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-outline-variant/30 text-on-surface-variant transition-colors hover:bg-surface-container hover:text-on-surface disabled:cursor-not-allowed disabled:opacity-40"
              aria-label="Previous guide step"
            >
              <Icon name="arrow_back" size="small" color="text-current" />
            </button>
            <button
              type="button"
              onClick={() => {
                if (isLastStep) {
                  onComplete();
                  return;
                }
                setStepIndex((current) => current + 1);
              }}
              className="inline-flex min-w-[112px] items-center justify-center gap-2 rounded-full bg-primary-container px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-void-black shadow-primary-container/20 transition-colors hover:bg-primary-container/90"
            >
              {isLastStep ? "Finish" : "Next"}
              {!isLastStep && (
                <Icon name="arrow_forward" size="small" color="text-current" />
              )}
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}

function findVisibleTourTarget(targetId: string) {
  const targets = Array.from(
    document.querySelectorAll<HTMLElement>(`[data-tour-id="${targetId}"]`),
  );

  return (
    targets.find((target) => {
      const rect = target.getBoundingClientRect();
      const style = window.getComputedStyle(target);
      return (
        rect.width > 0 &&
        rect.height > 0 &&
        style.display !== "none" &&
        style.visibility !== "hidden"
      );
    }) || null
  );
}

function buildPopoverStyle(targetBox: TargetBox | null): CSSProperties {
  if (typeof window === "undefined" || !targetBox) {
    return {
      left: "50%",
      top: "50%",
      transform: "translate(-50%, -50%)",
    };
  }

  const width = Math.min(POPOVER_WIDTH, window.innerWidth - POPOVER_MARGIN * 2);
  const canPlaceRight =
    targetBox.left + targetBox.width + POPOVER_MARGIN + width <=
    window.innerWidth - POPOVER_MARGIN;
  const canPlaceLeft =
    targetBox.left - POPOVER_MARGIN - width >= POPOVER_MARGIN;

  const left = canPlaceRight
    ? targetBox.left + targetBox.width + POPOVER_MARGIN
    : canPlaceLeft
      ? targetBox.left - POPOVER_MARGIN - width
      : clamp(
          targetBox.left + targetBox.width / 2 - width / 2,
          POPOVER_MARGIN,
          window.innerWidth - width - POPOVER_MARGIN,
        );

  const top = clamp(
    targetBox.top + targetBox.height / 2 - ESTIMATED_POPOVER_HEIGHT / 2,
    POPOVER_MARGIN,
    window.innerHeight - ESTIMATED_POPOVER_HEIGHT - POPOVER_MARGIN,
  );

  return {
    left,
    top,
    width,
  };
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), Math.max(min, max));
}
