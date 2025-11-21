/**
 * 流式响应处理客户端
 */

export interface StreamEvent {
  type: 'connected' | 'chunk' | 'done' | 'error';
  content?: string;
  message_id?: string;
  session_id?: string;
  input_tokens?: number;
  output_tokens?: number;
  total_tokens?: number;
  message?: string;
}

/**
 * 处理 SSE 流式响应
 */
export async function streamMessage(
  url: string,
  options: {
    method?: string;
    body: Record<string, any>;
    headers?: Record<string, string>;
    onEvent: (event: StreamEvent) => void;
    onError?: (error: Error) => void;
    onComplete?: () => void;
    signal?: AbortSignal;
  }
): Promise<void> {
  try {
    const response = await fetch(url, {
      method: options.method || 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      body: JSON.stringify(options.body),
      signal: options.signal,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    if (!response.body) {
      throw new Error('No response body');
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();

      if (done) {
        // 处理剩余的缓冲区内容
        if (buffer.trim()) {
          processLine(buffer, options.onEvent);
        }
        options.onComplete?.();
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');

      // 保留最后一行（可能不完整）
      buffer = lines.pop() || '';

      for (const line of lines) {
        processLine(line, options.onEvent);
      }
    }
  } catch (error) {
    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        // 取消请求，不报错
        options.onComplete?.();
      } else {
        options.onError?.(error);
      }
    }
  }
}

/**
 * 处理单行数据
 */
function processLine(line: string, onEvent: (event: StreamEvent) => void): void {
  const trimmed = line.trim();

  // 跳过空行
  if (!trimmed) {
    return;
  }

  // 解析事件类型
  const eventMatch = trimmed.match(/^event:\s*(.+?)$/m);
  if (eventMatch) {
    const eventType = eventMatch[1].trim();
    return;
  }

  // 解析数据
  const dataMatch = trimmed.match(/^data:\s*(.+)$/);
  if (!dataMatch) {
    return;
  }

  try {
    const jsonStr = dataMatch[1].trim();
    const data = JSON.parse(jsonStr) as StreamEvent;
    onEvent(data);
  } catch (error) {
    console.error('Failed to parse SSE data:', error, 'raw:', trimmed);
  }
}

/**
 * 创建可取消的流式请求控制器
 */
export class StreamController {
  private abortController: AbortController;
  private isRunning = false;

  constructor() {
    this.abortController = new AbortController();
  }

  async stream(
    url: string,
    options: Omit<Parameters<typeof streamMessage>[1], 'signal'>
  ): Promise<void> {
    this.isRunning = true;
    try {
      await streamMessage(url, {
        ...options,
        signal: this.abortController.signal,
      });
    } finally {
      this.isRunning = false;
    }
  }

  cancel(): void {
    if (this.isRunning) {
      this.abortController.abort();
    }
  }

  isActive(): boolean {
    return this.isRunning;
  }

  reset(): void {
    this.abortController = new AbortController();
  }
}

