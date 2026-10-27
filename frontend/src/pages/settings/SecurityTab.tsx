import React from "react";
import SettingsLayout from "@/components/layout/SettingsLayout";
import { authApi } from "@/api/auth";
import { credentialsApi } from "@/api/credentials";
import { userApi } from "@/api/user";
import { useAuth } from "@/context/AuthContext";
import { ProviderCredential, UserSession } from "@/types/api";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Card } from "../../components/ui/Card";
import {
  DeviceMobileIcon,
  DownloadSimpleIcon,
  LaptopIcon,
  MonitorIcon,
  WarningIcon,
  SignOutIcon,
} from "@phosphor-icons/react";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Dialog } from "../../components/ui/Dialog";

const SecurityTab = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  // Budget Alerts state
  const [budgetEnabled, setBudgetEnabled] = useState(true);
  const [budgetUSD, setBudgetUSD] = useState("50.00");
  const [threshold, setThreshold] = useState("80");
  const [alertEmail, setAlertEmail] = useState("");

  // Danger zone state
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [confirmDeleteText, setConfirmDeleteText] = useState("");

  const { data: alertSettings, refetch: refetchAlerts } = useQuery({
    queryKey: ["user", "budget-alerts"],
    queryFn: userApi.getAlertSettings,
  });

  const {
    data: sessions = [],
    isLoading: isLoadingSessions,
    refetch: refetchSessions,
  } = useQuery<UserSession[]>({
    queryKey: ["auth", "sessions"],
    queryFn: authApi.listSessions,
  });

  useEffect(() => {
    if (alertSettings) {
      setBudgetEnabled(alertSettings.is_budget_alert_enabled);
      setBudgetUSD(
        alertSettings.monthly_budget_usd ? alertSettings.monthly_budget_usd.toString() : "50.00"
      );
      setThreshold(
        alertSettings.threshold_percentage ? alertSettings.threshold_percentage.toString() : "80"
      );
      setAlertEmail(alertSettings.alert_email || user?.email || "");
    } else if (user?.email && !alertEmail) {
      setAlertEmail(user.email);
    }
  }, [alertSettings, user]);

  const handleLogout = async () => {
    await logout();
    toast.success("Logged out successfully");
    navigate({ to: "/auth/login" });
  };

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

  const revokeOthersMutation = useMutation({
    mutationFn: authApi.revokeOtherSessions,
    onSuccess: () => {
      toast.success("Signed out of all other devices!");
      refetchSessions();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to revoke sessions");
    },
  });

  const revokeSessionMutation = useMutation({
    mutationFn: (id: string) => authApi.revokeSession(id),
    onSuccess: () => {
      toast.success("Session revoked");
      refetchSessions();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to revoke session");
    },
  });

  const exportDataMutation = useMutation({
    mutationFn: userApi.exportData,
    onSuccess: (data) => {
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `aegis-account-export-${new Date().toISOString().slice(0, 10)}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success("Account data exported successfully!");
    },
    onError: () => {
      toast.error("Failed to export account data");
    },
  });

  const deleteAccountMutation = useMutation({
    mutationFn: userApi.deleteAccount,
    onSuccess: async () => {
      toast.success("Your account has been deleted.");
      await logout();
      navigate({ to: "/auth/register" });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to delete account");
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

  const otherSessionsCount = sessions.filter((s) => !s.is_current).length;

  return (
    <SettingsLayout
      title="Security"
      subTitle="Manage your account's security preferences and active sessions."
    >
      <Card hover={false} className="p-5 space-y-4 !bg-sidebar">
        <div className="space-y-0.5">
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">Password</h3>
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

      <Card hover={false} className="p-5 space-y-4 !bg-sidebar">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div className="space-y-0.5">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              Active Sessions & Devices
            </h3>
            <p className="text-xs text-zinc-400">
              Devices and browsers currently logged into your Aegis account
            </p>
          </div>
          {otherSessionsCount > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => revokeOthersMutation.mutate()}
              disabled={revokeOthersMutation.isPending}
              className="rounded-full text-[11px] text-red-400 hover:text-red-400 border-red-500/30 hover:bg-red-500/10 hover:border-red-500/50 self-start sm:self-auto"
            >
              {revokeOthersMutation.isPending ? "Signing out..." : "Sign Out Other Devices"}
            </Button>
          )}
        </div>

        <div className="space-y-2 pt-1">
          {isLoadingSessions ? (
            <div className="p-4 text-center text-xs text-zinc-500">Loading active sessions...</div>
          ) : sessions.length === 0 ? (
            <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/40 text-center text-xs text-zinc-400">
              Current session active
            </div>
          ) : (
            sessions.map((sess) => {
              const isDesktop =
                sess.device_name.toLowerCase().includes("windows") ||
                sess.device_name.toLowerCase().includes("macos") ||
                sess.device_name.toLowerCase().includes("linux");
              const isMobile =
                sess.device_name.toLowerCase().includes("iphone") ||
                sess.device_name.toLowerCase().includes("android") ||
                sess.device_name.toLowerCase().includes("ipad");

              return (
                <div
                  key={sess.id}
                  className="flex items-center justify-between py-2.5 border-b border-zinc-800/80 last:border-0 transition-colors"
                >
                  <div className="flex items-center gap-3 w-full">
                    <div className="w-8 h-8 rounded-lg bg-zinc-800 flex items-center justify-center text-zinc-300">
                      {isMobile ? (
                        <DeviceMobileIcon className="w-4 h-4" />
                      ) : isDesktop ? (
                        <LaptopIcon className="w-4 h-4" />
                      ) : (
                        <MonitorIcon className="w-4 h-4" />
                      )}
                    </div>
                    <div className="w-full">
                      <div className="flex items-center justify-between w-full gap-2">
                        <span className="text-xs font-semibold text-white">
                          {sess.device_name || "Web Client"}
                        </span>
                        {sess.is_current && (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium text-white border border-zinc-700">
                            Current session
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-2 text-[11px] text-zinc-400 mt-0.5">
                        <span>IP: {sess.ip_address || "127.0.0.1"}</span>
                        <span>•</span>
                        <span>
                          {sess.is_current
                            ? "Active now"
                            : `Last active ${new Date(sess.last_active_at).toLocaleDateString()}`}
                        </span>
                      </div>
                    </div>
                  </div>

                  {!sess.is_current && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => revokeSessionMutation.mutate(sess.id)}
                      disabled={revokeSessionMutation.isPending}
                      className="text-xs text-zinc-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg px-2.5 py-1"
                    >
                      Revoke
                    </Button>
                  )}
                </div>
              );
            })
          )}
        </div>
      </Card>

      <Card hover={false} className="p-5 space-y-4 !bg-sidebar border-red-950/40">
        <div className="space-y-0.5">
          <h3 className="text-sm font-semibold text-red-400 flex items-center gap-2">
            <WarningIcon className="w-4 h-4 text-red-400" />
            Danger Zone
          </h3>
          <p className="text-xs text-zinc-400">
            Irreversible actions and data export for your Aegis account
          </p>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3.5 rounded-xl border border-zinc-800 bg-zinc-900/40">
          <div>
            <h4 className="text-xs font-semibold text-white">Export Account Data</h4>
            <p className="text-[11px] text-zinc-400 mt-0.5">
              Download a complete JSON export of your chat threads, token usage records, and profile
            </p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => exportDataMutation.mutate()}
            disabled={exportDataMutation.isPending}
            className="rounded-full text-xs flex items-center gap-1.5 border-zinc-700 hover:bg-zinc-800 self-start sm:self-auto shrink-0"
          >
            <DownloadSimpleIcon className="w-3.5 h-3.5" />
            {exportDataMutation.isPending ? "Exporting..." : "Export Data"}
          </Button>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3.5 rounded-xl border border-red-900/30 bg-red-950/10">
          <div>
            <h4 className="text-xs font-semibold text-red-400">Delete Account</h4>
            <p className="text-[11px] text-zinc-400 mt-0.5">
              Permanently delete your Aegis account, API credentials, sessions, and chat history
            </p>
          </div>
          <Button
            variant="danger"
            size="sm"
            onClick={() => setDeleteModalOpen(true)}
            className="rounded-full text-xs self-start sm:self-auto shrink-0"
          >
            Delete Account
          </Button>
        </div>
      </Card>

      {/* Sign Out Button */}
      <div className="pt-2 flex justify-start">
        <Button
          onClick={handleLogout}
          className="bg-zinc-800 border border-zinc-700 text-zinc-200 rounded-full text-[13px] flex items-center gap-2 hover:bg-zinc-700 hover:text-white transition-colors"
        >
          <SignOutIcon className="w-4 h-4" /> Sign out
        </Button>
      </div>

      {/* Delete Account Confirmation Dialog */}
      <Dialog
        open={deleteModalOpen}
        onOpenChange={setDeleteModalOpen}
        title="Delete Aegis Account"
        description="This action is permanent and cannot be undone. All your stored BYOK provider credentials, chat threads, and usage history will be erased immediately."
      >
        <div className="space-y-4 pt-2">
          <p className="text-xs text-zinc-300">
            Please type <span className="font-mono font-bold text-red-400">DELETE</span> to confirm.
          </p>
          <Input
            value={confirmDeleteText}
            onChange={(e) => setConfirmDeleteText(e.target.value)}
            placeholder="DELETE"
            className="font-mono"
          />
          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setDeleteModalOpen(false);
                setConfirmDeleteText("");
              }}
              className="rounded-full text-xs"
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              size="sm"
              disabled={confirmDeleteText !== "DELETE" || deleteAccountMutation.isPending}
              onClick={() => deleteAccountMutation.mutate()}
              className="rounded-full text-xs"
            >
              {deleteAccountMutation.isPending ? "Deleting..." : "Permanently Delete"}
            </Button>
          </div>
        </div>
      </Dialog>
    </SettingsLayout>
  );
};

export default SecurityTab;
