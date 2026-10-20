import React, { useState, useEffect } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import { EnvelopeOpen } from "@phosphor-icons/react";
import { useMutation } from "@tanstack/react-query";
import { authApi } from "../../../api/auth";
import { Input } from "../../../components/ui/Input";
import { Button } from "../../../components/ui/Button";
import { toast } from "sonner";
import { useAuth } from "../../../context/AuthContext";

export const VerifyEmailPage: React.FC = () => {
  const [email] = useState(() => {
    const params = new URLSearchParams(window.location.search);
    return params.get("email") || sessionStorage.getItem("aegis_verify_email") || "";
  });
  const [code, setCode] = useState("");
  const [resendCooldown, setResendCooldown] = useState(60);
  const navigate = useNavigate();
  const { setAuthData, setToken, refreshUser } = useAuth();

  useEffect(() => {
    if (resendCooldown <= 0) return;
    const interval = setInterval(() => {
      setResendCooldown((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(interval);
  }, [resendCooldown]);

  const verifyEmailMutation = useMutation({
    mutationFn: async (otpCode: string) => {
      return await authApi.verifyEmail({ code: otpCode, email });
    },
    onSuccess: async (res) => {
      const jwt = res.data?.token?.access_token || res.data?.token?.token;
      const refreshToken = res.data?.token?.refresh_token;
      const userData = res.data?.user;

      if (jwt && userData) {
        setAuthData(jwt, userData, refreshToken);
      } else if (jwt) {
        setToken(jwt, refreshToken);
      }

      sessionStorage.removeItem("aegis_verify_email");
      toast.success("Email verified! Welcome to Aegis.");
      await refreshUser();
      navigate({ to: "/chat" });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Invalid or expired code.");
    },
  });

  const resendEmailMutation = useMutation({
    mutationFn: async () => {
      return await authApi.sendVerificationEmail(email);
    },
    onSuccess: () => {
      toast.success("A new verification code has been sent!");
      setResendCooldown(60);
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to resend code.");
    },
  });

  const handleVerify = (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) {
      toast.error("Email is missing. Please register or sign in again.");
      return;
    }
    if (code.length !== 6 || verifyEmailMutation.isPending) {
      toast.error("Please enter the 6-digit verification code.");
      return;
    }
    verifyEmailMutation.mutate(code.trim());
  };

  const handleResend = () => {
    if (resendCooldown > 0 || resendEmailMutation.isPending) return;
    if (!email) {
      toast.error("Email is missing. Please register or sign in again.");
      return;
    }
    resendEmailMutation.mutate();
  };

  return (
    <div className="w-full min-h-screen bg-background flex flex-col justify-center items-center py-12 px-4 sm:px-6 relative overflow-y-auto">
      <div className="w-full max-w-md my-auto flex flex-col items-center">
        <div className="text-center mb-8">
          <div className="w-12 h-12 rounded-2xl bg-zinc-900 border border-zinc-800 flex items-center justify-center mx-auto mb-4 shadow-lg">
            <EnvelopeOpen className="w-5 h-5 text-white" />
          </div>
          <h2 className="text-2xl font-medium text-white tracking-tight">Verify your email</h2>
          <p className="text-xs text-zinc-400 mt-2 max-w-xs mx-auto leading-relaxed">
            We sent a 6-digit code to{" "}
            <span className="text-white font-medium">{email || "your email address"}</span>. Enter it
            below to activate your account.
          </p>
        </div>

        <form onSubmit={handleVerify} className="space-y-4 w-full">
          <Input
            label="Verification Code"
            type="text"
            className="rounded-full text-center text-lg tracking-widest"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
            required
            autoFocus
            maxLength={6}
          />

          <Button
            type="submit"
            className="w-full py-2.5 font-semibold mt-2 rounded-full"
            disabled={verifyEmailMutation.isPending || code.length !== 6}
          >
            {verifyEmailMutation.isPending ? "Verifying..." : "Verify Email"}
          </Button>

          <div className="flex items-center justify-between text-xs text-zinc-400 pt-2">
            <Link to="/auth/login" className="hover:text-white transition-colors">
              Back to login
            </Link>
            <button
              type="button"
              onClick={handleResend}
              disabled={resendCooldown > 0 || resendEmailMutation.isPending}
              className={`font-medium transition-colors ${
                resendCooldown > 0 || resendEmailMutation.isPending
                  ? "text-zinc-600 cursor-not-allowed"
                  : "text-white hover:text-amber-400"
              }`}
            >
              {resendCooldown > 0
                ? `Resend in ${resendCooldown}s`
                : resendEmailMutation.isPending
                  ? "Sending..."
                  : "Resend code"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default VerifyEmailPage;
