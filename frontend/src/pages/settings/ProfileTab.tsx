import React, { useEffect, useState } from "react";
import { userApi } from "@/api/user";
import { useAuth } from "@/context/AuthContext";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { Card } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import SettingsLayout from "@/components/layout/SettingsLayout";
import { useAegis } from "@/context/AegisContext";
import { ArrowRightIcon } from "@phosphor-icons/react";

export const ProfileTab: React.FC = () => {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { supportedProviders } = useAegis();

  const [fullName, setFullName] = useState(user?.full_name || user?.name || "");

  // Sync input when user data is loaded or changes
  useEffect(() => {
    if (user) {
      setFullName(user.full_name || user.name || "");
    }
  }, [user]);

  const updateProfileMutation = useMutation({
    mutationFn: async (payload: { full_name: string }) => {
      return await userApi.updateProfile(payload);
    },
    onSuccess: (updatedUser) => {
      toast.success("Profile updated successfully!");
      queryClient.setQueryData(["auth", "me"], (old: any) => {
        if (!old) return old;
        const name = updatedUser.full_name || updatedUser.name;
        return {
          ...old,
          ...updatedUser,
          full_name: name,
          name: name,
        };
      });
      queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to update profile name");
    },
  });

  const handleSaveName = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = fullName.trim();
    if (!trimmed) {
      toast.error("Full name cannot be empty");
      return;
    }

    updateProfileMutation.mutate({ full_name: trimmed });
  };

  const isNameChanged =
    fullName.trim() !== (user?.full_name || user?.name || "").trim() && fullName.trim().length > 0;

  const initials =
    user?.full_name || user?.name
      ? (user.full_name || user.name || "")
          .split(" ")
          .filter(Boolean)
          .map((n) => n[0])
          .join("")
          .slice(0, 2)
          .toUpperCase()
      : user?.email
        ? user.email.slice(0, 2).toUpperCase()
        : "AE";

  const configuredProviders = Array.isArray(supportedProviders)
    ? supportedProviders.filter((p) => p.is_configured)
    : [];

  return (
    <SettingsLayout
      title="Profile"
      subTitle="Manage your personal details, and account credentials."
    >
      <div className="space-y-6">
        {/* Profile Identity Card */}
        <Card
          hover={false}
          className="p-6 bg-[#121214] border border-zinc-800/80 rounded-xl space-y-6"
        >
          <div className="flex items-center gap-5">
            <div className="relative group">
              <div className="w-20 h-20 rounded-full bg-[#27AF61] flex items-center justify-center text-white text-2xl font-semibold">
                {initials}
              </div>
            </div>

            <div className="flex flex-col items-start text-left space-y-1.5">
              <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2.5">
                <h2 className="text-lg font-semibold text-white">
                  {user?.full_name || user?.name}
                </h2>
              </div>

              <div className="flex items-center justify-center sm:justify-start gap-2 text-xs text-zinc-400">
                <span>{user?.email}</span>
              </div>
            </div>
          </div>
        </Card>

        {/* Edit Personal Information Form */}
        <Card
          hover={false}
          className="p-6 bg-[#121214] border border-zinc-800/80 rounded-xl space-y-5"
        >
          <div className="space-y-1">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              Personal Information
            </h3>
            <p className="text-xs text-zinc-400">
              Update your name. This will be reflected in your chats, workspace, and export reports.
            </p>
          </div>

          <form onSubmit={handleSaveName} className="space-y-4 pt-1">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-zinc-300">Full Name</label>
              <Input
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                placeholder="Full name"
                className="rounded-lg bg-zinc-900/70 border-zinc-800 text-sm focus:border-zinc-600 text-zinc-100"
                required
              />
            </div>

            {/* Email Address (Read-only with helper) */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-zinc-300">Email Address</label>
              <Input
                value={user?.email || ""}
                disabled
                className="rounded-lg bg-zinc-900/40 border-zinc-800/60 text-sm text-zinc-400 cursor-not-allowed opacity-80"
              />
            </div>

            <div className="flex items-center justify-end gap-3 pt-2">
              {isNameChanged && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setFullName(user?.full_name || user?.name || "");
                  }}
                  className="rounded-full text-xs text-zinc-400 hover:text-zinc-200 border-zinc-800 hover:bg-zinc-800/60"
                >
                  Cancel
                </Button>
              )}
              <Button
                type="submit"
                disabled={!isNameChanged || updateProfileMutation.isPending}
                className="rounded-full px-5 py-2 text-xs font-semibold bg-white text-zinc-950 hover:bg-zinc-200 transition-all shadow-sm disabled:opacity-50"
              >
                {updateProfileMutation.isPending ? "Saving..." : "Save changes"}
              </Button>
            </div>
          </form>
        </Card>

        <Card
          hover={false}
          className="p-5 bg-[#121214] border border-zinc-800/80 rounded-xl space-y-3"
        >
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                Connected Model Providers
              </h3>
              <p className="text-xs text-zinc-400">
                {configuredProviders.length > 0
                  ? `${configuredProviders.length} provider${configuredProviders.length > 1 ? "s" : ""} connected and active`
                  : "No BYOK model providers configured yet"}
              </p>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigate({ to: "/settings", search: { tab: "providers" } })}
              className="rounded-full text-xs flex items-center gap-1.5 border-zinc-700/80 hover:bg-zinc-800 text-zinc-200 self-start sm:self-auto"
            >
              Configure Providers <ArrowRightIcon className="w-3.5 h-3.5" />
            </Button>
          </div>

          {configuredProviders.length > 0 && (
            <div className="flex flex-wrap gap-1.5 pt-1">
              {configuredProviders.map((c) => (
                <span
                  key={c.id}
                  className="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-zinc-900 text-zinc-300 border border-zinc-800 capitalize"
                >
                  {c.name}
                </span>
              ))}
            </div>
          )}
        </Card>
      </div>
    </SettingsLayout>
  );
};

export default ProfileTab;
