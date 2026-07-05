"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Container } from "@/components/ui/Container";
import { Icon } from "@/components/ui/Icon";
import { AppTopBar } from "@/components/app/AppTopBar";
import {
  BeginnerTour,
  BEGINNER_TOUR_PENDING_STORAGE_KEY,
  beginnerTourSeenStorageKey,
} from "@/components/app/BeginnerTour";
import { useAuth } from "@/context/AuthContext";
import { removeToken } from "@/lib/auth";
import { appNavItems } from "@/components/app/appNav";

export function AppShell({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { user } = useAuth();
  const [tourOpen, setTourOpen] = useState(false);
  const tourSeenKey = useMemo(
    () => (user ? beginnerTourSeenStorageKey(user.id) : null),
    [user],
  );

  useEffect(() => {
    if (!user?.is_onboarded || pathname !== "/app/dashboard" || !tourSeenKey) {
      return;
    }

    const hasPendingReplay =
      localStorage.getItem(BEGINNER_TOUR_PENDING_STORAGE_KEY) === "true";
    const hasSeenTour = Boolean(localStorage.getItem(tourSeenKey));
    if (!hasPendingReplay && hasSeenTour) {
      return;
    }

    const timer = window.setTimeout(
      () => {
        localStorage.removeItem(BEGINNER_TOUR_PENDING_STORAGE_KEY);
        setTourOpen(true);
      },
      hasPendingReplay ? 250 : 900,
    );

    return () => window.clearTimeout(timer);
  }, [pathname, tourSeenKey, user?.is_onboarded]);

  const handleLogout = () => {
    removeToken();
    router.replace("/");
  };

  const markTourSeen = () => {
    if (tourSeenKey) {
      localStorage.setItem(tourSeenKey, new Date().toISOString());
    }
    localStorage.removeItem(BEGINNER_TOUR_PENDING_STORAGE_KEY);
    setTourOpen(false);
  };

  const handleStartGuide = () => {
    localStorage.setItem(BEGINNER_TOUR_PENDING_STORAGE_KEY, "true");
    if (pathname !== "/app/dashboard") {
      router.push("/app/dashboard");
      return;
    }

    localStorage.removeItem(BEGINNER_TOUR_PENDING_STORAGE_KEY);
    setTourOpen(true);
  };

  return (
    <div className="min-h-screen flex bg-background text-on-background font-body">
      <BeginnerTour
        open={tourOpen}
        onComplete={markTourSeen}
        onSkip={markTourSeen}
      />

      <aside className="hidden md:flex flex-col fixed left-0 top-0 h-full w-64 border-r border-outline-variant/20 bg-surface-container-lowest/40 backdrop-blur-md">
        <div className="px-7 pt-10 pb-12 flex items-center gap-2">
          <span className="text-lg font-light tracking-[0.15em] text-on-background">
            O(Alpha)
          </span>
        </div>

        <nav className="flex-1 px-4 space-y-1.5">
          {appNavItems.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                data-tour-id={`app-nav-${item.tourId}`}
                className={`group flex items-center gap-3.5 px-4 py-3 rounded-xl transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] ${
                  active
                    ? "bg-surface-container text-primary-container border border-outline-variant/30 shadow-sm"
                    : "text-on-surface-variant hover:bg-white/[0.02] hover:text-on-background"
                }`}
              >
                <div
                  className={`transition-transform duration-300 group-hover:scale-105 ${active ? "text-primary-container" : "text-on-surface-variant/70 group-hover:text-on-surface"}`}
                >
                  <Icon name={item.icon} size="small" />
                </div>
                <span className="text-sm font-light tracking-wide">
                  {item.label}
                </span>
              </Link>
            );
          })}
        </nav>

        <div className="px-6 pb-8 mt-auto">
          <button
            type="button"
            data-tour-id="app-guide-button"
            className="mb-3 w-full py-2.5 rounded-full border border-outline-variant/30 text-xs font-medium tracking-wider uppercase text-on-surface-variant hover:text-on-background hover:bg-surface-container-high hover:border-outline-variant/60 transition-all duration-300 ease-out inline-flex items-center justify-center gap-2"
            onClick={handleStartGuide}
          >
            <Icon name="help" size="small" color="text-current" />
            Guide
          </button>
          <button
            className="w-full py-2.5 rounded-full border border-outline-variant/30 text-xs font-medium tracking-wider uppercase text-on-surface-variant hover:text-on-background hover:bg-surface-container-high hover:border-outline-variant/60 transition-all duration-300 ease-out"
            onClick={handleLogout}
          >
            Log out
          </button>
        </div>
      </aside>

      <div className="flex-1 md:ml-64 flex flex-col min-w-0">
        <AppTopBar
          title={title}
          onSignOut={handleLogout}
          onStartGuide={handleStartGuide}
        />

        <main className="px-margin-mobile md:px-margin-desktop py-12 flex-grow min-w-0">
          <Container className="min-w-0">
            <div className="animate-in fade-in duration-700 ease-out">
              {children}
            </div>
          </Container>
        </main>
      </div>
    </div>
  );
}
