import React, { useState } from "react";
import { Link } from "@tanstack/react-router";
import { Settings, Sparkles } from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import { AuthModal } from "../modals/AuthModal";

export const Header: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [authMode, setAuthMode] = useState<"login" | "register">("login");

  const openAuth = (mode: "login" | "register") => {
    setAuthMode(mode);
    setAuthModalOpen(true);
  };

  return (
    <header className="h-14 px-6 flex items-center justify-between select-none bg-transparent sticky top-0 z-20">
      {/* Left side: Logo when unauthenticated */}
      <div className="flex items-center gap-2">
        <div className="w-8 h-8 rounded-full bg-zinc-900 border border-zinc-700/80 flex items-center justify-center text-white shadow-sm">
          <span className="font-bold text-lg tracking-tighter">Ø</span>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <>
          {/* Sign in Button */}
          <Link
            to="/auth/login"
            className="px-4 py-1.5 rounded-full bg-surface-card border border-surface-border text-xs font-semibold text-zinc-200 hover:bg-zinc-800 transition-all"
          >
            Sign in
          </Link>

          {/* Sign up Button */}
          <Link
            to="/auth/register"
            className="px-4 py-1.5 rounded-full bg-white text-zinc-950 text-xs font-semibold hover:bg-zinc-200 transition-all shadow-sm"
          >
            Sign up
          </Link>
        </>
      </div>

      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} initialMode={authMode} />
    </header>
  );
};
