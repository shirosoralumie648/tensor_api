'use client'

import axios from 'axios'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// 创建 API 实例
export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
apiClient.interceptors.request.use((config) => {
  // 添加认证令牌
  const token = localStorage.getItem('auth_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  return config
})

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // 处理 401 错误 - 重定向到登录
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }

    return Promise.reject(error)
  }
)

// 聊天 API
export const chatAPI = {
  // 发送消息
  sendMessage: async (messages: Array<{ role: string; content: string }>, options?: any) => {
    return apiClient.post('/api/chat/completions', {
      messages,
      ...options,
    })
  },

  // 获取会话列表
  listSessions: async (userId?: string) => {
    return apiClient.get('/api/sessions', {
      params: { userId },
    })
  },

  // 获取会话详情
  getSession: async (sessionId: string) => {
    return apiClient.get(`/api/sessions/${sessionId}`)
  },

  // 创建会话
  createSession: async (title: string, model?: string) => {
    return apiClient.post('/api/sessions', {
      title,
      model,
    })
  },

  // 更新会话
  updateSession: async (sessionId: string, data: any) => {
    return apiClient.put(`/api/sessions/${sessionId}`, data)
  },

  // 删除会话
  deleteSession: async (sessionId: string) => {
    return apiClient.delete(`/api/sessions/${sessionId}`)
  },

  // 分享会话
  shareSession: async (sessionId: string) => {
    return apiClient.post(`/api/sessions/${sessionId}/share`)
  },
}

// 模型 API
export const modelAPI = {
  // 获取可用模型
  listModels: async () => {
    return apiClient.get('/api/models')
  },

  // 获取模型详情
  getModel: async (modelId: string) => {
    return apiClient.get(`/api/models/${modelId}`)
  },
}

// 认证 API
export const authAPI = {
  // 登录
  login: async (email: string, password: string) => {
    return apiClient.post('/api/auth/login', {
      email,
      password,
    })
  },

  // 注册
  register: async (email: string, password: string, name?: string) => {
    return apiClient.post('/api/auth/register', {
      email,
      password,
      name,
    })
  },

  // 获取当前用户
  getCurrentUser: async () => {
    return apiClient.get('/api/auth/me')
  },

  // 登出
  logout: async () => {
    localStorage.removeItem('auth_token')
    return apiClient.post('/api/auth/logout')
  },
}

// 开发者 API
export const developerAPI = {
  // 获取 API 密钥
  listKeys: async () => {
    return apiClient.get('/api/developer/keys')
  },

  // 创建 API 密钥
  createKey: async (name: string, permissions?: string[]) => {
    return apiClient.post('/api/developer/keys', {
      name,
      permissions,
    })
  },

  // 删除 API 密钥
  deleteKey: async (keyId: string) => {
    return apiClient.delete(`/api/developer/keys/${keyId}`)
  },

  // 获取使用统计
  getUsageStats: async (startDate?: Date, endDate?: Date) => {
    return apiClient.get('/api/developer/usage', {
      params: {
        start_date: startDate?.toISOString(),
        end_date: endDate?.toISOString(),
      },
    })
  },

  // 获取账单信息
  getBilling: async () => {
    return apiClient.get('/api/developer/billing')
  },

  // 导出数据
  exportData: async (format: 'json' | 'csv') => {
    return apiClient.get('/api/developer/export', {
      params: { format },
      responseType: 'blob',
    })
  },
}

// 知识库 API
export const knowledgeAPI = {
  // 上传文档
  uploadDocument: async (file: File, knowledgeBaseId?: string) => {
    const formData = new FormData()
    formData.append('file', file)
    if (knowledgeBaseId) {
      formData.append('knowledge_base_id', knowledgeBaseId)
    }

    return apiClient.post('/api/knowledge/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  // 列出文档
  listDocuments: async (knowledgeBaseId?: string) => {
    return apiClient.get('/api/knowledge/documents', {
      params: { knowledge_base_id: knowledgeBaseId },
    })
  },

  // 删除文档
  deleteDocument: async (documentId: string) => {
    return apiClient.delete(`/api/knowledge/documents/${documentId}`)
  },

  // 检索文档
  searchDocuments: async (query: string, knowledgeBaseId?: string) => {
    return apiClient.get('/api/knowledge/search', {
      params: {
        query,
        knowledge_base_id: knowledgeBaseId,
      },
    })
  },
}

export default apiClient
