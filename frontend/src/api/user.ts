import { apiClient } from "./client";
import { UserAlertSettings } from "../types/api";

export const userApi = {
  getAlertSettings: async (): Promise<UserAlertSettings> => {
    const res = await apiClient.get<{ success: boolean; data: UserAlertSettings }>("/user/budget-alerts");
    return res.data.data;
  },

  updateAlertSettings: async (settings: Partial<UserAlertSettings>): Promise<UserAlertSettings> => {
    const res = await apiClient.put<{ success: boolean; data: UserAlertSettings }>("/user/budget-alerts", settings);
    return res.data.data;
  },

  exportData: async (): Promise<any> => {
    const res = await apiClient.post<{ success: boolean; data: any }>("/user/export-data");
    return res.data.data;
  },

  deleteAccount: async (): Promise<{ success: boolean; message: string }> => {
    const res = await apiClient.delete<{ success: boolean; message: string }>("/user/me");
    return res.data;
  },
};
