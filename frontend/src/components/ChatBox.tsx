'use client'

import { useState, useRef, useEffect } from 'react'
import { Send, StopCircle, Copy, Check, Download, Share2, Upload } from 'lucide-react'
import { useChat } from '@/hooks/useChat'
import { useChatStore } from '@/stores/chatStore'
import { MessageRenderer } from './MessageRenderer'
import { MessageExportDialog } from './MessageExportDialog'
import { SessionShareDialog } from './SessionShareDialog'
import { SessionImportDialog } from './SessionImportDialog'

export function ChatBox() {
  const { messages, loading, error, sendMessage, stop } = useChat()
  const { currentSessionId, currentModel, temperature, maxTokens } = useChatStore()
  const [input, setInput] = useState('')
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [showExportDialog, setShowExportDialog] = useState(false)
  const [showShareDialog, setShowShareDialog] = useState(false)
  const [showImportDialog, setShowImportDialog] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // 发送消息
  const handleSend = async () => {
    if (!input.trim() || loading) return

    const userInput = input
    setInput('')

    await sendMessage(userInput, {
      model: currentModel,
      temperature,
      maxTokens,
    })
  }

  // 复制消息
  const handleCopy = (content: string, messageId: string) => {
    navigator.clipboard.writeText(content)
    setCopiedId(messageId)
    setTimeout(() => setCopiedId(null), 2000)
  }

  // 处理回车发送
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex flex-col h-full bg-white dark:bg-dark-900">
      {/* 消息列表 */}
      <div className="flex-1 overflow-y-auto scrollbar-custom p-4 space-y-4">
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-400">
            <div className="text-center">
              <p className="text-lg font-semibold mb-2">开始对话</p>
              <p className="text-sm">选择模型并开始与 AI 对话</p>
            </div>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${
                message.role === 'user' ? 'justify-end' : 'justify-start'
              }`}
            >
              <div
                className={`max-w-xs lg:max-w-md xl:max-w-lg px-4 py-3 rounded-lg ${
                  message.role === 'user'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-200 dark:bg-dark-700 text-gray-900 dark:text-gray-100'
                }`}
              >
                {/* 消息内容 */}
                <MessageRenderer content={message.content} />

                {/* 消息工具栏 */}
                {message.role === 'assistant' && (
                  <div className="flex items-center gap-2 mt-3 pt-3 border-t border-gray-300 dark:border-dark-600">
                    <button
                      onClick={() => handleCopy(message.content, message.id)}
                      className="p-1 hover:bg-gray-300 dark:hover:bg-dark-600 rounded transition-colors"
                      title="复制"
                    >
                      {copiedId === message.id ? (
                        <Check className="w-4 h-4 text-green-500" />
                      ) : (
                        <Copy className="w-4 h-4" />
                      )}
                    </button>
                  </div>
                )}

                {/* 加载状态 */}
                {message.loading && (
                  <div className="flex items-center gap-2 mt-2">
                    <div className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full" />
                    <span className="text-sm">正在生成...</span>
                  </div>
                )}
              </div>
            </div>
          ))
        )}

        {/* 错误提示 */}
        {error && (
          <div className="bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-100 p-3 rounded">
            <p className="text-sm">{error}</p>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* 输入框 */}
      <div className="border-t border-gray-200 dark:border-dark-700 p-4 space-y-3">
        {/* 工具栏 */}
        {messages.length > 0 && (
          <div className="flex gap-2 flex-wrap">
            <button
              onClick={() => setShowExportDialog(true)}
              className="flex items-center gap-2 px-3 py-1.5 text-sm bg-gray-100 dark:bg-dark-700 hover:bg-gray-200 dark:hover:bg-dark-600 rounded transition-colors"
              title="导出对话"
            >
              <Download className="w-4 h-4" />
              导出
            </button>
            <button
              onClick={() => setShowShareDialog(true)}
              className="flex items-center gap-2 px-3 py-1.5 text-sm bg-gray-100 dark:bg-dark-700 hover:bg-gray-200 dark:hover:bg-dark-600 rounded transition-colors"
              title="分享对话"
            >
              <Share2 className="w-4 h-4" />
              分享
            </button>
            <button
              onClick={() => setShowImportDialog(true)}
              className="flex items-center gap-2 px-3 py-1.5 text-sm bg-gray-100 dark:bg-dark-700 hover:bg-gray-200 dark:hover:bg-dark-600 rounded transition-colors"
              title="导入对话"
            >
              <Upload className="w-4 h-4" />
              导入
            </button>
          </div>
        )}

        {/* 消息输入框 */}
        <div className="flex gap-3">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入消息... (Shift+Enter 换行)"
            className="input flex-1 resize-none"
            rows={3}
            disabled={loading}
          />
          <div className="flex flex-col gap-2">
            {loading ? (
              <button
                onClick={stop}
                className="btn btn-primary flex-1 flex items-center justify-center gap-2"
              >
                <StopCircle className="w-4 h-4" />
                停止
              </button>
            ) : (
              <button
                onClick={handleSend}
                disabled={!input.trim() || loading}
                className="btn btn-primary flex-1 flex items-center justify-center gap-2 disabled:opacity-50"
              >
                <Send className="w-4 h-4" />
                发送
              </button>
            )}
          </div>
        </div>
      </div>

      {/* 导出对话框 */}
      {showExportDialog && (
        <MessageExportDialog
          messages={messages}
          onClose={() => setShowExportDialog(false)}
        />
      )}

      {/* 分享对话框 */}
      {showShareDialog && currentSessionId && (
        <SessionShareDialog
          sessionId={currentSessionId}
          sessionTitle="对话"
          onClose={() => setShowShareDialog(false)}
        />
      )}

      {/* 导入对话框 */}
      {showImportDialog && (
        <SessionImportDialog
          onImport={async (data) => {
            // 处理导入逻辑
            console.log('导入数据:', data)
          }}
          onClose={() => setShowImportDialog(false)}
        />
      )}
    </div>
  )
}

