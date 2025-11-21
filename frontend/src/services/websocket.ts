/**
 * WebSocket 服务
 * 处理实时通信和流式响应
 */

export enum MessageType {
  // 系统消息
  CONNECT = 'connect',
  DISCONNECT = 'disconnect',
  ERROR = 'error',

  // 聊天消息
  MESSAGE = 'message',
  MESSAGE_START = 'message_start',
  MESSAGE_DELTA = 'message_delta',
  MESSAGE_STOP = 'message_stop',

  // 工具调用
  TOOL_CALL = 'tool_call',
  TOOL_RESULT = 'tool_result',
}

export interface WebSocketMessage {
  type: MessageType;
  id?: string;
  conversationId?: string;
  content?: string;
  data?: Record<string, any>;
  timestamp?: number;
  error?: string;
}

export interface StreamDelta {
  type: 'text' | 'tool_use';
  content: string;
  index: number;
}

export type WebSocketEventHandler = (message: WebSocketMessage) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string = '';
  private handlers: Map<MessageType, Set<WebSocketEventHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 3000;
  private messageQueue: WebSocketMessage[] = [];
  private isConnecting = false;

  /**
   * 连接到 WebSocket 服务器
   */
  connect(url: string, token?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      this.isConnecting = true;
      this.url = url;

      try {
        const wsUrl = token ? `${url}?token=${token}` : url;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log('[WebSocket] Connected');
          this.isConnecting = false;
          this.reconnectAttempts = 0;

          // 发送队列中的消息
          this.flushMessageQueue();

          // 发送连接事件
          this.emit({
            type: MessageType.CONNECT,
            timestamp: Date.now(),
          });

          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('[WebSocket] Failed to parse message:', error);
          }
        };

        this.ws.onerror = (event) => {
          console.error('[WebSocket] Error:', event);
          this.isConnecting = false;
          reject(new Error('WebSocket connection failed'));
        };

        this.ws.onclose = () => {
          console.log('[WebSocket] Disconnected');
          this.isConnecting = false;
          this.attemptReconnect();

          this.emit({
            type: MessageType.DISCONNECT,
            timestamp: Date.now(),
          });
        };
      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  /**
   * 尝试重新连接
   */
  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WebSocket] Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`[WebSocket] Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${delay}ms`);

    setTimeout(() => {
      if (!this.isConnecting && (!this.ws || this.ws.readyState !== WebSocket.OPEN)) {
        this.connect(this.url).catch(error => {
          console.error('[WebSocket] Reconnection failed:', error);
        });
      }
    }, delay);
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.messageQueue = [];
  }

  /**
   * 发送消息
   */
  send(message: WebSocketMessage): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      // 如果未连接，将消息加入队列
      this.messageQueue.push(message);
      return;
    }

    try {
      this.ws.send(JSON.stringify(message));
    } catch (error) {
      console.error('[WebSocket] Failed to send message:', error);
    }
  }

  /**
   * 发送聊天消息
   */
  sendMessage(conversationId: string, content: string): void {
    this.send({
      type: MessageType.MESSAGE,
      conversationId,
      content,
      timestamp: Date.now(),
    });
  }

  /**
   * 处理收到的消息
   */
  private handleMessage(message: WebSocketMessage): void {
    console.log('[WebSocket] Received message:', message.type);
    this.emit(message);
  }

  /**
   * 监听消息类型
   */
  on(type: MessageType, handler: WebSocketEventHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }

    this.handlers.get(type)!.add(handler);

    // 返回取消监听的函数
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  /**
   * 发出事件
   */
  private emit(message: WebSocketMessage): void {
    const handlers = this.handlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error('[WebSocket] Handler error:', error);
        }
      });
    }
  }

  /**
   * 发送队列中的消息
   */
  private flushMessageQueue(): void {
    while (this.messageQueue.length > 0) {
      const message = this.messageQueue.shift();
      if (message) {
        this.send(message);
      }
    }
  }

  /**
   * 获取连接状态
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  /**
   * 获取是否正在连接
   */
  isConnectingNow(): boolean {
    return this.isConnecting;
  }
}

export default new WebSocketService();

