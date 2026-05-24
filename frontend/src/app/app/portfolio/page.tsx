import { AppShell } from "@/components/app/AppShell";

export default function PortfolioPage() {
  return (
    <AppShell title="Portfolio">
      <div className="bg-surface-container-high/80 border border-outline-variant/40 rounded-2xl p-6">
        <h2 className="font-headline-lg text-headline-lg text-on-background mb-4">
          Total Asset Value
        </h2>
        <div className="font-headline-xl text-headline-xl text-on-background">
          $2,481,903.50
        </div>
      </div>
    </AppShell>
  );
}
