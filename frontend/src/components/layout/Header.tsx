import React, { useState } from "react";
import { Link } from "@tanstack/react-router";
import { Settings, Sparkles } from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import { AuthModal } from "../modals/AuthModal";

export const Header: React.FC = () => {
  const { isAuthenticated, user } = useAuth();
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
        <Link to="/" className="flex items-center gap-2 text-white">
          <div className="w-8 h-8 rounded-full bg-zinc-900 border border-zinc-700/80 flex items-center justify-center text-white shadow-sm">
            <span className="font-bold text-lg tracking-tighter">Ø</span>
          </div>
        </Link>
      </div>

      <div className="flex items-center gap-3">
        {!isAuthenticated ? (
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
        ) : (
          <Link
            to="/profile"
            className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-surface-card border border-surface-border text-xs font-medium text-zinc-200 hover:bg-zinc-800 transition-all"
          >
            <div className="w-5 h-5 rounded-full bg-zinc-800 border border-amber-500/40 flex items-center justify-center text-[10px] font-bold text-zinc-200">
              {user?.name
                ? user.name.slice(0, 2).toUpperCase()
                : user?.email
                ? user.email.slice(0, 2).toUpperCase()
                : "U"}
            </div>
            <span className="max-w-[120px] truncate">
              {user?.name || user?.email || "Account"}
            </span>
          </Link>
        )}
      </div>

      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} initialMode={authMode} />
    </header>
  );
};
