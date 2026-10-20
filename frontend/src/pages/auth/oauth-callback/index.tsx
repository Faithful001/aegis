import React, { useEffect, useState, useRef } from "react";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useAuth } from "../../../context/AuthContext";
import { authApi } from "../../../api/auth";
import { toast } from "sonner";
import { CircleNotchIcon, WarningCircleIcon, ArrowLeftIcon } from "@phosphor-icons/react";
import { Button } from "../../../components/ui/Button";

export const OAuthCallbackPage: React.FC = () => {
  const navigate = useNavigate();
  const { setAuthData } = useAuth();
  const searchParams = useSearch({ strict: false }) as Record<string, string> | undefined;

  const [status, setStatus] = useState<"loading" | "error">("loading");
  const [errorMessage, setErrorMessage] = useState<string>("");
  const executedRef = useRef(false);

  useEffect(() => {
    if (executedRef.current) return;
    executedRef.current = true;

    const handleCallback = async () => {
      // 1. Extract query params from searchParams or window.location.search
      const urlParams = new URLSearchParams(window.location.search);
      const code = searchParams?.code || urlParams.get("code");
      const state = searchParams?.state || urlParams.get("state");
      const errorParam = searchParams?.error || urlParams.get("error");
      const errorDesc = searchParams?.error_description || urlParams.get("error_description");

      // Check if provider sent an explicit error
      if (errorParam) {
        setStatus("error");
        setErrorMessage(errorDesc || errorParam || "Authentication was cancelled or denied");
        return;
      }

      if (!code) {
        setStatus("error");
        setErrorMessage("Missing authorization code from identity provider");
        return;
      }

      // Determine provider & mode
      const storedProvider = sessionStorage.getItem("aegis_oauth_provider");
      const paramProvider = searchParams?.provider || urlParams.get("provider");
      const provider = (paramProvider || storedProvider || "google").toLowerCase();
      const mode = sessionStorage.getItem("aegis_oauth_mode") || "auth";

      try {
        if (mode === "link") {
          // Linking external account to current authenticated session
          await authApi.linkAccount(provider, code, state || undefined);
          sessionStorage.removeItem("aegis_oauth_state");
          sessionStorage.removeItem("aegis_oauth_mode");
          sessionStorage.removeItem("aegis_oauth_provider");

          toast.success(
            `${provider.charAt(0).toUpperCase() + provider.slice(1)} account connected!`
          );
          navigate({
            to: "/settings",
            search: { tab: "connected-accounts" },
          });
        } else {
          // Standard login or register
          const res = await authApi.oauthCallback(provider, code, state || undefined);
          const tokenData = res.data?.token;
          const user = res.data?.user;

          const accessToken = tokenData?.access_token || tokenData?.token || "";
          const refreshToken = tokenData?.refresh_token;

          if (!accessToken) {
            throw new Error("Failed to retrieve access token from authentication server");
          }

          setAuthData(accessToken, user, refreshToken);

          sessionStorage.removeItem("aegis_oauth_state");
          sessionStorage.removeItem("aegis_oauth_mode");
          sessionStorage.removeItem("aegis_oauth_provider");

          toast.success("Signed in successfully!");
          navigate({ to: "/" });
        }
      } catch (err: any) {
        setStatus("error");
        const msg =
          err?.response?.data?.message ||
          err?.message ||
          `Failed to authenticate with ${provider}. Please try again.`;
        setErrorMessage(msg);
      }
    };

    handleCallback();
  }, [navigate, searchParams, setAuthData]);

  return (
    <div className="w-full h-screen bg-background flex flex-col justify-center items-center py-12 px-4 sm:px-6 relative overflow-y-auto">
      {status === "loading" ? (
        <div className="space-y-3 flex flex-col items-center justify-center">
          <div className="flex items-center justify-center text-zinc-400">
            <CircleNotchIcon className="w-6 h-6 animate-spin text-zinc-300" />
          </div>
          <h2 className="text-lg font-medium text-white tracking-tight">
            Completing authentication...
          </h2>
          <p className="text-xs text-zinc-500 max-w-xs mx-auto">
            Securely exchanging credentials and establishing your Aegis session.
          </p>
        </div>
      ) : (
        <div className="space-y-4 flex flex-col items-center justify-center">
          <div className="w-10 h-10 rounded-full bg-red-500/10 border border-red-500/20 text-red-400 flex items-center justify-center mx-auto">
            <WarningCircleIcon className="w-5 h-5" />
          </div>
          <div className="space-y-1.5 flex flex-col items-center justify-center">
            <h2 className="text-base font-medium text-white tracking-tight">
              Authentication Failed
            </h2>
            <p className="text-xs text-zinc-400 max-w-sm mx-auto leading-relaxed">{errorMessage}</p>
          </div>
          <div className="pt-2">
            <Button
              onClick={() => navigate({ to: "/auth/login" })}
              className="px-5 py-2 text-xs rounded-full inline-flex items-center gap-1.5"
            >
              <ArrowLeftIcon size={14} />
              Back to Sign In
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export default OAuthCallbackPage;
