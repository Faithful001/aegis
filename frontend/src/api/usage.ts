import { apiClient } from './client';
import { UsageMetric } from '../types/api';

export const usageApi = {
  getUserUsage: async (summary: boolean = true): Promise<UsageMetric> => {
    const res = await apiClient.get('/usage', {
      params: { summary: summary ? 'true' : 'false' },
    });
    return res.data.data || res.data;
  },
};
