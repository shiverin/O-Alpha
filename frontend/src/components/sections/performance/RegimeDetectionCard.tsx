"use client";

type RegimeLevel = {
  bull: number;
  volatile: number;
  bear: number;
};

type RegimeStatuses = {
  bullStatus: "SCALING" | "BUILDING" | "SOFT";
  volatileStatus: "ACTIVE" | "WATCH" | "CALM";
  bearStatus: "ELEVATED" | "HEDGED" | "LOW";
};

interface RegimeDetectionCardProps {
  levels: RegimeLevel;
  statuses: RegimeStatuses;
}

export default function RegimeDetectionCard({
  levels,
  statuses,
}: RegimeDetectionCardProps) {
  return (
    <div className="group relative flex flex-col h-full min-h-[340px] bg-surface-container-low border border-outline-variant/30 rounded-[32px] p-8 overflow-hidden hover:bg-surface-container transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:-translate-y-1 hover:shadow-[0_20px_40px_rgba(0,0,0,0.1)]">
      <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-container/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />

      <div className="flex items-center gap-2 mb-6">
        <span className="material-symbols-outlined text-sm text-on-surface-variant">
          radar
        </span>
        <span className="text-[10px] uppercase tracking-[0.2em] text-on-surface-variant font-medium">
          Regime Detection
        </span>
      </div>

      <p className="text-sm font-light leading-relaxed text-on-surface-variant mb-8 group-hover:text-on-surface transition-colors duration-300">
        Hidden Markov Models continuously classify market states, dynamically
        shifting agent posture.
      </p>

      <div className="flex-grow flex flex-col justify-around gap-4 pb-2">
        <div className="flex items-center justify-between">
          <span className="font-data-sm text-data-sm text-on-surface w-24">
            BULL
          </span>
          <div className="flex-grow mx-4 bg-outline-variant/20 h-[2px] relative rounded-full">
            <div
              className="absolute left-0 top-1/2 -translate-y-1/2 h-[2px] bg-primary-container transition-all duration-500 ease-out rounded-full"
              style={{ width: `${levels.bull}%` }}
            ></div>
          </div>
          <span className="font-data-sm text-data-sm text-primary-container w-20 text-right">
            {statuses.bullStatus}
          </span>
        </div>

        <div className="flex items-center justify-between">
          <span className="font-data-sm text-data-sm text-on-surface w-24">
            VOLATILE
          </span>
          <div className="flex-grow mx-4 bg-outline-variant/20 h-[2px] relative rounded-full">
            <div
              className="absolute left-0 top-1/2 -translate-y-1/2 h-[2px] bg-secondary-fixed transition-all duration-500 ease-out rounded-full"
              style={{ width: `${levels.volatile}%` }}
            ></div>
          </div>
          <span className="font-data-sm text-data-sm text-secondary-fixed w-20 text-right">
            {statuses.volatileStatus}
          </span>
        </div>

        <div className="flex items-center justify-between">
          <span className="font-data-sm text-data-sm text-on-surface w-24">
            BEAR
          </span>
          <div className="flex-grow mx-4 bg-outline-variant/20 h-[2px] relative rounded-full">
            <div
              className="absolute left-0 top-1/2 -translate-y-1/2 h-[2px] bg-on-surface-variant transition-all duration-500 ease-out rounded-full"
              style={{ width: `${levels.bear}%` }}
            ></div>
          </div>
          <span className="font-data-sm text-data-sm text-on-surface-variant w-20 text-right">
            {statuses.bearStatus}
          </span>
        </div>
      </div>
    </div>
  );
}
