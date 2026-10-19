import React, { createContext, useContext, useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { User } from "../types/api";
import { authApi } from "../api/auth";
import {
  getAccessToken,
  getRefreshToken,
  setTokens,
  clearTokens,
} from "../api/client";
import { router } from "../router";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => Promise<void>;
  setToken: (token: string, refreshToken?: string) => void;
  setAuthData: (token: string, user: any, refreshToken?: string) => void;
  refreshToken: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const queryClient = useQueryClient();
  const [token, setTokenState] = useState<string | null>(() => getAccessToken());

  const setToken = (newToken: string, newRefreshToken?: string) => {
    setTokens(newToken, newRefreshToken);
    setTokenState(newToken);
  };

  const logoutLocal = (redirectToHome = false) => {
    clearTokens();
    setTokenState(null);
    queryClient.setQueryData(["auth", "me"], null);
    if (redirectToHome && window.location.pathname !== "/") {
      router.navigate({ to: "/", replace: true });
    }
  };

  const formatUser = (rawUser: any): User | null => {
    if (!rawUser) return null;
    const fullName =
      rawUser.full_name ||
      [rawUser.first_name, rawUser.last_name].filter(Boolean).join(" ") ||
      rawUser.name ||
      rawUser.email;
    return {
      ...rawUser,
      full_name: fullName,
      name: fullName,
    };
  };

  const setAuthData = (newToken: string, rawUser: any, newRefreshToken?: string) => {
    setToken(newToken, newRefreshToken);
    const formatted = formatUser(rawUser);
    if (formatted) {
      queryClient.setQueryData(["auth", "me"], formatted);
    }
  };

  // TanStack Query for User profile
  const {
    data: user = null,
    isLoading: isUserLoading,
    refetch: refetchUserQuery,
  } = useQuery({
    queryKey: ["auth", "me"],
    queryFn: async () => {
      const res = await authApi.getMe();
      return formatUser(res.data);
    },
    enabled: !!token,
    staleTime: 1000 * 60 * 5,
  });

  const loginMutation = useMutation({
    mutationFn: async ({ email, password }: { email: string; password: string }) => {
      return await authApi.login(email, password);
    },
    onSuccess: (res) => {
      const jwt = res.data?.token.access_token;
      const refreshToken = res.data?.token.refresh_token;
      if (jwt) {
        const userData = res.data?.user;
        if (userData) {
          setAuthData(jwt, userData, refreshToken);
        } else {
          setToken(jwt, refreshToken);
          queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
        }
      }
    },
  });

  const registerMutation = useMutation({
    mutationFn: async ({
      email,
      password,
      name,
    }: {
      email: string;
      password: string;
      name: string;
    }) => {
      return await authApi.register(email, password, name);
    },
    onSuccess: (_data, variables) => {
      sessionStorage.setItem("aegis_verify_email", variables.email);
      router.navigate({
        to: "/auth/verify-email",
        search: { email: variables.email } as any,
      });
    },
  });

  // TanStack Mutation for Logout
  const logoutMutation = useMutation({
    mutationFn: async () => {
      await authApi.logout();
    },
    onSettled: () => {
      logoutLocal(true);
    },
  });

  useEffect(() => {
    // Listen for page changes / route navigations
    const unsubscribeRouter = router.history.subscribe(() => {
      const currentToken = getAccessToken();
      if (!currentToken) {
        logoutLocal(false);
      } else {
        queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
      }
    });

    // Listen for 401 unauthorized events emitted by API client when refresh fails
    const handleUnauthorized = () => {
      logoutLocal(true);
    };

    // Listen for token refresh events emitted by Axios interceptor
    const handleTokenRefreshed = (e: Event) => {
      const customEvent = e as CustomEvent<{
        accessToken: string;
        refreshToken?: string;
        user?: any;
      }>;
      if (customEvent.detail?.accessToken) {
        setTokenState(customEvent.detail.accessToken);
        if (customEvent.detail.user) {
          const formatted = formatUser(customEvent.detail.user);
          if (formatted) {
            queryClient.setQueryData(["auth", "me"], formatted);
          }
        }
      }
    };

    window.addEventListener("unauthorized", handleUnauthorized);
    window.addEventListener("aegis_token_refreshed", handleTokenRefreshed);

    return () => {
      unsubscribeRouter();
      window.removeEventListener("unauthorized", handleUnauthorized);
      window.removeEventListener("aegis_token_refreshed", handleTokenRefreshed);
    };
  }, [queryClient]);

  const login = async (email: string, password: string) => {
    await loginMutation.mutateAsync({ email, password });
  };

  const register = async (email: string, password: string, name: string) => {
    await registerMutation.mutateAsync({ email, password, name });
  };

  const logout = async () => {
    try {
      await logoutMutation.mutateAsync();
    } catch (e) {
      console.error("Logout error:", e);
      logoutLocal(true);
    }
  };

  const refreshToken = async () => {
    const currentRefresh = getRefreshToken();
    if (!currentRefresh) throw new Error("No refresh token available");
    const res = await authApi.refreshToken(currentRefresh);
    const newToken = res.data?.token?.access_token;
    const newRefresh = res.data?.token?.refresh_token;
    if (newToken) {
      setToken(newToken, newRefresh);
      if (res.data?.user) {
        const formatted = formatUser(res.data.user);
        if (formatted) {
          queryClient.setQueryData(["auth", "me"], formatted);
        }
      }
    }
  };

  const refreshUser = async () => {
    await refetchUserQuery();
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token,
        isLoading: !!token && isUserLoading,
        login,
        register,
        logout,
        setToken,
        setAuthData,
        refreshToken,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
};
