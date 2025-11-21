# Oblivious AI 平台完善开发计划

## 项目概述
Oblivious 是一个同时提供面向用户的 AI 服务和面向开发者的 API 中转服务的综合性平台。

**参考项目对比：**
- **LobeChat**: 优秀的前端交互、MCP插件系统、知识库RAG、会话管理
- **New API**: 完善的渠道管理、负载均衡、计费系统、适配器模式

---

## 一、核心功能完善 (Core Features Enhancement)

### 1.1 认证授权系统优化
**当前状态**: 基础的JWT认证中间件
**需要完善**:
- ✅ **参考 New API**: Token多级缓存机制（内存+Redis）
  - 文件: `backend/internal/middleware/auth.go`
  - 实现 `GetUserCache()` 方法，支持用户信息缓存
  - 减少数据库查询，提升认证性能

- ✅ **Token状态管理**:
  - 增加Token状态枚举：正常/已耗尽/已禁用/已过期
  - 实现Token续期机制
  - 支持Token的软删除和恢复

- ✅ **多种认证方式**:
  - 支持WebSocket认证（参考New API auth.go:182）
  - 支持Claude SDK格式认证
  - 支持Gemini API格式认证

- ✅ **权限控制**:
  - 实现基于角色的访问控制（RBAC）
  - 用户分组权限管理
  - API端点级别的权限控制

### 1.2 API中转与路由系统
**当前状态**: 基础的Relay Service
**需要完善**:
- ✅ **流式响应优化**（参考New API relay.go）:
  - 文件: `backend/cmd/relay/main.go`
  - 实现SSE流式响应处理
  - 支持流式数据的实时转发和错误处理
  - 添加连接超时和心跳机制

- ✅ **请求重试机制**:
  - 实现智能重试逻辑（参考New API relay.go:160）
  - 根据错误类型判断是否重试
  - 支持指数退避策略
  - 渠道失败时自动切换到备用渠道

- ✅ **请求体恢复机制**:
  - 解决请求体只能读取一次的问题
  - 实现请求体缓存和恢复
  - 支持大请求的流式处理

- ✅ **中继处理器抽象**:
  - 文件: `backend/internal/relay/handler.go`
  - 针对不同API类型实现专用处理器
  - Chat Completions / Embeddings / Images / Audio
  - 统一的错误处理和响应格式化

### 1.3 渠道管理与负载均衡
**当前状态**: 简单的渠道选择
**需要完善**:
- ✅ **渠道缓存系统**（参考New API channel.go）:
  - 文件: `backend/internal/model/channel.go`
  - 实现渠道信息的内存缓存
  - 定期从数据库同步渠道状态（SyncChannelCache）
  - 支持渠道的热更新，无需重启服务

- ✅ **智能渠道选择算法**:
  - 基于用户分组的渠道过滤
  - 支持模型名称映射和匹配
  - 加权随机选择（参考distributor.go:100）
  - 渠道健康检查和自动禁用

- ✅ **多密钥轮询机制**:
  - 支持单个渠道配置多个API密钥
  - 实现Random和Polling两种模式（channel.go:148）
  - 密钥级别的失败统计和自动切换

- ✅ **负载均衡策略**:
  - 支持权重配置
  - 实现轮询、随机、最小连接数等策略
  - 渠道优先级管理

- ✅ **渠道能力管理**:
  - 数据库表: `backend/migrations/000006_create_channels_table.up.sql`
  - 记录每个渠道支持的模型列表
  - 支持能力的动态更新和查询

### 1.4 计费与配额系统
**当前状态**: 基础的Billing Service
**需要完善**:
- ✅ **预扣费与后扣费机制**（参考New API quota.go）:
  - 文件: `backend/cmd/agent/main.go`
  - 实现请求前配额预扣除（PreConsumeQuota）
  - 失败时自动退还配额（ReturnPreConsumedQuota）
  - 请求完成后精确计费

- ✅ **Token计数系统**:
  - 集成tiktoken或类似库
  - 支持不同模型的token计算
  - 处理多模态内容（文本、图片、音频）的计费
  - 实现音频配额计算（quota.go:50）

- ✅ **定价与倍率系统**:
  - 文件: `backend/internal/model/pricing.go`
  - 支持按Token计费和按次计费两种模式
  - 用户分组倍率配置
  - 模型定价的动态加载和缓存（GetPricing）
  - 输入/输出Token不同定价

- ✅ **配额管理**:
  - 用户配额实时更新
  - Token配额独立管理
  - 配额预警和通知
  - 配额充值和过期管理

- ✅ **账单记录**:
  - 详细的使用记录（时间、模型、Token数、费用）
  - 支持账单导出和统计
  - 实现消息队列异步记账（RabbitMQ）

---

## 二、用户服务功能 (User-Facing Features)

### 2.1 AI对话系统优化
**当前状态**: Chat Service处理对话
**需要完善**:
- ✅ **流式对话优化**（参考LobeChat）:
  - 文件: `backend/cmd/chat/main.go`
  - 实现SSE流式响应
  - 支持打字机效果的实时输出
  - 处理流式中断和恢复

- ✅ **上下文管理**:
  - 实现消息历史压缩算法
  - 智能上下文窗口管理
  - 支持上下文总结和记忆
  - 多轮对话状态保持

- ✅ **消息格式化**:
  - 支持Markdown渲染
  - 代码高亮显示
  - LaTeX数学公式渲染
  - 图片、文件等多媒体消息

- ✅ **AI Agent系统**（参考LobeChat GeneralChatAgent.ts）:
  - 实现通用对话代理
  - 支持系统提示词配置
  - Agent能力扩展接口

### 2.2 会话管理
**当前状态**: 基础的对话记录
**需要完善**:
- ✅ **会话生命周期管理**（参考LobeChat session.ts）:
  - 文件: `backend/migrations/000004_create_sessions_table.up.sql`
  - 会话创建、更新、删除、归档
  - 会话分组和标签管理
  - 会话搜索和过滤

- ✅ **会话配置**:
  - 每个会话独立的Agent配置
  - 模型选择和参数配置（temperature、max_tokens等）
  - 系统提示词自定义
  - 会话级别的插件启用

- ✅ **消息管理**:
  - 数据库表: `backend/migrations/000005_create_messages_table.up.sql`
  - 消息的编辑和删除
  - 消息分支（多次重新生成）
  - 消息引用和回复
  - 消息导出（Markdown、JSON等格式）

- ✅ **会话共享**:
  - 会话分享链接生成
  - 公开会话浏览
  - 会话模板市场

### 2.3 知识库与RAG系统
**当前状态**: 未实现
**需要完善**:
- ✅ **文档上传与处理**（参考LobeChat file.ts）:
  - 新建文件: `backend/cmd/knowledge/main.go`
  - 支持PDF、Word、Excel、TXT等格式
  - 实现文档解析器（参考file-loaders）
  - 文件存储（MinIO/S3）

- ✅ **文本分块系统**（参考chunk.ts）:
  - 智能文本分块算法
  - 支持重叠分块提高检索质量
  - 保留文档结构信息
  - 分块元数据管理

- ✅ **向量化与存储**（参考rag.ts）:
  - 数据库表: `packages/database/src/schemas/rag.ts:69`
  - 集成Embedding模型（OpenAI/本地模型）
  - 向量数据库集成（PostgreSQL pgvector扩展）
  - 向量索引优化

- ✅ **语义检索**:
  - 实现向量相似度搜索
  - 混合检索（向量+关键词）
  - 检索结果重排序
  - 检索质量评估（参考ragEval.ts）

- ✅ **RAG流程集成**:
  - 在对话中自动触发知识库检索
  - 将检索结果注入上下文
  - 引用来源标注
  - 知识库更新通知

### 2.4 插件与工具调用
**当前状态**: 未实现
**需要完善**:
- ✅ **MCP插件系统**（参考LobeChat mcp.ts）:
  - 新建文件: `backend/cmd/plugin/main.go`
  - 支持MCP协议（Model Context Protocol）
  - stdio模式插件调用
  - http模式插件调用

- ✅ **Function Calling引擎**（参考createToolEngine.ts）:
  - 实现工具调用引擎
  - 支持OpenAI Function Calling格式
  - 工具参数验证和转换
  - 工具执行结果处理

- ✅ **插件管理**:
  - 插件安装、启用、禁用
  - 插件配置管理
  - 插件市场集成
  - 插件权限控制

- ✅ **内置工具**:
  - 网络搜索工具
  - 代码执行器
  - 文件操作工具
  - API调用工具

- ✅ **桌面端支持**（可选）:
  - Electron IPC通信（参考desktop相关文件）
  - 本地文件访问
  - 系统级插件调用

---

## 三、开发者服务功能 (Developer-Facing Features)

### 3.1 多模型适配器
**当前状态**: 基础转发
**需要完善**:
- ✅ **适配器模式实现**（参考New API adapter.go）:
  - 文件: `backend/internal/adapter/adapter.go`
  - 定义统一的Adaptor接口
  - ConvertRequest()：转换请求格式
  - DoRequest()：发送HTTP请求
  - DoResponse()：解析响应

- ✅ **支持的AI提供商**:
  - OpenAI（openai.go）
  - Anthropic Claude（claude.go）
  - Google Gemini（gemini.go）
  - Baidu 文心一言
  - Alibaba 通义千问
  - 设计扩展机制支持65+提供商

- ✅ **请求格式转换**:
  - OpenAI格式作为基准
  - 自动转换到目标提供商格式
  - 处理不同的参数映射
  - 多模态内容转换

- ✅ **响应统一化**:
  - 将不同提供商响应转为OpenAI格式
  - 统一错误处理和错误码
  - Token使用量提取和标准化

### 3.2 API密钥管理
**当前状态**: 简单的token验证
**需要完善**:
- ✅ **Token生命周期管理**（参考New API token.go）:
  - 数据库表: `backend/migrations/000003_create_tokens_table.up.sql`
  - Token创建、更新、禁用、删除
  - Token有效期管理
  - Token用途备注和分类

- ✅ **Token权限控制**:
  - 模型白名单/黑名单
  - API端点权限限制
  - IP白名单限制
  - 请求来源限制（Referer）

- ✅ **Token配额管理**:
  - 每个Token独立配额
  - 配额使用情况实时跟踪
  - 配额耗尽自动禁用
  - 配额预警通知

- ✅ **Token安全**:
  - Token加密存储
  - 访问频率限制
  - 异常访问检测
  - Token泄露风险预警

### 3.3 使用量统计与分析
**当前状态**: 基础记录
**需要完善**:
- ✅ **详细用量记录**:
  - 数据库表: `backend/migrations/000007_create_usage_logs_table.up.sql`
  - 记录每次API调用的详细信息
  - 请求时间、响应时间、延迟
  - 模型、Token数量、费用
  - 渠道信息、状态码

- ✅ **实时统计仪表盘**:
  - 实时请求数/QPS
  - Token使用量统计
  - 成本统计和趋势
  - 模型使用分布

- ✅ **历史数据分析**:
  - 按时间维度统计（小时、天、月）
  - 按模型维度统计
  - 按用户/Token维度统计
  - 成功率和错误率分析

- ✅ **数据导出**:
  - 支持CSV/Excel导出
  - 自定义报表生成
  - API调用日志导出

### 3.4 速率限制与配额控制
**当前状态**: 基础配额扣除
**需要完善**:
- ✅ **多级速率限制**:
  - 文件: `backend/internal/middleware/ratelimit.go`
  - 用户级别速率限制
  - Token级别速率限制
  - IP级别速率限制
  - 模型级别速率限制

- ✅ **限流算法**:
  - 基于Redis的滑动窗口算法
  - Token Bucket算法
  - Leaky Bucket算法
  - 分布式限流支持

- ✅ **配额控制策略**:
  - 日配额、月配额限制
  - 单次请求Token数限制
  - 并发请求数限制
  - 超限后的降级策略

- ✅ **限制响应**:
  - 返回429状态码
  - 提供Retry-After头
  - 限制信息详细描述
  - 自定义限制提示

---

## 四、数据层优化 (Data Layer)

### 4.1 数据库Schema完善
**当前状态**: 基础表结构
**需要完善**:
- ✅ **新增缺失表**:
  - `tokens表`: 开发者API密钥管理
  - `usage_logs表`: API调用详细日志
  - `knowledge_bases表`: 知识库管理
  - `documents表`: 文档元数据
  - `chunks表`: 文本分块存储
  - `embeddings表`: 向量嵌入（参考LobeChat rag.ts:69）
  - `plugins表`: 插件配置信息
  - `agent_configs表`: Agent配置

- ✅ **完善现有表**:
  - `users表`: 添加角色、分组、状态字段
  - `sessions表`: 添加配置、标签、归档字段
  - `messages表`: 添加工具调用、附件字段
  - `channels表`: 添加权重、优先级、健康状态

- ✅ **索引优化**:
  - 高频查询字段添加索引
  - 复合索引优化查询性能
  - 向量索引（pgvector）

- ✅ **数据关系优化**:
  - 外键约束完善
  - 级联删除策略
  - 软删除支持

### 4.2 缓存策略
**当前状态**: 未实现Redis缓存
**需要完善**:
- ✅ **Redis集成**（参考New API）:
  - 文件: `backend/internal/database/redis.go`
  - 初始化Redis客户端
  - 支持单机和集群模式
  - 连接池管理

- ✅ **缓存分类**:
  - **热数据缓存**: 用户信息、Token信息
  - **配置缓存**: 渠道信息、模型定价
  - **会话缓存**: 对话上下文、历史消息
  - **限流缓存**: 请求计数、配额使用

- ✅ **缓存策略**:
  - Cache-Aside模式
  - 设置合理的TTL
  - 缓存预热机制
  - 缓存穿透防护

- ✅ **缓存同步**:
  - 定期同步机制
  - 数据变更时主动失效
  - 分布式缓存一致性

### 4.3 消息队列
**当前状态**: 基础队列
**需要完善**:
- ✅ **RabbitMQ集成**:
  - 文件: `backend/pkg/queue/rabbitmq.go`
  - 初始化MQ连接
  - 声明交换机和队列
  - 实现重连机制

- ✅ **异步任务**:
  - **计费队列**: API调用后异步记账
  - **通知队列**: 配额预警、系统通知
  - **分析队列**: 数据统计和分析任务
  - **导出队列**: 文件导出任务

- ✅ **消费者实现**:
  - 文件: `backend/cmd/consumer/main.go`
  - 实现各类型消费者
  - 错误处理和重试机制
  - 死信队列处理

- ✅ **队列监控**:
  - 消息积压监控
  - 消费速率统计
  - 失败消息告警

---

## 五、前端应用 (Frontend)

### 5.1 用户界面
**当前状态**: 基础React应用
**需要完善**:
- ✅ **现代化UI框架**（参考LobeChat）:
  - 迁移到Next.js 14+ （App Router）
  - 使用TailwindCSS作为样式方案
  - 集成shadcn/ui组件库
  - 使用Lucide图标库

- ✅ **对话界面**:
  - 文件: `frontend/src/components/ChatPanel.tsx`
  - 流式打字效果
  - Markdown实时渲染
  - 代码高亮和复制功能
  - 多模态消息展示（图片、文件）

- ✅ **会话管理界面**:
  - 侧边栏会话列表
  - 会话创建和切换
  - 会话搜索和过滤
  - 会话设置面板

- ✅ **用户设置**:
  - 个人资料管理
  - 主题切换（深色/浅色）
  - 语言切换
  - 模型默认配置

- ✅ **知识库界面**:
  - 文档上传和管理
  - 知识库创建和编辑
  - 文档分块预览
  - 搜索结果展示

### 5.2 开发者控制台
**当前状态**: 未实现
**需要完善**:
- ✅ **控制台首页**:
  - 新建: `frontend/src/app/developer/page.tsx`
  - 概览仪表盘（请求数、费用、错误率）
  - 快速开始指南
  - API文档入口

- ✅ **API密钥管理**:
  - Token创建和删除
  - Token权限配置
  - Token使用统计
  - 密钥安全提示

- ✅ **使用量统计**:
  - 实时请求监控
  - 历史数据图表（ECharts/Recharts）
  - 按模型/时间筛选
  - 数据导出功能

- ✅ **账单管理**:
  - 配额余额展示
  - 账单详情列表
  - 消费趋势分析
  - 充值功能

- ✅ **文档中心**:
  - API参考文档
  - SDK示例代码
  - 快速开始教程
  - 最佳实践指南

### 5.3 实时通信
**当前状态**: 基础HTTP
**需要完善**:
- ✅ **SSE流式通信**:
  - 文件: `frontend/src/services/chat.ts`
  - 实现EventSource客户端
  - 处理流式数据解析
  - 处理连接中断和重连

- ✅ **WebSocket支持**（可选）:
  - 实现WebSocket客户端
  - 心跳机制
  - 自动重连
  - 分组和广播支持

- ✅ **状态管理**:
  - 使用Zustand/Redux管理全局状态
  - 实时消息同步
  - 乐观更新UI
  - 离线状态处理

---

## 六、运维与监控 (DevOps & Monitoring)

### 6.1 监控系统完善
**当前状态**: 基础Prometheus配置
**需要完善**:
- ✅ **Prometheus指标采集**（参考架构图8c）:
  - 文件: `backend/pkg/metrics/metrics.go`
  - 实现HTTP请求计数器
  - 请求延迟直方图
  - 渠道健康状态指标
  - 配额使用量指标
  - 错误率统计

- ✅ **Grafana仪表盘**（参考gateway.json）:
  - 创建Gateway服务仪表盘
  - QPS/TPS实时监控
  - 延迟P50/P95/P99监控
  - 错误率趋势图
  - 渠道负载分布图

- ✅ **告警规则**:
  - 文件: `infra/monitoring/prometheus/rules/alerts.yml`
  - 服务不可用告警
  - 高错误率告警
  - 响应时间过长告警
  - 配额耗尽告警

- ✅ **分布式追踪**（可选）:
  - 集成Jaeger/Zipkin
  - 实现OpenTelemetry
  - 请求链路追踪
  - 性能瓶颈分析

### 6.2 日志系统
**当前状态**: 基础日志
**需要完善**:
- ✅ **结构化日志**（参考logger.go）:
  - 文件: `backend/pkg/logger/logger.go`
  - 使用zap/logrus实现结构化日志
  - 统一日志格式（JSON）
  - 日志级别控制
  - 上下文信息注入（RequestID、UserID等）

- ✅ **日志聚合**（参考Loki配置）:
  - Loki日志收集系统
  - 配置文件: `infra/monitoring/loki/loki-config.yaml`
  - 日志标签和过滤
  - 日志保留策略（7天）

- ✅ **日志查询界面**:
  - Grafana Loki集成
  - 日志搜索和过滤
  - 日志流实时查看
  - 错误日志聚合分析

- ✅ **审计日志**:
  - 用户操作日志
  - API调用日志
  - 敏感操作记录
  - 日志不可篡改性保证

### 6.3 部署优化
**当前状态**: K8s基础配置
**需要完善**:
- ✅ **容器化优化**:
  - 多阶段构建Dockerfile
  - 文件: `deploy/docker/Dockerfile.backend`
  - 减小镜像体积
  - 使用非root用户运行
  - 健康检查配置

- ✅ **Kubernetes部署完善**:
  - HPA自动扩缩容配置（参考gateway-hpa.yaml:10）
  - 资源限制和请求配置
  - 存活探针和就绪探针
  - 滚动更新策略
  - PDB（Pod Disruption Budget）

- ✅ **Helm Chart**:
  - 创建: `deploy/helm/oblivious/`
  - 参数化配置管理
  - 多环境部署支持
  - 依赖管理

- ✅ **CI/CD流水线**:
  - GitHub Actions工作流
  - 文件: `.github/workflows/ci.yml`
  - 自动化测试
  - 自动化构建和推送镜像
  - 自动化部署到K8s

- ✅ **配置管理**:
  - ConfigMap管理应用配置
  - Secret管理敏感信息
  - 外部配置中心（可选：Consul/Etcd）

---

## 七、安全与性能 (Security & Performance)

### 7.1 安全加固
**当前状态**: 基础认证
**需要完善**:
- ✅ **输入验证**:
  - 文件: `backend/internal/middleware/validator.go`
  - 严格的请求参数验证
  - SQL注入防护
  - XSS攻击防护
  - CSRF Token验证

- ✅ **数据加密**:
  - 敏感数据加密存储（API密钥、渠道密钥）
  - 传输层HTTPS/TLS
  - 数据库连接加密
  - 备份数据加密

- ✅ **访问控制**:
  - 基于RBAC的权限系统
  - API端点级别权限控制
  - 数据行级别权限控制
  - IP白名单/黑名单

- ✅ **安全审计**:
  - 安全事件记录
  - 异常行为检测
  - API滥用监控
  - 定期安全扫描

- ✅ **依赖安全**:
  - 定期更新依赖包
  - 漏洞扫描（Dependabot）
  - 依赖版本锁定
  - 安全配置最佳实践

### 7.2 性能优化
**当前状态**: 基础实现
**需要完善**:
- ✅ **数据库优化**:
  - 查询优化和索引优化
  - 连接池配置调优
  - 读写分离（主从复制）
  - 分库分表策略（大表拆分）

- ✅ **缓存优化**:
  - 多级缓存策略（内存+Redis）
  - 缓存预热机制
  - 缓存雪崩防护
  - 缓存一致性保证

- ✅ **并发优化**:
  - Goroutine池管理
  - 并发请求控制
  - 防止Goroutine泄露
  - Context超时控制

- ✅ **网络优化**:
  - HTTP/2支持
  - Keep-Alive连接复用
  - 请求/响应压缩
  - CDN加速（静态资源）

- ✅ **代码优化**:
  - 热点代码路径优化
  - 内存分配优化
  - CPU Profile分析
  - 定期性能压测

---

## 八、开发优先级与时间规划

### Phase 1: 核心功能完善 (4-6周)
**目标**: 构建稳定可靠的API中转核心

#### Week 1-2: 认证与渠道系统
- ✅ 完善Token认证与缓存机制
- ✅ 实现渠道缓存和同步系统
- ✅ 实现智能渠道选择算法
- ✅ 多密钥轮询机制
- ✅ 新增`tokens表`和`channel能力`相关字段

#### Week 3-4: API中转与适配器
- ✅ 流式响应处理优化
- ✅ 请求重试机制和错误处理
- ✅ 实现适配器模式（OpenAI/Claude/Gemini）
- ✅ 请求/响应格式转换
- ✅ 统一错误码处理

#### Week 5-6: 计费与配额系统
- ✅ Token计数系统集成
- ✅ 预扣费/后扣费机制
- ✅ 定价与倍率系统
- ✅ 速率限制中间件
- ✅ 新增`usage_logs表`和计费相关逻辑

**交付成果**:
- 可用的API中转服务
- 支持65+AI提供商
- 完善的计费系统
- 基础监控指标

### Phase 2: 用户服务增强 (3-4周)
**目标**: 打造优秀的用户对话体验

#### Week 7-8: AI对话系统
- ✅ 优化Chat Service流式响应
- ✅ 实现上下文管理算法
- ✅ 消息格式化支持（Markdown/LaTeX）
- ✅ AI Agent系统实现
- ✅ 完善`messages表`添加工具调用字段

#### Week 9-10: 会话与前端界面
- ✅ 会话管理功能完善
- ✅ 前端迁移到Next.js + TailwindCSS
- ✅ 对话界面实现（参考LobeChat）
- ✅ 会话管理界面
- ✅ SSE流式通信实现

**交付成果**:
- 流畅的对话体验
- 现代化的前端界面
- 完善的会话管理
- 实时流式响应

### Phase 3: 开发者服务增强 (3-4周)
**目标**: 提供完善的开发者工具和文档

#### Week 11-12: 开发者控制台
- ✅ 开发者控制台前端实现
- ✅ API密钥管理界面
- ✅ 实时使用量统计仪表盘
- ✅ 账单管理界面
- ✅ 数据导出功能

#### Week 13-14: 知识库与RAG系统
- ✅ 文档上传和解析服务
- ✅ 文本分块算法
- ✅ 向量化与PostgreSQL pgvector集成
- ✅ 语义检索实现
- ✅ RAG流程集成到对话系统
- ✅ 新建`knowledge_bases`、`documents`、`chunks`、`embeddings`表

**交付成果**:
- 完整的开发者控制台
- 知识库RAG功能
- API文档和SDK
- 详细的使用统计

### Phase 4: 运维与优化 (2-3周)
**目标**: 提升系统稳定性、安全性和性能

#### Week 15-16: 监控与日志
- ✅ 完善Prometheus指标采集
- ✅ Grafana仪表盘配置
- ✅ 告警规则设置
- ✅ Loki日志聚合系统
- ✅ 分布式追踪（可选）

#### Week 17: 安全与性能
- ✅ 安全加固（输入验证、数据加密）
- ✅ 数据库优化（索引、查询优化）
- ✅ Redis缓存集成
- ✅ 性能压测和调优
- ✅ Helm Chart和CI/CD配置

**交付成果**:
- 完善的监控告警系统
- 结构化日志和日志查询
- 安全加固和性能优化
- 自动化部署流程

---

## 附录：技术栈建议

### 后端技术栈

**核心框架**:
- **语言**: Go 1.21+
- **Web框架**: Gin (HTTP路由和中间件)
- **ORM**: GORM (数据库操作)
- **配置管理**: Viper + godotenv

**数据存储**:
- **主数据库**: PostgreSQL 15+ (with pgvector扩展)
- **缓存**: Redis 7+
- **消息队列**: RabbitMQ 3.12+
- **对象存储**: MinIO / AWS S3

**AI相关**:
- **Token计数**: tiktoken-go
- **Embedding**: OpenAI API / 本地模型
- **向量搜索**: pgvector

**工具库**:
- **日志**: zap / logrus
- **请求客户端**: resty
- **校验**: go-playground/validator
- **JWT**: golang-jwt/jwt
- **密码加密**: bcrypt

**测试**:
- **单元测试**: testify
- **Mock**: gomock
- **API测试**: httptest

### 前端技术栈

**核心框架**:
- **框架**: Next.js 14+ (App Router)
- **语言**: TypeScript 5+
- **状态管理**: Zustand / Redux Toolkit
- **路由**: Next.js App Router

**UI框架**:
- **样式**: TailwindCSS 3+
- **组件库**: shadcn/ui
- **图标**: Lucide Icons
- **动画**: Framer Motion

**功能库**:
- **HTTP客户端**: axios
- **表单验证**: react-hook-form + zod
- **Markdown渲染**: react-markdown
- **代码高亮**: prism-react-renderer
- **图表**: ECharts / Recharts
- **日期处理**: date-fns

**开发工具**:
- **代码规范**: ESLint + Prettier
- **类型检查**: TypeScript
- **构建**: Turbopack (Next.js 14)
- **包管理**: pnpm

### 基础设施

**容器化**:
- **容器运行时**: Docker 24+
- **编排**: Kubernetes 1.28+
- **包管理器**: Helm 3+

**监控与日志**:
- **指标采集**: Prometheus
- **可视化**: Grafana
- **日志聚合**: Loki
- **追踪**: Jaeger (可选)

**DevOps**:
- **CI/CD**: GitHub Actions
- **镜像仓库**: Docker Hub / Harbor
- **配置管理**: ConfigMap + Secret
- **服务网格**: Istio (可选)

**云服务**:
- **部署**: 自建 K8s / AWS EKS / 阿里云 ACK
- **负载均衡**: Nginx Ingress Controller
- **域名解析**: CloudFlare
- **CDN**: CloudFlare / AWS CloudFront

---

## 总结

本开发计划通过对比**LobeChat**和**New API**两个优秀项目，为Oblivious AI平台提供了全面的功能完善路线图。

**关键亮点**:
1. ✅ **双服务架构**: 同时支持用户对话和开发者API中转
2. ✅ **强大的渠道管理**: 借鉴New API的渠道缓存和负载均衡
3. ✅ **优秀的用户体验**: 参考LobeChat的前端交互和功能设计
4. ✅ **完善的RAG系统**: 支持知识库和语义检索
5. ✅ **企业级运维**: 监控、日志、告警一体化

**建议开发顺序**: 核心功能 → 用户服务 → 开发者服务 → 运维优化

**预计完成时间**: 12-17周（约3-4个月）
