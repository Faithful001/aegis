import React, { useState, useEffect } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import { ArrowLeft, Key } from "@phosphor-icons/react";
import { useMutation } from "@tanstack/react-query";
import { authApi } from "../../../api/auth";
import { Input } from "../../../components/ui/Input";
import { Button } from "../../../components/ui/Button";
import { toast } from "sonner";

export const VerifyOTPPage: React.FC = () => {
  const [email, setEmail] = useState(() => {
    const params = new URLSearchParams(window.location.search);
    return params.get("email") || sessionStorage.getItem("aegis_reset_email") || "";
  });
  const [code, setCode] = useState("");
  const [resendCooldown, setResendCooldown] = useState(60);
  const navigate = useNavigate();

  useEffect(() => {
    if (resendCooldown <= 0) return;
    const interval = setInterval(() => {
      setResendCooldown((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(interval);
  }, [resendCooldown]);

  useEffect(() => {
    if (!email) {
      toast.error("Please enter your email to request an OTP code first.");
      navigate({ to: "/auth/forgot-password" });
    }
  }, [email, navigate]);

  const verifyOTPMutation = useMutation({
    mutationFn: async ({ email, code }: { email: string; code: string }) => {
      return await authApi.verifyResetOTP(email, code);
    },
    onSuccess: (res, variables) => {
      const resetToken = res.reset_token;
      sessionStorage.setItem("aegis_reset_token", resetToken);
      sessionStorage.setItem("aegis_reset_email", variables.email);
      toast.success("Code verified! Set your new password.");
      navigate({ to: "/auth/reset-password" });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Invalid or expired reset code");
    },
  });

  const resendOTPMutation = useMutation({
    mutationFn: async (targetEmail: string) => {
      return await authApi.forgotPassword(targetEmail);
    },
    onSuccess: () => {
      toast.success("A new 6-digit code has been sent!");
      setResendCooldown(60);
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to resend code");
    },
  });

  const handleVerify = (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) {
      toast.error("Email is missing. Please start over.");
      navigate({ to: "/auth/forgot-password" });
      return;
    }
    if (code.length !== 6 || verifyOTPMutation.isPending) {
      toast.error("Please enter the 6-digit verification code");
      return;
    }
    verifyOTPMutation.mutate({ email: email.trim(), code: code.trim() });
  };

  const handleResend = () => {
    if (resendCooldown > 0 || !email.trim() || resendOTPMutation.isPending) return;
    resendOTPMutation.mutate(email.trim());
  };

  return (
    <div className="w-full h-screen bg-background flex flex-col items-center justify-center max-w-md mx-auto px-4 relative">
      <Link
        to="/auth/forgot-password"
        className="fixed top-4 left-6 flex items-center gap-1.5 text-xs text-zinc-400 hover:text-white transition-colors group"
      >
        <ArrowLeft size={14} className="group-hover:-translate-x-0.5 transition-transform" />
        Back
      </Link>

      <div className="text-center mb-6">
        <div className="w-12 h-12 rounded-2xl bg-zinc-900 border border-zinc-700 flex items-center justify-center text-white text-xl font-bold mx-auto mb-3">
          <Key className="w-5 h-5" />
        </div>
        <h2 className="text-2xl font-medium text-white tracking-tight">Verify Reset Code</h2>
        <p className="text-xs text-zinc-400 mt-2 max-w-xs mx-auto leading-relaxed">
          Enter the 6-digit verification code sent to{" "}
          <span className="text-zinc-200 font-semibold">{email || "your email"}</span>
        </p>
      </div>

      <form onSubmit={handleVerify} className="space-y-4 w-full">
        <Input
          label="6-Digit OTP Code"
          type="text"
          // placeholder="123456"
          className="rounded-full text-center text-lg "
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
          required
          autoFocus
          maxLength={6}
        />

        <Button
          type="submit"
          className="w-full py-2.5 font-semibold mt-2 rounded-full"
          disabled={verifyOTPMutation.isPending || code.length !== 6}
        >
          {verifyOTPMutation.isPending ? "Verifying..." : "Verify Code"}
        </Button>

        <div className="flex items-center justify-between text-xs text-zinc-400 pt-2">
          <Link to="/auth/login" className="hover:text-white transition-colors">
            Cancel
          </Link>
          <button
            type="button"
            onClick={handleResend}
            disabled={resendCooldown > 0 || resendOTPMutation.isPending || !email.trim()}
            className={`font-medium ${
              resendCooldown > 0 || !email.trim() || resendOTPMutation.isPending
                ? "text-zinc-600 cursor-not-allowed"
                : "text-white"
            }`}
          >
            {resendCooldown > 0
              ? `Resend code in ${resendCooldown}s`
              : resendOTPMutation.isPending
                ? "Sending..."
                : "Resend code"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default VerifyOTPPage;
