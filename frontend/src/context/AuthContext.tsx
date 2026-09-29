import React, { createContext, useContext, useState, useEffect } from "react";
import { User } from "../types/api";
import { authApi } from "../api/auth";
import { router } from "../router";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, pass: string) => Promise<void>;
  register: (email: string, pass: string, name: string) => Promise<void>;
  logout: () => Promise<void>;
  setToken: (token: string) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setTokenState] = useState<string | null>(sessionStorage.getItem("aegis_jwt_token"));
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const setToken = (newToken: string) => {
    sessionStorage.setItem("aegis_jwt_token", newToken);
    setTokenState(newToken);
  };

  const logoutLocal = () => {
    sessionStorage.removeItem("aegis_jwt_token");
    setTokenState(null);
    setUser(null);
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

  const verifyAndFetchUser = async () => {
    const currentToken = sessionStorage.getItem("aegis_jwt_token");
    if (!currentToken) {
      logoutLocal();
      setIsLoading(false);
      return;
    }
    try {
      const u = await authApi.getMe();
      setUser(formatUser(u));
    } catch (e) {
      console.error("Failed to load user profile:", e);
      logoutLocal();
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    // Initial fetch on mount
    verifyAndFetchUser();

    // Listen for page changes / route navigations
    const unsubscribeRouter = router.history.subscribe(() => {
      const currentToken = sessionStorage.getItem("aegis_jwt_token");
      if (!currentToken) {
        logoutLocal();
      } else {
        verifyAndFetchUser();
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
  }, []);

  const login = async (email: string, pass: string) => {
    const res = await authApi.login(email, pass);
    const jwt = res.data?.access_token || res.data?.token || res.access_token || res.token;
    if (jwt) {
      setToken(jwt);
      const rawUser = res.data?.user || res.user;
      if (rawUser) {
        setUser(formatUser(rawUser));
      } else {
        try {
          const profile = await authApi.getMe();
          setUser(formatUser(profile));
        } catch (e) {
          console.error("Failed to load user profile after login:", e);
        }
      }
    }
  };

  const register = async (email: string, pass: string, name: string) => {
    const res = await authApi.register(email, pass, name);
    const jwt = res.data?.access_token || res.data?.token || res.access_token || res.token;
    if (jwt) {
      setToken(jwt);
      const rawUser = res.data?.user || res.user;
      if (rawUser) {
        setUser(formatUser(rawUser));
      } else {
        try {
          const profile = await authApi.getMe();
          setUser(formatUser(profile));
        } catch (e) {
          console.error("Failed to load user profile after register:", e);
        }
      }
    }
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } catch (e) {
      console.error("Logout error:", e);
    } finally {
      logoutLocal();
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token,
        isLoading,
        login,
        register,
        logout,
        setToken,
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
