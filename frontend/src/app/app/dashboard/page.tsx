import { AppShell } from "@/components/app/AppShell";

export default function DashboardPage() {
  return (
    <AppShell title="System Overview">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-gutter">
        <div className="bg-surface-container-high/80 border border-outline-variant/40 rounded-2xl p-6">
          <div className="font-data-sm text-data-sm text-on-surface-variant mb-2">
            Agent Status
          </div>
          <div className="font-headline-lg text-headline-lg text-on-background">
            Optimizing
          </div>
          <div className="mt-6 h-2 bg-surface-container-highest rounded-full overflow-hidden">
            <div className="h-full w-3/4 bg-primary-container rounded-full"></div>
          </div>
        </div>
        <div className="bg-surface-container-high/80 border border-outline-variant/40 rounded-2xl p-6">
          <div className="font-data-sm text-data-sm text-on-surface-variant mb-2">
            P&L (24h)
          </div>
          <div className="font-headline-lg text-headline-lg text-primary-container">
            +$12,450.89
          </div>
        </div>
      </div>
    </AppShell>
  );
}
