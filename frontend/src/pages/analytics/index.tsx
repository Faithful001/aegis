import React, { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { MainLayout } from "../../components/layout/MainLayout";
import { useAuth } from "../../context/AuthContext";
import { Card } from "../../components/ui/Card";
import { usageApi } from "../../api/usage";
import { UsageMetric } from "../../types/api";
import { TokenActivityHeatmap } from "../../components/analytics/TokenActivityHeatmap";
import { cn } from "../../components/ui/Button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../components/ui/Select";
import { Skeleton } from "../../components/ui/Skeleton";

type Granularity = "daily" | "weekly" | "cumulative";
type RangeKey = "1D" | "7D" | "30D" | "3M" | "12M" | "ALL";

interface RangePreset {
  label: string;
  key: RangeKey;
  range: () => [string | undefined, string | undefined];
  since: () => string;
}

const now = () => new Date();
const iso = (d: Date) => d.toISOString();

const subtractDays = (n: number) => {
  const d = now();
  d.setDate(d.getDate() - n);
  return d;
};

const subtractMonths = (n: number) => {
  const d = now();
  d.setMonth(d.getMonth() - n);
  return d;
};

const startOfDay = (d: Date) => {
  d.setHours(0, 0, 0, 0);
  return d;
};

const RANGE_PRESETS: RangePreset[] = [
  {
    key: "1D",
    label: "Today",
    range: () => [iso(startOfDay(now())), iso(now())],
    since: () => iso(startOfDay(now())),
  },
  {
    key: "7D",
    label: "7D",
    range: () => [iso(subtractDays(7)), iso(now())],
    since: () => iso(subtractDays(7)),
  },
  {
    key: "30D",
    label: "30D",
    range: () => [iso(subtractDays(30)), iso(now())],
    since: () => iso(subtractDays(30)),
  },
  {
    key: "3M",
    label: "3M",
    range: () => [iso(subtractMonths(3)), iso(now())],
    since: () => iso(subtractMonths(3)),
  },
  {
    key: "12M",
    label: "12M",
    range: () => [iso(subtractMonths(12)), iso(now())],
    since: () => iso(subtractMonths(12)),
  },
  {
    key: "ALL",
    label: "All time",
    range: () => [undefined, undefined],
    since: () => iso(subtractMonths(12)),
  },
];

const DEFAULT_RANGE: RangeKey = "7D";

export const AnalyticsPage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const [granularity, setGranularity] = useState<Granularity>("daily");
  const [activeRange, setActiveRange] = useState<RangeKey>(DEFAULT_RANGE);

  const preset = useMemo(() => RANGE_PRESETS.find((p) => p.key === activeRange)!, [activeRange]);

  const [startTime, endTime] = useMemo(() => preset.range(), [preset]);
  const since = useMemo(() => preset.since(), [preset]);

  const {
    data: metrics,
    isFetching,
    isLoading,
    isPending,
  } = useQuery<UsageMetric>({
    queryKey: ["usage", "metrics", activeRange],
    queryFn: () => usageApi.getUserUsage(true, startTime, endTime),
    enabled: isAuthenticated,
    refetchInterval: 30_000,
    placeholderData: (prev) => prev,
  });

  const loading = isLoading || isPending;

  const fmt = (n?: number) => (n ? n.toLocaleString() : "0");
  const fmtUSD = (n?: number) =>
    n ? (n < 0.01 ? `$${n.toFixed(6)}` : `$${n.toFixed(4)}`) : "$0.00";

  return (
    <MainLayout>
      <div className="max-w-5xl mx-auto p-8 space-y-6">
        {/* Header + filter row */}
        <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
          <div>
            <h1 className="text-xl font-bold text-white tracking-tight">Usage &amp; Metering</h1>
            <p className="text-xs text-zinc-400 mt-1">
              Token usage, request volume, and cost estimates.
            </p>
          </div>

          {/* Range dropdown */}
          <div className="self-start sm:self-auto">
            <Select value={activeRange} onValueChange={(val) => setActiveRange(val as RangeKey)}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="Select range" />
              </SelectTrigger>
              <SelectContent align="end">
                {RANGE_PRESETS.map((p) => (
                  <SelectItem key={p.key} value={p.key}>
                    {p.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Metric cards */}
        <div
          className={cn(
            "grid grid-cols-2 lg:grid-cols-4 gap-4 transition-opacity duration-300",
            isFetching && "opacity-60"
          )}
        >
          <MetricCard
            label="Total Requests"
            value={fmt(metrics?.total_requests)}
            loading={loading}
          />
          <MetricCard label="Input Tokens" value={fmt(metrics?.input_tokens)} loading={loading} />
          <MetricCard label="Output Tokens" value={fmt(metrics?.output_tokens)} loading={loading} />
          <MetricCard
            label="Est. Cost"
            value={fmtUSD(metrics?.total_cost_usd)}
            sub={`${fmt(metrics?.total_tokens)} tokens`}
            loading={loading}
          />
        </div>

        {/* Token Activity Heatmap */}
        <Card className="p-6">
          <TokenActivityHeatmap
            granularity={granularity}
            onGranularityChange={setGranularity}
            // since={since}
          />
        </Card>
      </div>
    </MainLayout>
  );
};

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------
function MetricCard({
  label,
  value,
  sub,
  loading,
}: {
  label: string;
  value: string;
  sub?: string;
  loading?: boolean;
}) {
  return (
    <Card className="p-5 space-y-1.5">
      <span className="text-zinc-400 text-[11px] uppercase tracking-wider font-medium">
        {label}
      </span>
      {loading ? (
        <Skeleton className="h-8 w-28 my-0.5" />
      ) : (
        <div className="text-2xl font-semibold text-white tabular-nums">{value}</div>
      )}
      {loading && sub !== undefined ? (
        <Skeleton className="h-3 w-20" />
      ) : sub ? (
        <p className="text-[10px] text-zinc-500">{sub}</p>
      ) : null}
    </Card>
  );
}

export default AnalyticsPage;
