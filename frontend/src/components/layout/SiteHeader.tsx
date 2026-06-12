"use client";

import { useState, Dispatch, SetStateAction } from "react";
import { useRouter } from "next/navigation";

import { Bars3Icon, XMarkIcon } from "@heroicons/react/24/solid";

import Link from "next/link";
import { BrandMark } from "../BrandMark";
import { LoginModal } from "../auth/LoginModal";
import { PillButton } from "../PillButton";
import { isAuthenticated } from "@/lib/auth";

type NavLink = {
  label: string;
  href: string;
  active?: boolean;
};

type SiteHeaderProps = {
  activePath?: string;
  loginModalOpen?: boolean;
  onLoginModalOpenChange?: Dispatch<SetStateAction<boolean>>;
};

const navLinks: NavLink[] = [
  { label: "Product", href: "/" },
  { label: "Performance", href: "/performance" },
  { label: "Pricing", href: "/pricing" },
  { label: "Mission", href: "/mission" },
];

export function SiteHeader({
  activePath,
  loginModalOpen: externalLoginOpen,
  onLoginModalOpenChange,
}: SiteHeaderProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [internalLoginOpen, setInternalLoginOpen] = useState(false);

  const loginOpen =
    externalLoginOpen !== undefined ? externalLoginOpen : internalLoginOpen;
  const setLoginOpen = onLoginModalOpenChange || setInternalLoginOpen;

  const router = useRouter();

  const handleLaunchApp = () => {
    if (isAuthenticated()) {
      router.push("/app/dashboard");
      return;
    }
    setLoginOpen(true);
  };

  const links = navLinks.map((link) => ({
    ...link,
    active: activePath ? link.href === activePath : link.active,
  }));

  return (
    <nav className="fixed top-0 w-full z-50 bg-background/80 backdrop-blur-xl border-b border-outline-variant/30">
      <div className="flex justify-between items-center px-16 md:px-margin-desktop py-4 max-w-[1440px] mx-auto">
        <div className="flex items-center gap-2 md:gap-12">
          <Link href="/" className="flex items-center gap-2 md:gap-3">
            <BrandMark className="-ml-0.5" logoSize="md" showText={false} />
          </Link>
          <div className="hidden md:flex gap-8 items-center pt-1">
            {links.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                className={`
                  ${
                    link.active
                      ? "font-body-md text-body-md text-primary-container border-b-2 border-primary-container pb-1"
                      : "font-body-md text-body-md text-on-surface hover:text-primary-container"
                  }
                  transition-colors duration-300
                `}
              >
                {link.label}
              </Link>
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
            <XMarkIcon
              className="h-5 w-5"
              style={{ color: "#ffffff", fill: "#ffffff" }}
            />
          ) : (
            <Bars3Icon
              className="h-5 w-5"
              style={{ color: "#ffffff", fill: "#ffffff" }}
            />
          )}
        </button>
        <div className="hidden md:flex items-center gap-3">
          <PillButton
            variant="outline"
            size="sm"
            onClick={() => setLoginOpen(true)}
          >
            Login
          </PillButton>
          <PillButton
            className="scale-95 active:scale-90"
            size="sm"
            onClick={handleLaunchApp}
          >
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
            "px-6 py-4 flex flex-col gap-3 transition-all duration-300 ease-out " +
            (menuOpen
              ? "opacity-100 translate-y-0"
              : "opacity-0 -translate-y-2 pointer-events-none")
          }
        >
          {links.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              onClick={() => setMenuOpen(false)}
              className={`
                ${
                  link.active
                    ? "font-body-md text-body-md text-primary-container"
                    : "font-body-md text-body-md text-on-surface-variant"
                }
              `}
            >
              {link.label}
            </Link>
          ))}
          <div className="pt-2 border-t border-outline-variant/30 flex flex-col gap-3">
            <PillButton
              variant="outline"
              size="sm"
              onClick={() => {
                setMenuOpen(false);
                setLoginOpen(true);
              }}
            >
              Login
            </PillButton>
            <PillButton
              size="sm"
              onClick={() => {
                setMenuOpen(false);
                handleLaunchApp();
              }}
            >
              Launch App
            </PillButton>
          </div>
        </div>
      </div>
      <LoginModal
        isOpen={loginOpen}
        onClose={() => setLoginOpen(false)}
        redirectPath="/app/dashboard"
      />
    </nav>
  );
}
