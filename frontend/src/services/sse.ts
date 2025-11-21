/**
 * SSE (Server-Sent Events) 客户端服务
 * 用于处理实时通信和流式数据接收
 */

export interface SSEConfig {
  url: string
  headers?: Record<string, string>
  timeout?: number
  onMessage?: (data: any) => void
  onError?: (error: Error) => void
  onClose?: () => void
}

export interface SSEClient {
  connect: () => Promise<void>
  disconnect: () => void
  isConnected: () => boolean
  reconnect: () => Promise<void>
}

/**
 * 创建 SSE 客户端
 */
export function createSSEClient(config: SSEConfig): SSEClient {
  let eventSource: EventSource | null = null
  let isConnecting = false
  let reconnectAttempts = 0
  const maxReconnectAttempts = 5
  const baseReconnectDelay = 1000

  // 计算重连延迟（指数退避）
  const getReconnectDelay = () => {
    return Math.min(baseReconnectDelay * Math.pow(2, reconnectAttempts), 30000)
  }

  // 连接
  const connect = async () => {
    if (eventSource || isConnecting) {
      return
    }

    isConnecting = true

    try {
      // 构建 URL
      const url = new URL(config.url, window.location.origin)

      // 创建 EventSource
      eventSource = new EventSource(url.toString())

      // 监听消息
      eventSource.addEventListener('message', (event) => {
        try {
          const data = JSON.parse(event.data)
          config.onMessage?.(data)
          reconnectAttempts = 0 // 重置重连计数
        } catch (err) {
          console.error('Failed to parse SSE message:', err)
        }
      })

      // 监听自定义事件
      eventSource.addEventListener('error', (event) => {
        const error = event as any
        if (error.readyState === EventSource.CLOSED) {
          disconnect()
          if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++
            const delay = getReconnectDelay()
            console.log(`Reconnecting in ${delay}ms... (attempt ${reconnectAttempts})`)
            setTimeout(() => {
              reconnect()
            }, delay)
          }
        }
        config.onError?.(new Error('SSE connection error'))
      })

      // 处理打开事件
      eventSource.addEventListener('open', () => {
        console.log('SSE connected')
      })

      isConnecting = false
    } catch (err) {
      isConnecting = false
      config.onError?.(err instanceof Error ? err : new Error(String(err)))
      throw err
    }
  }

  // 断开连接
  const disconnect = () => {
    if (eventSource) {
      eventSource.close()
      eventSource = null
      config.onClose?.()
    }
    isConnecting = false
  }

  // 检查连接状态
  const isConnected = () => {
    return eventSource !== null && eventSource.readyState === EventSource.OPEN
  }

  // 重新连接
  const reconnect = async () => {
    disconnect()
    await connect()
  }

  return {
    connect,
    disconnect,
    isConnected,
    reconnect,
  }
}

/**
 * SSE 消息队列管理
 */
export class SSEMessageQueue {
  private queue: any[] = []
  private maxSize: number

  constructor(maxSize: number = 100) {
    this.maxSize = maxSize
  }

  push(message: any) {
    this.queue.push({
      ...message,
      timestamp: Date.now(),
    })
    // 维持队列大小
    if (this.queue.length > this.maxSize) {
      this.queue.shift()
    }
  }

  getAll() {
    return [...this.queue]
  }

  clear() {
    this.queue = []
  }

  getLastN(n: number) {
    return this.queue.slice(-n)
  }

  size() {
    return this.queue.length
  }
}

/**
 * 带重试的 SSE 流请求
 */
export async function* streamWithRetry(
  url: string,
  options?: {
    maxRetries?: number
    headers?: Record<string, string>
    signal?: AbortSignal
  }
) {
  const maxRetries = options?.maxRetries ?? 3
  let retryCount = 0

  while (retryCount <= maxRetries) {
    try {
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Accept': 'text/event-stream',
          ...(options?.headers || {}),
        },
        signal: options?.signal,
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      if (!response.body) {
        throw new Error('Response body is empty')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')

          buffer = lines.pop() || ''

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const data = line.slice(6).trim()
              if (data) {
                try {
                  yield JSON.parse(data)
                } catch {
                  yield { raw: data }
                }
              }
            }
          }
        }
      } finally {
        reader.releaseLock()
      }

      return // 成功完成，不需要重试
    } catch (error) {
      retryCount++
      if (retryCount > maxRetries) {
        throw error
      }

      // 指数退避
      const delay = Math.min(1000 * Math.pow(2, retryCount - 1), 10000)
      await new Promise((resolve) => setTimeout(resolve, delay))
    }
  }
}

/**
 * 离线支持 - 本地存储消息
 */
export class OfflineMessageStore {
  private storageKey = 'offline_messages'

  save(messages: any[]) {
    try {
      localStorage.setItem(this.storageKey, JSON.stringify(messages))
    } catch (err) {
      console.error('Failed to save offline messages:', err)
    }
  }

  load(): any[] {
    try {
      const data = localStorage.getItem(this.storageKey)
      return data ? JSON.parse(data) : []
    } catch (err) {
      console.error('Failed to load offline messages:', err)
      return []
    }
  }

  clear() {
    try {
      localStorage.removeItem(this.storageKey)
    } catch (err) {
      console.error('Failed to clear offline messages:', err)
    }
  }

  append(messages: any[]) {
    const existing = this.load()
    this.save([...existing, ...messages])
  }
}

