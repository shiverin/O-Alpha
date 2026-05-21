import type { ReactNode } from "react";

import { SiteFooter } from "./SiteFooter";
import { SiteHeader } from "./SiteHeader";

type LandingShellProps = {
  children: ReactNode;
  activePath?: string;
  className?: string;
};

export function LandingShell({ children, activePath, className }: LandingShellProps) {
  return (
    <div
      className={`text-on-background min-h-screen flex flex-col relative overflow-x-hidden bg-background ${className ?? ""}`}
    >
      <SiteHeader activePath={activePath} />
      {children}
      <SiteFooter />
    </div>
  );
}
