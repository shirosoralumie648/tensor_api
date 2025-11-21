'use client'

import { useState } from 'react'
import {
  BarChart3,
  Key,
  TrendingUp,
  Settings,
  LogOut,
  Menu,
  X,
} from 'lucide-react'

type TabType = 'dashboard' | 'keys' | 'usage' | 'settings'

interface DeveloperConsoleProps {
  children?: React.ReactNode
  activeTab?: TabType
}

export function DeveloperConsole({ children, activeTab = 'dashboard' }: DeveloperConsoleProps) {
  const [isOpen, setIsOpen] = useState(true)
  const [currentTab, setCurrentTab] = useState<TabType>(activeTab)

  const navItems = [
    {
      id: 'dashboard',
      label: '仪表盘',
      icon: BarChart3,
      description: '查看概览统计',
    },
    {
      id: 'keys',
      label: 'API 密钥',
      icon: Key,
      description: '管理 API 密钥',
    },
    {
      id: 'usage',
      label: '使用统计',
      icon: TrendingUp,
      description: '查看使用情况',
    },
    {
      id: 'settings',
      label: '设置',
      icon: Settings,
      description: '应用设置',
    },
  ] as const

  return (
    <div className="flex h-screen bg-white dark:bg-dark-900">
      {/* 侧边栏 */}
      <div
        className={`${
          isOpen ? 'w-console-sidebar' : 'w-0'
        } transition-all duration-300 overflow-hidden bg-gray-50 dark:bg-dark-800 border-r border-gray-200 dark:border-dark-700 flex flex-col`}
      >
        {/* 侧边栏头部 */}
        <div className="p-4 border-b border-gray-200 dark:border-dark-700">
          <h1 className="text-lg font-bold text-gray-900 dark:text-white">
            开发者控制台
          </h1>
          <p className="text-xs text-gray-500 mt-1">管理您的 API</p>
        </div>

        {/* 导航菜单 */}
        <nav className="flex-1 p-4 space-y-2 overflow-y-auto">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <button
                key={item.id}
                onClick={() => setCurrentTab(item.id)}
                className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
                  currentTab === item.id
                    ? 'bg-primary-600 text-white'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-dark-700'
                }`}
              >
                <div className="flex items-center gap-3">
                  <Icon className="w-5 h-5" />
                  <div>
                    <p className="font-medium">{item.label}</p>
                    <p className="text-xs opacity-75">{item.description}</p>
                  </div>
                </div>
              </button>
            )
          })}
        </nav>

        {/* 侧边栏底部 */}
        <div className="p-4 border-t border-gray-200 dark:border-dark-700 space-y-2">
          <button className="w-full flex items-center gap-2 px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-dark-700 rounded transition-colors text-sm">
            <Settings className="w-4 h-4" />
            账户设置
          </button>
          <button className="w-full flex items-center gap-2 px-4 py-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900 rounded transition-colors text-sm">
            <LogOut className="w-4 h-4" />
            退出登录
          </button>
        </div>
      </div>

      {/* 主内容区 */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* 顶部栏 */}
        <div className="h-header border-b border-gray-200 dark:border-dark-700 flex items-center px-6 gap-4 bg-white dark:bg-dark-900">
          <button
            onClick={() => setIsOpen(!isOpen)}
            className="lg:hidden p-2 hover:bg-gray-100 dark:hover:bg-dark-800 rounded"
          >
            {isOpen ? (
              <X className="w-6 h-6" />
            ) : (
              <Menu className="w-6 h-6" />
            )}
          </button>

          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
            {navItems.find((item) => item.id === currentTab)?.label}
          </h2>

          {/* 右侧操作 */}
          <div className="ml-auto flex items-center gap-4">
            <button className="px-4 py-2 text-sm bg-primary-600 text-white rounded hover:bg-primary-700 transition-colors">
              帮助文档
            </button>
          </div>
        </div>

        {/* 主体内容 */}
        <div className="flex-1 overflow-y-auto">
          {children}
        </div>
      </div>

      {/* 移动端侧边栏悬浮 */}
      {isOpen && (
        <div
          className="lg:hidden fixed inset-0 bg-black bg-opacity-50 z-30"
          onClick={() => setIsOpen(false)}
        />
      )}
    </div>
  )
}

