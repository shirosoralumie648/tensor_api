'use client'

import Link from 'next/link'
import { ArrowRight, Zap, Shield, Gauge } from 'lucide-react'

export default function Home() {
  return (
    <div className="w-full min-h-screen">
      {/* 导航栏 */}
      <nav className="sticky top-0 z-50 bg-white dark:bg-dark-900 border-b border-gray-200 dark:border-dark-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Zap className="w-6 h-6 text-primary-600" />
            <span className="text-xl font-bold text-gray-900 dark:text-white">Oblivious AI</span>
          </div>
          <div className="flex items-center gap-4">
            <Link href="/chat" className="btn btn-primary">
              开始使用
              <ArrowRight className="w-4 h-4 ml-2" />
            </Link>
          </div>
        </div>
      </nav>

      {/* 主要内容 */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* 英雄部分 */}
        <section className="text-center py-12 md:py-20">
          <h1 className="text-4xl md:text-6xl font-bold text-gray-900 dark:text-white mb-6">
            企业级 AI API 中转平台
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-400 mb-8 max-w-2xl mx-auto">
            支持多个 AI 模型提供商的统一接入、完整管理系统和企业级功能
          </p>
          <div className="flex justify-center gap-4">
            <Link href="/chat" className="btn btn-primary px-8 py-3 text-lg">
              进入对话
            </Link>
            <Link href="/console" className="btn btn-secondary px-8 py-3 text-lg">
              开发者控制台
            </Link>
          </div>
        </section>

        {/* 特性部分 */}
        <section className="grid grid-cols-1 md:grid-cols-3 gap-8 py-16">
          <FeatureCard
            icon={<Zap className="w-8 h-8" />}
            title="多模型支持"
            description="支持 OpenAI、Claude、Gemini 等多个 AI 提供商的统一接入"
          />
          <FeatureCard
            icon={<Shield className="w-8 h-8" />}
            title="企业级安全"
            description="API 密钥管理、权限控制、访问审计等完整的安全体系"
          />
          <FeatureCard
            icon={<Gauge className="w-8 h-8" />}
            title="完整管理系统"
            description="实时统计、使用分析、限流配额、成本账单等管理功能"
          />
        </section>

        {/* CTA 部分 */}
        <section className="bg-gradient-primary rounded-lg p-12 text-center text-white my-16">
          <h2 className="text-3xl font-bold mb-4">准备好开始了吗？</h2>
          <p className="text-lg opacity-90 mb-8">
            立即注册使用 Oblivious AI，体验企业级 AI API 中转服务
          </p>
          <Link href="/chat" className="inline-flex items-center gap-2 bg-white text-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-100 transition-colors">
            立即开始
            <ArrowRight className="w-5 h-5" />
          </Link>
        </section>
      </main>

      {/* 页脚 */}
      <footer className="bg-gray-50 dark:bg-dark-950 border-t border-gray-200 dark:border-dark-700 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
            <div>
              <h3 className="font-semibold text-gray-900 dark:text-white mb-4">产品</h3>
              <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                <li><a href="#" className="hover:text-primary-600">对话</a></li>
                <li><a href="#" className="hover:text-primary-600">知识库</a></li>
                <li><a href="#" className="hover:text-primary-600">插件</a></li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 dark:text-white mb-4">开发者</h3>
              <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                <li><a href="#" className="hover:text-primary-600">API 文档</a></li>
                <li><a href="#" className="hover:text-primary-600">部署指南</a></li>
                <li><a href="#" className="hover:text-primary-600">示例代码</a></li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 dark:text-white mb-4">资源</h3>
              <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                <li><a href="#" className="hover:text-primary-600">博客</a></li>
                <li><a href="#" className="hover:text-primary-600">常见问题</a></li>
                <li><a href="#" className="hover:text-primary-600">联系我们</a></li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 dark:text-white mb-4">法律</h3>
              <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                <li><a href="#" className="hover:text-primary-600">服务条款</a></li>
                <li><a href="#" className="hover:text-primary-600">隐私政策</a></li>
                <li><a href="#" className="hover:text-primary-600">安全政策</a></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-gray-200 dark:border-dark-700 pt-8 flex justify-between items-center text-sm text-gray-600 dark:text-gray-400">
            <p>&copy; 2024 Oblivious AI. 保留所有权利。</p>
            <p>Made with ❤️ for developers</p>
          </div>
        </div>
      </footer>
    </div>
  )
}

function FeatureCard({ icon, title, description }: {
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <div className="card card-lg">
      <div className="text-primary-600 mb-4">
        {icon}
      </div>
      <h3 className="text-lg-title mb-2">{title}</h3>
      <p className="text-sm-gray">{description}</p>
    </div>
  )
}
