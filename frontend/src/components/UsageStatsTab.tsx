'use client'

import { useState } from 'react'
import { Download, DateRange } from 'lucide-react'

interface UsageData {
  date: string
  requests: number
  tokens: number
  cost: number
}

export function UsageStatsTab() {
  const [startDate, setStartDate] = useState('2024-11-01')
  const [endDate, setEndDate] = useState('2024-11-21')
  const [timeRange, setTimeRange] = useState<'day' | 'week' | 'month'>('month')

  // 模拟数据
  const usageData: UsageData[] = [
    { date: '2024-11-21', requests: 2543, tokens: 125000, cost: 0.45 },
    { date: '2024-11-20', requests: 1890, tokens: 98000, cost: 0.35 },
    { date: '2024-11-19', requests: 2156, tokens: 108000, cost: 0.39 },
    { date: '2024-11-18', requests: 1654, tokens: 82000, cost: 0.30 },
    { date: '2024-11-17', requests: 2234, tokens: 112000, cost: 0.40 },
  ]

  const totalRequests = usageData.reduce((sum, d) => sum + d.requests, 0)
  const totalTokens = usageData.reduce((sum, d) => sum + d.tokens, 0)
  const totalCost = usageData.reduce((sum, d) => sum + d.cost, 0)
  const avgRequests = Math.round(totalRequests / usageData.length)

  const handleExport = () => {
    const csv = [
      ['日期', '请求数', 'Token 数', '成本'],
      ...usageData.map((d) => [d.date, d.requests, d.tokens, `$${d.cost.toFixed(2)}`]),
    ]
      .map((row) => row.join(','))
      .join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `usage-stats-${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      {/* 时间范围选择 */}
      <div className="bg-gray-50 dark:bg-dark-700 p-4 rounded-lg space-y-4">
        <div className="flex items-center gap-4">
          <div className="flex gap-2">
            {(['day', 'week', 'month'] as const).map((range) => (
              <button
                key={range}
                onClick={() => setTimeRange(range)}
                className={`px-3 py-1.5 text-sm rounded transition-colors ${
                  timeRange === range
                    ? 'bg-primary-600 text-white'
                    : 'bg-white dark:bg-dark-800 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-dark-700'
                }`}
              >
                {range === 'day' ? '日' : range === 'week' ? '周' : '月'}
              </button>
            ))}
          </div>

          {/* 自定义日期 */}
          <div className="flex items-center gap-2 ml-auto">
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="input px-3 py-1.5 text-sm"
            />
            <span className="text-gray-500">-</span>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="input px-3 py-1.5 text-sm"
            />
            <button
              onClick={handleExport}
              className="px-3 py-1.5 bg-primary-600 hover:bg-primary-700 text-white rounded text-sm flex items-center gap-1"
            >
              <Download className="w-4 h-4" />
              导出
            </button>
          </div>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {[
          { label: '总请求数', value: totalRequests.toLocaleString(), unit: '次' },
          { label: '总 Token 数', value: (totalTokens / 1000).toFixed(0), unit: 'K' },
          { label: '总成本', value: `$${totalCost.toFixed(2)}`, unit: '' },
          { label: '平均请求', value: avgRequests.toLocaleString(), unit: '次/天' },
        ].map((stat) => (
          <div
            key={stat.label}
            className="bg-white dark:bg-dark-800 border border-gray-200 dark:border-dark-700 rounded-lg p-4"
          >
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
              {stat.label}
            </p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white">
              {stat.value}
              <span className="text-sm text-gray-500 ml-1">{stat.unit}</span>
            </p>
          </div>
        ))}
      </div>

      {/* 模型分布 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          模型使用分布
        </h3>
        <div className="space-y-3">
          {[
            { model: 'GPT-4', percentage: 45, requests: 1100 },
            { model: 'GPT-3.5 Turbo', percentage: 35, requests: 850 },
            { model: 'Claude 3', percentage: 15, requests: 370 },
            { model: 'Gemini', percentage: 5, requests: 120 },
          ].map((item) => (
            <div key={item.model}>
              <div className="flex items-center justify-between mb-1">
                <p className="text-sm font-medium text-gray-900 dark:text-white">
                  {item.model}
                </p>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  {item.percentage}% ({item.requests.toLocaleString()} 次)
                </p>
              </div>
              <div className="w-full bg-gray-200 dark:bg-dark-700 rounded-full h-2">
                <div
                  className="bg-primary-600 h-2 rounded-full transition-all"
                  style={{ width: `${item.percentage}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 使用详情表 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          详细记录
        </h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th className="px-4 py-3 text-left text-gray-700 dark:text-gray-300 font-medium">
                  日期
                </th>
                <th className="px-4 py-3 text-left text-gray-700 dark:text-gray-300 font-medium">
                  请求数
                </th>
                <th className="px-4 py-3 text-left text-gray-700 dark:text-gray-300 font-medium">
                  Token 数
                </th>
                <th className="px-4 py-3 text-left text-gray-700 dark:text-gray-300 font-medium">
                  成本
                </th>
                <th className="px-4 py-3 text-left text-gray-700 dark:text-gray-300 font-medium">
                  平均延迟
                </th>
              </tr>
            </thead>
            <tbody>
              {usageData.map((data) => (
                <tr
                  key={data.date}
                  className="border-t border-gray-200 dark:border-dark-700 hover:bg-gray-50 dark:hover:bg-dark-700"
                >
                  <td className="px-4 py-3">{data.date}</td>
                  <td className="px-4 py-3">{data.requests.toLocaleString()}</td>
                  <td className="px-4 py-3">{(data.tokens / 1000).toFixed(0)}K</td>
                  <td className="px-4 py-3">${data.cost.toFixed(2)}</td>
                  <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                    {Math.random() * 200 + 100 | 0}ms
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

