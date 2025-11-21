/**
 * 类型定义
 */

// API 响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 用户类型
export interface User {
  id: string
  email: string
  name: string
  avatar?: string
  createdAt: Date
  updatedAt: Date
}

// 会话类型
export interface Session {
  id: string
  userId: string
  title: string
  model: string
  description?: string
  createdAt: Date
  updatedAt: Date
  messageCount: number
  isStarred: boolean
}

// 消息类型
export interface Message {
  id: string
  sessionId: string
  role: 'user' | 'assistant' | 'system'
  content: string
  tokens?: number
  createdAt: Date
  updatedAt: Date
}

// AI 模型类型
export interface AIModel {
  id: string
  name: string
  provider: string
  description?: string
  maxTokens: number
  costPerToken: number
  isActive: boolean
  createdAt: Date
}

// API 密钥类型
export interface ApiKey {
  id: string
  userId: string
  name: string
  key: string
  keyPrefix: string
  permissions: string[]
  rateLimit?: number
  lastUsed?: Date
  createdAt: Date
  expiresAt?: Date
}

// 使用统计类型
export interface UsageStats {
  date: Date
  requests: number
  tokens: number
  cost: number
  errors: number
  avgLatency: number
}

// 账单类型
export interface BillingInfo {
  userId: string
  balance: number
  totalSpent: number
  totalTokens: number
  currentMonth: BillingMonth
  history: BillingMonth[]
}

export interface BillingMonth {
  month: string
  spent: number
  tokens: number
  requests: number
}

// 文档类型
export interface Document {
  id: string
  name: string
  size: number
  type: string
  status: 'uploading' | 'processing' | 'ready' | 'error'
  uploadedAt: Date
  chunks: number
}

// 知识库类型
export interface KnowledgeBase {
  id: string
  name: string
  description?: string
  documents: Document[]
  createdAt: Date
  updatedAt: Date
}

// 搜索结果类型
export interface SearchResult {
  id: string
  title: string
  content: string
  score: number
  source?: string
}

// 分享链接类型
export interface ShareLink {
  id: string
  sessionId: string
  token: string
  createdBy: string
  createdAt: Date
  expiresAt?: Date
  accessCount: number
}

// 插件类型
export interface Plugin {
  id: string
  name: string
  description?: string
  version: string
  enabled: boolean
  config?: Record<string, any>
  createdAt: Date
  updatedAt: Date
}

// 工具类型
export interface Tool {
  id: string
  name: string
  description?: string
  parameters?: Record<string, any>
  result?: string
}

// 通知类型
export interface Notification {
  id: string
  userId: string
  type: 'info' | 'success' | 'warning' | 'error'
  title: string
  message: string
  read: boolean
  createdAt: Date
}

// 分页类型
export interface Paginated<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  hasMore: boolean
}

// 错误类型
export interface ErrorDetail {
  code: string
  message: string
  details?: Record<string, any>
}

// 聊天选项类型
export interface ChatOptions {
  model?: string
  temperature?: number
  maxTokens?: number
  topP?: number
  frequencyPenalty?: number
  presencePenalty?: number
  systemPrompt?: string
}

// 导出类型
export type ExportFormat = 'json' | 'csv' | 'markdown' | 'pdf'

// 主题类型
export type Theme = 'light' | 'dark' | 'auto'

