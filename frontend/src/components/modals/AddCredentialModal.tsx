import React, { useState } from 'react';
import { Dialog } from '../ui/Dialog';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { useAegis } from '../../context/AegisContext';
import { orgsApi } from '../../api/orgs';
import { toast } from 'sonner';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export const AddCredentialModal: React.FC<Props> = ({ open, onOpenChange, onSuccess }) => {
  const [provider, setProvider] = useState<'openai' | 'openrouter' | 'anthropic'>('openrouter');
  const [apiKey, setApiKey] = useState('');
  const [loading, setLoading] = useState(false);
  const { activeOrg } = useAegis();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeOrg) {
      toast.error('No organization selected');
      return;
    }
    if (!apiKey.trim()) return;
    setLoading(true);
    try {
      await orgsApi.saveCredential(activeOrg.id, provider, apiKey.trim());
      toast.success(`${provider.toUpperCase()} API Credential saved!`);
      setApiKey('');
      onOpenChange(false);
      if (onSuccess) onSuccess();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || 'Failed to save credential');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title="Add BYOK Provider Key"
      description={`Connect your upstream LLM provider API key to ${activeOrg?.name || 'your organization'}.`}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-1.5">Provider</label>
          <div className="grid grid-cols-3 gap-2">
            {[
              { id: 'openrouter', name: 'OpenRouter' },
              { id: 'openai', name: 'OpenAI' },
              { id: 'anthropic', name: 'Anthropic' },
            ].map((p) => (
              <button
                key={p.id}
                type="button"
                onClick={() => setProvider(p.id as any)}
                className={`py-2 px-3 text-xs font-medium rounded-xl border transition-all ${
                  provider === p.id
                    ? 'bg-zinc-100 text-zinc-950 border-white'
                    : 'bg-surface-card text-zinc-400 border-surface-border hover:text-white'
                }`}
              >
                {p.name}
              </button>
            ))}
          </div>
        </div>

        <Input
          label={`${provider.toUpperCase()} API Key`}
          type="password"
          placeholder={`sk-...`}
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          required
        />

        <div className="flex justify-end gap-3 pt-3">
          <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? 'Saving...' : 'Save Credential'}
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
