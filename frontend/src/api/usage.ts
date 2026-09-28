import { apiClient } from './client';
import { UsageMetric } from '../types/api';

export const usageApi = {
  getOrgUsage: async (orgId: string): Promise<UsageMetric> => {
    const res = await apiClient.get(`/organizations/${orgId}/usage`);
    return res.data.data || res.data;
  },

  getProjectUsage: async (projectId: string): Promise<UsageMetric> => {
    const res = await apiClient.get(`/projects/${projectId}/usage`);
    return res.data.data || res.data;
  },
};
