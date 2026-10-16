import React, { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
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

export const AddCredentialModal: React.FC<Props> = ({
  open,
  onOpenChange,
  onSuccess,
  defaultProvider = 'mistral',
}) => {
  const [provider, setProvider] = useState<string>(defaultProvider);
  const [apiKey, setApiKey] = useState('');
  const [baseUrl, setBaseUrl] = useState('');
  const { refreshCredentials, supportedProviders } = useAegis();

  const providersList = supportedProviders.length > 0 ? supportedProviders : [];

  const selectedProviderMeta =
    providersList.find((p) => p.id === provider) ||
    providersList[0] || {
      id: provider,
      name: provider,
      placeholder: 'Enter API key...',
    };

  const saveMutation = useMutation({
    mutationFn: async ({
      provider,
      apiKey,
      baseUrl,
    }: {
      provider: string;
      apiKey: string;
      baseUrl?: string;
    }) => {
      return await credentialsApi.save(provider, apiKey, baseUrl);
    },
    onSuccess: async () => {
      toast.success(`${selectedProviderMeta.name} credential saved to your vault!`);
      setApiKey('');
      setBaseUrl('');
      await refreshCredentials();
      onOpenChange(false);
      if (onSuccess) onSuccess();
    },
    onError: (err: any) => {
      toast.error(
        err?.response?.data?.error || err?.response?.data?.message || 'Failed to save credential'
      );
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!apiKey.trim() || saveMutation.isPending) return;
    saveMutation.mutate({
      provider,
      apiKey: apiKey.trim(),
      baseUrl: baseUrl.trim() || undefined,
    });
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
            {providersList.map((p) => (
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
          <Button type="submit" disabled={saveMutation.isPending}>
            {saveMutation.isPending ? 'Saving...' : 'Save Credential'}
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
