import React, { useState } from 'react';
import { Dialog } from '../ui/Dialog';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { useAegis } from '../../context/AegisContext';
import { credentialsApi } from '../../api/credentials';
import { toast } from 'sonner';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  defaultProvider?: string;
}

const SUPPORTED_PROVIDERS = [
  { id: 'mistral', name: 'Mistral AI', placeholder: 'your_mistral_api_key' },
  { id: 'openai', name: 'OpenAI', placeholder: 'sk-proj-...' },
  { id: 'anthropic', name: 'Anthropic', placeholder: 'sk-ant-...' },
  { id: 'openrouter', name: 'OpenRouter', placeholder: 'sk-or-...' },
  { id: 'gemini', name: 'Google Gemini', placeholder: 'AIzaSy...' },
];

export const AddCredentialModal: React.FC<Props> = ({
  open,
  onOpenChange,
  onSuccess,
  defaultProvider = 'mistral',
}) => {
  const [provider, setProvider] = useState<string>(defaultProvider);
  const [apiKey, setApiKey] = useState('');
  const [baseUrl, setBaseUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const { refreshCredentials } = useAegis();

  const selectedProviderMeta =
    SUPPORTED_PROVIDERS.find((p) => p.id === provider) || SUPPORTED_PROVIDERS[0];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!apiKey.trim()) return;
    setLoading(true);
    try {
      await credentialsApi.save(provider, apiKey.trim(), baseUrl.trim() || undefined);
      toast.success(`${selectedProviderMeta.name} credential saved to your vault!`);
      setApiKey('');
      setBaseUrl('');
      await refreshCredentials();
      onOpenChange(false);
      if (onSuccess) onSuccess();
    } catch (err: any) {
      toast.error(err?.response?.data?.error || err?.response?.data?.message || 'Failed to save credential');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title="Add BYOK Provider Key"
      description="Store your model provider API key encrypted in your personal Aegis vault."
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-1.5">Provider</label>
          <div className="grid grid-cols-3 gap-2">
            {SUPPORTED_PROVIDERS.map((p) => (
              <button
                key={p.id}
                type="button"
                onClick={() => setProvider(p.id)}
                className={`py-2 px-2.5 text-xs font-medium rounded-xl border transition-all truncate ${
                  provider === p.id
                    ? 'bg-zinc-100 text-zinc-950 border-white font-semibold'
                    : 'bg-surface-card text-zinc-400 border-surface-border hover:text-white'
                }`}
              >
                {p.name}
              </button>
            ))}
          </div>
        </div>

        <Input
          label={`${selectedProviderMeta.name} API Key`}
          type="password"
          placeholder={selectedProviderMeta.placeholder}
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          required
        />

        <Input
          label="Custom Base URL (Optional)"
          type="text"
          placeholder="https://api.mistral.ai/v1 (leave empty for default)"
          value={baseUrl}
          onChange={(e) => setBaseUrl(e.target.value)}
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
