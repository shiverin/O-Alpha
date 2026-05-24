import { AppShell } from "@/components/app/AppShell";

export default function ActivityPage() {
  return (
    <AppShell title="Activity">
      <div className="bg-surface-container-high/80 border border-outline-variant/40 rounded-2xl p-6">
        <h2 className="font-headline-lg text-headline-lg text-on-background mb-4">
          Execution Stream
        </h2>
        <p className="font-body-md text-body-md text-on-surface-variant">
          Real-time audit log of systematic trading actions, strategy
          recalibrations, and critical system alerts.
        </p>
      </div>
    </AppShell>
  );
}
