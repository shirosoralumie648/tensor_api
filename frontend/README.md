# Oblivious AI Frontend

企业级 AI API 中转平台的前端应用

## 技术栈

- **框架**: Next.js 14
- **UI 库**: React 18 + Tailwind CSS
- **状态管理**: Zustand
- **HTTP 客户端**: Axios
- **实时通信**: Server-Sent Events (SSE)
- **可视化**: Recharts
- **代码高亮**: Highlight.js + KaTeX

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

访问 [http://localhost:3000](http://localhost:3000)

### 生产构建

```bash
npm run build
npm run start
```

## 项目结构

```
frontend/
├── src/
│   ├── app/                 # Next.js 应用目录
│   │   ├── layout.tsx       # 根布局
│   │   ├── page.tsx         # 首页
│   │   └── globals.css      # 全局样式
│   ├── components/          # React 组件
│   ├── pages/               # 页面组件
│   ├── hooks/               # 自定义 Hooks
│   ├── stores/              # Zustand 状态管理
│   ├── services/            # API 服务
│   ├── types/               # TypeScript 类型定义
│   ├── utils/               # 工具函数
│   └── styles/              # 样式文件
├── public/                  # 静态资源
├── package.json
├── next.config.js
├── tailwind.config.ts
├── tsconfig.json
└── postcss.config.js
```

## 主要功能

### 5.1 用户界面 (5.1.1-5.1.3)
- ✅ Next.js 框架搭建
- ✅ 对话界面实现
- ✅ 会话管理界面

### 5.2 开发者控制台 (5.2.1-5.2.3)
- ⏳ 控制台框架
- ⏳ API 密钥管理
- ⏳ 使用统计与账单

### 5.3 实时通信 (5.3.1-5.3.2)
- ⏳ SSE 流式客户端
- ⏳ 状态管理

## 开发规范

### 代码风格

使用 Prettier 和 ESLint 进行代码格式化和检查

```bash
npm run format      # 格式化代码
npm run lint        # 检查代码
npm run type-check  # TypeScript 类型检查
```

### 组件编写

- 使用函数式组件
- 优先使用 TypeScript
- 遵循 React Hook 规范
- 使用 Tailwind CSS 进行样式化

### 命名约定

- 组件文件: `PascalCase` (如 `ChatBox.tsx`)
- 页面文件: 小写 (如 `chat.tsx`)
- Hook 文件: `useXxx` (如 `useChat.ts`)
- 类型文件: 根据导出类型命名

## 环境变量

创建 `.env.local` 文件：

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

详见 `.env.example`

## 构建配置

### Tailwind CSS

自定义配置位于 `tailwind.config.ts`，包括：
- 扩展颜色和间距
- 自定义动画
- 插件配置

### Next.js 配置

`next.config.js` 包括：
- 图片优化
- API 重写
- 构建优化

## 测试

```bash
npm run test        # 运行测试
npm run test:watch  # 监听模式
```

## 性能优化

- 代码分割和懒加载
- 图片优化
- 缓存策略
- 压缩和最小化

## 部署

### Vercel

```bash
# 配置 Vercel CLI
npm install -g vercel

# 部署
vercel
```

### Docker

```dockerfile
# 见根目录 Dockerfile
```

## 常见问题

### Q: 如何添加新页面？

A: 在 `src/app` 中创建新的文件夹和 `page.tsx` 文件

### Q: 如何调用后端 API？

A: 使用 `src/services` 中的 API 服务函数

### Q: 如何管理全局状态？

A: 使用 Zustand stores，位于 `src/stores`

## 许可证

MIT

## 联系方式

- 文档: [API Docs](http://localhost:8080/docs)
- 问题: [GitHub Issues](https://github.com/oblivious-ai/frontend/issues)

