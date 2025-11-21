/**
 * 流式聊天 Hook
 * 处理实时消息流和响应
 */

import { useState, useCallback, useRef, useEffect } from 'react';
import websocketService, { MessageType, WebSocketMessage } from '@/services/websocket';

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  isStreaming?: boolean;
  timestamp: Date;
}

export interface UseStreamingChatOptions {
  conversationId: string;
  onConnected?: () => void;
  onDisconnected?: () => void;
  onError?: (error: Error) => void;
}

export function useStreamingChat(options: UseStreamingChatOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const unsubscribersRef = useRef<Array<() => void>>([]);

  // 连接 WebSocket
  useEffect(() => {
    const connect = async () => {
      try {
        // TODO: 从认证服务获取 token
        await websocketService.connect('ws://localhost:8000/ws');
        setIsConnected(true);
        options.onConnected?.();
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        setError(error);
        options.onError?.(error);
      }
    };

    connect();

    // 设置事件监听
    const unsubConnect = websocketService.on(MessageType.CONNECT, () => {
      setIsConnected(true);
      setError(null);
      options.onConnected?.();
    });

    const unsubDisconnect = websocketService.on(MessageType.DISCONNECT, () => {
      setIsConnected(false);
      options.onDisconnected?.();
    });

    const unsubError = websocketService.on(MessageType.ERROR, (message: WebSocketMessage) => {
      const err = new Error(message.error || 'Unknown error');
      setError(err);
      options.onError?.(err);
    });

    const unsubMessageStart = websocketService.on(MessageType.MESSAGE_START, (message: WebSocketMessage) => {
      setIsLoading(true);
      // 添加助手消息占位符
      const assistantMessage: ChatMessage = {
        id: message.id || `msg-${Date.now()}`,
        role: 'assistant',
        content: '',
        isStreaming: true,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, assistantMessage]);
    });

    const unsubMessageDelta = websocketService.on(MessageType.MESSAGE_DELTA, (message: WebSocketMessage) => {
      // 更新最后一条助手消息
      setMessages(prev => {
        const lastMessage = prev[prev.length - 1];
        if (lastMessage && lastMessage.role === 'assistant') {
          return [
            ...prev.slice(0, -1),
            {
              ...lastMessage,
              content: lastMessage.content + (message.content || ''),
            },
          ];
        }
        return prev;
      });
    });

    const unsubMessageStop = websocketService.on(MessageType.MESSAGE_STOP, () => {
      setIsLoading(false);
      // 标记最后一条消息流完成
      setMessages(prev => {
        const lastMessage = prev[prev.length - 1];
        if (lastMessage && lastMessage.role === 'assistant') {
          return [
            ...prev.slice(0, -1),
            {
              ...lastMessage,
              isStreaming: false,
            },
          ];
        }
        return prev;
      });
    });

    unsubscribersRef.current = [
      unsubConnect,
      unsubDisconnect,
      unsubError,
      unsubMessageStart,
      unsubMessageDelta,
      unsubMessageStop,
    ];

    return () => {
      unsubscribersRef.current.forEach(unsub => unsub());
      websocketService.disconnect();
    };
  }, [options]);

  // 发送消息
  const sendMessage = useCallback(
    (content: string) => {
      if (!isConnected) {
        setError(new Error('Not connected to WebSocket'));
        return;
      }

      // 添加用户消息
      const userMessage: ChatMessage = {
        id: `msg-${Date.now()}`,
        role: 'user',
        content,
        timestamp: new Date(),
      };

      setMessages(prev => [...prev, userMessage]);
      setError(null);

      // 通过 WebSocket 发送
      websocketService.sendMessage(options.conversationId, content);
    },
    [isConnected, options.conversationId]
  );

  // 清空消息
  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  return {
    messages,
    isConnected,
    isLoading,
    error,
    sendMessage,
    clearMessages,
  };
}

