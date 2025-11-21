'use client'

import { useState } from 'react'
import { Download, X, Copy, Check } from 'lucide-react'

interface SessionExportModalProps {
  sessionId: string
  sessionTitle: string
  messages: Array<{ role: string; content: string; createdAt?: Date }>
  isOpen: boolean
  onClose: () => void
}

export function SessionExportModal({
  sessionId,
  sessionTitle,
  messages,
  isOpen,
  onClose,
}: SessionExportModalProps) {
  const [exportFormat, setExportFormat] = useState<'json' | 'markdown' | 'txt'>('markdown')
  const [copied, setCopied] = useState(false)

  if (!isOpen) return null

  // 生成导出内容
  const generateExportContent = () => {
    switch (exportFormat) {
      case 'markdown':
        return generateMarkdown()
      case 'json':
        return generateJSON()
      case 'txt':
        return generateTXT()
      default:
        return ''
    }
  }

  const generateMarkdown = () => {
    let content = `# ${sessionTitle}\n\n`
    content += `**导出时间**: ${new Date().toLocaleString()}\n\n`
    content += '---\n\n'

    messages.forEach((msg, idx) => {
      if (msg.role === 'user') {
        content += `## 💬 用户消息 #${idx}\n\n${msg.content}\n\n`
      } else {
        content += `## 🤖 AI 回复 #${idx}\n\n${msg.content}\n\n`
      }
      content += '---\n\n'
    })

    return content
  }

  const generateJSON = () => {
    return JSON.stringify(
      {
        session: {
          id: sessionId,
          title: sessionTitle,
          exportedAt: new Date().toISOString(),
        },
        messages,
      },
      null,
      2
    )
  }

  const generateTXT = () => {
    let content = `${sessionTitle}\n`
    content += `导出时间: ${new Date().toLocaleString()}\n`
    content += '='.repeat(80) + '\n\n'

    messages.forEach((msg, idx) => {
      content += `[${msg.role.toUpperCase()} #${idx}]\n${msg.content}\n\n`
    })

    return content
  }

  // 复制到剪贴板
  const handleCopy = () => {
    navigator.clipboard.writeText(generateExportContent())
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  // 下载文件
  const handleDownload = () => {
    const content = generateExportContent()
    const fileExtension = {
      markdown: 'md',
      json: 'json',
      txt: 'txt',
    }[exportFormat]

    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${sessionTitle}-${Date.now()}.${fileExtension}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const exportContent = generateExportContent()

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-dark-800 rounded-lg shadow-xl max-w-2xl w-full mx-4">
        {/* 头部 */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-dark-700">
          <h2 className="text-xl font-semibold">导出对话</h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 dark:hover:bg-dark-700 rounded"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* 主体 */}
        <div className="p-6 space-y-4">
          {/* 格式选择 */}
          <div>
            <label className="block text-sm font-medium mb-2">导出格式</label>
            <div className="flex gap-3">
              {(['markdown', 'json', 'txt'] as const).map((format) => (
                <button
                  key={format}
                  onClick={() => setExportFormat(format)}
                  className={`px-4 py-2 rounded transition-colors ${
                    exportFormat === format
                      ? 'bg-primary-600 text-white'
                      : 'bg-gray-100 dark:bg-dark-700 hover:bg-gray-200 dark:hover:bg-dark-600'
                  }`}
                >
                  {format.toUpperCase()}
                </button>
              ))}
            </div>
          </div>

          {/* 预览 */}
          <div>
            <label className="block text-sm font-medium mb-2">预览</label>
            <div className="bg-gray-50 dark:bg-dark-900 p-4 rounded border border-gray-200 dark:border-dark-700 max-h-64 overflow-y-auto">
              <pre className="text-xs whitespace-pre-wrap break-words text-gray-600 dark:text-gray-300">
                {exportContent.slice(0, 500)}
                {exportContent.length > 500 && '\n...(内容过长)'}
              </pre>
            </div>
          </div>

          {/* 统计信息 */}
          <div className="text-sm text-gray-500">
            <p>消息数: {messages.length}</p>
            <p>内容大小: {(exportContent.length / 1024).toFixed(2)} KB</p>
          </div>
        </div>

        {/* 底部 */}
        <div className="flex gap-3 p-6 border-t border-gray-200 dark:border-dark-700">
          <button
            onClick={handleCopy}
            className="flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-dark-700 hover:bg-gray-200 dark:hover:bg-dark-600 rounded transition-colors"
          >
            {copied ? (
              <>
                <Check className="w-4 h-4" />
                已复制
              </>
            ) : (
              <>
                <Copy className="w-4 h-4" />
                复制内容
              </>
            )}
          </button>

          <button
            onClick={handleDownload}
            className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white hover:bg-primary-700 rounded transition-colors"
          >
            <Download className="w-4 h-4" />
            下载文件
          </button>

          <button
            onClick={onClose}
            className="flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-dark-700 hover:bg-gray-200 dark:hover:bg-dark-600 rounded transition-colors ml-auto"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  )
}

