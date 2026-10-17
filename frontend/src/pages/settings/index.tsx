import React, { useState, useMemo, useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { MainLayout } from "../../components/layout/MainLayout";
import { Card } from "../../components/ui/Card";
import { Badge } from "../../components/ui/Badge";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import {
  Trash2,
  User,
  Blocks,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  LogOut,
  Check,
  RefreshCw,
  Cpu,
  ChevronLeft,
  ChevronRight,
  Search,
  Lock,
} from "lucide-react";
import { useAegis } from "../../context/AegisContext";
import { useAuth } from "../../context/AuthContext";
import { credentialsApi } from "../../api/credentials";
import { authApi } from "../../api/auth";
import { toast } from "sonner";
import { DiscoveredModel, PaginatedResult } from "../../types/api";

type SettingsTab = "profile" | "providers";

const NAV_ITEMS: { id: SettingsTab; label: string; icon: React.ReactNode }[] = [
  { id: "profile", label: "Profile", icon: <User className="w-4 h-4" /> },
  { id: "providers", label: "Providers", icon: <Blocks className="w-4 h-4" /> },
];

export const SettingsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<SettingsTab>("profile");

  return (
    <MainLayout>
      <div className="relative flex-1 overflow-y-auto h-full">
        {/* ---- Floating Tab Switcher ---- */}
        <div className="sticky top-6 z-20 flex justify-center pointer-events-none">
          <nav className="pointer-events-auto inline-flex items-center gap-1 p-1 rounded-full bg-zinc-900/80 border border-zinc-700/50 shadow-xl backdrop-blur-md">
            {NAV_ITEMS.map((item) => (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`flex items-center gap-2 px-4 py-1.5 rounded-full text-sm font-medium transition-all ${
                  activeTab === item.id
                    ? "bg-zinc-700 text-white shadow-sm"
                    : "text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60"
                }`}
              >
                {item.icon}
                {item.label}
              </button>
            ))}
          </nav>
        </div>

        {/* ---- Content Area ---- */}
        <div className="max-w-3xl mx-auto px-8 pt-6 pb-12">
          {activeTab === "profile" && <ProfileTab />}
          {activeTab === "providers" && <ProvidersTab />}
        </div>
      </div>
    </MainLayout>
  );
};

/* ================================================================== */
/*  Profile Tab                                                         */
/* ================================================================== */
const ProfileTab: React.FC = () => {
  const { user, logout, refreshUser } = useAuth();
  const navigate = useNavigate();

  const [verificationCode, setVerificationCode] = useState("");
  const [verifyCooldown, setVerifyCooldown] = useState(0);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  useEffect(() => {
    if (verifyCooldown <= 0) return;
    const interval = setInterval(() => {
      setVerifyCooldown((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(interval);
  }, [verifyCooldown]);

  const handleLogout = async () => {
    await logout();
    toast.success("Logged out successfully");
    navigate({ to: "/auth/login" });
  };

  const confirmVerificationMutation = useMutation({
    mutationFn: async (code: string) =>
      await authApi.verifyEmail({ code, email: user?.email ?? "" }),
    onSuccess: async () => {
      toast.success("Email verified successfully!");
      setVerificationCode("");
      await refreshUser();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Invalid or expired code");
    },
  });

  const changePasswordMutation = useMutation({
    mutationFn: async ({ currentPass, newPass }: { currentPass: string; newPass: string }) =>
      await authApi.changePassword(currentPass, newPass),
    onSuccess: () => {
      toast.success("Password changed successfully!");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    },
    onError: (err: any) => {
      toast.error(
        err?.response?.data?.message ||
          "Failed to change password. Please check your current password."
      );
    },
  });

  const handleChangePassword = (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword.length < 8) {
      toast.error("New password must be at least 8 characters");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("New passwords do not match");
      return;
    }
    if (changePasswordMutation.isPending) return;
    changePasswordMutation.mutate({ currentPass: currentPassword, newPass: newPassword });
  };

  const initials =
    user?.first_name && user?.last_name
      ? user.first_name.slice(0, 1).toUpperCase() + user.last_name.slice(0, 1).toUpperCase()
      : user?.email
        ? user.email.slice(0, 2).toUpperCase()
        : "AE";

  return (
    <div className="space-y-8">
      <div className="flex flex-col items-center text-center space-y-3 pt-4">
        <div className="w-20 h-20 rounded-full bg-[#27AF61] flex items-center justify-center text-zinc-100 text-3xl">
          {initials}
        </div>
        <div>
          <h2 className="text-lg font-bold text-white">
            {user?.first_name + " " + user?.last_name || "Aegis User"}
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">{user?.email || "user@aegis.dev"}</p>
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {[
          { value: "0", label: "Lifetime tokens" },
          { value: "0", label: "Active days" },
        ].map((stat) => (
          <div
            key={stat.label}
            className="p-4 bg-zinc-900/60 rounded-xl border border-zinc-800 text-center space-y-1"
          >
            <p className="text-lg font-semibold text-zinc-100">{stat.value}</p>
            <p className="text-[10px] font-medium text-zinc-500 uppercase tracking-wider">
              {stat.label}
            </p>
          </div>
        ))}
      </div>

      <Card hover={false} className="p-5 space-y-4 !bg-sidebar">
        <div className="space-y-0.5">
          <h3 className="text-sm font-semibold text-white">Security & Password</h3>
          <p className="text-xs text-zinc-400">
            Update your account password. Choose a strong password with at least 8 characters.
          </p>
        </div>
        <form onSubmit={handleChangePassword} className="space-y-3 pt-2">
          <Input
            label="Current Password"
            type="password"
            placeholder="••••••••"
            className="rounded-full"
            showToggle
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            required
          />
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Input
              label="New Password"
              type="password"
              placeholder="•••••••• (min 8 chars)"
              className="rounded-full"
              showToggle
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
            />
            <Input
              label="Confirm New Password"
              type="password"
              placeholder="••••••••"
              className="rounded-full"
              showToggle
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
          </div>
          <div className="flex justify-end pt-1">
            <Button
              type="submit"
              disabled={
                changePasswordMutation.isPending ||
                !currentPassword ||
                !newPassword ||
                !confirmPassword
              }
              className="rounded-full px-5 py-2 font-semibold text-xs"
            >
              {changePasswordMutation.isPending ? "Updating..." : "Update Password"}
            </Button>
          </div>
        </form>
      </Card>

      <Button
        onClick={handleLogout}
        className="bg-white rounded-full text-[13px] flex items-center gap-2 hover:bg-white/70"
      >
        <LogOut className="w-4 h-4" /> Sign out
      </Button>
    </div>
  );
};

/* ================================================================== */
/*  Provider Model Drawer (paginated, lazy-fetched)                     */
/* ================================================================== */
const MODELS_PER_PAGE = 10;

const ProviderModelDrawer: React.FC<{ providerId: string; providerName: string }> = ({
  providerId,
  providerName,
}) => {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [togglingId, setTogglingId] = useState<string | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1);
    }, 250);
    return () => clearTimeout(timer);
  }, [search]);

  const { data, isLoading, isFetching } = useQuery<PaginatedResult<DiscoveredModel>>({
    queryKey: ["provider-models", providerId, page, debouncedSearch],
    queryFn: () => credentialsApi.listModels(providerId, page, MODELS_PER_PAGE, debouncedSearch),
    staleTime: 1000 * 60 * 5,
  });

  const toggleMutation = useMutation({
    mutationFn: async ({ modelId, currentActive }: { modelId: string; currentActive: boolean }) => {
      setTogglingId(modelId);
      return credentialsApi.toggleModel(providerId, modelId, !currentActive);
    },
    onSuccess: (updatedModel) => {
      toast.success(
        `${updatedModel.name || updatedModel.id} is now ${updatedModel.is_active ? "active" : "inactive"}`
      );
      queryClient.invalidateQueries({ queryKey: ["provider-models", providerId] });
      queryClient.invalidateQueries({ queryKey: ["credentials", "models"] });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to toggle model status");
    },
    onSettled: () => {
      setTogglingId(null);
    },
  });

  const models = data?.items || [];
  const total = data?.total || 0;
  const totalPages = Math.max(1, data?.total_pages || 1);

  return (
    <div className="px-5 pb-4 pt-3 space-y-3">
      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-500" />
        <input
          type="text"
          placeholder="Filter models…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-9 pr-8 py-1.5 text-xs bg-zinc-900 border border-zinc-700/60 text-zinc-200 rounded-lg placeholder:text-zinc-600 focus:outline-none focus:border-zinc-500 transition-all"
        />
        {/* {(isFetching || isLoading) && (
          <RefreshCw className="absolute right-3 top-1/2 -translate-y-1/2 w-3 h-3 text-zinc-500 animate-spin" /> 
        )} */}
      </div>

      {/* Model rows */}
      <>
        {isLoading ? (
          <div className="flex items-center gap-2 text-zinc-500 text-xs py-4 px-5">
            <RefreshCw className="w-3.5 h-3.5 animate-spin" /> Loading models…
          </div>
        ) : total === 0 && !debouncedSearch ? (
          <p className="text-xs text-zinc-500 italic px-5 py-4">
            No models found for {providerName}.
          </p>
        ) : (
          <div className="space-y-0.5">
            {models.length === 0 ? (
              <p className="text-xs text-zinc-500 italic py-2">No models match your search.</p>
            ) : (
              models.map((model) => (
                <div
                  key={model.id}
                  className="flex items-center justify-between px-3 py-2 rounded-lg hover:bg-zinc-800/40 transition-colors"
                >
                  <div className="min-w-0">
                    <span className="text-xs font-medium text-zinc-200 truncate block">
                      {model.name}
                    </span>
                    <span className="text-[10px] text-zinc-500 font-mono truncate block">
                      {model.id}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-3">
                    {model.fast && <Badge variant="info">Fast</Badge>}
                    <button
                      type="button"
                      role="switch"
                      aria-checked={model.is_active}
                      onClick={() =>
                        toggleMutation.mutate({
                          modelId: model.id,
                          currentActive: !!model.is_active,
                        })
                      }
                      disabled={togglingId === model.id}
                      className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full p-0.5 transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500/50 ${
                        model.is_active
                          ? "bg-emerald-500 shadow-sm shadow-emerald-500/30"
                          : "bg-zinc-700/80 hover:bg-zinc-600/80"
                      } ${togglingId === model.id ? "opacity-60 cursor-not-allowed" : ""}`}
                      title={
                        model.is_active
                          ? "Model is active. Click to deactivate"
                          : "Model is inactive. Click to activate"
                      }
                    >
                      <span className="sr-only">Toggle model active status</span>
                      <span
                        className={`pointer-events-none flex items-center justify-center h-4 w-4 transform rounded-full bg-white shadow-md transition-transform duration-200 ease-in-out ${
                          model.is_active ? "translate-x-4" : "translate-x-0"
                        }`}
                      >
                        {togglingId === model.id && (
                          <RefreshCw className="w-2.5 h-2.5 text-zinc-600 animate-spin" />
                        )}
                      </span>
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between pt-1">
          <span className="text-[10px] text-zinc-500">
            {total} models · page {page}/{totalPages}
          </span>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1 || isFetching}
              className="p-1 rounded text-zinc-400 hover:text-zinc-100 disabled:opacity-30 hover:bg-zinc-800 transition-colors"
            >
              <ChevronLeft className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages || isFetching}
              className="p-1 rounded text-zinc-400 hover:text-zinc-100 disabled:opacity-30 hover:bg-zinc-800 transition-colors"
            >
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

/* ================================================================== */
/*  Providers & Models Tab                                              */
/* ================================================================== */
const ProvidersTab: React.FC = () => {
  const { supportedProviders, refreshCredentials } = useAegis();
  const providersList = Array.isArray(supportedProviders) ? supportedProviders : [];

  const [configuringProvider, setConfiguringProvider] = useState<string | null>(null);
  const [expandedModels, setExpandedModels] = useState<string | null>(null);
  const [apiKeyInput, setApiKeyInput] = useState("");
  const [baseUrlInput, setBaseUrlInput] = useState("");

  const saveCredentialMutation = useMutation({
    mutationFn: async ({
      providerId,
      apiKey,
      baseUrl,
    }: {
      providerId: string;
      apiKey: string;
      baseUrl?: string;
    }) => credentialsApi.save(providerId, apiKey, baseUrl),
    onSuccess: async (_, { providerId }) => {
      const provider = providersList.find((p) => p.id === providerId);
      toast.success(`${provider?.name || providerId} credential saved! Discovering models…`);
      setApiKeyInput("");
      setBaseUrlInput("");
      setConfiguringProvider(null);
      await refreshCredentials();
    },
    onError: (err: any) => {
      toast.error(
        err?.response?.data?.error || err?.response?.data?.message || "Failed to save credential"
      );
    },
  });

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

  const handleSave = (providerId: string) => {
    if (!apiKeyInput.trim() || saveCredentialMutation.isPending) return;
    saveCredentialMutation.mutate({
      providerId,
      apiKey: apiKeyInput.trim(),
      baseUrl: baseUrlInput.trim() || undefined,
    });
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-white tracking-tight">Providers</h1>
        <p className="text-xs text-zinc-400 mt-1">
          Bring your own key. Connect any model provider — then browse its available models.
        </p>
      </div>

      <Card hover={false} className="p-0 divide-y divide-surface-border/60 overflow-hidden">
        {providersList.map((provider) => {
          const isConfigured = provider.is_configured;
          const isConfigOpen = configuringProvider === provider.id;
          const isModelsOpen = expandedModels === provider.id;

          return (
            <div key={provider.id}>
              {/* ---- Provider row ---- */}
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
                  {/* Models toggle — only when configured */}
                  {isConfigured && (
                    <button
                      onClick={() =>
                        setExpandedModels((prev) => (prev === provider.id ? null : provider.id))
                      }
                      className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-100 px-2.5 py-1.5 rounded-lg hover:bg-zinc-800/60 transition-all"
                    >
                      <Cpu className="w-3.5 h-3.5" />
                      Models
                      {isModelsOpen ? (
                        <ChevronUp className="w-3 h-3" />
                      ) : (
                        <ChevronDown className="w-3 h-3" />
                      )}
                    </button>
                  )}

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

                  {/* Configure toggle */}
                  <Button
                    variant={isConfigured ? "ghost" : "secondary"}
                    size="sm"
                    onClick={() => {
                      if (isConfigOpen) {
                        setConfiguringProvider(null);
                        setApiKeyInput("");
                        setBaseUrlInput("");
                      } else {
                        setConfiguringProvider(provider.id);
                        setApiKeyInput("");
                        setBaseUrlInput("");
                      }
                    }}
                    className={
                      isConfigured ? "text-emerald-400 hover:text-emerald-300 gap-1.5" : "gap-1.5"
                    }
                  >
                    {isConfigured ? (
                      <>
                        <Check className="w-3.5 h-3.5" /> Configured
                      </>
                    ) : (
                      "Configure"
                    )}
                  </Button>
                </div>
              </div>

              {/* ---- Inline configure panel ---- */}
              {isConfigOpen && (
                <div className="px-5 pb-5 pt-3 bg-zinc-900/40 border-t border-surface-border/40">
                  <div className="max-w-md space-y-3">
                    <Input
                      type="password"
                      placeholder={provider.placeholder}
                      value={apiKeyInput}
                      onChange={(e) => setApiKeyInput(e.target.value)}
                      showToggle
                    />
                    <a
                      href={provider.key_link}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
                    >
                      Get API key <ExternalLink className="w-3 h-3" />
                    </a>
                    <div className="flex items-center gap-3 pt-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setConfiguringProvider(null);
                          setApiKeyInput("");
                          setBaseUrlInput("");
                        }}
                      >
                        Cancel
                      </Button>
                      <Button
                        size="sm"
                        onClick={() => handleSave(provider.id)}
                        disabled={saveCredentialMutation.isPending || !apiKeyInput.trim()}
                      >
                        {saveCredentialMutation.isPending ? "Saving…" : "Save"}
                      </Button>
                    </div>
                  </div>
                </div>
              )}

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
    </div>
  );
};

export default SettingsPage;
