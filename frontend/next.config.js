/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  
  // 静态导出模式（用于 Docker 部署）
  output: process.env.BUILD_MODE === 'standalone' ? 'standalone' : undefined,
  
  // 环境变量
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    NEXT_PUBLIC_APP_NAME: 'Oblivious',
  },
  
  // 图片优化
  images: {
    domains: ['localhost', 'your-cdn-domain.com'],
    unoptimized: process.env.BUILD_MODE === 'export',
  },
  
  // 国际化
  i18n: {
    locales: ['zh-CN', 'en-US'],
    defaultLocale: 'zh-CN',
  },
  
  // Webpack 配置
  webpack: (config, { isServer }) => {
    // 客户端 bundle 优化
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
        net: false,
        tls: false,
      };
    }
    return config;
  },
  
  // 实验性功能
  experimental: {
    optimizePackageImports: ['antd', '@ant-design/icons'],
  },
};

module.exports = nextConfig;

