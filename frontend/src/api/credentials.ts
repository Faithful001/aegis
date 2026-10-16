import { apiClient } from './client';
import { ProviderCredential, DiscoveredModel, ProviderMeta } from '../types/api';

export const credentialsApi = {
  getProviders: async (): Promise<ProviderMeta[]> => {
    const res = await apiClient.get('/providers');
    if (Array.isArray(res.data?.data)) return res.data.data;
    if (Array.isArray(res.data)) return res.data;
    return [];
  },

  list: async (): Promise<ProviderCredential[]> => {
    const res = await apiClient.get('/credentials');
    if (Array.isArray(res.data?.data)) return res.data.data;
    if (Array.isArray(res.data)) return res.data;
    return [];
  },

  save: async (provider: string, apiKey: string, baseUrl?: string): Promise<ProviderCredential> => {
    const res = await apiClient.post('/credentials', {
      provider,
      api_key: apiKey,
      base_url: baseUrl,
    });
    return res.data?.data ?? res.data;
  },

  delete: async (provider: string): Promise<void> => {
    await apiClient.delete(`/credentials/${provider}`);
  },

  listModels: async (provider?: string): Promise<DiscoveredModel[]> => {
    const endpoint = provider ? `/credentials/${provider}/models` : '/models';
    const res = await apiClient.get(endpoint);
    if (Array.isArray(res.data?.data)) return res.data.data;
    if (Array.isArray(res.data)) return res.data;
    return [];
  },
};
