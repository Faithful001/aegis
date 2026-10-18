import { credentialsApi } from "@/api/credentials";
import { ProviderMeta } from "@/types/api";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import { Card } from "../../components/ui/Card";
import { ChevronDown, ChevronUp, Lock, Trash2 } from "lucide-react";
import { Button } from "../../components/ui/Button";
import { useAegis } from "@/context/AegisContext";
import { ProviderModelDrawer } from "../../components/settings/ProviderModelDrawer";
import { ConfigureProviderModal } from "../../components/modals/ConfigureProviderModal";
import SettingsLayout from "@/components/layout/SettingsLayout";

export const ProvidersTab: React.FC = () => {
  const { supportedProviders, refreshCredentials } = useAegis();
  const providersList = Array.isArray(supportedProviders) ? supportedProviders : [];

  const [configuringProvider, setConfiguringProvider] = useState<ProviderMeta | null>(null);
  const [expandedModels, setExpandedModels] = useState<string | null>(null);

  const deleteCredentialMutation = useMutation({
    mutationFn: async (providerId: string) => credentialsApi.delete(providerId),
    onSuccess: async (_, providerId) => {
      const provider = providersList.find((p) => p.id === providerId);
      toast.success(`${provider?.name || providerId} credential removed`);
      setExpandedModels(null);
      await refreshCredentials();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || "Failed to remove credential");
    },
  });

  return (
    <SettingsLayout
      title="Providers"
      subTitle="Bring your own key. Connect any model provider, then browse its available models."
    >
      <Card hover={false} className="p-0 divide-y divide-surface-border/60 overflow-hidden">
        {providersList.map((provider) => {
          const isConfigured = provider.is_configured;
          const isModelsOpen = expandedModels === provider.id;

          return (
            <div key={provider.id}>
              <div className="flex items-center justify-between px-5 py-4 hover:bg-surface-hover/50 transition-colors">
                <div className="min-w-0">
                  <span className="text-sm font-semibold text-zinc-100">{provider.name}</span>
                  {isConfigured && provider.masked_api_key && (
                    <p className="text-[11px] text-zinc-500 font-mono mt-0.5 flex items-center gap-1">
                      <Lock className="w-2.5 h-2.5" />
                      {provider.masked_api_key}
                    </p>
                  )}
                  <p className="text-[11px] text-zinc-600 mt-0.5">{provider.desc}</p>
                </div>

                <div className="flex items-center gap-2 shrink-0 ml-4">
                  {/* Delete */}
                  {isConfigured && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => deleteCredentialMutation.mutate(provider.id)}
                      disabled={deleteCredentialMutation.isPending}
                      className="text-zinc-500 hover:text-rose-400 hover:bg-rose-500/10"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  )}

                  {/* Configure button */}
                  {!isConfigured && (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => setConfiguringProvider(provider)}
                      className="rounded-full gap-1.5"
                    >
                      Configure
                    </Button>
                  )}

                  {/* Models toggle — only when configured */}
                  {isConfigured && (
                    <button
                      onClick={() =>
                        setExpandedModels((prev) => (prev === provider.id ? null : provider.id))
                      }
                      className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-100 px-2.5 py-1.5 rounded-lg hover:bg-zinc-800/60 transition-all"
                    >
                      Models
                      {isModelsOpen ? (
                        <ChevronUp className="w-3 h-3" />
                      ) : (
                        <ChevronDown className="w-3 h-3" />
                      )}
                    </button>
                  )}
                </div>
              </div>

              {/* ---- Model drawer ---- */}
              {isModelsOpen && isConfigured && (
                <div className="border-t border-surface-border/40 bg-zinc-900/20">
                  <ProviderModelDrawer providerId={provider.id} providerName={provider.name} />
                </div>
              )}
            </div>
          );
        })}
      </Card>

      {/* Configure Provider Modal */}
      <ConfigureProviderModal
        open={!!configuringProvider}
        onOpenChange={(open) => !open && setConfiguringProvider(null)}
        provider={configuringProvider}
      />
    </SettingsLayout>
  );
};
