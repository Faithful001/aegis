import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { authApi } from "@/api/auth";
import SettingsLayout from "@/components/layout/SettingsLayout";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Dialog } from "@/components/ui/Dialog";
import { GoogleIcon, GitHubIcon } from "@/components/auth/OAuthButtons";
import { CheckCircleIcon, WarningCircleIcon, LinkBreakIcon, PlusIcon } from "@phosphor-icons/react";
import { toast } from "sonner";
import { useNavigate } from "@tanstack/react-router";

export const ConnectedAccountsTab: React.FC = () => {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const [disconnectingProvider, setDisconnectingProvider] = useState<string | null>(null);
  const [connectingProvider, setConnectingProvider] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["user", "connected-accounts"],
    queryFn: authApi.getConnectedAccounts,
  });

  const unlinkMutation = useMutation({
    mutationFn: (provider: string) => authApi.unlinkAccount(provider),
    onSuccess: (_, provider) => {
      toast.success(`${provider.charAt(0).toUpperCase() + provider.slice(1)} account disconnected`);
      queryClient.invalidateQueries({ queryKey: ["user", "connected-accounts"] });
      setDisconnectingProvider(null);
    },
    onError: (err: any) => {
      toast.error(
        err?.response?.data?.message || "Failed to disconnect account. Please try again."
      );
      setDisconnectingProvider(null);
    },
  });

  const handleConnect = async (provider: "google" | "github") => {
    try {
      setConnectingProvider(provider);
      const state = Math.random().toString(36).substring(2, 15);
      sessionStorage.setItem("aegis_oauth_state", state);
      sessionStorage.setItem("aegis_oauth_mode", "link");
      sessionStorage.setItem("aegis_oauth_provider", provider);

      const res = await authApi.getOAuthUrl(provider, state);
      if (res.data?.url) {
        window.location.href = res.data.url;
      } else {
        toast.error(`Unable to initialize ${provider} connection`);
        setConnectingProvider(null);
      }
    } catch (err: any) {
      toast.error(err?.response?.data?.message || `Failed to start ${provider} connection`);
      setConnectingProvider(null);
    }
  };

  const accounts = data?.accounts || [];
  const hasPassword = data?.has_password ?? true;
  const connectedCount = accounts.filter((a) => a.connected).length;

  const providerConfigs = [
    {
      id: "google",
      name: "Google",
      icon: <GoogleIcon className="w-5 h-5 shrink-0" />,
      description: "Sign in with your Google Workspace or personal account.",
    },
    {
      id: "github",
      name: "GitHub",
      icon: <GitHubIcon className="w-5 h-5 shrink-0 text-white" />,
      description: "Sign in with your GitHub developer identity.",
    },
  ];

  return (
    <SettingsLayout
      title="Connected accounts"
      subTitle="Link third-party identities to your Aegis profile for single-click authentication."
    >
      <div className="space-y-6">
        {/* {!hasPassword && connectedCount <= 1 && (
          <div className="flex items-start gap-3 p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 text-xs leading-relaxed">
            <WarningCircleIcon className="w-5 h-5 shrink-0 text-amber-400 mt-0.5" />
            <div className="space-y-1">
              <span className="font-semibold block text-amber-200">
                Password not set
              </span>
              <p className="text-amber-300/80">
                You currently sign in via third-party OAuth. To prevent losing access to your account, you cannot disconnect your last connected provider unless you set a password first.
              </p>
              <button
                onClick={() =>
                  navigate({
                    to: "/settings",
                    search: { tab: "security" },
                  })
                }
                className="inline-block mt-1 font-medium underline text-amber-200 hover:text-white"
              >
                Set a password in Security & sign-in &rarr;
              </button>
            </div>
          </div>
        )} */}

        <Card hover={false} className="p-0 divide-y divide-surface-border/60 overflow-hidden">
          {providerConfigs.map((cfg) => {
            const acc = accounts.find((a) => a.provider.toLowerCase() === cfg.id);
            const isConnected = !!acc?.connected;
            const canDisconnect = hasPassword || connectedCount > 1;

            return (
              <div
                key={cfg.id}
                className="flex flex-col sm:flex-row sm:items-center justify-between p-5 gap-4 hover:bg-zinc-900/30 transition-colors"
              >
                <div className="flex items-start sm:items-center gap-4">
                  <div className="w-11 h-11 rounded-xl bg-zinc-900 border border-zinc-800 flex items-center justify-center shrink-0 shadow-sm">
                    {cfg.icon}
                  </div>
                  <div className="space-y-1 min-w-0">
                    <div className="flex items-center gap-2.5">
                      <h4 className="text-sm font-medium text-white">{cfg.name}</h4>
                      {isConnected ? (
                        <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          Connected
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-zinc-800 text-zinc-400 border border-zinc-700/50">
                          Not connected
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-zinc-400 line-clamp-1">
                      {isConnected && acc?.email ? (
                        <span className="text-zinc-300 font-mono text-[11px]">{acc.email}</span>
                      ) : (
                        cfg.description
                      )}
                    </p>
                  </div>
                </div>

                <div className="flex items-center justify-end gap-2.5 shrink-0 self-end sm:self-center">
                  {isConnected ? (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => setDisconnectingProvider(cfg.id)}
                      className="text-xs text-zinc-400 hover:text-red-400 hover:border-red-500/30 hover:bg-red-500/5 rounded-full flex items-center gap-1.5"
                    >
                      <LinkBreakIcon size={14} />
                      Disconnect
                    </Button>
                  ) : (
                    <Button
                      size="sm"
                      onClick={() => handleConnect(cfg.id as "google" | "github")}
                      disabled={connectingProvider === cfg.id}
                      className="text-xs rounded-full flex items-center gap-1.5 font-medium"
                    >
                      <PlusIcon size={14} />
                      {connectingProvider === cfg.id ? "Connecting..." : "Connect"}
                    </Button>
                  )}
                </div>
              </div>
            );
          })}
        </Card>
      </div>

      {/* Disconnect Confirmation Dialog */}
      <Dialog
        open={!!disconnectingProvider}
        onOpenChange={(open) => !open && setDisconnectingProvider(null)}
        title={`Disconnect ${
          disconnectingProvider
            ? disconnectingProvider.charAt(0).toUpperCase() + disconnectingProvider.slice(1)
            : ""
        }`}
      >
        <div className="space-y-4 pt-2">
          {!hasPassword && connectedCount <= 1 ? (
            <div className="space-y-3">
              <div className="p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs leading-relaxed">
                <span className="font-semibold block text-red-200 mb-1">
                  Cannot disconnect this account
                </span>
                This is currently your only sign-in method and you have not configured a password.
                Disconnecting it would lock you out of your Aegis account.
              </div>
              <p className="text-xs text-zinc-400">
                Please visit the{" "}
                <button
                  onClick={() => {
                    setDisconnectingProvider(null);
                    navigate({ to: "/settings", search: { tab: "security" } });
                  }}
                  className="text-white underline font-medium"
                >
                  Security & sign-in
                </button>{" "}
                tab to set a password first.
              </p>
              <div className="flex justify-end pt-2">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setDisconnectingProvider(null)}
                  className="rounded-full text-xs"
                >
                  Understood
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-xs text-zinc-300 leading-relaxed">
                Are you sure you want to disconnect this {disconnectingProvider} account? You will
                no longer be able to use it to sign in to Aegis unless you reconnect it.
              </p>
              <div className="flex justify-end items-center gap-2.5 pt-2">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setDisconnectingProvider(null)}
                  className="rounded-full text-xs"
                >
                  Cancel
                </Button>
                <Button
                  variant="danger"
                  size="sm"
                  disabled={unlinkMutation.isPending}
                  onClick={() => {
                    if (disconnectingProvider) {
                      unlinkMutation.mutate(disconnectingProvider);
                    }
                  }}
                  className="rounded-full text-xs"
                >
                  {unlinkMutation.isPending ? "Disconnecting..." : "Disconnect"}
                </Button>
              </div>
            </div>
          )}
        </div>
      </Dialog>
    </SettingsLayout>
  );
};

export default ConnectedAccountsTab;
