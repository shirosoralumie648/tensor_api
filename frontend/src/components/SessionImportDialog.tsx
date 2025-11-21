'use client'

import { useState, useRef } from 'react'
import { Upload, X, AlertCircle, CheckCircle } from 'lucide-react'

interface SessionImportDialogProps {
  onImport: (data: any) => Promise<void>
  onClose: () => void
}

type ImportStatus = 'idle' | 'uploading' | 'processing' | 'success' | 'error'

export function SessionImportDialog({
  onImport,
  onClose,
}: SessionImportDialogProps) {
  const [status, setStatus] = useState<ImportStatus>('idle')
  const [file, setFile] = useState<File | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [progress, setProgress] = useState(0)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // 处理文件选择
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (!selectedFile) return

    // 验证文件类型
    const validTypes = ['application/json', 'text/plain', 'text/csv']
    if (!validTypes.includes(selectedFile.type)) {
      setError('仅支持 JSON、TXT、CSV 文件格式')
      return
    }

    // 验证文件大小 (限制 10MB)
    if (selectedFile.size > 10 * 1024 * 1024) {
      setError('文件大小不超过 10MB')
      return
    }

    setFile(selectedFile)
    setError(null)
    setSuccess(null)
  }

  // 解析文件
  const parseFile = async (file: File): Promise<any> => {
    const text = await file.text()

    try {
      if (file.name.endsWith('.json')) {
        return JSON.parse(text)
      } else if (file.name.endsWith('.csv')) {
        return parseCSV(text)
      } else {
        return parseText(text)
      }
    } catch (err) {
      throw new Error('文件解析失败')
    }
  }

  // 解析 CSV
  const parseCSV = (text: string): any[] => {
    const lines = text.split('\n')
    const headers = lines[0].split(',')
    const messages = []

    for (let i = 1; i < lines.length; i++) {
      if (!lines[i].trim()) continue

      const values = lines[i].split(',')
      messages.push({
        role: values[0],
        content: values[1]?.replace(/^"(.*)"$/, '$1'),
        timestamp: new Date(values[2]).toISOString(),
      })
    }

    return messages
  }

  // 解析纯文本
  const parseText = (text: string): any[] => {
    const messages = []
    const sections = text.split(/\n={60,}\n/)

    for (const section of sections) {
      if (!section.trim()) continue

      const lines = section.trim().split('\n')
      const firstLine = lines[0]
      const match = firstLine.match(/\[(.+?)\]\s+(.+?):/)

      if (match) {
        messages.push({
          role: match[2].toLowerCase() === 'user' ? 'user' : 'assistant',
          content: lines.slice(1).join('\n'),
          timestamp: new Date(match[1]).toISOString(),
        })
      }
    }

    return messages
  }

  // 导入文件
  const handleImport = async () => {
    if (!file) return

    setStatus('uploading')
    setProgress(0)

    try {
      // 解析文件
      const data = await parseFile(file)
      setProgress(50)

      // 调用导入回调
      await onImport(data)
      setProgress(100)

      setStatus('success')
      setSuccess(`成功导入 ${Array.isArray(data) ? data.length : 1} 条消息`)

      // 2 秒后关闭对话框
      setTimeout(() => {
        onClose()
      }, 2000)
    } catch (err) {
      setStatus('error')
      setError(err instanceof Error ? err.message : '导入失败')
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-dark-800 rounded-lg max-w-md w-full mx-4 shadow-xl">
        {/* 头部 */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-dark-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            导入对话
          </h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 dark:hover:bg-dark-700 rounded"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 内容 */}
        <div className="p-6 space-y-4">
          {/* 文件上传区 */}
          <div
            className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-all ${
              file
                ? 'border-primary-600 bg-primary-50 dark:bg-dark-600'
                : 'border-gray-300 dark:border-dark-600 hover:border-primary-400'
            }`}
            onClick={() => fileInputRef.current?.click()}
          >
            <input
              ref={fileInputRef}
              type="file"
              onChange={handleFileChange}
              accept=".json,.txt,.csv"
              className="hidden"
            />

            {file ? (
              <div>
                <p className="font-medium text-gray-900 dark:text-white">
                  {file.name}
                </p>
                <p className="text-sm text-gray-500">
                  {(file.size / 1024).toFixed(2)} KB
                </p>
              </div>
            ) : (
              <>
                <Upload className="w-12 h-12 mx-auto text-gray-400 mb-2" />
                <p className="font-medium text-gray-900 dark:text-white mb-1">
                  选择文件或拖放
                </p>
                <p className="text-sm text-gray-500">
                  支持 JSON、TXT、CSV 格式 (最大 10MB)
                </p>
              </>
            )}
          </div>

          {/* 支持的格式 */}
          <div className="bg-gray-50 dark:bg-dark-700 p-4 rounded space-y-2 text-sm">
            <p className="font-medium text-gray-900 dark:text-white">
              支持的文件格式:
            </p>
            <ul className="text-gray-600 dark:text-gray-400 space-y-1">
              <li>✓ JSON - 保留所有元数据</li>
              <li>✓ CSV - 从表格导入</li>
              <li>✓ TXT - 从文本导入</li>
            </ul>
          </div>

          {/* 错误信息 */}
          {error && (
            <div className="bg-red-50 dark:bg-red-900 text-red-800 dark:text-red-100 p-3 rounded flex items-start gap-2">
              <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
              <p className="text-sm">{error}</p>
            </div>
          )}

          {/* 成功信息 */}
          {success && (
            <div className="bg-green-50 dark:bg-green-900 text-green-800 dark:text-green-100 p-3 rounded flex items-start gap-2">
              <CheckCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
              <p className="text-sm">{success}</p>
            </div>
          )}

          {/* 进度条 */}
          {status === 'uploading' && (
            <div>
              <div className="w-full bg-gray-200 dark:bg-dark-700 rounded-full h-2">
                <div
                  className="bg-primary-600 h-2 rounded-full transition-all"
                  style={{ width: `${progress}%` }}
                />
              </div>
              <p className="text-xs text-gray-500 text-center mt-2">
                {progress}%
              </p>
            </div>
          )}
        </div>

        {/* 底部操作 */}
        <div className="flex gap-3 p-4 border-t border-gray-200 dark:border-dark-700 bg-gray-50 dark:bg-dark-700 rounded-b-lg">
          <button
            onClick={onClose}
            disabled={status === 'uploading'}
            className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-dark-800 border border-gray-300 dark:border-dark-600 rounded hover:bg-gray-50 dark:hover:bg-dark-700 disabled:opacity-50"
          >
            取消
          </button>
          <button
            onClick={handleImport}
            disabled={!file || status === 'uploading' || status === 'success'}
            className="flex-1 px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded hover:bg-primary-700 disabled:opacity-50"
          >
            {status === 'uploading'
              ? `导入中 (${progress}%)`
              : status === 'success'
                ? '已完成'
                : '导入'}
          </button>
        </div>
      </div>
    </div>
  )
}

