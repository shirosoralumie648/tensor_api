/**
 * 格式化工具函数
 */

// 格式化时间
export function formatTime(date: Date | string, format = 'HH:mm:ss'): string {
  const d = typeof date === 'string' ? new Date(date) : date

  const pad = (n: number) => String(n).padStart(2, '0')

  const formats: Record<string, string> = {
    'YYYY': String(d.getFullYear()),
    'MM': pad(d.getMonth() + 1),
    'DD': pad(d.getDate()),
    'HH': pad(d.getHours()),
    'mm': pad(d.getMinutes()),
    'ss': pad(d.getSeconds()),
  }

  let result = format
  for (const [key, value] of Object.entries(formats)) {
    result = result.replace(key, value)
  }

  return result
}

// 相对时间
export function formatRelativeTime(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffSec < 60) {
    return '刚刚'
  } else if (diffMin < 60) {
    return `${diffMin}分钟前`
  } else if (diffHour < 24) {
    return `${diffHour}小时前`
  } else if (diffDay < 7) {
    return `${diffDay}天前`
  } else {
    return formatTime(d, 'MM-DD HH:mm')
  }
}

// 格式化文件大小
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化数字
export function formatNumber(num: number, decimals = 2): string {
  return num.toFixed(decimals)
}

// 格式化货币
export function formatCurrency(amount: number, currency = 'CNY'): string {
  const symbols: Record<string, string> = {
    'CNY': '¥',
    'USD': '$',
    'EUR': '€',
    'GBP': '£',
  }

  const symbol = symbols[currency] || currency
  return `${symbol}${amount.toFixed(2)}`
}

// 格式化百分比
export function formatPercent(value: number, decimals = 1): string {
  return `${(value * 100).toFixed(decimals)}%`
}

// 格式化 Token 数量
export function formatTokens(tokens: number): string {
  if (tokens < 1000) {
    return `${tokens} tokens`
  }
  return `${(tokens / 1000).toFixed(1)}k tokens`
}

// 格式化代码 (支持 Markdown)
export function formatCode(code: string, language?: string): string {
  if (!language) {
    return code
  }

  return `\`\`\`${language}\n${code}\n\`\`\``
}

// 截断文本
export function truncateText(text: string, length: number, suffix = '...'): string {
  if (text.length <= length) {
    return text
  }
  return text.slice(0, length) + suffix
}

// 首字母大写
export function capitalize(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1)
}

// 蛇形命名转驼峰
export function toCamelCase(str: string): string {
  return str.replace(/_([a-z])/g, (g) => g[1].toUpperCase())
}

// 驼峰转蛇形
export function toSnakeCase(str: string): string {
  return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
}

// 格式化错误消息
export function formatErrorMessage(error: any): string {
  if (typeof error === 'string') {
    return error
  }

  if (error?.response?.data?.message) {
    return error.response.data.message
  }

  if (error?.message) {
    return error.message
  }

  return '发生未知错误'
}

