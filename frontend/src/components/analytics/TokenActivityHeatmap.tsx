import React, { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { usageApi } from "../../api/usage";
import { TokenActivityBucket, TokenActivityResponse } from "../../types/api";
import { useAuth } from "../../context/AuthContext";
import { Tooltip } from "../ui/Tooltip";

type Granularity = "daily" | "weekly" | "cumulative";

const toDateStr = (d: Date) => d.toISOString().slice(0, 10);

function formatTooltip(dateStr: string, requests: number, tokens: number): string {
  const d = new Date(dateStr + "T00:00:00");
  const month = d.toLocaleString("default", { month: "long" });
  const day = d.getDate();
  const suffix =
    day === 1 || day === 21 || day === 31
      ? "st"
      : day === 2 || day === 22
        ? "nd"
        : day === 3 || day === 23
          ? "rd"
          : "th";
  const reqLabel = requests === 1 ? "1 request" : `${requests} requests`;
  return `${reqLabel} on ${month} ${day}${suffix}`;
}

function buildDailyGrid(): string[][] {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const start = new Date(today);
  start.setDate(today.getDate() - 52 * 7);
  const dayOfWeek = (start.getDay() + 6) % 7;
  start.setDate(start.getDate() - dayOfWeek);
  const weeks: string[][] = [];
  let cursor = new Date(start);
  while (cursor <= today) {
    const week: string[] = [];
    for (let d = 0; d < 7; d++) {
      week.push(toDateStr(cursor));
      cursor.setDate(cursor.getDate() + 1);
    }
    weeks.push(week);
  }
  return weeks;
}

function intensity(tokens: number, max: number): 0 | 1 | 2 | 3 | 4 {
  if (!tokens || !max) return 0;
  const ratio = tokens / max;
  if (ratio < 0.15) return 1;
  if (ratio < 0.4) return 2;
  if (ratio < 0.7) return 3;
  return 4;
}

const INTENSITY_CLASSES: Record<0 | 1 | 2 | 3 | 4, string> = {
  0: "bg-zinc-800/60",
  1: "bg-emerald-900/70",
  2: "bg-emerald-700/80",
  3: "bg-emerald-500/90",
  4: "bg-emerald-400",
};

const MONTH_LABELS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

interface Props {
  granularity: Granularity;
  onGranularityChange: (g: Granularity) => void;
  since?: string; // ISO 8601 — drives the heatmap time window
}

export const TokenActivityHeatmap: React.FC<Props> = ({ granularity, onGranularityChange, since }) => {
  const { isAuthenticated } = useAuth();

  const { data, isLoading } = useQuery<TokenActivityResponse>({
    queryKey: ["usage", "activity", granularity, since],
    queryFn: () => usageApi.getTokenActivity(granularity, since),
    enabled: isAuthenticated,
    staleTime: 60_000,
    placeholderData: { granularity, buckets: [], max_tokens: 0 },
  });

  const tokenMap = useMemo<Record<string, { tokens: number; requests: number }>>(() => {
    const map: Record<string, { tokens: number; requests: number }> = {};
    (data?.buckets ?? []).forEach((b: TokenActivityBucket) => {
      map[b.date] = { tokens: b.tokens, requests: b.requests };
    });
    return map;
  }, [data]);

  const maxTokens = data?.max_tokens ?? 0;
  const grid = useMemo(() => buildDailyGrid(), []);

  const monthBreaks = useMemo<{ col: number; label: string }[]>(() => {
    const breaks: { col: number; label: string }[] = [];
    let lastMonth = -1;
    grid.forEach((week, col) => {
      const m = new Date(week[0]).getMonth();
      if (m !== lastMonth) {
        breaks.push({ col, label: MONTH_LABELS[m] });
        lastMonth = m;
      }
    });
    return breaks;
  }, [grid]);

  const totalCols = grid.length;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-sm font-semibold text-white tracking-tight">Token activity</span>
        <div className="flex items-center gap-3 text-xs font-medium">
          {(["daily", "weekly", "cumulative"] as Granularity[]).map((g) => (
            <button
              key={g}
              id={`token-activity-${g}`}
              onClick={() => onGranularityChange(g)}
              className={[
                "capitalize transition-colors",
                granularity === g ? "text-white font-bold" : "text-zinc-400 hover:text-zinc-200",
              ].join(" ")}
            >
              {g.charAt(0).toUpperCase() + g.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <div className="overflow-x-auto pb-1">
        <div style={{ minWidth: `${totalCols * 15}px` }}>
          <div
            className="grid gap-[3px]"
            style={{
              gridTemplateColumns: `repeat(${totalCols}, 11px)`,
              gridTemplateRows: "repeat(7, 11px)",
              gridAutoFlow: "column",
            }}
          >
            {grid.map((week, col) =>
              week.map((dateStr, row) => {
                const entry = tokenMap[dateStr];
                const tokens = entry?.tokens ?? 0;
                const requests = entry?.requests ?? 0;
                const level = intensity(tokens, maxTokens) as 0 | 1 | 2 | 3 | 4;
                const cls = isLoading ? "bg-zinc-800/40 animate-pulse" : INTENSITY_CLASSES[level];
                const tip = formatTooltip(dateStr, requests, tokens);
                return (
                  <Tooltip key={`${col}-${row}`} content={tip} placement="top">
                    <div
                      className={`w-[11px] h-[11px] rounded-sm transition-colors cursor-default ${cls}`}
                    />
                  </Tooltip>
                );
              })
            )}
          </div>
          <div className="relative mt-[6px]" style={{ height: "14px" }}>
            {monthBreaks.map(({ col, label }) => (
              <span
                key={label + col}
                className="absolute text-[10px] text-zinc-500 font-mono"
                style={{ left: `${col * 14}px` }}
              >
                {label}
              </span>
            ))}
          </div>
        </div>
      </div>

      <div className="flex items-center gap-1.5 pt-1">
        <span className="text-[10px] text-zinc-500 mr-1">Less</span>
        {([0, 1, 2, 3, 4] as const).map((l) => (
          <div key={l} className={`w-[11px] h-[11px] rounded-sm ${INTENSITY_CLASSES[l]}`} />
        ))}
        <span className="text-[10px] text-zinc-500 ml-1">More</span>
      </div>
    </div>
  );
};

export default TokenActivityHeatmap;
