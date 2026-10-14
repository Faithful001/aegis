import React, { useState, useMemo, useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";
import { MainLayout } from "../../components/layout/MainLayout";
import { Card } from "../../components/ui/Card";
import { Badge } from "../../components/ui/Badge";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import {
  Key,
  Trash2,
  User,
  Blocks,
  Cpu,
  Search,
  ExternalLink,
  LogOut,
  Mail,
  Shield,
  Calendar,
  X,
  Check,
  RefreshCw,
  ShieldCheck,
  AlertCircle,
  Lock,
  CheckCircle2,
} from "lucide-react";
import { useAegis } from "../../context/AegisContext";
import { useAuth } from "../../context/AuthContext";
import { credentialsApi } from "../../api/credentials";
import { authApi } from "../../api/auth";
import { toast } from "sonner";

/* ------------------------------------------------------------------ */
/*  Provider definitions with branding                                  */
/* ------------------------------------------------------------------ */
const PROVIDERS = [
  {
    id: "anthropic",
    name: "Anthropic",
    icon: "A",
    iconStyle: "bg-[#CC785C]/15 text-[#CC785C] font-bold",
    desc: "Claude models via api.anthropic.com",
    placeholder: "sk-ant-...",
    keyLink: "https://console.anthropic.com/account/keys",
  },
  {
    id: "openai",
    name: "OpenAI",
    icon: "◎",
    iconStyle: "bg-emerald-500/10 text-emerald-400",
    desc: "GPT models via api.openai.com",
    placeholder: "sk-proj-...",
    keyLink: "https://platform.openai.com/api-keys",
  },
  {
    id: "gemini",
    name: "Google Gemini",
    icon: "✦",
    iconStyle: "bg-blue-500/10 text-blue-400",
    desc: "Gemini models via Google AI Studio",
    placeholder: "AIzaSy...",
    keyLink: "https://aistudio.google.com/app/apikey",
  },
  {
    id: "mistral",
    name: "Mistral",
    icon: "M",
    iconStyle: "bg-orange-500/10 text-orange-400 font-bold",
    desc: "Mistral models via api.mistral.ai",
    placeholder: "your_mistral_api_key",
    keyLink: "https://console.mistral.ai/api-keys/",
  },
  {
    id: "openrouter",
    name: "OpenRouter",
    icon: "⊙",
    iconStyle: "bg-purple-500/10 text-purple-400",
    desc: "Universal router for 200+ models",
    placeholder: "sk-or-...",
    keyLink: "https://openrouter.ai/keys",
  },
] as const;

/* ------------------------------------------------------------------ */
/*  Settings nav tabs                                                   */
/* ------------------------------------------------------------------ */
type SettingsTab = "profile" | "providers" | "models";

const NAV_ITEMS: { id: SettingsTab; label: string; icon: React.ReactNode }[] = [
  { id: "profile", label: "Profile", icon: <User className="w-4 h-4" /> },
  { id: "providers", label: "Providers", icon: <Blocks className="w-4 h-4" /> },
  { id: "models", label: "Models", icon: <Cpu className="w-4 h-4" /> },
];

/* ================================================================== */
/*  Root Component                                                      */
/* ================================================================== */
export const SettingsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<SettingsTab>("profile");

  return (
    <MainLayout>
      <div className="flex min-h-full" style={{ height: "calc(100vh - 0px)" }}>
        {/* ---- Settings Sidebar ---- */}
        <nav className="w-56 shrink-0 border-r border-surface-border bg-surface/40 p-4 space-y-6">
          <div className="space-y-1">
            <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-wider px-3">
              Personal
            </span>
            {NAV_ITEMS.map((item) => (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`w-full flex items-center gap-2.5 px-3 py-2 rounded-xl text-sm font-medium transition-all ${
                  activeTab === item.id
                    ? "bg-zinc-800 text-white"
                    : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40"
                }`}
              >
                {item.icon}
                {item.label}
              </button>
            ))}
          </div>
        </nav>

        {/* ---- Content Area ---- */}
        <div className="flex-1 overflow-y-auto">
          <div className="max-w-3xl mx-auto p-8">
            {activeTab === "profile" && <ProfileTab />}
            {activeTab === "providers" && <ProvidersTab />}
            {activeTab === "models" && <ModelsTab />}
          </div>
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

  // Email verification state
  const [showVerifyInput, setShowVerifyInput] = useState(false);
  const [verificationCode, setVerificationCode] = useState("");
  const [sendingVerification, setSendingVerification] = useState(false);
  const [verifyingEmail, setVerifyingEmail] = useState(false);
  const [verifyCooldown, setVerifyCooldown] = useState(0);

  // Change password state
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [changingPassword, setChangingPassword] = useState(false);

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

  const handleSendVerificationEmail = async () => {
    setSendingVerification(true);
    try {
      await authApi.sendVerificationEmail();
      toast.success("6-digit verification code sent to your email!");
      setShowVerifyInput(true);
      setVerifyCooldown(60);
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to send verification email");
    } finally {
      setSendingVerification(false);
    }
  };

  const handleConfirmVerification = async (e: React.FormEvent) => {
    e.preventDefault();
    if (verificationCode.length !== 6) {
      toast.error("Please enter a valid 6-digit code");
      return;
    }
    setVerifyingEmail(true);
    try {
      await authApi.verifyEmail(verificationCode.trim());
      toast.success("Email verified successfully!");
      setShowVerifyInput(false);
      setVerificationCode("");
      await refreshUser();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Invalid or expired code");
    } finally {
      setVerifyingEmail(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword.length < 8) {
      toast.error("New password must be at least 8 characters");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("New passwords do not match");
      return;
    }
    setChangingPassword(true);
    try {
      await authApi.changePassword(currentPassword, newPassword);
      toast.success("Password changed successfully!");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to change password. Please check your current password.");
    } finally {
      setChangingPassword(false);
    }
  };

  const initials = user?.name
    ? user.name.slice(0, 2).toUpperCase()
    : user?.email
      ? user.email.slice(0, 2).toUpperCase()
      : "AE";

  return (
    <div className="space-y-8">
      {/* Avatar + Name */}
      <div className="flex flex-col items-center text-center space-y-3 pt-4">
        <div className="w-20 h-20 rounded-full bg-zinc-800 border-2 border-amber-500/40 flex items-center justify-center text-zinc-100 text-2xl font-bold shadow-glow-yellow">
          {initials}
        </div>
        <div>
          <h2 className="text-lg font-bold text-white">{user?.name || "Aegis User"}</h2>
          <p className="text-xs text-zinc-400 flex items-center justify-center gap-1.5 mt-0.5">
            <Mail className="w-3.5 h-3.5 text-zinc-500" /> {user?.email || "user@aegis.dev"}
          </p>
        </div>
      </div>

      {/* Email Verification Status Card */}
      <Card hover={false} className="p-5 space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <div className="flex items-center gap-2">
              <h3 className="text-sm font-semibold text-white">Email Verification</h3>
              {user?.is_verified ? (
                <span className="inline-flex items-center gap-1 text-[11px] font-medium text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">
                  <CheckCircle2 className="w-3 h-3" /> Verified
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 text-[11px] font-medium text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded-full border border-amber-500/20">
                  <AlertCircle className="w-3 h-3" /> Unverified
                </span>
              )}
            </div>
            <p className="text-xs text-zinc-400">
              {user?.is_verified
                ? "Your email address is verified and active."
                : "Verify your email address to secure your account and receive alerts."}
            </p>
          </div>

          {!user?.is_verified && !showVerifyInput && (
            <Button
              onClick={handleSendVerificationEmail}
              disabled={sendingVerification || verifyCooldown > 0}
              className="text-xs py-1.5 px-3 rounded-full"
            >
              {sendingVerification
                ? "Sending..."
                : verifyCooldown > 0
                ? `Resend in ${verifyCooldown}s`
                : "Verify Email"}
            </Button>
          )}
        </div>

        {!user?.is_verified && showVerifyInput && (
          <form onSubmit={handleConfirmVerification} className="pt-2 border-t border-zinc-800 space-y-3">
            <div className="flex items-end gap-3">
              <div className="flex-1">
                <Input
                  label="6-Digit Verification Code"
                  type="text"
                  placeholder="123456"
                  className="rounded-full text-center tracking-[0.4em] font-mono font-bold"
                  value={verificationCode}
                  onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                  required
                  maxLength={6}
                />
              </div>
              <Button
                type="submit"
                disabled={verifyingEmail || verificationCode.length !== 6}
                className="py-2.5 px-4 rounded-full font-semibold"
              >
                {verifyingEmail ? "Verifying..." : "Confirm"}
              </Button>
            </div>
            <div className="flex items-center justify-between text-xs text-zinc-400">
              <button
                type="button"
                onClick={() => setShowVerifyInput(false)}
                className="hover:text-white transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleSendVerificationEmail}
                disabled={verifyCooldown > 0 || sendingVerification}
                className={`font-medium ${
                  verifyCooldown > 0 ? "text-zinc-600 cursor-not-allowed" : "text-amber-400 hover:text-amber-300"
                }`}
              >
                {verifyCooldown > 0 ? `Resend code in ${verifyCooldown}s` : "Resend code"}
              </button>
            </div>
          </form>
        )}
      </Card>

      {/* Change Password Card */}
      <Card hover={false} className="p-5 space-y-4">
        <div className="space-y-0.5">
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">
            <Lock className="w-4 h-4 text-amber-400" /> Security & Password
          </h3>
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
              disabled={changingPassword || !currentPassword || !newPassword || !confirmPassword}
              className="rounded-full px-5 py-2 font-semibold text-xs"
            >
              {changingPassword ? "Updating..." : "Update Password"}
            </Button>
          </div>
        </form>
      </Card>

      {/* Usage Limits */}
      <Card hover={false} className="p-5 space-y-3">
        <h3 className="text-xs font-bold text-zinc-300 uppercase tracking-wider">Usage limits</h3>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-semibold text-white">Weekly usage limit</p>
            <p className="text-[11px] text-zinc-500">Resets every Monday at 00:00 UTC</p>
          </div>
          <div className="flex items-center gap-3">
            <div className="w-32 h-2 rounded-full bg-zinc-800 overflow-hidden">
              <div className="h-full bg-zinc-300 rounded-full" style={{ width: "100%" }} />
            </div>
            <span className="text-xs font-medium text-zinc-300 whitespace-nowrap">100% left</span>
          </div>
        </div>
      </Card>

      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {[
          { value: "0", label: "Lifetime tokens" },
          { value: "0", label: "Active days" },
          { value: "BYOK", label: "Auth mode" },
          { value: "JWT", label: "Session type" },
        ].map((stat) => (
          <div
            key={stat.label}
            className="p-4 bg-zinc-900/60 rounded-xl border border-zinc-800 text-center space-y-1"
          >
            <p className="text-lg font-bold text-zinc-100">{stat.value}</p>
            <p className="text-[10px] font-medium text-zinc-500 uppercase tracking-wider">
              {stat.label}
            </p>
          </div>
        ))}
      </div>

      {/* Account Details */}
      <Card hover={false} className="p-5 space-y-4">
        <h3 className="text-xs font-bold text-zinc-300 uppercase tracking-wider">Account</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="p-4 bg-zinc-900/60 rounded-xl border border-zinc-800 space-y-1">
            <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">
              Role
            </span>
            <p className="text-sm font-semibold text-amber-400 flex items-center gap-1.5">
              <Shield className="w-4 h-4 text-amber-400" /> Developer
            </p>
          </div>
          <div className="p-4 bg-zinc-900/60 rounded-xl border border-zinc-800 space-y-1">
            <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">
              Member Since
            </span>
            <p className="text-sm font-semibold text-zinc-200 flex items-center gap-1.5">
              <Calendar className="w-4 h-4 text-zinc-400" /> September 2026
            </p>
          </div>
        </div>
      </Card>

      {/* Sign Out */}
      <Card hover={false} className="p-5">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-semibold text-white">Sign out</p>
            <p className="text-[11px] text-zinc-500">
              Aegis keeps your encrypted BYOK keys safe in your account.
            </p>
          </div>
          <Button variant="danger" onClick={handleLogout} className="flex items-center gap-2">
            <LogOut className="w-4 h-4" /> Sign out
          </Button>
        </div>
      </Card>
    </div>
  );
};

/* ================================================================== */
/*  Providers Tab                                                       */
/* ================================================================== */
const ProvidersTab: React.FC = () => {
  const { credentials, refreshCredentials, hasCredential, refreshModels } = useAegis();
  const [configuringProvider, setConfiguringProvider] = useState<string | null>(null);
  const [apiKeyInput, setApiKeyInput] = useState("");
  const [baseUrlInput, setBaseUrlInput] = useState("");
  const [saving, setSaving] = useState(false);
  const [deletingProvider, setDeletingProvider] = useState<string | null>(null);

  const handleSave = async (providerId: string) => {
    if (!apiKeyInput.trim()) return;
    setSaving(true);
    try {
      await credentialsApi.save(providerId, apiKeyInput.trim(), baseUrlInput.trim() || undefined);
      const provider = PROVIDERS.find((p) => p.id === providerId);
      toast.success(`${provider?.name || providerId} credential saved! Discovering models...`);
      setApiKeyInput("");
      setBaseUrlInput("");
      setConfiguringProvider(null);
      await refreshCredentials();
      await refreshModels();
    } catch (err: any) {
      toast.error(
        err?.response?.data?.error || err?.response?.data?.message || "Failed to save credential"
      );
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (providerId: string) => {
    setDeletingProvider(providerId);
    try {
      await credentialsApi.delete(providerId);
      const provider = PROVIDERS.find((p) => p.id === providerId);
      toast.success(`${provider?.name || providerId} credential removed`);
      await refreshCredentials();
      await refreshModels();
    } catch (err: any) {
      toast.error(err?.response?.data?.error || "Failed to remove credential");
    } finally {
      setDeletingProvider(null);
    }
  };

  const getCredentialForProvider = (providerId: string) =>
    credentials.find((c) => c.provider.toLowerCase() === providerId.toLowerCase());

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white tracking-tight">Providers</h1>
        <p className="text-sm text-zinc-400 mt-1">
          Bring your own key. Connect any model provider with an API key.
        </p>
      </div>

      <Card hover={false} className="p-0 divide-y divide-surface-border/60 overflow-hidden">
        {PROVIDERS.map((provider) => {
          const cred = getCredentialForProvider(provider.id);
          const isConfigured = !!cred;
          const isOpen = configuringProvider === provider.id;

          return (
            <div key={provider.id}>
              {/* Provider row */}
              <div className="flex items-center justify-between px-5 py-4 hover:bg-surface-hover/50 transition-colors">
                <div className="flex items-center gap-3">
                  <div
                    className={`w-9 h-9 rounded-xl flex items-center justify-center text-sm ${provider.iconStyle}`}
                  >
                    {provider.icon}
                  </div>
                  <div>
                    <span className="text-sm font-semibold text-zinc-100">{provider.name}</span>
                    {isConfigured && (
                      <p className="text-[11px] text-zinc-500 font-mono mt-0.5">
                        {cred?.masked_api_key || cred?.api_key_masked || "••••••••"}
                      </p>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {isConfigured && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDelete(provider.id)}
                      disabled={deletingProvider === provider.id}
                      className="text-zinc-500 hover:text-rose-400 hover:bg-rose-500/10"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  )}
                  <Button
                    variant={isConfigured ? "ghost" : "secondary"}
                    size="sm"
                    onClick={() => {
                      if (isOpen) {
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

              {/* Inline configure panel */}
              {isOpen && (
                <div className="px-5 pb-5 pt-1 bg-zinc-900/40 border-t border-surface-border/40">
                  <div className="max-w-md space-y-4">
                    <div className="flex items-center gap-2.5">
                      <div
                        className={`w-7 h-7 rounded-lg flex items-center justify-center text-xs ${provider.iconStyle}`}
                      >
                        {provider.icon}
                      </div>
                      <div>
                        <p className="text-sm font-semibold text-zinc-100">{provider.name}</p>
                        <p className="text-[11px] text-zinc-500">{provider.desc}</p>
                      </div>
                    </div>

                    <Input
                      type="password"
                      placeholder={provider.placeholder}
                      value={apiKeyInput}
                      onChange={(e) => setApiKeyInput(e.target.value)}
                      showToggle
                    />

                    <a
                      href={provider.keyLink}
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
                        disabled={saving || !apiKeyInput.trim()}
                      >
                        {saving ? "Saving..." : "Save"}
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </Card>
    </div>
  );
};

/* ================================================================== */
/*  Models Tab                                                          */
/* ================================================================== */
const ModelsTab: React.FC = () => {
  const { hasCredential, selectedModel, setSelectedModel, models, refreshModels, isLoadingModels } =
    useAegis();
  const [searchQuery, setSearchQuery] = useState("");
  const [enabledModels, setEnabledModels] = useState<Set<string>>(() => {
    return new Set(models.map((m) => m.id));
  });

  // Keep enabled models synced as new models are discovered
  useEffect(() => {
    setEnabledModels((prev) => {
      const next = new Set(prev);
      models.forEach((m) => {
        if (!prev.has(m.id)) {
          next.add(m.id);
        }
      });
      return next;
    });
  }, [models]);

  const filteredModels = useMemo(() => {
    if (!searchQuery.trim()) return models;
    const q = searchQuery.toLowerCase();
    return models.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.id.toLowerCase().includes(q) ||
        m.provider.toLowerCase().includes(q) ||
        (m.description && m.description.toLowerCase().includes(q))
    );
  }, [searchQuery, models]);

  const toggleModel = (modelId: string) => {
    setEnabledModels((prev) => {
      const next = new Set(prev);
      if (next.has(modelId)) {
        next.delete(modelId);
        if (selectedModel === modelId && next.size > 0) {
          setSelectedModel(Array.from(next)[0]);
        }
      } else {
        next.add(modelId);
      }
      return next;
    });
  };

  const handleScan = async () => {
    await refreshModels();
    toast.success("Models scanned and updated from providers");
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Models</h1>
          <p className="text-sm text-zinc-400 mt-1">
            Dynamically discovered models from your configured upstream provider API keys.
          </p>
        </div>
        <Button
          variant="secondary"
          size="sm"
          onClick={handleScan}
          disabled={isLoadingModels}
          className="gap-2 shrink-0"
        >
          <RefreshCw
            className={`w-3.5 h-3.5 ${isLoadingModels ? "animate-spin text-sky-400" : ""}`}
          />
          {isLoadingModels ? "Scanning..." : "Scan Providers"}
        </Button>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
        <input
          type="text"
          placeholder="Search discovered models by name, ID, or provider..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2.5 text-sm bg-surface-card border border-surface-border text-zinc-100 rounded-xl placeholder:text-zinc-500 focus:outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500 transition-all duration-150"
        />
      </div>

      {/* Model list */}
      <Card hover={false} className="p-0 divide-y divide-surface-border/60 overflow-hidden">
        {filteredModels.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-sm text-zinc-500">No models match your search.</p>
          </div>
        ) : (
          filteredModels.map((model) => {
            const providerConfigured = hasCredential(model.provider);
            const isEnabled = enabledModels.has(model.id);

            return (
              <div
                key={model.id}
                className="flex items-center justify-between px-5 py-4 hover:bg-surface-hover/50 transition-colors"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-9 h-9 rounded-xl bg-zinc-800 border border-zinc-700/60 flex items-center justify-center flex-shrink-0">
                    <Cpu className="w-4 h-4 text-zinc-400" />
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-semibold text-zinc-100 truncate">
                        {model.name}
                      </span>
                      <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-zinc-800 border border-zinc-700/60 text-zinc-400">
                        {model.provider}
                      </span>
                      {model.fast && <Badge variant="info">Fast</Badge>}
                      {!providerConfigured && (
                        <span className="text-[10px] text-amber-500/90 font-medium">
                          Requires {model.provider} key
                        </span>
                      )}
                    </div>
                    <p className="text-[11px] text-zinc-500 truncate mt-0.5">
                      {model.id} {model.description ? `· ${model.description}` : ""}
                    </p>
                  </div>
                </div>

                {/* Toggle switch */}
                <button
                  onClick={() => toggleModel(model.id)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 flex-shrink-0 ml-4 ${
                    isEnabled && providerConfigured
                      ? "bg-blue-500"
                      : isEnabled && !providerConfigured
                        ? "bg-zinc-600"
                        : "bg-zinc-700"
                  }`}
                  aria-label={`Toggle ${model.name}`}
                >
                  <span
                    className={`inline-block h-4.5 w-4.5 transform rounded-full bg-white shadow-sm transition-transform duration-200 ${
                      isEnabled ? "translate-x-5.5" : "translate-x-1"
                    }`}
                    style={{
                      width: "18px",
                      height: "18px",
                      transform: isEnabled ? "translateX(22px)" : "translateX(4px)",
                    }}
                  />
                </button>
              </div>
            );
          })
        )}
      </Card>
    </div>
  );
};

export default SettingsPage;
