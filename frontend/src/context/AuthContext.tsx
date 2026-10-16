import React, { createContext, useContext, useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { User } from "../types/api";
import { authApi } from "../api/auth";
import { router } from "../router";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => Promise<void>;
  setToken: (token: string) => void;
  setAuthData: (token: string, user: any) => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const queryClient = useQueryClient();
  const [token, setTokenState] = useState<string | null>(() =>
    sessionStorage.getItem("aegis_jwt_token")
  );

  const setToken = (newToken: string) => {
    sessionStorage.setItem("aegis_jwt_token", newToken);
    setTokenState(newToken);
  };

  const logoutLocal = () => {
    sessionStorage.removeItem("aegis_jwt_token");
    setTokenState(null);
    queryClient.setQueryData(["auth", "me"], null);
  };

  const formatUser = (rawUser: any): User | null => {
    if (!rawUser) return null;
    const fullName =
      [rawUser.first_name, rawUser.last_name].filter(Boolean).join(" ") ||
      rawUser.name ||
      rawUser.email;
    return {
      ...rawUser,
      name: fullName,
    };
  };

  const setAuthData = (newToken: string, rawUser: any) => {
    setToken(newToken);
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
      if (jwt) {
        const userData = res.data?.user;
        if (userData) {
          setAuthData(jwt, userData);
        } else {
          setToken(jwt);
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
      logoutLocal();
    },
  });

  useEffect(() => {
    // Listen for page changes / route navigations
    const unsubscribeRouter = router.history.subscribe(() => {
      const currentToken = sessionStorage.getItem("aegis_jwt_token");
      if (!currentToken) {
        logoutLocal();
      } else {
        queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
      }
    });

    // Listen for 401 or 403 unauthorized events emitted by API client
    const handleUnauthorized = () => {
      logoutLocal();
    };

    window.addEventListener("unauthorized", handleUnauthorized);

    return () => {
      unsubscribeRouter();
      window.removeEventListener("unauthorized", handleUnauthorized);
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
      logoutLocal();
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
