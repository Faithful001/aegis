import { apiClient } from './client';
import { ProviderCredential } from '../types/api';

export const credentialsApi = {
  list: async (): Promise<ProviderCredential[]> => {
    const res = await apiClient.get('/credentials');
    return res.data.data || res.data || [];
  },

  save: async (provider: string, apiKey: string, baseUrl?: string): Promise<ProviderCredential> => {
    const res = await apiClient.post('/credentials', {
      provider,
      api_key: apiKey,
      base_url: baseUrl,
    });
    return res.data.data || res.data;
  },

  delete: async (provider: string): Promise<void> => {
    await apiClient.delete(`/credentials/${provider}`);
  },
};
