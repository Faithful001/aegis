import React, { useState } from "react";
import { Dialog } from "../ui/Dialog";
import { Input } from "../ui/Input";
import { Button } from "../ui/Button";
import { useAuth } from "../../context/AuthContext";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { Link } from "@tanstack/react-router";

interface AuthModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialMode?: "login" | "register";
}

export const AuthModal: React.FC<AuthModalProps> = ({
  open,
  onOpenChange,
  initialMode = "login",
}) => {
  const [mode, setMode] = useState<"login" | "register">(initialMode);
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const { login, register } = useAuth();

  const resetForm = () => {
    setEmail("");
    setPassword("");
    setName("");
  };

  const loginMutation = useMutation({
    mutationFn: () => login(email, password),
    onSuccess: () => {
      toast.success("Signed in successfully!");
      onOpenChange(false);
      resetForm();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Authentication failed. Please try again.");
    },
  });

  const registerMutation = useMutation({
    mutationFn: () => register(email, password, name),
    onSuccess: () => {
      toast.success("Account created successfully!");
      onOpenChange(false);
      resetForm();
      // AuthContext.registerMutation.onSuccess handles redirect to /auth/verify-email
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Authentication failed. Please try again.");
    },
  });

  const isPending = loginMutation.isPending || registerMutation.isPending;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (mode === "login") {
      loginMutation.mutate();
    } else {
      registerMutation.mutate();
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={mode === "login" ? "Sign in to Aegis" : "Create Aegis Account"}
      description="Please authenticate to send prompts, access AI models, and manage projects."
    >
      <div className="space-y-4">
        {/* Toggle Mode */}
        <div className="flex bg-surface-card p-1 rounded-full border border-surface-border">
          <button
            type="button"
            onClick={() => setMode("login")}
            className={`flex-1 py-1.5 text-xs font-semibold rounded-full transition-all ${
              mode === "login"
                ? "bg-zinc-800 text-white shadow-sm"
                : "text-zinc-400 hover:text-zinc-200"
            }`}
          >
            Sign In
          </button>
          <button
            type="button"
            onClick={() => setMode("register")}
            className={`flex-1 py-1.5 text-xs font-semibold rounded-full transition-all ${
              mode === "register"
                ? "bg-zinc-800 text-white shadow-sm"
                : "text-zinc-400 hover:text-zinc-200"
            }`}
          >
            Sign Up
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3.5">
          {mode === "register" && (
            <Input
              label="Full Name"
              placeholder="e.g. Faithful Eromosele"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          )}

          <Input
            label="Email Address"
            type="email"
            placeholder="name@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />

          <Input
            label="Password"
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          <div className="flex justify-end pt-1">
            <Link
              to="/auth/forgot-password"
              className="text-xs text-zinc-400 hover:text-white transition-colors"
            >
              Forgot password?
            </Link>
          </div>

          <div className="flex justify-end gap-3 pt-3">
            <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending} className="px-6 font-semibold">
              {isPending
                ? mode === "login"
                  ? "Signing in..."
                  : "Creating Account..."
                : mode === "login"
                  ? "Sign In"
                  : "Create Account"}
            </Button>
          </div>
        </form>
      </div>
    </Dialog>
  );
};
