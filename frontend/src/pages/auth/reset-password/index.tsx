import React, { useState, useEffect } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import { ArrowLeftIcon, LockIcon } from "@phosphor-icons/react";
import { useMutation } from "@tanstack/react-query";
import { authApi } from "../../../api/auth";
import { Input } from "../../../components/ui/Input";
import { Button } from "../../../components/ui/Button";
import { toast } from "sonner";

export const ResetPasswordPage: React.FC = () => {
  const [token, setToken] = useState(() => {
    const params = new URLSearchParams(window.location.search);
    return params.get("token") || sessionStorage.getItem("aegis_reset_token") || "";
  });
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    if (!token) {
      toast.error("No reset token found. Please verify your OTP code first.");
      navigate({ to: "/auth/verify-otp" });
    }
  }, [token, navigate]);

  const resetPasswordMutation = useMutation({
    mutationFn: async ({ token, newPassword }: { token: string; newPassword: string }) => {
      return await authApi.resetPassword(token, newPassword);
    },
    onSuccess: () => {
      sessionStorage.removeItem("aegis_reset_token");
      sessionStorage.removeItem("aegis_reset_email");
      toast.success("Password reset successfully! Please sign in with your new password.");
      navigate({ to: "/auth/login" });
    },
    onError: (err: any) => {
      toast.error(
        err?.response?.data?.message || "Failed to reset password. Token may have expired."
      );
    },
  });

  const handleResetPassword = (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) {
      toast.error("Reset token missing or expired");
      return;
    }
    if (newPassword.length < 8) {
      toast.error("Password must be at least 8 characters");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }
    resetPasswordMutation.mutate({ token, newPassword });
  };

  return (
    <div className="w-full min-h-screen bg-background flex flex-col justify-center items-center py-12 px-4 sm:px-6 relative overflow-y-auto">
      <Link
        to="/auth/verify-otp"
        className="absolute top-6 left-6 flex items-center gap-1.5 text-xs text-zinc-400 hover:text-white transition-colors group z-10"
      >
        <ArrowLeftIcon size={14} className="group-hover:-translate-x-0.5 transition-transform" />
        Back to OTP Verification
      </Link>

      <div className="w-full max-w-md my-auto flex flex-col items-center">
        <div className="text-center mb-6">
          <div className="w-12 h-12 rounded-2xl bg-zinc-900 border border-zinc-700 flex items-center justify-center text-white text-xl font-bold mx-auto mb-3 shadow-lg">
            <LockIcon className="w-5 h-5 text-amber-400" />
          </div>
          <h2 className="text-2xl font-medium text-white tracking-tight">Create New Password</h2>
          <p className="text-xs text-zinc-400 mt-1 max-w-xs mx-auto">
            Your reset code was verified. Please choose a strong new password for your account.
          </p>
        </div>

        <form onSubmit={handleResetPassword} className="space-y-4 w-full">
          <Input
            label="New Password"
            type="password"
            placeholder="•••••••• (min 8 chars)"
            className="rounded-full"
            showToggle
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            required
            autoFocus
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

          <Button
            type="submit"
            className="w-full py-2.5 font-semibold mt-2 rounded-full"
            disabled={resetPasswordMutation.isPending || !newPassword || !confirmPassword}
          >
            {resetPasswordMutation.isPending ? "Updating Password..." : "Update Password"}
          </Button>

          <p className="text-xs text-center text-zinc-400 mt-4">
            Need a new code?{" "}
            <Link to="/auth/forgot-password" className="text-amber-400 hover:underline font-medium">
              Start over
            </Link>
          </p>
        </form>
      </div>
    </div>
  );
};

export default ResetPasswordPage;
