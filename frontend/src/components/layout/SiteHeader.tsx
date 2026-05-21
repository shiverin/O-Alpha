"use client";

import { useState } from "react";

import { Bars3Icon, XMarkIcon } from "@heroicons/react/24/solid";

import { BrandMark } from "../BrandMark";
import { PillButton } from "../PillButton";

type NavLink = {
  label: string;
  href: string;
  active?: boolean;
};

type SiteHeaderProps = {
  activePath?: string;
};

const navLinks: NavLink[] = [
  { label: "Product", href: "/" },
  { label: "Performance", href: "/performance" },
  { label: "Pricing", href: "/pricing" },
  { label: "Mission", href: "/mission" },
];

export function SiteHeader({ activePath }: SiteHeaderProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const links = navLinks.map((link) => ({
    ...link,
    active: activePath ? link.href === activePath : link.active,
  }));

  return (
    <nav className="fixed top-0 w-full z-50 bg-background/80 backdrop-blur-xl border-b border-outline-variant/30">
      <div className="flex justify-between items-center px-margin-mobile md:px-margin-desktop py-4 max-w-[1440px] mx-auto">
        <div className="flex items-center gap-2 md:gap-12">
          <a className="flex items-center gap-2 md:gap-3" href="/">
            <BrandMark className="-ml-0.5" />
          </a>
          <div className="hidden md:flex gap-8 items-center pt-1">
            {links.map((link) => (
              <a
                key={link.label}
                className={
                  link.active
                    ? "font-body-md text-body-md text-primary-container border-b-2 border-primary-container pb-1 transition-colors duration-300"
                    : "font-body-md text-body-md text-on-surface hover:text-primary-container transition-colors duration-300"
                }
                href={link.href}
              >
                {link.label}
              </a>
            ))}
          </div>
        </div>
        <button
          className="md:hidden inline-flex h-11 w-11 items-center justify-center rounded-full border border-outline-variant/50 bg-surface-container-high/80 text-white transition-colors hover:border-primary-container/60 hover:bg-surface-container-highest/80"
          aria-label="Toggle navigation menu"
          aria-expanded={menuOpen}
          onClick={() => setMenuOpen((open) => !open)}
        >
          {menuOpen ? (
            <XMarkIcon className="h-5 w-5" style={{ color: "#ffffff", fill: "#ffffff" }} />
          ) : (
            <Bars3Icon className="h-5 w-5" style={{ color: "#ffffff", fill: "#ffffff" }} />
          )}
        </button>
        <div className="hidden md:flex items-center gap-3">
          <PillButton variant="outline" size="sm">
            Login
          </PillButton>
          <PillButton className="scale-95 active:scale-90" size="sm">
            Launch App
          </PillButton>
        </div>
      </div>
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
          {links.map((link) => (
            <a
              key={link.label}
              className={
                link.active
                  ? "font-body-md text-body-md text-primary-container"
                  : "font-body-md text-body-md text-on-surface-variant"
              }
              href={link.href}
              onClick={() => setMenuOpen(false)}
            >
              {link.label}
            </a>
          ))}
          <div className="pt-2 border-t border-outline-variant/30 flex flex-col gap-3">
            <PillButton variant="outline" size="sm">
              Login
            </PillButton>
            <PillButton size="sm">Launch App</PillButton>
          </div>
        </div>
      </div>
    </nav>
  );
}
