'use client'

import { useState } from 'react'
import { Download, FileText, Copy, Check } from 'lucide-react'
import { useChatStore } from '@/stores/chatStore'

interface ExportFormat {
  format: 'markdown' | 'html' | 'json' | 'txt'
  label: string
  icon: React.ReactNode
}

export function MessageExport() {
  const [isOpen, setIsOpen] = useState(false)
  const [copied, setCopied] = useState<string | null>(null)
  const { currentSession, messages } = useChatStore()

  const formats: ExportFormat[] = [
    {
      format: 'markdown',
      label: 'Markdown',
      icon: <FileText className="w-4 h-4" />,
    },
    {
      format: 'html',
      label: 'HTML',
      icon: <FileText className="w-4 h-4" />,
    },
    {
      format: 'json',
      label: 'JSON',
      icon: <FileText className="w-4 h-4" />,
    },
    {
      format: 'txt',
      label: '纯文本',
      icon: <FileText className="w-4 h-4" />,
    },
  ]

  // 导出为 Markdown
  const exportMarkdown = () => {
    let content = `# ${currentSession?.title || '对话'}\n\n`
    content += `导出时间: ${new Date().toLocaleString()}\n\n`

    messages.forEach((msg) => {
      if (msg.role === 'user') {
        content += `## 用户\n\n${msg.content}\n\n`
      } else {
        content += `## AI\n\n${msg.content}\n\n`
      }
    })

    downloadFile(content, 'conversation.md', 'text/markdown')
  }

  // 导出为 HTML
  const exportHTML = () => {
    let content = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>${currentSession?.title || '对话'}</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; line-height: 1.6; max-width: 900px; margin: 0 auto; padding: 20px; }
    .container { background: #f9f9f9; padding: 20px; border-radius: 8px; }
    h1 { color: #333; }
    .message { margin: 15px 0; padding: 12px; border-radius: 6px; }
    .user { background: #e3f2fd; border-left: 4px solid #2196f3; }
    .assistant { background: #f5f5f5; border-left: 4px solid #666; }
    .role { font-weight: bold; color: #666; margin-bottom: 8px; }
    .content { color: #333; }
    code { background: #e8e8e8; padding: 2px 6px; border-radius: 3px; font-family: monospace; }
    pre { background: #333; color: #fff; padding: 12px; border-radius: 4px; overflow-x: auto; }
  </style>
</head>
<body>
  <div class="container">
    <h1>${currentSession?.title || '对话'}</h1>
    <p>导出时间: ${new Date().toLocaleString()}</p>
`

    messages.forEach((msg) => {
      const role = msg.role === 'user' ? '用户' : 'AI'
      const className = msg.role === 'user' ? 'user' : 'assistant'
      content += `
    <div class="message ${className}">
      <div class="role">${role}</div>
      <div class="content">${msg.content.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</div>
    </div>
`
    })

    content += `
  </div>
</body>
</html>`

    downloadFile(content, 'conversation.html', 'text/html')
  }

  // 导出为 JSON
  const exportJSON = () => {
    const data = {
      title: currentSession?.title,
      exportTime: new Date().toISOString(),
      messages: messages.map((msg) => ({
        role: msg.role,
        content: msg.content,
        timestamp: msg.timestamp,
      })),
    }

    downloadFile(JSON.stringify(data, null, 2), 'conversation.json', 'application/json')
  }

  // 导出为纯文本
  const exportText = () => {
    let content = `${currentSession?.title || '对话'}\n`
    content += `${'='.repeat(50)}\n\n`
    content += `导出时间: ${new Date().toLocaleString()}\n\n`

    messages.forEach((msg) => {
      const role = msg.role === 'user' ? '[用户]' : '[AI]'
      content += `${role}\n${msg.content}\n${'─'.repeat(50)}\n\n`
    })

    downloadFile(content, 'conversation.txt', 'text/plain')
  }

  // 下载文件
  const downloadFile = (content: string, filename: string, mimeType: string) => {
    const blob = new Blob([content], { type: mimeType })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  // 复制到剪贴板
  const copyToClipboard = async (format: string) => {
    let content = ''

    if (format === 'markdown') {
      messages.forEach((msg) => {
        content += `**${msg.role === 'user' ? '用户' : 'AI'}**\n${msg.content}\n\n`
      })
    } else if (format === 'txt') {
      messages.forEach((msg) => {
        content += `[${msg.role === 'user' ? '用户' : 'AI'}]\n${msg.content}\n\n`
      })
    }

    if (content) {
      await navigator.clipboard.writeText(content)
      setCopied(format)
      setTimeout(() => setCopied(null), 2000)
    }
  }

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 hover:bg-gray-100 dark:hover:bg-dark-700 rounded transition-colors"
        title="导出对话"
      >
        <Download className="w-4 h-4" />
        <span className="text-sm">导出</span>
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full mt-2 bg-white dark:bg-dark-700 rounded-lg shadow-lg z-10 min-w-max">
          <div className="p-2 border-b border-gray-200 dark:border-dark-600">
            <p className="text-xs font-semibold px-2 py-1 text-gray-600 dark:text-gray-400">
              下载格式
            </p>
          </div>

          <div className="space-y-1 p-2">
            {formats.map((fmt) => (
              <button
                key={fmt.format}
                onClick={() => {
                  if (fmt.format === 'markdown') exportMarkdown()
                  else if (fmt.format === 'html') exportHTML()
                  else if (fmt.format === 'json') exportJSON()
                  else if (fmt.format === 'txt') exportText()
                  setIsOpen(false)
                }}
                className="w-full flex items-center gap-3 px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-dark-600 rounded"
              >
                {fmt.icon}
                <span>{fmt.label}</span>
                <Download className="w-3 h-3 ml-auto opacity-50" />
              </button>
            ))}
          </div>

          <div className="border-t border-gray-200 dark:border-dark-600 p-2">
            <p className="text-xs font-semibold px-2 py-1 text-gray-600 dark:text-gray-400">
              复制
            </p>
          </div>

          <div className="space-y-1 p-2">
            <button
              onClick={() => copyToClipboard('markdown')}
              className="w-full flex items-center gap-3 px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-dark-600 rounded"
            >
              {copied === 'markdown' ? (
                <Check className="w-4 h-4 text-green-500" />
              ) : (
                <Copy className="w-4 h-4" />
              )}
              <span>
                {copied === 'markdown' ? '已复制' : '复制为 Markdown'}
              </span>
            </button>

            <button
              onClick={() => copyToClipboard('txt')}
              className="w-full flex items-center gap-3 px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-dark-600 rounded"
            >
              {copied === 'txt' ? (
                <Check className="w-4 h-4 text-green-500" />
              ) : (
                <Copy className="w-4 h-4" />
              )}
              <span>
                {copied === 'txt' ? '已复制' : '复制为纯文本'}
              </span>
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

