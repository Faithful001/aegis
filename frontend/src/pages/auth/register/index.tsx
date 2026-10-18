import React, { useState } from "react";
import { Link } from "@tanstack/react-router";
import { ArrowLeft } from "@phosphor-icons/react";
import { useMutation } from "@tanstack/react-query";
import { useAuth } from "../../../context/AuthContext";
import { Input } from "../../../components/ui/Input";
import { Button } from "../../../components/ui/Button";
import { toast } from "sonner";

export const RegisterPage: React.FC = () => {
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const { register } = useAuth();

  const registerMutation = useMutation({
    mutationFn: () => register(email, password, name),
    onSuccess: () => {
      toast.success("Account created successfully!");
      // AuthContext.registerMutation.onSuccess handles redirect to /auth/verify-email
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Registration failed");
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    registerMutation.mutate();
  };

  return (
    <div className="w-full h-screen bg-background flex flex-col items-center justify-center max-w-md mx-auto relative">
      <Link
        to="/"
        className="fixed top-4 left-6 flex items-center gap-1.5 text-xs text-zinc-400 hover:text-white transition-colors group"
      >
        <ArrowLeft size={14} className="group-hover:-translate-x-0.5 transition-transform" />
        Back to Home
      </Link>
      <div className="text-center mb-6">
        <div className="w-12 h-12 rounded-2xl bg-zinc-900 border border-zinc-700 flex items-center justify-center text-white text-xl font-bold mx-auto mb-3">
          Ø
        </div>
        <h2 className="text-2xl font-medium text-white tracking-tight">Create Aegis Account</h2>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4 w-full">
        <Input
          label="Full Name"
          placeholder=""
          className="rounded-full"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <Input
          label="Email Address"
          type="email"
          placeholder=""
          className="rounded-full"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
        <Input
          label="Password"
          type="password"
          placeholder="••••••••"
          className="rounded-full"
          showToggle
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />

        <Button
          type="submit"
          className="w-full py-2.5 font-semibold mt-2 rounded-full"
          disabled={registerMutation.isPending}
        >
          {registerMutation.isPending ? "Creating Account..." : "Register"}
        </Button>
      </form>

      <p className="text-xs text-center text-zinc-400 mt-6">
        Already have an account?{" "}
        <Link to="/auth/login" className="text-white hover:underline font-medium">
          Sign in
        </Link>
      </p>
    </div>
  );
};

export default RegisterPage;
