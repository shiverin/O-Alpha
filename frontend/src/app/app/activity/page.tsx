"use client";

import { useState } from "react";
import { AppShell } from "@/components/app/AppShell";
import { Icon } from "@/components/ui/Icon";
import { mockExecutionStream, mockSystemAlerts } from "@/lib/mockAppData";

export default function ActivityPage() {
  const [activeFilter, setActiveFilter] = useState("ALL");

  return (
    <AppShell title="Activity Console">
      <div className="w-full bg-transparent flex flex-col gap-6 md:gap-10 animate-in fade-in duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]">
        {/* =========================================
            SECTION 1: OVERHEAD AUDIT STATUS META
        ========================================= */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4 pb-2">
          <div>
            <h1 className="text-2xl sm:text-3xl font-light tracking-tight text-on-surface">
              Execution Stream
            </h1>
            <p className="text-xs sm:text-sm font-light text-on-surface-variant/70 max-w-2xl mt-1">
              Real-time audit log of systematic trading actions, strategy
              recalibrations, and critical system alerts.
            </p>
          </div>

          <div className="flex items-center gap-3 w-full sm:w-auto">
            <button className="flex-1 sm:flex-none justify-center px-5 py-2 rounded-full border border-outline-variant/30 text-xs font-mono font-medium tracking-wide text-on-surface hover:bg-surface-container transition-all duration-300">
              Export CSV
            </button>
          </div>
        </div>

        {/* =========================================
            SECTION 2: RESPONSIVE ASYMMETRIC BENTO DOCK
        ========================================= */}
        <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 md:gap-8 items-start">
          {/* LEFT: MASTER EXECUTION STREAM TIMELINE (SPAN 12 OVER TRANSIT COHORTS -> SPAN 8 ON XL) */}
          <div className="md:col-span-12 xl:col-span-8 flex flex-col gap-4 sm:gap-6">
            {/* Context Filter Panel */}
            <div className="flex flex-wrap items-center justify-between gap-4 p-4 rounded-2xl bg-surface-container-low border border-outline-variant/20 backdrop-blur-md">
              <div className="flex gap-2">
                {["ALL", "FILLS", "ERRORS"].map((filter) => (
                  <button
                    key={filter}
                    onClick={() => setActiveFilter(filter)}
                    className={`px-4 py-1.5 rounded-full font-mono text-[11px] tracking-wide transition-all duration-300 ${
                      activeFilter === filter
                        ? "bg-white/[0.04] border border-outline-variant/60 text-on-surface shadow-sm"
                        : "border border-transparent text-on-surface-variant/60 hover:text-on-surface"
                    }`}
                  >
                    {filter === "ALL"
                      ? "All Actions"
                      : filter === "FILLS"
                        ? "Fills Only"
                        : "Errors"}
                  </button>
                ))}
              </div>
              <div className="flex items-center gap-2 font-mono text-[11px] tracking-wide text-on-surface-variant/60 select-none">
                <span className="material-symbols-outlined text-[16px]">
                  filter_list
                </span>
                <span>Filter by Asset</span>
              </div>
            </div>

            {/* Structured Cyber Terminal Display Box */}
            <div className="group relative rounded-[24px] bg-surface-container-low border border-outline-variant/30 overflow-hidden hover:shadow-[0_20px_40px_rgba(0,0,0,0.2)] transition-all duration-700">
              {/* Top accent glow pipeline border */}
              <div className="absolute top-0 inset-x-0 h-[1px] bg-gradient-to-r from-transparent via-primary-fixed-dim/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />

              {/* Linear tracking alignment backdrop matrix */}
              <div
                className="absolute inset-0 opacity-[0.03] pointer-events-none"
                style={{
                  backgroundImage:
                    "linear-gradient(rgba(255,255,255,0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.05) 1px, transparent 1px)",
                  backgroundSize: "20px 20px",
                }}
              />

              <div className="overflow-x-auto w-full">
                <table className="w-full text-left border-collapse min-w-[750px]">
                  <thead>
                    <tr className="border-b border-outline-variant/20 bg-void-black/30 font-mono text-[10px] tracking-wider text-on-surface-variant/50 uppercase">
                      <th className="py-4 px-6 font-medium">TIMESTAMP (UTC)</th>
                      <th className="py-4 px-6 font-medium">ACTION</th>
                      <th className="py-4 px-6 font-medium">ASSET</th>
                      <th className="py-4 px-6 font-medium text-right">
                        PRICE
                      </th>
                      <th className="py-4 px-6 font-medium text-right">SIZE</th>
                      <th className="py-4 px-6 font-medium text-right">
                        SLIPPAGE
                      </th>
                      <th className="py-4 px-6 font-medium">STATUS</th>
                    </tr>
                  </thead>
                  <tbody className="font-mono text-[11px] tracking-wide text-on-surface/90 divide-y divide-outline-variant/10">
                    {mockExecutionStream.map((log, index) => (
                      <tr
                        key={index}
                        className="transition-colors duration-150 hover:bg-white/[0.01] cursor-default"
                      >
                        <td className="py-4 px-6 text-on-surface-variant/60">
                          {log.timestamp}
                        </td>
                        <td className="py-4 px-6">
                          <span className={log.actionColorClass}>
                            {log.action}
                          </span>
                        </td>
                        <td className="py-4 px-6 font-medium text-on-surface">
                          {log.asset}
                        </td>
                        <td className="py-4 px-6 text-right text-on-surface-variant">
                          {log.price}
                        </td>
                        <td className="py-4 px-6 text-right text-on-surface-variant">
                          {log.size}
                        </td>
                        <td
                          className={`py-4 px-6 text-right ${log.action.startsWith("BUY") ? "text-primary-fixed-dim" : "text-on-surface-variant/40"}`}
                        >
                          {log.slippage}
                        </td>
                        <td className="py-4 px-6">
                          <span
                            className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-sm border font-medium text-[9px] tracking-wider ${log.statusColorClass}`}
                          >
                            {log.status === "FILLED" && (
                              <span className="w-1 h-1 rounded-full bg-primary-fixed-dim shadow-[0_0_6px_#00dbe9]" />
                            )}
                            {log.status === "PENDING" && (
                              <span className="w-1 h-1 rounded-full bg-secondary-fixed" />
                            )}
                            {log.status}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* History Expansion Control */}
              <div className="p-4 border-t border-outline-variant/20 bg-void-black/20 flex justify-center">
                <button className="text-primary-fixed-dim font-mono text-[11px] tracking-wider uppercase hover:text-primary transition-colors flex items-center gap-1.5 duration-300">
                  Load More History
                  <span className="material-symbols-outlined text-[16px] mt-0.5">
                    expand_more
                  </span>
                </button>
              </div>
            </div>
          </div>

          {/* RIGHT: RISK WARNINGS & TELEMETRY TIMELINES (SPAN 12 OVER TRANSIT COHORTS -> SPAN 4 ON XL) */}
          <div className="md:col-span-12 xl:col-span-4 flex flex-col gap-6 md:gap-8 w-full">
            {/* ALERT CONTROL BOX MONITOR */}
            <div className="rounded-[24px] bg-surface-container-low border border-outline-variant/30 p-6 relative overflow-hidden group">
              <div className="absolute top-0 right-0 w-32 h-32 bg-error/5 rounded-full blur-3xl pointer-events-none" />

              <div className="flex items-center gap-3 mb-6 border-b border-outline-variant/20 pb-4">
                <div className="text-error">
                  <Icon name="warning" />
                </div>
                <h2 className="text-sm font-light tracking-wide text-on-surface">
                  System Alerts
                </h2>
              </div>

              <div className="flex flex-col gap-4">
                {mockSystemAlerts.map((alert) => (
                  <div
                    key={alert.id}
                    className={`flex gap-4 p-4 rounded-xl bg-void-black/20 border border-outline-variant/10 border-l-2 ${alert.borderClass}`}
                  >
                    <div className="mt-0.5 shrink-0 text-on-surface-variant/60">
                      <Icon name={alert.iconName} size="small" />
                    </div>
                    <div>
                      <div className="font-mono text-[11px] font-medium tracking-wide text-on-surface mb-1">
                        {alert.title}
                      </div>
                      <p className="text-xs font-light leading-relaxed text-on-surface-variant/70">
                        {alert.description}
                      </p>
                      <div className="font-mono text-[9px] text-on-surface-variant/40 tracking-wider mt-2.5">
                        {alert.timeLabel}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </AppShell>
  );
}
