'use client'

import { useState, useCallback, useRef } from 'react'
import axios from 'axios'

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  loading?: boolean
}

export interface ChatOptions {
  model?: string
  temperature?: number
  maxTokens?: number
  systemPrompt?: string
}

export function useChat(sessionId?: string) {
  const [messages, setMessages] = useState<Message[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  // 发送消息
  const sendMessage = useCallback(
    async (content: string, options?: ChatOptions) => {
      try {
        setError(null)
        setLoading(true)

        // 添加用户消息
        const userMessage: Message = {
          id: `msg_${Date.now()}`,
          role: 'user',
          content,
          timestamp: new Date(),
        }
        setMessages((prev) => [...prev, userMessage])

        // 创建流式请求
        abortControllerRef.current = new AbortController()

        const response = await axios.post(
          '/api/chat/completions',
          {
            messages: [...messages, userMessage].map((m) => ({
              role: m.role,
              content: m.content,
            })),
            model: options?.model || 'gpt-3.5-turbo',
            temperature: options?.temperature || 0.7,
            max_tokens: options?.maxTokens || 2000,
            stream: true,
          },
          {
            responseType: 'stream',
            signal: abortControllerRef.current.signal,
          }
        )

        // 处理流式响应
        let assistantContent = ''
        const reader = response.data.getReader()

        const assistantMessage: Message = {
          id: `msg_${Date.now()}_assistant`,
          role: 'assistant',
          content: '',
          timestamp: new Date(),
          loading: true,
        }

        setMessages((prev) => [...prev, assistantMessage])

        // 读取流数据
        const decoder = new TextDecoder()
        let done = false

        while (!done) {
          const { value, done: streamDone } = await reader.read()
          done = streamDone

          if (value) {
            const chunk = decoder.decode(value)
            assistantContent += chunk

            // 更新消息内容
            setMessages((prev) => {
              const updated = [...prev]
              const lastMessage = updated[updated.length - 1]
              if (lastMessage.role === 'assistant') {
                lastMessage.content = assistantContent
              }
              return updated
            })
          }
        }

        // 完成消息
        setMessages((prev) => {
          const updated = [...prev]
          const lastMessage = updated[updated.length - 1]
          if (lastMessage.role === 'assistant') {
            lastMessage.loading = false
          }
          return updated
        })

        setLoading(false)
      } catch (err) {
        if (err instanceof axios.Cancel) {
          setError('请求已取消')
        } else if (axios.isAxiosError(err)) {
          setError(err.response?.data?.message || '发送消息失败')
        } else {
          setError('未知错误')
        }
        setLoading(false)
      }
    },
    [messages]
  )

  // 停止生成
  const stop = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      setLoading(false)
    }
  }, [])

  // 清空消息
  const clear = useCallback(() => {
    setMessages([])
    setError(null)
  }, [])

  // 删除消息
  const deleteMessage = useCallback((messageId: string) => {
    setMessages((prev) => prev.filter((m) => m.id !== messageId))
  }, [])

  // 编辑消息
  const editMessage = useCallback((messageId: string, content: string) => {
    setMessages((prev) =>
      prev.map((m) => (m.id === messageId ? { ...m, content } : m))
    )
  }, [])

  return {
    messages,
    loading,
    error,
    sendMessage,
    stop,
    clear,
    deleteMessage,
    editMessage,
  }
}

