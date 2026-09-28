import React from 'react';
import { MainLayout } from '../../components/layout/MainLayout';
import { Card } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Sliders, Grid, Database, Lock, Server } from 'lucide-react';

export const SettingsPage: React.FC = () => {
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto p-8 space-y-8">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Settings & Plugins</h1>
          <p className="text-xs text-zinc-400 mt-1">Configure Aegis cluster connectivity, worker nodes, and extensions.</p>
        </div>

        <Card className="p-6 space-y-6">
          <div className="flex items-center gap-3 pb-4 border-b border-surface-border">
            <Server className="w-5 h-5 text-amber-400" />
            <div>
              <h3 className="text-sm font-bold text-white">Go Control Plane Node</h3>
              <p className="text-xs text-zinc-400">http://localhost:8080 (Operational)</p>
            </div>
          </div>

          <div className="space-y-4 text-xs">
            <div className="flex items-center justify-between py-2 border-b border-surface-border/50">
              <div>
                <span className="font-semibold text-zinc-200">Rate Limiter</span>
                <p className="text-zinc-500">Redis Token Bucket rate limiting on /v1/chat/completions</p>
              </div>
              <Badge variant="success">Enabled</Badge>
            </div>

            <div className="flex items-center justify-between py-2 border-b border-surface-border/50">
              <div>
                <span className="font-semibold text-zinc-200">Admission Control</span>
                <p className="text-zinc-500">Queue admission control for worker protection</p>
              </div>
              <Badge variant="success">Active</Badge>
            </div>

            <div className="flex items-center justify-between py-2">
              <div>
                <span className="font-semibold text-zinc-200">Kafka Event Outbox</span>
                <p className="text-zinc-500">Durable asynchronous metering & billing events</p>
              </div>
              <Badge variant="info">Connected</Badge>
            </div>
          </div>
        </Card>

        {/* Installed Plugins section matching bottom sidebar item */}
        <Card className="p-6 space-y-4">
          <div className="flex items-center gap-3 pb-3 border-b border-surface-border">
            <Grid className="w-5 h-5 text-sky-400" />
            <h3 className="text-sm font-bold text-white">Installed Platform Plugins</h3>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {[
              { title: 'Grok-UI Aesthetic Extension', desc: 'Futuristic dark chat canvas design system', active: true },
              { title: 'OpenRouter BYOK Connector', desc: 'Multi-provider upstream API routing', active: true },
              { title: 'Kafka Metering Consumer', desc: 'Usage tracking and billing event ingestion', active: true },
              { title: 'Prometheus Metrics Exporter', desc: '/metrics endpoint metrics collector', active: true },
            ].map((p, i) => (
              <div key={i} className="p-3 bg-zinc-900 border border-zinc-800 rounded-xl space-y-1">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-zinc-200">{p.title}</span>
                  <Badge variant="success">Active</Badge>
                </div>
                <p className="text-[11px] text-zinc-500">{p.desc}</p>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </MainLayout>
  );
};

export default SettingsPage;
