import { useState, useEffect } from "react";

type RegimeLevel = {
  bull: number;
  volatile: number;
  bear: number;
};

type RegimeStatus = {
  bullStatus: "SCALING" | "BUILDING" | "SOFT";
  volatileStatus: "ACTIVE" | "WATCH" | "CALM";
  bearStatus: "ELEVATED" | "HEDGED" | "LOW";
};

export function useRegimeSimulation() {
  const [regimeLevels, setRegimeLevels] = useState<RegimeLevel>({
    bull: 74,
    volatile: 50,
    bear: 24,
  });

  useEffect(() => {
    const clamp = (value: number, min: number, max: number) =>
      Math.max(min, Math.min(max, value));

    const interval = window.setInterval(() => {
      setRegimeLevels((current) => {
        const shock = () =>
          Math.random() < 0.28 ? Math.random() * 40 - 20 : 0;

        return {
          bull: Math.round(
            clamp(current.bull + (Math.random() * 26 - 13) + shock(), 18, 98),
          ),
          volatile: Math.round(
            clamp(
              current.volatile + (Math.random() * 30 - 15) + shock(),
              6,
              95,
            ),
          ),
          bear: Math.round(
            clamp(current.bear + (Math.random() * 26 - 13) + shock(), 4, 88),
          ),
        };
      });
    }, 700);

    return () => window.clearInterval(interval);
  }, []);

  const bullStatus: "SCALING" | "BUILDING" | "SOFT" =
    regimeLevels.bull >= 72
      ? "SCALING"
      : regimeLevels.bull >= 58
        ? "BUILDING"
        : "SOFT";

  const volatileStatus: "ACTIVE" | "WATCH" | "CALM" =
    regimeLevels.volatile >= 58
      ? "ACTIVE"
      : regimeLevels.volatile >= 36
        ? "WATCH"
        : "CALM";

  const bearStatus: "ELEVATED" | "HEDGED" | "LOW" =
    regimeLevels.bear >= 32
      ? "ELEVATED"
      : regimeLevels.bear >= 18
        ? "HEDGED"
        : "LOW";

  const statuses: RegimeStatus = {
    bullStatus,
    volatileStatus,
    bearStatus,
  };

  return { regimeLevels, ...statuses };
}
