"use client";

import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { format } from "date-fns";
import type { EquityPoint } from "@/lib/api";

interface Props {
  data: EquityPoint[];
}

export function EquityCurveChart({ data }: Props) {
  const chartData = data.map((p) => ({
    ...p,
    label: format(new Date(p.time), "MMM d, yyyy"),
  }));

  return (
    <div className="h-80 w-full rounded-lg border border-slate-700 bg-panel p-4">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis
            dataKey="label"
            tick={{ fill: "#94a3b8", fontSize: 11 }}
            minTickGap={40}
          />
          <YAxis
            tick={{ fill: "#94a3b8", fontSize: 11 }}
            tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`}
          />
          <Tooltip
            contentStyle={{
              background: "#1a2332",
              border: "1px solid #334155",
              borderRadius: 8,
            }}
            formatter={(value: number) => [
              `$${value.toLocaleString(undefined, { maximumFractionDigits: 0 })}`,
              "Equity",
            ]}
          />
          <Line
            type="monotone"
            dataKey="equity"
            stroke="#3b82f6"
            strokeWidth={2}
            dot={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
