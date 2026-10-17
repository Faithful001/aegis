import { apiClient } from './client';
import { UsageMetric, TokenActivityResponse } from '../types/api';

export const usageApi = {
  getUserUsage: async (
    summary: boolean = true,
    startTime?: string,
    endTime?: string
  ): Promise<UsageMetric> => {
    const res = await apiClient.get('/usage', {
      params: {
        summary: summary ? 'true' : 'false',
        ...(startTime ? { start_time: startTime } : {}),
        ...(endTime ? { end_time: endTime } : {}),
      },
    });
    return res.data.data || res.data;
  },

  getTokenActivity: async (
    granularity: 'daily' | 'weekly' | 'cumulative' = 'daily',
    since?: string
  ): Promise<TokenActivityResponse> => {
    const res = await apiClient.get('/usage/activity', {
      params: { granularity, ...(since ? { since } : {}) },
    });
    return res.data.data || res.data;
  },
};
