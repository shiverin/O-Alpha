import { AppShell } from "@/components/app/AppShell";

export default function AgentSettingsPage() {
  return (
    <AppShell title="Agent Settings">
      <div className="bg-surface-container-high/80 border border-outline-variant/40 rounded-2xl p-6">
        <h2 className="font-headline-lg text-headline-lg text-on-background mb-4">
          Configuration Matrix
        </h2>
        <p className="font-body-md text-body-md text-on-surface-variant">
          Adjust hyperparameters and execution logic for the O(Alpha) trading
          agent. Changes propagate to active strategies instantly.
        </p>
      </div>
    </AppShell>
  );
}
