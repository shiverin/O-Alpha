import type { ReactNode, Dispatch, SetStateAction } from "react";

import { SiteFooter } from "./SiteFooter";
import { SiteHeader } from "./SiteHeader";

type LandingShellProps = {
  children: ReactNode;
  activePath?: string;
  className?: string;
  loginModalOpen?: boolean;
  onLoginModalOpenChange?: Dispatch<SetStateAction<boolean>>;
};

export function LandingShell({ children, activePath, className, loginModalOpen, onLoginModalOpenChange }: LandingShellProps) {
  return (
    <div
      className={`text-on-background min-h-screen flex flex-col relative overflow-x-hidden bg-background ${className ?? ""}`}
    >
      <SiteHeader activePath={activePath} loginModalOpen={loginModalOpen} onLoginModalOpenChange={onLoginModalOpenChange} />
      {children}
      <SiteFooter />
    </div>
  );
}
