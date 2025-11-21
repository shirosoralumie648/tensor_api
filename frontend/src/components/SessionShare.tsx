'use client'

import { useState } from 'react'
import { Share2, Copy, Check, X, Eye, Lock, Link as LinkIcon } from 'lucide-react'
import { useChatStore } from '@/stores/chatStore'

interface ShareLink {
  id: string
  url: string
  permission: 'view' | 'comment'
  expiresAt: Date | null
  createdAt: Date
}

export function SessionShare() {
  const [isOpen, setIsOpen] = useState(false)
  const [copied, setCopied] = useState(false)
  const [shareLinks, setShareLinks] = useState<ShareLink[]>([])
  const [permission, setPermission] = useState<'view' | 'comment'>('view')
  const [expiresIn, setExpiresIn] = useState<'never' | '1h' | '1d' | '7d'>('never')
  const { currentSession } = useChatStore()

  // 生成分享链接
  const generateShareLink = () => {
    const id = `share_${Date.now()}`
    let expiresAt = null

    if (expiresIn !== 'never') {
      const now = new Date()
      if (expiresIn === '1h') now.setHours(now.getHours() + 1)
      else if (expiresIn === '1d') now.setDate(now.getDate() + 1)
      else if (expiresIn === '7d') now.setDate(now.getDate() + 7)
      expiresAt = now
    }

    const newLink: ShareLink = {
      id,
      url: `${window.location.origin}/share/${id}`,
      permission,
      expiresAt,
      createdAt: new Date(),
    }

    setShareLinks([newLink, ...shareLinks])
  }

  // 复制链接
  const copyLink = (url: string) => {
    navigator.clipboard.writeText(url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  // 删除分享链接
  const deleteLink = (id: string) => {
    setShareLinks(shareLinks.filter((link) => link.id !== id))
  }

  // 检查链接是否过期
  const isLinkExpired = (expiresAt: Date | null) => {
    if (!expiresAt) return false
    return new Date() > expiresAt
  }

  // 格式化时间
  const formatTime = (date: Date) => {
    return date.toLocaleString()
  }

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 hover:bg-gray-100 dark:hover:bg-dark-700 rounded transition-colors"
        title="分享会话"
      >
        <Share2 className="w-4 h-4" />
        <span className="text-sm">分享</span>
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full mt-2 bg-white dark:bg-dark-700 rounded-lg shadow-lg z-10 w-96">
          <div className="p-4 border-b border-gray-200 dark:border-dark-600 flex items-center justify-between">
            <h3 className="font-semibold">分享会话</h3>
            <button
              onClick={() => setIsOpen(false)}
              className="p-1 hover:bg-gray-100 dark:hover:bg-dark-600 rounded"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* 生成分享链接 */}
          <div className="p-4 space-y-3 border-b border-gray-200 dark:border-dark-600">
            <div>
              <label className="block text-sm font-medium mb-2">权限设置</label>
              <select
                value={permission}
                onChange={(e) => setPermission(e.target.value as 'view' | 'comment')}
                className="input w-full text-sm"
              >
                <option value="view">仅查看</option>
                <option value="comment">可评论</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">过期时间</label>
              <select
                value={expiresIn}
                onChange={(e) => setExpiresIn(e.target.value as any)}
                className="input w-full text-sm"
              >
                <option value="never">永不过期</option>
                <option value="1h">1小时</option>
                <option value="1d">1天</option>
                <option value="7d">7天</option>
              </select>
            </div>

            <button
              onClick={generateShareLink}
              className="btn btn-primary w-full flex items-center justify-center gap-2"
            >
              <LinkIcon className="w-4 h-4" />
              生成分享链接
            </button>
          </div>

          {/* 分享链接列表 */}
          <div className="p-4">
            {shareLinks.length === 0 ? (
              <p className="text-sm text-gray-500 text-center">还没有分享链接</p>
            ) : (
              <div className="space-y-3 max-h-64 overflow-y-auto">
                {shareLinks.map((link) => {
                  const expired = isLinkExpired(link.expiresAt)

                  return (
                    <div
                      key={link.id}
                      className={`p-3 rounded-lg border ${
                        expired
                          ? 'bg-red-50 dark:bg-red-900 border-red-200 dark:border-red-700'
                          : 'bg-gray-50 dark:bg-dark-600 border-gray-200 dark:border-dark-500'
                      }`}
                    >
                      {/* 链接头部 */}
                      <div className="flex items-start justify-between mb-2">
                        <div>
                          <div className="flex items-center gap-2">
                            {link.permission === 'view' ? (
                              <Eye className="w-4 h-4 text-blue-500" />
                            ) : (
                              <Eye className="w-4 h-4 text-green-500" />
                            )}
                            <span className="text-xs font-semibold">
                              {link.permission === 'view' ? '仅查看' : '可评论'}
                            </span>
                          </div>
                          <p className="text-xs text-gray-500 mt-1">
                            创建: {formatTime(link.createdAt)}
                          </p>
                        </div>

                        <button
                          onClick={() => deleteLink(link.id)}
                          className="p-1 hover:bg-red-100 dark:hover:bg-red-800 rounded text-red-600"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </div>

                      {/* 过期状态 */}
                      {expired && (
                        <div className="mb-2 p-2 bg-red-100 dark:bg-red-800 text-red-700 dark:text-red-100 text-xs rounded">
                          🔒 已过期
                        </div>
                      )}

                      {link.expiresAt && !expired && (
                        <div className="mb-2 p-2 bg-yellow-50 dark:bg-yellow-900 text-yellow-700 dark:text-yellow-100 text-xs rounded">
                          ⏰ 过期: {formatTime(link.expiresAt)}
                        </div>
                      )}

                      {!link.expiresAt && (
                        <div className="mb-2 p-2 bg-blue-50 dark:bg-blue-900 text-blue-700 dark:text-blue-100 text-xs rounded">
                          ♾️ 永不过期
                        </div>
                      )}

                      {/* 链接和复制按钮 */}
                      <div className="flex gap-2">
                        <input
                          type="text"
                          value={link.url}
                          readOnly
                          className="input flex-1 text-xs py-1"
                        />
                        <button
                          onClick={() => copyLink(link.url)}
                          className="px-2 py-1 bg-primary-600 hover:bg-primary-700 text-white rounded text-xs flex items-center gap-1"
                        >
                          {copied ? (
                            <Check className="w-3 h-3" />
                          ) : (
                            <Copy className="w-3 h-3" />
                          )}
                          {copied ? '已复制' : '复制'}
                        </button>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>

          {/* 底部提示 */}
          <div className="p-4 bg-blue-50 dark:bg-blue-900 border-t border-gray-200 dark:border-dark-600">
            <p className="text-xs text-blue-700 dark:text-blue-100">
              💡 分享的链接可以让其他人访问这个对话。
            </p>
          </div>
        </div>
      )}
    </div>
  )
}

