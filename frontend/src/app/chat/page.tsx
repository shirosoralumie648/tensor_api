'use client'

import { useEffect, useState } from 'react'
import { SessionSidebar } from '@/components/SessionSidebar'
import { ChatBox } from '@/components/ChatBox'
import { useChatStore } from '@/stores/chatStore'

export default function ChatPage() {
  const { currentSessionId, setCurrentSession, sessions } = useChatStore()
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)

  // 如果没有选中会话，自动选择第一个或创建新的
  useEffect(() => {
    if (!currentSessionId && sessions.length > 0) {
      setCurrentSession(sessions[0].id)
    }
  }, [currentSessionId, sessions, setCurrentSession])

  return (
    <div className="flex h-screen bg-white dark:bg-dark-900">
      {/* 侧边栏 - 移动端隐藏 */}
      <div
        className={`${
          isSidebarOpen ? 'w-sidebar' : 'w-0'
        } transition-all duration-300 hidden md:block overflow-hidden`}
      >
        <SessionSidebar />
      </div>

      {/* 主容器 */}
      <div className="flex-1 flex flex-col">
        {/* 顶部栏 */}
        <div className="h-header border-b border-gray-200 dark:border-dark-700 flex items-center px-4 gap-4">
          {/* 移动端侧边栏切换 */}
          <button
            onClick={() => setIsSidebarOpen(!isSidebarOpen)}
            className="md:hidden p-2 hover:bg-gray-100 dark:hover:bg-dark-800 rounded"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 6h16M4 12h16M4 18h16"
              />
            </svg>
          </button>

          <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
            对话
          </h1>

          {/* 右侧操作区 */}
          <div className="ml-auto flex items-center gap-4">
            {/* 模型选择 */}
            <select
              defaultValue="gpt-3.5-turbo"
              className="input px-3 py-2 text-sm"
            >
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
              <option value="gpt-4">GPT-4</option>
              <option value="claude-3">Claude 3</option>
              <option value="gemini-pro">Gemini Pro</option>
            </select>

            {/* 用户菜单 */}
            <button className="flex items-center gap-2 px-3 py-2 hover:bg-gray-100 dark:hover:bg-dark-800 rounded">
              <div className="w-8 h-8 bg-primary-600 rounded-full flex items-center justify-center text-white text-sm font-semibold">
                U
              </div>
            </button>
          </div>
        </div>

        {/* 聊天主体 */}
        <ChatBox />
      </div>

      {/* 移动端侧边栏悬浮 */}
      {isSidebarOpen && (
        <div
          className="md:hidden fixed inset-0 bg-black bg-opacity-50 z-40"
          onClick={() => setIsSidebarOpen(false)}
        >
          <div
            className="absolute inset-y-0 left-0 w-sidebar bg-gray-50 dark:bg-dark-800"
            onClick={(e) => e.stopPropagation()}
          >
            <SessionSidebar />
          </div>
        </div>
      )}
    </div>
  )
}
