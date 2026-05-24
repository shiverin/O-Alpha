"use client";

import { AppShell } from "@/components/app/AppShell";
import { PillButton } from "@/components/PillButton";
import { Container } from "@/components/ui/Container";
import { useState } from "react";

export default function AgentSettingsPage() {
  const [riskProfile, setRiskProfile] = useState("moderate");
  const [leverage, setLeverage] = useState(1);
  const [maxPositions, setMaxPositions] = useState(5);
  const [stopLoss, setStopLoss] = useState(2.0);
  const [takeProfit, setTakeProfit] = useState(4.0);
  const [rebalanceFreq, setRebalanceFreq] = useState("daily");
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = async () => {
    setIsSaving(true);
    // In a real implementation, this would save to backend
    // For now, we'll simulate an API call
    await new Promise(resolve => setTimeout(resolve, 1000));
    setIsSaving(false);
    alert("Settings saved successfully!");
  };

  return (
    <AppShell title="Agent Settings">
      <Container className="py-8">
        <section className="mb-8">
          <h2 className="font-headline-lg text-headline-lg text-on-background mb-4">
            Agent Configuration
          </h2>
          <p className="font-body-md text-body-md text-on-surface-variant">
            Adjust hyperparameters and execution logic for the O(Alpha) trading
            agent. Changes propagate to active strategies instantly.
          </p>
        </section>

        <div className="space-y-6">
          {/* Risk Profile */}
          <div className="space-y-3">
            <h3 className="font-headline-md text-headline-md text-on-background">
              Risk Profile
            </h3>
            <div className="flex flex-wrap gap-2">
              <PillButton
                variant={riskProfile === "conservative" ? "solid" : "outline"}
                size="sm"
                onClick={() => setRiskProfile("conservative")}
              >
                Conservative
              </PillButton>
              <PillButton
                variant={riskProfile === "moderate" ? "solid" : "outline"}
                size="sm"
                onClick={() => setRiskProfile("moderate")}
              >
                Moderate
              </PillButton>
              <PillButton
                variant={riskProfile === "aggressive" ? "solid" : "outline"}
                size="sm"
                onClick={() => setRiskProfile("aggressive")}
              >
                Aggressive
              </PillButton>
            </div>
          </div>

          {/* Leverage */}
          <div className="space-y-3">
            <h3 className="font-headline-md text-headline-md text-on-background">
              Leverage
            </h3>
            <p className="font-data-sm text-data-sm text-on-surface-variant">
              Maximum leverage to apply to positions
            </p>
            <div className="flex items-center gap-4">
              <input
                type="range"
                min="1"
                max="5"
                value={leverage}
                onChange={(e) => setLeverage(parseInt(e.target.value))}
                className="w-full"
              />
              <span className="font-data-sm text-data-sm">{leverage}x</span>
            </div>
          </div>

          {/* Max Positions */}
          <div className="space-y-3">
            <h3 className="font-headline-md text-headline-md text-on-background">
              Max Concurrent Positions
            </h3>
            <p className="font-data-sm text-data-sm text-on-surface-variant">
              Maximum number of simultaneous positions
            </p>
            <div className="flex items-center gap-4">
              <input
                type="range"
                min="1"
                max="20"
                value={maxPositions}
                onChange={(e) => setMaxPositions(parseInt(e.target.value))}
                className="w-full"
              />
              <span className="font-data-sm text-data-sm">{maxPositions}</span>
            </div>
          </div>

          {/* Stop Loss */}
          <div className="space-y-3">
            <h3 className="font-headline-md text-headline-md text-on-background">
              Stop Loss (%)
            </h3>
            <p className="font-data-sm text-data-sm text-on-surface-variant">
              Automatic stop loss trigger
            </p>
            <div className="flex items-center gap-4">
              <input
                type="range"
                min="0.5"
                max="10"
                step="0.5"
                value={stopLoss}
                onChange={(e) => setStopLoss(parseFloat(e.target.value))}
                className="w-full"
              />
              <span className="font-data-sm text-data-sm">{stopLoss}%</span>
            </div>
          </div>

          {/* Take Profit */}
          <div className="space-y-3">
            <h3 className="font-headline-md text-headline-md text-on-background">
              Take Profit (%)
            </h3>
            <p className="font-data-sm text-data-sm text-on-surface-variant">
              Automatic take profit trigger
            </p>
            <div className="flex items-center gap-4">
              <input
                type="range"
                min="1"
                max="20"
                step="0.5"
                value={takeProfit}
                onChange={(e) => setTakeProfit(parseFloat(e.target.value))}
                className="w-full"
              />
              <span className="font-data-sm text-data-sm">{takeProfit}%</span>
            </div>
          </div>

          {/* Rebalance Frequency */}
          <div className="space-y-3">
            <h3 className="font-headline-md text-headline-md text-on-background">
              Rebalance Frequency
            </h3>
            <p className="font-data-sm text-data-sm text-on-surface-variant">
              How often to rebalance the portfolio
            </p>
            <div className="flex flex-wrap gap-2">
              <PillButton
                variant={rebalanceFreq === "hourly" ? "solid" : "outline"}
                size="sm"
                onClick={() => setRebalanceFreq("hourly")}
              >
                Hourly
              </PillButton>
              <PillButton
                variant={rebalanceFreq === "daily" ? "solid" : "outline"}
                size="sm"
                onClick={() => setRebalanceFreq("daily")}
              >
                Daily
              </PillButton>
              <PillButton
                variant={rebalanceFreq === "weekly" ? "solid" : "outline"}
                size="sm"
                onClick={() => setRebalanceFreq("weekly")}
              >
                Weekly
              </PillButton>
            </div>
          </div>
        </div>

        <div className="mt-8 pt-4 border-t border-outline-variant/30">
          <PillButton
            variant="solid"
            size="md"
            onClick={handleSave}
            disabled={isSaving}
          >
            {isSaving ? "Saving..." : "Save Settings"}
          </PillButton>
        </div>
      </Container>
    </AppShell>
  );
}