import { apiClient, getAccessToken, getRefreshToken, clearTokens } from "./client";
import { User, AuthResponse } from "../types/api";
import { BaseResponse } from ".";

export const authApi = {
  login: async (email: string, password: string): Promise<BaseResponse<AuthResponse>> => {
    const res = await apiClient.post("/auth/login", { email, password });
    return res.data;
  },

  register: async (email: string, password: string, name: string): Promise<BaseResponse<void>> => {
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

  refreshToken: async (refreshToken: string): Promise<BaseResponse<AuthResponse>> => {
    const res = await apiClient.post("/auth/refresh", { refresh_token: refreshToken });
    return res.data;
  },

  logout: async (): Promise<void> => {
    try {
      const accessToken = getAccessToken();
      const refreshToken = getRefreshToken();
      await apiClient.post("/auth/logout", {
        access_token: accessToken || "",
        refresh_token: refreshToken || "",
      });
    } finally {
      clearTokens();
    }
  },

  getMe: async (): Promise<BaseResponse<User>> => {
    const res = await apiClient.get("/user/me");
    return res.data;
  },

  sendVerificationEmail: async (email?: string): Promise<{ success: boolean; message: string }> => {
    const res = await apiClient.post("/auth/send-verification-email", { email });
    return res.data;
  },

  verifyEmail: async (payload: {
    code: string;
    email: string;
  }): Promise<BaseResponse<AuthResponse>> => {
    const res = await apiClient.post("/auth/verify-email", payload);
    return res.data;
  },

  forgotPassword: async (email: string): Promise<{ success: boolean; message: string }> => {
    const res = await apiClient.post("/auth/forgot-password", { email });
    return res.data;
  },

  verifyResetOTP: async (
    email: string,
    code: string
  ): Promise<{ success: boolean; reset_token: string; token_type: string; message: string }> => {
    const res = await apiClient.post("/auth/verify-reset-otp", { email, code });
    return res.data;
  },

  resetPassword: async (
    token: string,
    password: string
  ): Promise<{ success: boolean; message: string }> => {
    const res = await apiClient.post("/auth/reset-password", { token, password });
    return res.data;
  },

  changePassword: async (
    currentPassword: string,
    newPassword: string
  ): Promise<{ success: boolean; message: string }> => {
    const res = await apiClient.post("/auth/change-password", {
      current_password: currentPassword,
      new_password: newPassword,
    });
    return res.data;
  },
};
