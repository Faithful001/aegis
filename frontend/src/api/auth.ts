import { apiClient } from "./client";
import { User, AuthResponse } from "../types/api";

export const authApi = {
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const res = await apiClient.post("/auth/login", { email, password });
    return res.data;
  },

  register: async (email: string, password: string, name: string): Promise<AuthResponse> => {
    const trimmed = name.trim();
    const parts = trimmed.split(/\s+/);
    const firstName = parts[0] || "User";
    const lastName = parts.slice(1).join(" ") || parts[0] || "User";
    const res = await apiClient.post("/auth/register", {
      email,
      password,
      first_name: firstName,
      last_name: lastName,
    });
    return res.data;
  },

  logout: async (): Promise<void> => {
    try {
      await apiClient.post("/auth/logout");
    } finally {
      sessionStorage.removeItem("aegis_jwt_token");
    }
  },

  getMe: async (): Promise<User> => {
    const res = await apiClient.get("/user/me");
    return res.data.data;
  },
};
