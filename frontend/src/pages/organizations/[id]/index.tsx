import React, { useState, useEffect } from 'react';
import { MainLayout } from '../../../components/layout/MainLayout';
import { useAegis } from '../../../context/AegisContext';
import { Card } from '../../../components/ui/Card';
import { Button } from '../../../components/ui/Button';
import { Badge } from '../../../components/ui/Badge';
import { Key, Plus, Trash2, ShieldCheck, Landmark } from 'lucide-react';
import { orgsApi } from '../../../api/orgs';
import { ProviderCredential } from '../../../types/api';
import { AddCredentialModal } from '../../../components/modals/AddCredentialModal';
import { toast } from 'sonner';

interface Props {
  id?: string;
}

export const DynamicOrgDetailsPage: React.FC<Props> = ({ id }) => {
  const { activeOrg } = useAegis();
  const [credentials, setCredentials] = useState<ProviderCredential[]>([]);
  const [addCredOpen, setAddCredOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const fetchCreds = async () => {
    if (!activeOrg) return;
    setLoading(true);
    try {
      const list = await orgsApi.listCredentials(activeOrg.id);
      setCredentials(list);
    } catch (e: any) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCreds();
  }, [activeOrg]);

  const handleDelete = async (provider: string) => {
    if (!activeOrg) return;
    try {
      await orgsApi.deleteCredential(activeOrg.id, provider);
      toast.success(`${provider.toUpperCase()} credential removed`);
      fetchCreds();
    } catch (e: any) {
      toast.error('Failed to delete credential');
    }
  };

  return (
    <MainLayout>
      <div className="max-w-5xl mx-auto p-8 space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-mono text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/30">
                Organization Vault
              </span>
              <Badge variant="warning">BYOK Active</Badge>
            </div>
            <h1 className="text-2xl font-bold text-white tracking-tight">
              {activeOrg?.name || 'Organization Details'}
            </h1>
            <p className="text-xs text-zinc-400 mt-1">
              Manage Bring-Your-Own-Key (BYOK) providers for OpenAI, OpenRouter, and Anthropic.
            </p>
          </div>

          <Button onClick={() => setAddCredOpen(true)} className="flex items-center gap-2">
            <Plus className="w-4 h-4" /> Add Provider Key
          </Button>
        </div>

        {/* Credentials List */}
        <Card className="p-6 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-surface-border">
            <div className="flex items-center gap-2">
              <Key className="w-4 h-4 text-amber-400" />
              <h3 className="text-sm font-bold text-white">Stored Provider Credentials</h3>
            </div>
            <span className="text-xs text-zinc-500 font-mono">{credentials.length} Providers Vaulted</span>
          </div>

          {credentials.length === 0 ? (
            <div className="text-center py-8 space-y-3">
              <Landmark className="w-8 h-8 text-zinc-600 mx-auto" />
              <p className="text-xs text-zinc-400">No BYOK credentials stored for this organization yet.</p>
              <Button size="sm" onClick={() => setAddCredOpen(true)}>
                Add OpenAI / OpenRouter Key
              </Button>
            </div>
          ) : (
            <div className="divide-y divide-surface-border">
              {credentials.map((cred) => (
                <div key={cred.provider} className="py-3.5 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-xl bg-zinc-900 border border-zinc-800 flex items-center justify-center text-amber-400 font-bold uppercase text-xs">
                      {cred.provider.slice(0, 2)}
                    </div>
                    <div>
                      <h4 className="text-sm font-semibold text-white uppercase">{cred.provider}</h4>
                      <p className="text-xs text-zinc-500 font-mono">
                        Key: {cred.api_key_masked || '••••••••••••••••'}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Badge variant="success">Vaulted</Badge>
                    <Button
                      size="icon"
                      variant="ghost"
                      onClick={() => handleDelete(cred.provider)}
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

        <AddCredentialModal open={addCredOpen} onOpenChange={setAddCredOpen} onSuccess={fetchCreds} />
      </div>
    </MainLayout>
  );
};

export default DynamicOrgDetailsPage;
