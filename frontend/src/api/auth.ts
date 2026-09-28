import { apiClient } from './client';
import { User, AuthResponse } from '../types/api';

export const authApi = {
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const res = await apiClient.post('/auth/login', { email, password });
    return res.data;
  },

  register: async (email: string, password: string, name?: string): Promise<AuthResponse> => {
    const res = await apiClient.post('/auth/register', { email, password, name });
    return res.data;
  },

  logout: async (): Promise<void> => {
    try {
      await apiClient.post('/auth/logout');
    } finally {
      localStorage.removeItem('aegis_jwt_token');
    }
  },

  getProfile: async (): Promise<User> => {
    const res = await apiClient.get('/user/profile');
    return res.data.data;
  },
};
