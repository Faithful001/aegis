import { apiClient } from './client';
import { Project, APIKey } from '../types/api';

export const projectsApi = {
  listByOrg: async (orgId: string): Promise<Project[]> => {
    const res = await apiClient.get(`/organizations/${orgId}/projects`);
    return res.data.data || res.data || [];
  },

  get: async (id: string): Promise<Project> => {
    const res = await apiClient.get(`/projects/${id}`);
    return res.data.data || res.data;
  },

  create: async (orgId: string, name: string, description?: string): Promise<Project> => {
    const res = await apiClient.post(`/organizations/${orgId}/projects`, { name, description });
    return res.data.data || res.data;
  },

  listApiKeys: async (projectId: string): Promise<APIKey[]> => {
    const res = await apiClient.get(`/projects/${projectId}/api-keys`);
    return res.data.data || res.data || [];
  },

  createApiKey: async (projectId: string, name: string): Promise<APIKey> => {
    const res = await apiClient.post(`/projects/${projectId}/api-keys`, { name });
    return res.data.data || res.data;
  },

  revokeApiKey: async (keyId: string): Promise<void> => {
    await apiClient.delete(`/api-keys/${keyId}`);
  },
};
