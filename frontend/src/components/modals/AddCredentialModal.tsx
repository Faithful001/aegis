import React, { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Dialog } from "../ui/Dialog";
import { Input } from "../ui/Input";
import { Button } from "../ui/Button";
import { useAegis } from "../../context/AegisContext";
import { credentialsApi } from "../../api/credentials";
import { toast } from "sonner";
import { validateApiKey, validateBaseUrl } from "../../utils/validation";

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
  defaultProvider = "mistral",
}) => {
  const [provider, setProvider] = useState<string>(defaultProvider);
  const [apiKey, setApiKey] = useState("");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKeyError, setApiKeyError] = useState<string | null>(null);
  const [baseUrlError, setBaseUrlError] = useState<string | null>(null);
  const [touched, setTouched] = useState<{ apiKey?: boolean; baseUrl?: boolean }>({});

  const { refreshCredentials, supportedProviders } = useAegis();

  const providersList = supportedProviders.length > 0 ? supportedProviders : [];

  const selectedProviderMeta =
    providersList.find((p) => p.id === provider) ||
    providersList[0] || {
      id: provider,
      name: provider,
      placeholder: "Enter API key...",
    };

  const handleProviderSelect = (pId: string) => {
    setProvider(pId);
    if (touched.apiKey) {
      setApiKeyError(validateApiKey(pId, apiKey, baseUrl));
    }
  };

  const handleApiKeyChange = (val: string) => {
    setApiKey(val);
    if (touched.apiKey) {
      setApiKeyError(validateApiKey(provider, val, baseUrl));
    }
  };

  const handleBaseUrlChange = (val: string) => {
    setBaseUrl(val);
    if (touched.baseUrl) {
      setBaseUrlError(validateBaseUrl(val));
    }
    if (touched.apiKey) {
      setApiKeyError(validateApiKey(provider, apiKey, val));
    }
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
      setApiKey("");
      setBaseUrl("");
      setApiKeyError(null);
      setBaseUrlError(null);
      setTouched({});
      await refreshCredentials();
      onOpenChange(false);
      if (onSuccess) onSuccess();
    },
    onError: (err: any) => {
      const serverMsg =
        err?.response?.data?.error || err?.response?.data?.message || "Failed to save credential";
      toast.error(serverMsg);

      const lowerMsg = serverMsg.toLowerCase();
      if (lowerMsg.includes("api key") || lowerMsg.includes("key")) {
        setApiKeyError(serverMsg);
      } else if (lowerMsg.includes("base url") || lowerMsg.includes("url")) {
        setBaseUrlError(serverMsg);
      }
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (saveMutation.isPending) return;

    const keyErr = validateApiKey(provider, apiKey, baseUrl);
    const urlErr = validateBaseUrl(baseUrl);

    setApiKeyError(keyErr);
    setBaseUrlError(urlErr);
    setTouched({ apiKey: true, baseUrl: true });

    if (keyErr || urlErr) {
      toast.error(keyErr || urlErr || "Please fix validation errors");
      return;
    }

    saveMutation.mutate({
      provider,
      apiKey: apiKey.trim(),
      baseUrl: baseUrl.trim() || undefined,
    });
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) {
          setApiKeyError(null);
          setBaseUrlError(null);
          setTouched({});
        }
        onOpenChange(v);
      }}
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
                onClick={() => handleProviderSelect(p.id)}
                className={`py-2 px-2.5 text-xs font-medium rounded-xl border transition-all truncate ${
                  provider === p.id
                    ? "bg-zinc-100 text-zinc-950 border-white font-semibold"
                    : "bg-surface-card text-zinc-400 border-surface-border hover:text-white"
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
          onChange={(e) => handleApiKeyChange(e.target.value)}
          onBlur={() => {
            setTouched((prev) => ({ ...prev, apiKey: true }));
            setApiKeyError(validateApiKey(provider, apiKey, baseUrl));
          }}
          error={apiKeyError || undefined}
          showToggle
          required
        />

        <Input
          label="Custom Base URL (Optional)"
          type="text"
          placeholder="https://api.mistral.ai/v1 (leave empty for default)"
          value={baseUrl}
          onChange={(e) => handleBaseUrlChange(e.target.value)}
          onBlur={() => {
            setTouched((prev) => ({ ...prev, baseUrl: true }));
            setBaseUrlError(validateBaseUrl(baseUrl));
          }}
          error={baseUrlError || undefined}
        />

        <div className="flex justify-end gap-3 pt-3">
          <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={saveMutation.isPending || !apiKey.trim()}>
            {saveMutation.isPending ? "Saving..." : "Save Credential"}
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
