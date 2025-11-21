'use client'

import { useState } from 'react'
import { Plus, Trash2, Edit2, MoreVertical, Search } from 'lucide-react'
import { useChatStore, type Session } from '@/stores/chatStore'

export function SessionSidebar() {
  const { sessions, currentSessionId, setCurrentSession, addSession, removeSession, updateSession } = useChatStore()
  const [searchQuery, setSearchQuery] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [showContextMenu, setShowContextMenu] = useState<string | null>(null)

  // 过滤会话
  const filteredSessions = sessions.filter((s) =>
    s.title.toLowerCase().includes(searchQuery.toLowerCase())
  )

  // 创建新会话
  const handleNewSession = () => {
    const newSession: Session = {
      id: `session_${Date.now()}`,
      title: '新对话',
      model: 'gpt-3.5-turbo',
      createdAt: new Date(),
      updatedAt: new Date(),
      messageCount: 0,
    }
    addSession(newSession)
    setCurrentSession(newSession.id)
  }

  // 编辑会话标题
  const handleEditTitle = (session: Session) => {
    setEditingId(session.id)
    setEditTitle(session.title)
    setShowContextMenu(null)
  }

  // 保存编辑
  const handleSaveTitle = (sessionId: string) => {
    if (editTitle.trim()) {
      updateSession(sessionId, { title: editTitle })
    }
    setEditingId(null)
    setEditTitle('')
  }

  // 删除会话
  const handleDeleteSession = (sessionId: string) => {
    removeSession(sessionId)
    setShowContextMenu(null)
  }

  return (
    <div className="w-sidebar bg-gray-50 dark:bg-dark-800 border-r border-gray-200 dark:border-dark-700 flex flex-col h-full">
      {/* 头部 */}
      <div className="p-4 space-y-3">
        <button
          onClick={handleNewSession}
          className="btn btn-primary w-full flex items-center justify-center gap-2"
        >
          <Plus className="w-4 h-4" />
          新对话
        </button>

        {/* 搜索框 */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="搜索会话..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input pl-9"
          />
        </div>
      </div>

      {/* 会话列表 */}
      <div className="flex-1 overflow-y-auto scrollbar-custom">
        {filteredSessions.length === 0 ? (
          <div className="p-4 text-center text-gray-400 text-sm">
            {sessions.length === 0 ? '暂无对话' : '未找到匹配的对话'}
          </div>
        ) : (
          <div className="space-y-1 p-2">
            {filteredSessions.map((session) => (
              <div
                key={session.id}
                className={`group relative p-3 rounded-lg cursor-pointer transition-colors ${
                  currentSessionId === session.id
                    ? 'bg-primary-600 text-white'
                    : 'hover:bg-gray-200 dark:hover:bg-dark-700 text-gray-900 dark:text-gray-100'
                }`}
              >
                {editingId === session.id ? (
                  /* 编辑模式 */
                  <input
                    type="text"
                    autoFocus
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    onBlur={() => handleSaveTitle(session.id)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') handleSaveTitle(session.id)
                      if (e.key === 'Escape') setEditingId(null)
                    }}
                    className="w-full bg-transparent border-b border-current focus:outline-none"
                  />
                ) : (
                  /* 显示模式 */
                  <>
                    <div onClick={() => setCurrentSession(session.id)}>
                      <p className="font-medium truncate">{session.title}</p>
                      <p className="text-xs opacity-75">
                        {session.messageCount} 条消息
                      </p>
                    </div>

                    {/* 上下文菜单按钮 */}
                    <button
                      onClick={() =>
                        setShowContextMenu(
                          showContextMenu === session.id ? null : session.id
                        )
                      }
                      className="absolute right-2 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      <MoreVertical className="w-4 h-4" />
                    </button>

                    {/* 上下文菜单 */}
                    {showContextMenu === session.id && (
                      <div className="absolute right-0 top-full mt-1 bg-white dark:bg-dark-700 rounded-lg shadow-lg z-10 min-w-max">
                        <button
                          onClick={() => handleEditTitle(session)}
                          className="flex items-center gap-2 w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-600"
                        >
                          <Edit2 className="w-4 h-4" />
                          编辑
                        </button>
                        <button
                          onClick={() => handleDeleteSession(session.id)}
                          className="flex items-center gap-2 w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900"
                        >
                          <Trash2 className="w-4 h-4" />
                          删除
                        </button>
                      </div>
                    )}
                  </>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 底部信息 */}
      <div className="p-4 border-t border-gray-200 dark:border-dark-700 text-xs text-gray-500">
        <p>{sessions.length} 个对话</p>
      </div>
    </div>
  )
}

