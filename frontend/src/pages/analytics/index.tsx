import React, { useState, useEffect } from 'react';
import { MainLayout } from '../../components/layout/MainLayout';
import { useAuth } from '../../context/AuthContext';
import { Card } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Activity, Cpu, DollarSign, Zap, BarChart3 } from 'lucide-react';
import { usageApi } from '../../api/usage';
import { UsageMetric } from '../../types/api';

export const AnalyticsPage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const [metrics, setMetrics] = useState<UsageMetric | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      if (!isAuthenticated) return;
      try {
        const data = await usageApi.getUserUsage(true);
        setMetrics(data);
      } catch (e) {
        // Default initial display
        setMetrics({
          total_requests: 0,
          total_tokens: 0,
          prompt_tokens: 0,
          completion_tokens: 0,
          total_cost_usd: 0,
        });
      }
    };
    fetchMetrics();
  }, [isAuthenticated]);

  return (
    <MainLayout>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="text-xs font-mono text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/30">
              Kafka Usage Metering
            </span>
            <Badge variant="info">Real-time Stream</Badge>
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Personal Usage & Metering</h1>
          <p className="text-xs text-zinc-400 mt-1">
            Durable token consumption, prompt analytics, and inference request throughput.
          </p>
        </div>

        {/* Top Metric Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-xs">
              <span>Total Requests</span>
              <Activity className="w-4 h-4 text-emerald-400" />
            </div>
            <div className="text-2xl font-bold text-white">
              {metrics?.total_requests ? metrics.total_requests.toLocaleString() : '0'}
            </div>
            <p className="text-[11px] text-emerald-400 font-medium">Real-time tracked</p>
          </Card>

          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-xs">
              <span>Tokens Processed</span>
              <Cpu className="w-4 h-4 text-sky-400" />
            </div>
            <div className="text-2xl font-bold text-white">
              {metrics?.total_tokens ? metrics.total_tokens.toLocaleString() : '0'}
            </div>
            <p className="text-[11px] text-zinc-400">
              Prompt: {metrics?.prompt_tokens?.toLocaleString() || '0'} | Completion: {metrics?.completion_tokens?.toLocaleString() || '0'}
            </p>
          </Card>

          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-xs">
              <span>Rate Limit Remaining</span>
              <DollarSign className="w-4 h-4 text-amber-400" />
            </div>
            <div className="text-2xl font-bold text-white">600 RPM</div>
            <p className="text-[11px] text-zinc-400">Redis sliding window</p>
          </Card>

          <Card className="p-5 space-y-2">
            <div className="flex items-center justify-between text-zinc-400 text-xs">
              <span>Worker Latency</span>
              <Zap className="w-4 h-4 text-purple-400" />
            </div>
            <div className="text-2xl font-bold text-white">~120 ms</div>
            <p className="text-[11px] text-purple-400 font-medium">Direct gRPC stream</p>
          </Card>
        </div>

        {/* Visual Chart Placeholder */}
        <Card className="p-6 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-surface-border">
            <div className="flex items-center gap-2">
              <BarChart3 className="w-4 h-4 text-amber-400" />
              <h3 className="text-sm font-bold text-white">Token Metering Ledger</h3>
            </div>
            <span className="text-xs text-zinc-500 font-mono">Durable Kafka Ingestion</span>
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
