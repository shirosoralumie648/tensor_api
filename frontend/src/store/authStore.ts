import { create } from 'zustand';
import { apiClient, LoginRequest, LoginResponse } from '@/services/api';

export interface User {
  id: number;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string;
  role: number;
  quota: number;
  total_quota?: number;
  used_quota?: number;
}

interface AuthStore {
  user: User | null;
  isLoading: boolean;
  error: string | null;
  
  // 认证操作
  login: (credentials: LoginRequest) => Promise<void>;
  register: (data: { username: string; email: string; password: string }) => Promise<void>;
  logout: () => void;
  updateProfile: (data: Partial<User>) => Promise<void>;
  
  // Token 管理
  setTokens: (accessToken: string, refreshToken: string) => void;
  clearTokens: () => void;
  
  // 状态管理
  setUser: (user: User | null) => void;
  setError: (error: string | null) => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: typeof window !== 'undefined' ? (() => {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  })() : null,
  isLoading: false,
  error: null,

  login: async (credentials: LoginRequest) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.login(credentials);
      if (response.success && response.data) {
        const { access_token, refresh_token, user } = response.data;
        apiClient.setTokens(access_token, refresh_token);
        
        if (typeof window !== 'undefined') {
          localStorage.setItem('user', JSON.stringify(user));
        }
        
        set({ user, isLoading: false });
      } else {
        throw new Error(response.error?.message || '登录失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '登录失败';
      set({ error: errorMsg, isLoading: false });
      throw error;
    }
  },

  register: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.register(data);
      if (response.success) {
        // 注册成功后自动登录
        await useAuthStore.getState().login({
          username: data.username,
          password: data.password,
        });
      } else {
        throw new Error(response.error?.message || '注册失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '注册失败';
      set({ error: errorMsg, isLoading: false });
      throw error;
    }
  },

  logout: () => {
    apiClient.clearTokens();
    if (typeof window !== 'undefined') {
      localStorage.removeItem('user');
    }
    set({ user: null, error: null });
  },

  updateProfile: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.updateUserProfile(data);
      if (response.success) {
        const updatedUser = { ...useAuthStore.getState().user, ...data } as User;
        if (typeof window !== 'undefined') {
          localStorage.setItem('user', JSON.stringify(updatedUser));
        }
        set({ user: updatedUser, isLoading: false });
      } else {
        throw new Error(response.error?.message || '更新失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '更新失败';
      set({ error: errorMsg, isLoading: false });
      throw error;
    }
  },

  setTokens: (accessToken: string, refreshToken: string) => {
    apiClient.setTokens(accessToken, refreshToken);
  },

  clearTokens: () => {
    apiClient.clearTokens();
    set({ user: null });
  },

  setUser: (user: User | null) => {
    set({ user });
    if (user && typeof window !== 'undefined') {
      localStorage.setItem('user', JSON.stringify(user));
    }
  },

  setError: (error: string | null) => {
    set({ error });
  },
}));

