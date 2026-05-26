"use client";

import { useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { Container } from "@/components/ui/Container";
import { Icon } from "@/components/ui/Icon";
import { appNavItems } from "@/components/app/appNav";

type AppTopBarProps = {
  title: string;
  onSignOut: () => void;
};

export function AppTopBar({ title, onSignOut }: AppTopBarProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-40 bg-background/60 backdrop-blur-xl border-b border-outline-variant/20">
      <Container>
        <div className="relative h-20 flex items-center">
          <h1 className="max-w-[calc(100%-120px)] truncate pr-4 text-xl md:text-2xl font-light tracking-wide text-on-background">
            {title}
          </h1>

          <button
            className="md:hidden absolute right-0 top-1/2 -translate-y-1/2 h-10 w-10 inline-flex items-center justify-center rounded-full border border-outline-variant/50 bg-surface-container-high/80 text-on-background transition-colors hover:border-primary-container/60 hover:bg-surface-container-highest/80"
            aria-label="Toggle navigation menu"
            aria-expanded={menuOpen}
            onClick={() => setMenuOpen((open) => !open)}
          >
            <Icon name={menuOpen ? "close" : "menu"} size="small" color="text-on-background" />
          </button>

          <button
            className="hidden md:block absolute right-0 top-1/2 -translate-y-1/2 px-5 py-2 rounded-full border border-outline-variant/30 text-xs font-medium tracking-wide text-on-surface-variant hover:text-on-background hover:bg-surface-container transition-all duration-300 ease-out"
            onClick={onSignOut}
          >
            Sign Out
          </button>
        </div>
      </Container>
      <div
        className={
          "md:hidden overflow-hidden transition-all duration-300 ease-out " +
          (menuOpen
            ? "max-h-[420px] border-t border-outline-variant/30 bg-background/95 backdrop-blur-xl"
            : "max-h-0 border-transparent bg-transparent")
        }
      >
        <div
          className={
            "px-margin-mobile py-4 flex flex-col gap-3 transition-all duration-300 ease-out " +
            (menuOpen
              ? "opacity-100 translate-y-0"
              : "opacity-0 -translate-y-2 pointer-events-none")
          }
        >
          {appNavItems.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setMenuOpen(false)}
                className={
                  "flex items-center gap-3 rounded-xl px-3 py-2 transition-colors duration-300 " +
                  (active
                    ? "text-primary-container bg-surface-container"
                    : "text-on-surface-variant hover:text-on-background hover:bg-surface-container-high/40")
                }
              >
                <Icon name={item.icon} size="small" color={active ? "text-primary-container" : "text-on-surface-variant"} />
                <span className="text-sm font-light tracking-wide">{item.label}</span>
              </Link>
            );
          })}
          <div className="pt-2 border-t border-outline-variant/30">
            <button
              className="w-full mt-2 px-4 py-2 rounded-full border border-outline-variant/30 text-xs font-medium tracking-wide text-on-surface-variant hover:text-on-background hover:bg-surface-container transition-all duration-300 ease-out"
              onClick={() => {
                setMenuOpen(false);
                onSignOut();
              }}
            >
              Sign Out
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}
