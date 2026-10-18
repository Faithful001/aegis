import React, { useState } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import { ArrowLeftIcon, EnvelopeIcon } from "@phosphor-icons/react";
import { useMutation } from "@tanstack/react-query";
import { authApi } from "../../../api/auth";
import { Input } from "../../../components/ui/Input";
import { Button } from "../../../components/ui/Button";
import { toast } from "sonner";

export const ForgotPasswordPage: React.FC = () => {
  const [email, setEmail] = useState("");
  const navigate = useNavigate();

  const forgotPasswordMutation = useMutation({
    mutationFn: async (targetEmail: string) => {
      return await authApi.forgotPassword(targetEmail);
    },
    onSuccess: (_, targetEmail) => {
      sessionStorage.setItem("aegis_reset_email", targetEmail);
      toast.success("If an account exists, a 6-digit code has been sent!");
      navigate({ to: "/auth/verify-otp", search: { email: targetEmail } as any });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Failed to send reset code");
    },
  });

  const handleRequestOTP = (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || forgotPasswordMutation.isPending) return;
    forgotPasswordMutation.mutate(email.trim());
  };

  return (
    <div className="w-full h-screen bg-background flex flex-col items-center justify-center max-w-md mx-auto px-4 relative">
      <Link
        to="/auth/login"
        className="fixed top-4 left-6 flex items-center gap-1.5 text-xs text-zinc-400 hover:text-white transition-colors group"
      >
        <ArrowLeftIcon size={14} className="group-hover:-translate-x-0.5 transition-transform" />
        Back to Sign In
      </Link>

      <div className="text-center mb-6">
        <div className="w-12 h-12 rounded-2xl bg-zinc-900 border border-zinc-700 flex items-center justify-center text-white text-xl font-bold mx-auto mb-3">
          <EnvelopeIcon className="w-5 h-5 text-zinc-300" />
        </div>
        <h2 className="text-2xl font-medium text-white tracking-tight">Reset your password</h2>
        <p className="text-xs text-zinc-400 mt-1 max-w-xs mx-auto">
          Enter your email address and we'll send you a 6-digit OTP to reset your password.
        </p>
      </div>

      <form onSubmit={handleRequestOTP} className="space-y-4 w-full">
        <Input
          label="Email Address"
          type="email"
          placeholder="name@example.com"
          className="rounded-full"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          autoFocus
        />

        <Button
          type="submit"
          className="w-full py-2.5 font-semibold mt-2 rounded-full"
          disabled={forgotPasswordMutation.isPending}
        >
          {forgotPasswordMutation.isPending ? "Sending Code..." : "Send Reset Code"}
        </Button>

        <p className="text-xs text-center text-zinc-400 mt-4">
          Remember your password?{" "}
          <Link to="/auth/login" className="text-white hover:underline font-medium">
            Sign in
          </Link>
        </p>
      </form>
    </div>
  );
};

export default ForgotPasswordPage;
