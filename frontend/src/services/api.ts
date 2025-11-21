import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: number;
    message: string;
  };
  message?: string;
  timestamp?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: {
    id: number;
    username: string;
    email: string;
    display_name: string;
    avatar_url: string;
    role: number;
    quota: number;
  };
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

class ApiClient {
  private client: AxiosInstance;
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器 - 添加 Authorization Header
    this.client.interceptors.request.use(
      (config) => {
        const token = this.getAccessToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // 响应拦截器 - 处理错误和 Token 过期
    this.client.interceptors.response.use(
      (response) => response.data,
      async (error: AxiosError<ApiResponse>) => {
        const originalRequest = error.config;

        // Token 过期，尝试刷新
        if (error.response?.status === 401 && originalRequest) {
          const refreshToken = this.getRefreshToken();
          if (refreshToken) {
            try {
              const response = await this.refreshAccessToken(refreshToken);
              if (response.success && response.data) {
                const { access_token, refresh_token } = response.data;
                this.setTokens(access_token, refresh_token);
                
                // 重试原始请求
                if (originalRequest.headers) {
                  originalRequest.headers.Authorization = `Bearer ${access_token}`;
                }
                return this.client(originalRequest);
              }
            } catch (refreshError) {
              // 刷新失败，清除 Token 并重定向到登录
              this.clearTokens();
              if (typeof window !== 'undefined') {
                window.location.href = '/login';
              }
              return Promise.reject(refreshError);
            }
          } else {
            // 没有 refresh token，清除并重定向
            this.clearTokens();
            if (typeof window !== 'undefined') {
              window.location.href = '/login';
            }
          }
        }

        return Promise.reject(error.response?.data || error.message);
      }
    );

    // 从 localStorage 加载已保存的 Token
    if (typeof window !== 'undefined') {
      this.loadTokens();
    }
  }

  // Token 管理
  private getAccessToken(): string | null {
    return this.accessToken || (typeof window !== 'undefined' ? localStorage.getItem('access_token') : null);
  }

  private getRefreshToken(): string | null {
    return this.refreshToken || (typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null);
  }

  private loadTokens(): void {
    this.accessToken = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
  }

  public setTokens(accessToken: string, refreshToken: string): void {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
    if (typeof window !== 'undefined') {
      localStorage.setItem('access_token', accessToken);
      localStorage.setItem('refresh_token', refreshToken);
    }
  }

  public clearTokens(): void {
    this.accessToken = null;
    this.refreshToken = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
    }
  }

  // 用户认证接口
  public async register(data: RegisterRequest): Promise<ApiResponse> {
    return this.client.post('/api/v1/register', data);
  }

  public async login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    return this.client.post('/api/v1/login', data);
  }

  public async refreshAccessToken(refreshToken: string): Promise<ApiResponse<{ access_token: string; refresh_token: string }>> {
    return this.client.post('/api/v1/refresh', { refresh_token: refreshToken });
  }

  // 用户信息接口
  public async getUserProfile(): Promise<ApiResponse> {
    return this.client.get('/api/v1/user/profile');
  }

  public async updateUserProfile(data: any): Promise<ApiResponse> {
    return this.client.put('/api/v1/user/profile', data);
  }

  // 对话接口
  public async createSession(data: any): Promise<ApiResponse> {
    return this.client.post('/api/v1/chat/sessions', data);
  }

  public async getSessions(): Promise<ApiResponse> {
    return this.client.get('/api/v1/chat/sessions');
  }

  public async getSession(id: string): Promise<ApiResponse> {
    return this.client.get(`/api/v1/chat/sessions/${id}`);
  }

  public async updateSession(id: string, data: any): Promise<ApiResponse> {
    return this.client.put(`/api/v1/chat/sessions/${id}`, data);
  }

  public async deleteSession(id: string): Promise<ApiResponse> {
    return this.client.delete(`/api/v1/chat/sessions/${id}`);
  }

  public async getMessages(sessionId: string): Promise<ApiResponse> {
    return this.client.get(`/api/v1/chat/sessions/${sessionId}/messages`);
  }

  public async sendMessage(data: any): Promise<ApiResponse> {
    return this.client.post('/api/v1/chat/messages', data);
  }

  // 使用 Fetch API 发送流式消息（SSE）- Week 7 新增
  public async sendMessageStreamFetch(
    data: any,
    onChunk: (chunk: any) => void,
    onError?: (error: Error) => void,
    onComplete?: () => void
  ): Promise<void> {
    const accessToken = this.getAccessToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/chat/messages/stream`, {
        method: 'POST',
        headers,
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error('Response body is empty');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();

        if (done) {
          if (buffer.trim()) {
            const match = buffer.match(/data: (.*)/);
            if (match) {
              try {
                const chunk = JSON.parse(match[1]);
                if (chunk.type === 'complete' || chunk.status === 'completed') {
                  onChunk(chunk);
                }
              } catch (e) {
                // 忽略解析错误
              }
            }
          }
          onComplete?.();
          break;
        }

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n\n');

        // 处理完整的消息
        for (let i = 0; i < lines.length - 1; i++) {
          const line = lines[i].trim();
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.substring(6));
              onChunk(data);
            } catch (e) {
              // 忽略解析错误
            }
          }
        }

        // 保留未完成的行
        buffer = lines[lines.length - 1];
      }
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      onError?.(err);
      throw err;
    }
  }

  // 计费接口（待启用）
  public async getBillingHistory(page: number = 1, pageSize: number = 20): Promise<ApiResponse> {
    return this.client.get('/api/v1/billing/history', {
      params: { page, page_size: pageSize },
    });
  }

  public async getQuotaHistory(page: number = 1, pageSize: number = 20): Promise<ApiResponse> {
    return this.client.get('/api/v1/billing/quota-history', {
      params: { page, page_size: pageSize },
    });
  }

  public async getInvoices(page: number = 1, pageSize: number = 20): Promise<ApiResponse> {
    return this.client.get('/api/v1/billing/invoices', {
      params: { page, page_size: pageSize },
    });
  }

  // 通用请求方法
  public get<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.client.get(url, config);
  }

  public post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.client.post(url, data, config);
  }

  public put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.client.put(url, data, config);
  }

  public delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.client.delete(url, config);
  }

  // 获取原始 axios 实例（用于特殊情况）
  public getClient(): AxiosInstance {
    return this.client;
  }
}

export const apiClient = new ApiClient();

