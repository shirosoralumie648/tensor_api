'use client'

import { useState } from 'react'
import { Download, X } from 'lucide-react'
import type { Message } from '@/types'

interface MessageExportDialogProps {
  messages: Message[]
  onClose: () => void
}

type ExportFormat = 'json' | 'markdown' | 'csv' | 'txt'

export function MessageExportDialog({
  messages,
  onClose,
}: MessageExportDialogProps) {
  const [format, setFormat] = useState<ExportFormat>('markdown')
  const [loading, setLoading] = useState(false)

  // 转换为 Markdown 格式
  const toMarkdown = () => {
    return messages
      .map((msg) => {
        const role = msg.role === 'user' ? '👤 用户' : '🤖 AI'
        return `## ${role}\n\n${msg.content}\n\n---\n`
      })
      .join('\n')
  }

  // 转换为 JSON 格式
  const toJSON = () => {
    return JSON.stringify(messages, null, 2)
  }

  // 转换为 CSV 格式
  const toCSV = () => {
    const headers = ['Role', 'Content', 'Timestamp']
    const rows = messages.map((msg) => [
      msg.role,
      `"${msg.content.replace(/"/g, '""')}"`,
      new Date(msg.timestamp).toISOString(),
    ])

    return [headers, ...rows].map((row) => row.join(',')).join('\n')
  }

  // 转换为纯文本格式
  const toText = () => {
    return messages
      .map((msg) => {
        const role = msg.role === 'user' ? 'User' : 'Assistant'
        const time = new Date(msg.timestamp).toLocaleString('zh-CN')
        return `[${time}] ${role}:\n${msg.content}\n`
      })
      .join('\n' + '='.repeat(60) + '\n\n')
  }

  // 获取转换函数
  const getConverter = () => {
    const converters = {
      markdown: toMarkdown,
      json: toJSON,
      csv: toCSV,
      txt: toText,
    }
    return converters[format]
  }

  // 获取文件扩展名
  const getFileExtension = () => {
    const extensions = {
      markdown: 'md',
      json: 'json',
      csv: 'csv',
      txt: 'txt',
    }
    return extensions[format]
  }

  // 导出文件
  const handleExport = async () => {
    setLoading(true)
    try {
      const converter = getConverter()
      const content = converter()
      const extension = getFileExtension()
      const filename = `conversation_${new Date().getTime()}.${extension}`

      // 创建 Blob 对象
      const blob = new Blob([content], {
        type:
          format === 'json'
            ? 'application/json'
            : format === 'csv'
              ? 'text/csv'
              : 'text/plain',
      })

      // 创建下载链接
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)

      onClose()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-dark-800 rounded-lg max-w-md w-full mx-4 shadow-xl">
        {/* 头部 */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-dark-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            导出对话
          </h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 dark:hover:bg-dark-700 rounded"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 内容 */}
        <div className="p-4 space-y-4">
          {/* 消息统计 */}
          <div className="bg-gray-50 dark:bg-dark-700 p-3 rounded">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              将导出 <span className="font-semibold">{messages.length}</span> 条消息
            </p>
          </div>

          {/* 格式选择 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              导出格式
            </label>
            <div className="grid grid-cols-2 gap-2">
              {(
                [
                  { value: 'markdown', label: 'Markdown', desc: '最佳格式化' },
                  { value: 'json', label: 'JSON', desc: '完整数据' },
                  { value: 'csv', label: 'CSV', desc: '电子表格' },
                  { value: 'txt', label: 'Text', desc: '纯文本' },
                ] as const
              ).map((option) => (
                <button
                  key={option.value}
                  onClick={() => setFormat(option.value)}
                  className={`p-3 rounded border-2 transition-all text-left ${
                    format === option.value
                      ? 'border-primary-600 bg-primary-50 dark:bg-dark-600'
                      : 'border-gray-200 dark:border-dark-600 hover:border-primary-300'
                  }`}
                >
                  <p className="font-medium text-sm">{option.label}</p>
                  <p className="text-xs text-gray-500">{option.desc}</p>
                </button>
              ))}
            </div>
          </div>

          {/* 格式描述 */}
          <div className="bg-blue-50 dark:bg-blue-900 text-blue-900 dark:text-blue-100 p-3 rounded text-sm">
            {format === 'markdown' && '✨ 推荐用于分享和阅读'}
            {format === 'json' && '📊 保留所有元数据和结构'}
            {format === 'csv' && '📈 可在 Excel 等工具中打开'}
            {format === 'txt' && '📄 简单纯文本格式'}
          </div>
        </div>

        {/* 底部操作 */}
        <div className="flex gap-3 p-4 border-t border-gray-200 dark:border-dark-700 bg-gray-50 dark:bg-dark-700 rounded-b-lg">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-dark-800 border border-gray-300 dark:border-dark-600 rounded hover:bg-gray-50 dark:hover:bg-dark-700"
          >
            取消
          </button>
          <button
            onClick={handleExport}
            disabled={loading}
            className="flex-1 px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded hover:bg-primary-700 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Download className="w-4 h-4" />
            {loading ? '处理中...' : '导出'}
          </button>
        </div>
      </div>
    </div>
  )
}

