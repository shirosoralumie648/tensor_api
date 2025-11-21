'use client'

import { useState } from 'react'
import { Copy, Check, Link2, X, Lock, Globe } from 'lucide-react'

interface SessionShareDialogProps {
  sessionId: string
  sessionTitle: string
  onClose: () => void
}

type ShareLevel = 'private' | 'link' | 'public'

export function SessionShareDialog({
  sessionId,
  sessionTitle,
  onClose,
}: SessionShareDialogProps) {
  const [shareLevel, setShareLevel] = useState<ShareLevel>('link')
  const [copied, setCopied] = useState(false)
  const [expiryDays, setExpiryDays] = useState(7)
  const [allowComments, setAllowComments] = useState(true)
  const [loading, setLoading] = useState(false)

  // 生成分享链接
  const shareUrl = `${typeof window !== 'undefined' ? window.location.origin : ''}/shared/${sessionId}?token=${btoa(sessionId)}`

  // 复制链接
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      console.error('复制失败')
    }
  }

  // 生成分享
  const handleShare = async () => {
    setLoading(true)
    try {
      // 调用 API 生成分享链接
      const response = await fetch('/api/sessions/share', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sessionId,
          shareLevel,
          expiryDays: shareLevel === 'link' ? expiryDays : null,
          allowComments,
        }),
      })

      if (response.ok) {
        const data = await response.json()
        console.log('分享链接已生成:', data)
        // 显示成功提示
      }
    } finally {
      setLoading(false)
    }
  }

  // 分享到社交媒体
  const handleShareToSocial = (platform: 'twitter' | 'wechat') => {
    const text = `我想分享这个 AI 对话: "${sessionTitle}"`
    const urls = {
      twitter: `https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}&url=${encodeURIComponent(shareUrl)}`,
      wechat: `weixin://profile/wxhb://wxpay?action=openapp&appid=wx...&url=${encodeURIComponent(shareUrl)}`,
    }
    window.open(urls[platform], '_blank', 'width=600,height=400')
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-dark-800 rounded-lg max-w-2xl w-full mx-4 shadow-xl">
        {/* 头部 */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-dark-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            分享对话
          </h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 dark:hover:bg-dark-700 rounded"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 内容 */}
        <div className="p-6 space-y-6">
          {/* 分享级别选择 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
              分享级别
            </label>
            <div className="space-y-2">
              {(
                [
                  {
                    value: 'private' as ShareLevel,
                    label: '仅自己可见',
                    icon: Lock,
                    desc: '不分享给任何人',
                  },
                  {
                    value: 'link' as ShareLevel,
                    label: '链接分享',
                    icon: Link2,
                    desc: '有链接的人可以查看',
                  },
                  {
                    value: 'public' as ShareLevel,
                    label: '公开分享',
                    icon: Globe,
                    desc: '任何人都可以找到',
                  },
                ] as const
              ).map(({ value, label, icon: Icon, desc }) => (
                <label
                  key={value}
                  className={`flex items-center p-3 border-2 rounded-lg cursor-pointer transition-all ${
                    shareLevel === value
                      ? 'border-primary-600 bg-primary-50 dark:bg-dark-600'
                      : 'border-gray-200 dark:border-dark-600 hover:border-primary-300'
                  }`}
                >
                  <input
                    type="radio"
                    name="shareLevel"
                    value={value}
                    checked={shareLevel === value}
                    onChange={(e) => setShareLevel(e.target.value as ShareLevel)}
                    className="w-4 h-4"
                  />
                  <Icon className="w-5 h-5 mx-3 text-gray-600 dark:text-gray-400" />
                  <div className="flex-1">
                    <p className="font-medium text-sm">{label}</p>
                    <p className="text-xs text-gray-500">{desc}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* 链接分享选项 */}
          {shareLevel === 'link' && (
            <>
              {/* 有效期 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  链接有效期
                </label>
                <select
                  value={expiryDays}
                  onChange={(e) => setExpiryDays(Number(e.target.value))}
                  className="input w-full"
                >
                  <option value={1}>1 天</option>
                  <option value={7}>7 天</option>
                  <option value={30}>30 天</option>
                  <option value={0}>永不过期</option>
                </select>
              </div>

              {/* 权限设置 */}
              <div>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={allowComments}
                    onChange={(e) => setAllowComments(e.target.checked)}
                    className="w-4 h-4"
                  />
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    允许访问者评论
                  </span>
                </label>
              </div>
            </>
          )}

          {/* 分享链接 */}
          {shareLevel !== 'private' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                分享链接
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={shareUrl}
                  readOnly
                  className="input flex-1 bg-gray-50 dark:bg-dark-700"
                />
                <button
                  onClick={handleCopy}
                  className="px-4 py-2 bg-gray-100 dark:bg-dark-700 hover:bg-gray-200 dark:hover:bg-dark-600 rounded transition-colors flex items-center gap-2"
                >
                  {copied ? (
                    <>
                      <Check className="w-4 h-4 text-green-500" />
                      已复制
                    </>
                  ) : (
                    <>
                      <Copy className="w-4 h-4" />
                      复制
                    </>
                  )}
                </button>
              </div>
            </div>
          )}

          {/* 社交分享 */}
          {shareLevel !== 'private' && (
            <div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                分享到社交媒体
              </p>
              <div className="flex gap-3">
                <button
                  onClick={() => handleShareToSocial('twitter')}
                  className="flex-1 py-2 px-4 bg-blue-500 hover:bg-blue-600 text-white rounded transition-colors text-sm font-medium"
                >
                  Twitter
                </button>
                <button
                  onClick={() => handleShareToSocial('wechat')}
                  className="flex-1 py-2 px-4 bg-green-500 hover:bg-green-600 text-white rounded transition-colors text-sm font-medium"
                >
                  微信
                </button>
              </div>
            </div>
          )}

          {/* 统计信息 */}
          <div className="bg-gray-50 dark:bg-dark-700 p-4 rounded">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">访问次数</p>
                <p className="text-lg font-semibold text-gray-900 dark:text-white">
                  0
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">最后访问</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">从未</p>
              </div>
            </div>
          </div>
        </div>

        {/* 底部操作 */}
        <div className="flex gap-3 p-4 border-t border-gray-200 dark:border-dark-700 bg-gray-50 dark:bg-dark-700 rounded-b-lg">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-dark-800 border border-gray-300 dark:border-dark-600 rounded hover:bg-gray-50 dark:hover:bg-dark-700"
          >
            关闭
          </button>
          <button
            onClick={handleShare}
            disabled={loading}
            className="flex-1 px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded hover:bg-primary-700 disabled:opacity-50"
          >
            {loading ? '处理中...' : '保存设置'}
          </button>
        </div>
      </div>
    </div>
  )
}

