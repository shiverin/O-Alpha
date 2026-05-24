"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Container } from '@/components/ui/Container';
import { Panel } from '@/components/ui/Panel';
import { Icon } from '@/components/ui/Icon';

const navItems = [
  { label: "Overview", href: "/app/dashboard", icon: "dashboard" },
  { label: "Agent Settings", href: "/app/agent-settings", icon: "settings_input_component" },
  { label: "Portfolio", href: "/app/portfolio", icon: "pie_chart" },
  { label: "Activity", href: "/app/activity", icon: "history" },
];

export function AppShell({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = () => {
    localStorage.removeItem("oa-auth");
    router.replace("/");
  };

  return (
    <div className="min-h-screen flex bg-background text-on-background">
      <aside className="hidden md:flex flex-col fixed left-0 top-0 h-full w-64 border-r border-outline-variant/30 bg-surface-container-lowest">
        <Panel className="px-6 py-8">
          <div className="font-headline-lg text-headline-lg text-on-background">
            O(Alpha)
          </div>
          <div className="font-data-sm text-data-sm text-on-surface-variant mt-2">
            Neural Core Active
          </div>
        </Panel>
        <nav className="flex-1 px-3 space-y-1">
          {navItems.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={
                  active
                    ? "flex items-center gap-3 px-4 py-3 rounded-xl bg-primary-container/10 text-primary-container"
                    : "flex items-center gap-3 px-4 py-3 rounded-xl text-on-surface-variant hover:bg-surface-container-high/70 hover:text-on-background"
                }
              >
                <Icon name={item.icon} size="small" />
                <span className="font-body-md text-body-md">{item.label}</span>
              </Link>
            );
          })}
        </nav>
        <div className="px-6 pb-6 mt-auto">
          <button
            className="w-full py-2 rounded-full border border-outline-variant/40 font-body-md text-body-md text-on-background hover:bg-surface-container-high transition-colors"
            onClick={handleLogout}
          >
            Log out
          </button>
        </div>
      </aside>
      <div className="flex-1 md:ml-64">
        <header className="sticky top-0 z-40 bg-background/80 backdrop-blur-xl border-b border-outline-variant/30">
          <Container>
            <div className="flex items-center justify-between h-16">
              <div className="font-headline-lg text-headline-lg text-on-background">
                {title}
              </div>
              <button
                className="px-4 py-2 rounded-full border border-outline-variant/40 text-on-surface-variant hover:text-on-background hover:bg-surface-container-high transition-colors"
                onClick={handleLogout}
              >
                Log out
              </button>
            </div>
          </Container>
        </header>
        <main className="px-margin-mobile md:px-margin-desktop py-10">
          <Container>
            {children}
          </Container>
        </main>
      </div>
    </div>
  );
}
