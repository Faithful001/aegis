import { apiClient } from "./client";
import { ProviderCredential, DiscoveredModel, ProviderMeta, PaginatedResult } from "../types/api";

export const credentialsApi = {
  getProviders: async (): Promise<ProviderMeta[]> => {
    const res = await apiClient.get("/providers");
    if (Array.isArray(res.data?.data)) return res.data.data;
    if (Array.isArray(res.data)) return res.data;
    return [];
  },

  list: async (): Promise<ProviderCredential[]> => {
    const res = await apiClient.get("/credentials");
    if (Array.isArray(res.data?.data)) return res.data.data;
    if (Array.isArray(res.data)) return res.data;
    return [];
  },

  save: async (provider: string, apiKey: string, baseUrl?: string): Promise<ProviderCredential> => {
    const res = await apiClient.post("/credentials", {
      provider,
      api_key: apiKey,
      base_url: baseUrl,
    });
    return res.data?.data ?? res.data;
  },

  delete: async (provider: string): Promise<void> => {
    await apiClient.delete(`/credentials/${provider}`);
  },

  listAllModels: async (isActive?: boolean): Promise<DiscoveredModel[]> => {
    const params: Record<string, any> = {};
    if (isActive !== undefined) {
      params.is_active = isActive;
    }
    const res = await apiClient.get("/models", { params });
    if (Array.isArray(res.data?.data)) return res.data.data;
    if (Array.isArray(res.data)) return res.data;
    return [];
  },

  toggleModel: async (
    provider: string,
    modelId: string,
    isActive?: boolean
  ): Promise<DiscoveredModel> => {
    const res = await apiClient.post("/models/toggle", {
      provider,
      model_id: modelId,
      is_active: isActive,
    });
    return res.data?.data ?? res.data;
  },

  listModels: async (
    provider: string,
    page = 1,
    limit = 10,
    search?: string
  ): Promise<PaginatedResult<DiscoveredModel>> => {
    const params: Record<string, any> = { page, limit };
    if (search && search.trim()) {
      params.search = search.trim();
    }
    const res = await apiClient.get(`/credentials/${provider}/models`, { params });
    const payload = res.data?.data ?? res.data;
    if (payload && Array.isArray(payload.items)) {
      return payload;
    }
    if (Array.isArray(payload)) {
      return {
        items: payload,
        total: payload.length,
        page,
        limit,
        total_pages: Math.max(1, Math.ceil(payload.length / limit)),
      };
    }
    return {
      items: [],
      total: 0,
      page,
      limit,
      total_pages: 1,
    };
  },
};
