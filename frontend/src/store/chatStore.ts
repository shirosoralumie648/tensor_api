import { create } from 'zustand';
import { apiClient } from '@/services/api';

export interface Message {
  id: string;
  session_id: string;
  role: 'user' | 'assistant';
  content: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  cost: number;
  created_at: string;
  updated_at: string;
}

export interface Session {
  id: string;
  user_id: number;
  title: string;
  description?: string;
  model: string;
  temperature: number;
  top_p: number;
  max_tokens?: number;
  system_role: string;
  context_length: number;
  pinned: boolean;
  archived: boolean;
  created_at: string;
  updated_at: string;
}

interface ChatStore {
  sessions: Session[];
  currentSession: Session | null;
  messages: Message[];
  isLoading: boolean;
  isSending: boolean;
  isStreaming: boolean;
  streamingMessage: string;
  error: string | null;

  // 会话操作
  createSession: (data: Partial<Session>) => Promise<void>;
  loadSessions: () => Promise<void>;
  selectSession: (sessionId: string) => Promise<void>;
  updateSession: (sessionId: string, data: Partial<Session>) => Promise<void>;
  deleteSession: (sessionId: string) => Promise<void>;

  // 消息操作
  loadMessages: (sessionId: string) => Promise<void>;
  sendMessage: (content: string) => Promise<void>;
  sendMessageStream: (content: string) => Promise<void>;

  // 状态管理
  setError: (error: string | null) => void;
  clearMessages: () => void;
}

export const useChatStore = create<ChatStore>((set, get) => ({
  sessions: [],
  currentSession: null,
  messages: [],
  isLoading: false,
  isSending: false,
  isStreaming: false,
  streamingMessage: '',
  error: null,

  createSession: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.createSession({
        title: data.title || '新对话',
        model: data.model || 'gpt-3.5-turbo',
        temperature: data.temperature || 0.7,
        system_role: data.system_role || '你是一个有帮助的助手',
        context_length: data.context_length || 4,
      });

      if (response.success && response.data) {
        const newSession = response.data;
        const sessions = get().sessions;
        set({
          sessions: [newSession, ...sessions],
          currentSession: newSession,
          messages: [],
          isLoading: false,
        });
      } else {
        throw new Error(response.error?.message || '创建会话失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '创建会话失败';
      set({ error: errorMsg, isLoading: false });
      throw error;
    }
  },

  loadSessions: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getSessions();
      if (response.success && response.data) {
        const sessions = Array.isArray(response.data) ? response.data : response.data.sessions || [];
        set({ sessions, isLoading: false });
      } else {
        throw new Error(response.error?.message || '加载会话失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '加载会话失败';
      set({ error: errorMsg, isLoading: false });
      throw error;
    }
  },

  selectSession: async (sessionId: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getSession(sessionId);
      if (response.success && response.data) {
        const session = response.data;
        set({ currentSession: session, isLoading: false });
        
        // 加载会话消息
        await get().loadMessages(sessionId);
      } else {
        throw new Error(response.error?.message || '加载会话失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '加载会话失败';
      set({ error: errorMsg, isLoading: false });
      throw error;
    }
  },

  updateSession: async (sessionId: string, data) => {
    try {
      const response = await apiClient.updateSession(sessionId, data);
      if (response.success && response.data) {
        const sessions = get().sessions.map((s) =>
          s.id === sessionId ? { ...s, ...data } : s
        );
        set({ sessions });
        if (get().currentSession?.id === sessionId) {
          set({ currentSession: { ...get().currentSession!, ...data } });
        }
      } else {
        throw new Error(response.error?.message || '更新会话失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '更新会话失败';
      set({ error: errorMsg });
      throw error;
    }
  },

  deleteSession: async (sessionId: string) => {
    try {
      const response = await apiClient.deleteSession(sessionId);
      if (response.success) {
        const sessions = get().sessions.filter((s) => s.id !== sessionId);
        set({
          sessions,
          currentSession: get().currentSession?.id === sessionId ? null : get().currentSession,
          messages: get().currentSession?.id === sessionId ? [] : get().messages,
        });
      } else {
        throw new Error(response.error?.message || '删除会话失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '删除会话失败';
      set({ error: errorMsg });
      throw error;
    }
  },

  loadMessages: async (sessionId: string) => {
    set({ isLoading: true });
    try {
      const response = await apiClient.getMessages(sessionId);
      if (response.success && response.data) {
        const messages = Array.isArray(response.data) ? response.data : response.data.messages || [];
        set({ messages, isLoading: false });
      } else {
        throw new Error(response.error?.message || '加载消息失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '加载消息失败';
      set({ error: errorMsg, isLoading: false });
      throw error;
    }
  },

  sendMessage: async (content: string) => {
    const currentSession = get().currentSession;
    if (!currentSession) {
      set({ error: '请先选择或创建一个会话' });
      return;
    }

    set({ isSending: true, error: null });
    try {
      const response = await apiClient.sendMessage({
        session_id: currentSession.id,
        content,
      });

      if (response.success && response.data) {
        // 获取返回的消息（包括用户消息和 AI 响应）
        // 假设返回的是 AI 的响应消息，我们需要重新加载消息列表
        await get().loadMessages(currentSession.id);
        set({ isSending: false });
      } else {
        throw new Error(response.error?.message || '发送消息失败');
      }
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '发送消息失败';
      set({ error: errorMsg, isSending: false });
      throw error;
    }
  },

  // Week 7 新增：流式消息发送
  sendMessageStream: async (content: string) => {
    const currentSession = get().currentSession;
    if (!currentSession) {
      set({ error: '请先选择或创建一个会话' });
      return;
    }

    set({ isStreaming: true, isSending: true, error: null, streamingMessage: '' });
    try {
      await apiClient.sendMessageStreamFetch(
        {
          session_id: currentSession.id,
          content,
        },
        (chunk) => {
          if (chunk.type === 'chunk' && chunk.content) {
            // 累积流式消息内容
            set({ streamingMessage: get().streamingMessage + chunk.content });
          } else if (chunk.type === 'complete') {
            // 重新加载消息以显示完整的消息和 token 信息
            set({ isStreaming: false });
            // 延迟加载以等待数据库写入
            setTimeout(() => {
              get().loadMessages(currentSession.id);
            }, 500);
          }
        },
        (error) => {
          set({ error: error.message, isStreaming: false, isSending: false });
        },
        () => {
          set({ isStreaming: false, isSending: false });
        }
      );
    } catch (error: any) {
      const errorMsg = error.error?.message || error.message || '发送消息失败';
      set({ error: errorMsg, isStreaming: false, isSending: false });
      throw error;
    }
  },

  setError: (error: string | null) => {
    set({ error });
  },

  clearMessages: () => {
    set({ messages: [] });
  },
}));

