'use client'

import { useState } from 'react'
import { BarChart3, Key, TrendingUp, Wallet } from 'lucide-react'
import { TokenManagementTab } from '@/components/TokenManagementTab'
import { UsageStatsTab } from '@/components/UsageStatsTab'

export default function DeveloperPage() {
  const [activeTab, setActiveTab] = useState<'overview' | 'tokens' | 'usage' | 'billing'>('overview')

  const stats = [
    {
      label: '总 API 调用',
      value: '2,543,210',
      change: '+12.5%',
      icon: BarChart3,
    },
    {
      label: '当月花费',
      value: '$1,234.56',
      change: '+5.2%',
      icon: Wallet,
    },
    {
      label: '活跃密钥',
      value: '5',
      change: '-1',
      icon: Key,
    },
    {
      label: '平均延迟',
      value: '245ms',
      change: '-8.3%',
      icon: TrendingUp,
    },
  ]

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-dark-900">
      {/* 头部 */}
      <div className="bg-white dark:bg-dark-800 border-b border-gray-200 dark:border-dark-700">
        <div className="max-w-7xl mx-auto px-4 py-6">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            开发者控制台
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            管理 API 密钥、查看使用统计和账单信息
          </p>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {stats.map((stat) => {
            const Icon = stat.icon
            return (
              <div
                key={stat.label}
                className="bg-white dark:bg-dark-800 rounded-lg p-6 shadow-sm border border-gray-200 dark:border-dark-700"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {stat.label}
                    </p>
                    <p className="text-2xl font-bold text-gray-900 dark:text-white mt-2">
                      {stat.value}
                    </p>
                  </div>
                  <Icon className="w-8 h-8 text-primary-600 opacity-20" />
                </div>
                <p
                  className={`text-sm mt-4 ${
                    stat.change.startsWith('+')
                      ? 'text-red-600'
                      : 'text-green-600'
                  }`}
                >
                  {stat.change}
                </p>
              </div>
            )
          })}
        </div>
      </div>

      {/* 主容器 */}
      <div className="max-w-7xl mx-auto px-4 pb-12">
        {/* 导航标签 */}
        <div className="bg-white dark:bg-dark-800 rounded-lg shadow-sm border border-gray-200 dark:border-dark-700 mb-6">
          <div className="flex border-b border-gray-200 dark:border-dark-700">
            {(
              [
                { id: 'overview', label: '概览' },
                { id: 'tokens', label: 'API 密钥' },
                { id: 'usage', label: '使用统计' },
                { id: 'billing', label: '账单' },
              ] as const
            ).map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex-1 py-4 px-6 text-center font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'text-primary-600 border-b-2 border-primary-600'
                    : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-300'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* 标签内容 */}
          <div className="p-6">
            {activeTab === 'overview' && <OverviewTab />}
            {activeTab === 'tokens' && <TokenManagementTab />}
            {activeTab === 'usage' && <UsageStatsTab />}
            {activeTab === 'billing' && <BillingTab />}
          </div>
        </div>
      </div>
    </div>
  )
}

// 概览标签
function OverviewTab() {
  return (
    <div className="space-y-6">
      {/* 快速开始 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          快速开始
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            {
              title: '获取 API 密钥',
              description: '前往密钥管理页面创建新的 API 密钥',
              action: '前往',
            },
            {
              title: '阅读文档',
              description: '查看完整的 API 文档和代码示例',
              action: '查看',
            },
            {
              title: '加入社区',
              description: '加入 Discord 社区获取支持和讨论',
              action: '加入',
            },
          ].map((item) => (
            <div
              key={item.title}
              className="border border-gray-200 dark:border-dark-700 rounded-lg p-4 hover:shadow-lg transition-shadow"
            >
              <h4 className="font-medium text-gray-900 dark:text-white mb-2">
                {item.title}
              </h4>
              <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                {item.description}
              </p>
              <button className="text-sm text-primary-600 hover:text-primary-700 font-medium">
                {item.action} →
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* 最近活动 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          最近活动
        </h3>
        <div className="space-y-3">
          {[
            { action: '创建新密钥', time: '2小时前' },
            { action: 'API 调用 10,234 次', time: '1天前' },
            { action: '更新密钥权限', time: '3天前' },
          ].map((activity, i) => (
            <div
              key={i}
              className="flex items-center justify-between py-3 border-b border-gray-200 dark:border-dark-700 last:border-b-0"
            >
              <p className="text-gray-900 dark:text-white">{activity.action}</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                {activity.time}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}


// 账单标签
function BillingTab() {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
        账单详情
      </h3>
      <p className="text-gray-600 dark:text-gray-400">
        账单信息和发票将在这里显示
      </p>
    </div>
  )
}
