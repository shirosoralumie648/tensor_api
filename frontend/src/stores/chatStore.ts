'use client'

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface Session {
  id: string
  title: string
  model: string
  createdAt: Date
  updatedAt: Date
  messageCount: number
}

export interface ChatStoreState {
  sessions: Session[]
  currentSessionId: string | null
  currentModel: string
  temperature: number
  maxTokens: number

  // Actions
  addSession: (session: Session) => void
  removeSession: (sessionId: string) => void
  updateSession: (sessionId: string, updates: Partial<Session>) => void
  setCurrentSession: (sessionId: string | null) => void
  setCurrentModel: (model: string) => void
  setTemperature: (temperature: number) => void
  setMaxTokens: (maxTokens: number) => void
  clearSessions: () => void
}

export const useChatStore = create<ChatStoreState>()(
  persist(
    (set) => ({
      sessions: [],
      currentSessionId: null,
      currentModel: 'gpt-3.5-turbo',
      temperature: 0.7,
      maxTokens: 2000,

      addSession: (session) =>
        set((state) => ({
          sessions: [...state.sessions, session],
        })),

      removeSession: (sessionId) =>
        set((state) => ({
          sessions: state.sessions.filter((s) => s.id !== sessionId),
          currentSessionId:
            state.currentSessionId === sessionId ? null : state.currentSessionId,
        })),

      updateSession: (sessionId, updates) =>
        set((state) => ({
          sessions: state.sessions.map((s) =>
            s.id === sessionId ? { ...s, ...updates, updatedAt: new Date() } : s
          ),
        })),

      setCurrentSession: (sessionId) =>
        set({
          currentSessionId: sessionId,
        }),

      setCurrentModel: (model) =>
        set({
          currentModel: model,
        }),

      setTemperature: (temperature) =>
        set({
          temperature: Math.max(0, Math.min(2, temperature)),
        }),

      setMaxTokens: (maxTokens) =>
        set({
          maxTokens: Math.max(1, Math.min(4000, maxTokens)),
        }),

      clearSessions: () =>
        set({
          sessions: [],
          currentSessionId: null,
        }),
    }),
    {
      name: 'chat-store',
      version: 1,
    }
  )
)

