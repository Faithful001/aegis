import React, { useState, useEffect } from "react";
import { useMutation } from "@tanstack/react-query";
import { Dialog } from "../ui/Dialog";
import { Input } from "../ui/Input";
import { Button } from "../ui/Button";
import { ExternalLink, RefreshCw } from "lucide-react";
import { useAegis } from "../../context/AegisContext";
import { credentialsApi } from "../../api/credentials";
import { ProviderMeta } from "../../types/api";
import { toast } from "sonner";
import { validateApiKey, validateBaseUrl } from "../../utils/validation";

interface ConfigureProviderModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  provider: ProviderMeta | null;
  onSuccess?: () => void;
}

export const ConfigureProviderModal: React.FC<ConfigureProviderModalProps> = ({
  open,
  onOpenChange,
  provider,
  onSuccess,
}) => {
  const [apiKey, setApiKey] = useState("");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKeyError, setApiKeyError] = useState<string | null>(null);
  const [baseUrlError, setBaseUrlError] = useState<string | null>(null);
  const [touched, setTouched] = useState<{ apiKey?: boolean; baseUrl?: boolean }>({});

  const { refreshCredentials } = useAegis();

  useEffect(() => {
    if (open) {
      setApiKey("");
      setBaseUrl("");
      setApiKeyError(null);
      setBaseUrlError(null);
      setTouched({});
    }
  }, [open, provider?.id]);

  const handleApiKeyChange = (val: string) => {
    setApiKey(val);
    if (touched.apiKey && provider) {
      setApiKeyError(validateApiKey(provider.id, val, baseUrl));
    }
  };

  const handleBaseUrlChange = (val: string) => {
    setBaseUrl(val);
    if (touched.baseUrl) {
      setBaseUrlError(validateBaseUrl(val));
    }
    // Base URL change might affect whether provider prefix check applies to API key
    if (touched.apiKey && provider) {
      setApiKeyError(validateApiKey(provider.id, apiKey, val));
    }
  };

  const handleApiKeyBlur = () => {
    if (!provider) return;
    setTouched((prev) => ({ ...prev, apiKey: true }));
    setApiKeyError(validateApiKey(provider.id, apiKey, baseUrl));
  };

  const handleBaseUrlBlur = () => {
    setTouched((prev) => ({ ...prev, baseUrl: true }));
    setBaseUrlError(validateBaseUrl(baseUrl));
  };

  const saveMutation = useMutation({
    mutationFn: async ({
      providerId,
      apiKey,
      baseUrl,
    }: {
      providerId: string;
      apiKey: string;
      baseUrl?: string;
    }) => {
      return await credentialsApi.save(providerId, apiKey, baseUrl);
    },
    onSuccess: async () => {
      toast.success(`${provider?.name || "Provider"} credential saved! Discovering models…`);
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

      // If server returned a field-specific error message, highlight the relevant field
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
    if (!provider || saveMutation.isPending) return;

    // Run full validation
    const keyErr = validateApiKey(provider.id, apiKey, baseUrl);
    const urlErr = validateBaseUrl(baseUrl);

    setApiKeyError(keyErr);
    setBaseUrlError(urlErr);
    setTouched({ apiKey: true, baseUrl: true });

    if (keyErr || urlErr) {
      toast.error(keyErr || urlErr || "Please fix the validation errors");
      return;
    }

    saveMutation.mutate({
      providerId: provider.id,
      apiKey: apiKey.trim(),
      baseUrl: baseUrl.trim() || undefined,
    });
  };

  if (!provider) return null;

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Configure ${provider.name}`}
      description={provider.desc || "Store your provider credentials securely in Aegis."}
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-1">
        <div className="space-y-2">
          <Input
            label={`${provider.name} API Key`}
            type="password"
            placeholder={provider.placeholder || "Enter API key…"}
            value={apiKey}
            onChange={(e) => handleApiKeyChange(e.target.value)}
            onBlur={handleApiKeyBlur}
            error={apiKeyError || undefined}
            showToggle
            required
            autoFocus
          />

          {provider.key_link && (
            <div className="flex items-center justify-between text-xs pt-0.5 px-0.5">
              <span className="text-zinc-500">Need an API key?</span>
              <a
                href={provider.key_link}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 transition-colors text-emerald-400 hover:text-emerald-300"
              >
                Get {provider.name} key <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          )}
        </div>

        <div className="space-y-1">
          <Input
            label="Custom Base URL (Optional)"
            type="text"
            placeholder="https://api.example.com/v1 (leave empty for default)"
            value={baseUrl}
            onChange={(e) => handleBaseUrlChange(e.target.value)}
            onBlur={handleBaseUrlBlur}
            error={baseUrlError || undefined}
          />
          <p className="text-[11px] text-zinc-500 px-0.5">
            Only set this if you use a custom gateway, proxy, or self-hosted deployment.
          </p>
        </div>

        <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-surface-border">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="rounded-full px-4"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            size="sm"
            className="rounded-full px-4 py-2 flex items-center gap-1.5"
            disabled={saveMutation.isPending || !apiKey.trim()}
          >
            {saveMutation.isPending ? (
              <>
                <RefreshCw className="w-3.5 h-3.5 animate-spin" /> Saving…
              </>
            ) : (
              <>Save</>
            )}
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
