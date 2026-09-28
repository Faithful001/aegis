import React, { useState } from 'react';
import { MainLayout } from '../../../components/layout/MainLayout';
import { useAegis } from '../../../context/AegisContext';
import { Card } from '../../../components/ui/Card';
import { Button } from '../../../components/ui/Button';
import { Badge } from '../../../components/ui/Badge';
import { Key, Plus, Trash2, Copy, Shield, Check, Cpu } from 'lucide-react';
import { CreateKeyModal } from '../../../components/modals/CreateKeyModal';
import { projectsApi } from '../../../api/projects';
import { toast } from 'sonner';

interface Props {
  id?: string;
}

export const DynamicProjectDetailsPage: React.FC<Props> = ({ id }) => {
  const { activeProject, apiKeys, refreshKeys } = useAegis();
  const [createKeyOpen, setCreateKeyOpen] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleRevoke = async (keyId: string) => {
    try {
      await projectsApi.revokeApiKey(keyId);
      toast.success('API Key revoked');
      await refreshKeys();
    } catch (e: any) {
      toast.error('Failed to revoke API key');
    }
  };

  const copyPrefix = (prefix: string, keyId: string) => {
    navigator.clipboard.writeText(`aegis_${prefix}_xxxxxxxxxxxx`);
    setCopiedId(keyId);
    toast.success('Key prefix copied!');
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <MainLayout>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-mono text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/30">
                Project Details
              </span>
              <Badge variant="success">Active</Badge>
            </div>
            <h1 className="text-2xl font-bold text-white tracking-tight">
              {activeProject?.name || 'Project Details'}
            </h1>
            <p className="text-xs text-zinc-400 mt-1">
              {activeProject?.description || 'Manage project API keys and data plane inference access.'}
            </p>
          </div>

          <Button onClick={() => setCreateKeyOpen(true)} className="flex items-center gap-2">
            <Plus className="w-4 h-4" /> Generate API Key
          </Button>
        </div>

        {/* API Keys Table */}
        <Card className="p-6 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-surface-border">
            <div className="flex items-center gap-2">
              <Key className="w-4 h-4 text-amber-400" />
              <h3 className="text-sm font-bold text-white">Project API Keys</h3>
            </div>
            <span className="text-xs text-zinc-500 font-mono">{apiKeys.length} Keys Active</span>
          </div>

          {apiKeys.length === 0 ? (
            <div className="text-center py-8 space-y-3">
              <Key className="w-8 h-8 text-zinc-600 mx-auto" />
              <p className="text-xs text-zinc-400">No API Keys generated for this project yet.</p>
              <Button size="sm" onClick={() => setCreateKeyOpen(true)}>
                Create First Key
              </Button>
            </div>
          ) : (
            <div className="divide-y divide-surface-border">
              {apiKeys.map((key) => (
                <div key={key.id} className="py-3 flex items-center justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-semibold text-zinc-200">{key.name}</span>
                      <span className="text-[11px] font-mono text-amber-400 bg-zinc-900 px-2 py-0.5 rounded border border-zinc-800">
                        {key.key_prefix}...
                      </span>
                    </div>
                    <p className="text-[11px] text-zinc-500">
                      Created {key.created_at ? new Date(key.created_at).toLocaleDateString() : 'Recently'}
                    </p>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      size="icon"
                      variant="ghost"
                      onClick={() => copyPrefix(key.key_prefix, key.id)}
                      className="text-zinc-400 hover:text-white"
                    >
                      {copiedId === key.id ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                    </Button>
                    <Button
                      size="icon"
                      variant="ghost"
                      onClick={() => handleRevoke(key.id)}
                      className="text-zinc-500 hover:text-red-400"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>

        {/* Inference Endpoint Reference */}
        <Card className="p-6 space-y-3">
          <div className="flex items-center gap-2 text-sm font-bold text-white">
            <Cpu className="w-4 h-4 text-sky-400" />
            <span>OpenAI-Compatible Inference Endpoint</span>
          </div>
          <p className="text-xs text-zinc-400">
            Send chat completions directly to Aegis Go worker scheduler using any OpenAI SDK or standard cURL:
          </p>
          <div className="p-3 bg-zinc-950 rounded-xl border border-zinc-800 font-mono text-xs text-zinc-300 overflow-x-auto">
            curl -X POST http://localhost:8080/v1/chat/completions \<br />
            &nbsp;&nbsp;-H "Authorization: Bearer YOUR_AEGIS_API_KEY" \<br />
            &nbsp;&nbsp;-H "Content-Type: application/json" \<br />
            &nbsp;&nbsp;-d &#39;&#123; "model": "aegis-mistral-7b", "messages": [&#123;"role": "user", "content": "Hello Aegis!"&#125;] &#125;&#39;
          </div>
        </Card>

        <CreateKeyModal open={createKeyOpen} onOpenChange={setCreateKeyOpen} />
      </div>
    </MainLayout>
  );
};

export default DynamicProjectDetailsPage;
