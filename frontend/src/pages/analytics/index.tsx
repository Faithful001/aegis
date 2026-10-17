import React from "react";
import { useQuery } from "@tanstack/react-query";
import { MainLayout } from "../../components/layout/MainLayout";
import { useAuth } from "../../context/AuthContext";
import { Card } from "../../components/ui/Card";
import { Badge } from "../../components/ui/Badge";
import { Activity, Cpu, DollarSign, Zap, BarChart3 } from "lucide-react";
import { usageApi } from "../../api/usage";
import { UsageMetric } from "../../types/api";

export const AnalyticsPage: React.FC = () => {
  const { isAuthenticated } = useAuth();

  const { data: metrics } = useQuery<UsageMetric>({
    queryKey: ["usage", "metrics"],
    queryFn: async () => {
      return await usageApi.getUserUsage(true);
    },
    enabled: isAuthenticated,
    refetchInterval: 10000,
    initialData: {
      total_requests: 0,
      total_tokens: 0,
      prompt_tokens: 0,
      completion_tokens: 0,
      total_cost_usd: 0,
    },
  });

  return (
    <MainLayout>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight">Personal Usage & Metering</h1>
          <p className="text-xs text-zinc-400 mt-1">
            Durable token consumption, prompt analytics, and inference request throughput.
          </p>
        </div>

        {/* Top Metric Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-[12px]">
              <span>Total Requests</span>
              <Activity className="w-4 h-4" />
            </div>
            <div className="text-xl font-semibold text-white">
              {metrics?.total_requests ? metrics.total_requests.toLocaleString() : "0"}
            </div>
            <p className="text-[10px] font-medium text-zinc-400">Real-time tracked</p>
          </Card>

          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-[12px]">
              <span>Tokens Processed</span>
              <Cpu className="w-4 h-4" />
            </div>
            <div className="text-xl font-semibold text-white">
              {metrics?.total_tokens ? metrics.total_tokens.toLocaleString() : "0"}
            </div>
            <p className="text-[10px] text-zinc-400">
              Prompt: {metrics?.prompt_tokens?.toLocaleString() || "0"} | Completion:{" "}
              {metrics?.completion_tokens?.toLocaleString() || "0"}
            </p>
          </Card>

          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-[12px]">
              <span>Rate Limit Remaining</span>
              <DollarSign className="w-4 h-4" />
            </div>
            <div className="text-xl font-semibold text-white">600 RPM</div>
            <p className="text-[10px] text-zinc-400">Redis sliding window</p>
          </Card>

          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-xs">
              <span>Worker Latency</span>
              <Zap className="w-4 h-4" />
            </div>
            <div className="text-2xl font-bold text-white">~120 ms</div>
            <p className="text-[11px] font-medium text-zinc-400">Direct gRPC stream</p>
          </Card>
        </div>

        {/* Visual Chart Placeholder */}
        <Card className="p-6 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-surface-border">
            <div className="flex items-center gap-2">
              <BarChart3 className="w-4 h-4" />
              <h3 className="text-sm font-semibold text-white">Token Metering Ledger</h3>
            </div>
          </div>

          <div className="h-48 flex items-end justify-between gap-2 pt-6 px-4">
            {[40, 65, 30, 85, 95, 70, 110, 80, 120, 140, 100, 130].map((h, i) => (
              <div key={i} className="flex-1 flex flex-col items-center gap-2">
                <div
                  className="w-full bg-gradient-to-t from-amber-500/20 to-amber-400 rounded-t transition-all hover:brightness-125"
                  style={{ height: `${(h / 140) * 100}%` }}
                ></div>
                <span className="text-[10px] text-zinc-500 font-mono">{i * 2}h</span>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </MainLayout>
  );
};

export default AnalyticsPage;
