import { apiClient } from './client';
import { Organization, OrgMember, ProviderCredential } from '../types/api';

export const orgsApi = {
  list: async (): Promise<Organization[]> => {
    const res = await apiClient.get('/organizations');
    return res.data.data || res.data || [];
  },

  get: async (id: string): Promise<Organization> => {
    const res = await apiClient.get(`/organizations/${id}`);
    return res.data.data || res.data;
  },

  create: async (name: string, slug?: string): Promise<Organization> => {
    const res = await apiClient.post('/organizations', { name, slug });
    return res.data.data || res.data;
  },

  listMembers: async (orgId: string): Promise<OrgMember[]> => {
    const res = await apiClient.get(`/organizations/${orgId}/members`);
    return res.data.data || res.data || [];
  },

  addMember: async (orgId: string, email: string, role: string): Promise<OrgMember> => {
    const res = await apiClient.post(`/organizations/${orgId}/members`, { email, role });
    return res.data.data || res.data;
  },

  listCredentials: async (orgId: string): Promise<ProviderCredential[]> => {
    const res = await apiClient.get(`/organizations/${orgId}/credentials`);
    return res.data.data || res.data || [];
  },

  saveCredential: async (orgId: string, provider: string, apiKey: string): Promise<ProviderCredential> => {
    const res = await apiClient.post(`/organizations/${orgId}/credentials`, { provider, api_key: apiKey });
    return res.data.data || res.data;
  },

  deleteCredential: async (orgId: string, provider: string): Promise<void> => {
    await apiClient.delete(`/organizations/${orgId}/credentials/${provider}`);
  },
};
