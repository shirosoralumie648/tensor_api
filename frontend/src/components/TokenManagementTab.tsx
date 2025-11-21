'use client'

import { useState } from 'react'
import { Copy, Eye, EyeOff, Trash2, Edit2, Plus } from 'lucide-react'

interface Token {
  id: string
  name: string
  key: string
  displayKey: string
  status: 'active' | 'inactive' | 'expired'
  createdAt: string
  lastUsed: string
  permissions: string[]
  rateLimit: number
}

export function TokenManagementTab() {
  const [tokens, setTokens] = useState<Token[]>([
    {
      id: '1',
      name: 'Production Key',
      key: 'sk-prod-xxxxxxxxxxxxx',
      displayKey: 'sk-prod-***xxx',
      status: 'active',
      createdAt: '2024-01-15',
      lastUsed: '5分钟前',
      permissions: ['chat', 'embedding'],
      rateLimit: 10000,
    },
    {
      id: '2',
      name: 'Development Key',
      key: 'sk-dev-yyyyyyyyyyyyyy',
      displayKey: 'sk-dev-***yyy',
      status: 'active',
      createdAt: '2024-02-01',
      lastUsed: '1小时前',
      permissions: ['chat', 'image'],
      rateLimit: 1000,
    },
  ])

  const [showCreateModal, setShowCreateModal] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set())

  // 复制密钥
  const handleCopy = (key: string, id: string) => {
    navigator.clipboard.writeText(key)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  // 切换密钥显示
  const toggleKeyVisibility = (id: string) => {
    const newVisible = new Set(visibleKeys)
    if (newVisible.has(id)) {
      newVisible.delete(id)
    } else {
      newVisible.add(id)
    }
    setVisibleKeys(newVisible)
  }

  // 删除密钥
  const handleDelete = (id: string) => {
    if (confirm('确定要删除此密钥吗？此操作无法撤销。')) {
      setTokens(tokens.filter((t) => t.id !== id))
    }
  }

  return (
    <div className="space-y-4">
      {/* 头部 */}
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
          API 密钥管理
        </h3>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn btn-primary px-4 py-2 text-sm flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          创建新密钥
        </button>
      </div>

      {/* 密钥列表 */}
      <div className="space-y-3">
        {tokens.map((token) => (
          <div
            key={token.id}
            className="bg-gray-50 dark:bg-dark-700 border border-gray-200 dark:border-dark-600 rounded-lg p-4"
          >
            {/* 顶部：名称 + 状态 */}
            <div className="flex items-start justify-between mb-3">
              <div>
                <h4 className="font-medium text-gray-900 dark:text-white">
                  {token.name}
                </h4>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  创建于 {token.createdAt}
                </p>
              </div>
              <span
                className={`px-2 py-1 rounded text-xs font-medium ${
                  token.status === 'active'
                    ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
                    : 'bg-gray-200 text-gray-800 dark:bg-dark-600 dark:text-gray-200'
                }`}
              >
                {token.status === 'active' ? '活跃' : '非活跃'}
              </span>
            </div>

            {/* 密钥 */}
            <div className="bg-white dark:bg-dark-800 rounded p-3 mb-3 flex items-center justify-between">
              <code className="text-sm text-gray-700 dark:text-gray-300 font-mono">
                {visibleKeys.has(token.id) ? token.key : token.displayKey}
              </code>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => toggleKeyVisibility(token.id)}
                  className="p-1 hover:bg-gray-100 dark:hover:bg-dark-700 rounded"
                >
                  {visibleKeys.has(token.id) ? (
                    <EyeOff className="w-4 h-4 text-gray-500" />
                  ) : (
                    <Eye className="w-4 h-4 text-gray-500" />
                  )}
                </button>
                <button
                  onClick={() => handleCopy(token.key, token.id)}
                  className="p-1 hover:bg-gray-100 dark:hover:bg-dark-700 rounded"
                >
                  <Copy className="w-4 h-4 text-gray-500" />
                </button>
              </div>
            </div>

            {/* 权限和限流 */}
            <div className="grid grid-cols-2 gap-4 mb-3">
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                  权限
                </p>
                <div className="flex flex-wrap gap-1">
                  {token.permissions.map((perm) => (
                    <span
                      key={perm}
                      className="px-2 py-1 bg-primary-100 dark:bg-primary-900 text-primary-700 dark:text-primary-100 rounded text-xs"
                    >
                      {perm}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">
                  速率限制
                </p>
                <p className="text-sm font-medium text-gray-900 dark:text-white">
                  {token.rateLimit.toLocaleString()} req/min
                </p>
              </div>
            </div>

            {/* 最后使用 */}
            <div className="mb-3 text-sm text-gray-600 dark:text-gray-400">
              最后使用: {token.lastUsed}
            </div>

            {/* 操作按钮 */}
            <div className="flex gap-2 pt-3 border-t border-gray-200 dark:border-dark-600">
              <button className="flex-1 px-3 py-2 text-sm bg-white dark:bg-dark-800 hover:bg-gray-50 dark:hover:bg-dark-700 border border-gray-300 dark:border-dark-600 rounded flex items-center justify-center gap-1 text-gray-700 dark:text-gray-300">
                <Edit2 className="w-4 h-4" />
                编辑
              </button>
              <button
                onClick={() => handleDelete(token.id)}
                className="flex-1 px-3 py-2 text-sm bg-red-50 dark:bg-red-900 hover:bg-red-100 dark:hover:bg-red-800 border border-red-200 dark:border-red-800 rounded flex items-center justify-center gap-1 text-red-700 dark:text-red-100"
              >
                <Trash2 className="w-4 h-4" />
                删除
              </button>
            </div>
          </div>
        ))}
      </div>

      {tokens.length === 0 && (
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <p>暂无 API 密钥</p>
          <p className="text-sm">点击"创建新密钥"开始</p>
        </div>
      )}

      {/* 创建密钥模态框 */}
      {showCreateModal && (
        <CreateTokenModal onClose={() => setShowCreateModal(false)} />
      )}
    </div>
  )
}

// 创建密钥模态框
function CreateTokenModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('')
  const [permissions, setPermissions] = useState<string[]>(['chat'])

  const availablePermissions = [
    { id: 'chat', label: 'Chat API' },
    { id: 'embedding', label: 'Embedding API' },
    { id: 'image', label: 'Image API' },
    { id: 'audio', label: 'Audio API' },
  ]

  const handlePermissionToggle = (perm: string) => {
    setPermissions((prev) =>
      prev.includes(perm) ? prev.filter((p) => p !== perm) : [...prev, perm]
    )
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-dark-800 rounded-lg max-w-md w-full mx-4 shadow-xl">
        <div className="p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            创建新密钥
          </h2>

          {/* 密钥名称 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              密钥名称
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如：Production API Key"
              className="input w-full"
            />
          </div>

          {/* 权限选择 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              权限
            </label>
            <div className="space-y-2">
              {availablePermissions.map((perm) => (
                <label
                  key={perm.id}
                  className="flex items-center gap-2 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={permissions.includes(perm.id)}
                    onChange={() => handlePermissionToggle(perm.id)}
                    className="w-4 h-4"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    {perm.label}
                  </span>
                </label>
              ))}
            </div>
          </div>

          {/* 提示 */}
          <div className="bg-blue-50 dark:bg-blue-900 text-blue-900 dark:text-blue-100 p-3 rounded text-sm">
            ⚠️ 密钥一旦创建就无法再次查看，请妥善保存
          </div>
        </div>

        {/* 操作按钮 */}
        <div className="flex gap-3 p-4 border-t border-gray-200 dark:border-dark-700 bg-gray-50 dark:bg-dark-700 rounded-b-lg">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-dark-800 border border-gray-300 dark:border-dark-600 rounded hover:bg-gray-50"
          >
            取消
          </button>
          <button
            onClick={() => {
              // 创建密钥逻辑
              onClose()
            }}
            className="flex-1 px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded hover:bg-primary-700 disabled:opacity-50"
            disabled={!name}
          >
            创建
          </button>
        </div>
      </div>
    </div>
  )
}

