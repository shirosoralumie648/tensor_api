# 项目修复工作计划（细化版）

## 一、总体策略与治理
1. **技术栈归一化**
   1.1 在 README、docs/ARCHITECTURE.md、WORK_PLAN 中声明：正式版本仅包含 `backend/` + `frontend/`。  
   1.2 将 `lobe-chat-next`、`new-api-main` 移动到 `legacy/` 或单独仓库，保留使用说明；若未来需要引用，改为 git submodule 且默认不参与构建。  
   1.3 清理 go.mod / package.json 中与这些目录相关的依赖，并在 CI 中添加检测脚本，阻止重新引入。
2. **环境配置与一键启动**
   2.1 编写 `.env.example`（后端）和 `.env.local.example`（前端），字段包含：数据库、Redis、RabbitMQ、MinIO、JWT、第三方模型 API Key、对象存储桶名、外部回调地址。  
   2.2 新建 `scripts/setup_env.sh`：校验依赖 → 拷贝示例 env → 生成随机密钥（JWT、加密密钥）→ 输出下一步提示。  
   2.3 更新 `deploy/docker-compose.yml`：定义 Postgres、Redis、RabbitMQ、MinIO、Jaeger、Prometheus、Grafana、backend、frontend；为每个服务配置 volume、健康检查、默认账号密码。  
   2.4 在 docs/QUICK_START.md 编写 “10 分钟启动” 教程：clone→setup_env→docker compose up→访问 `http://localhost:3000`。
3. **CI/CD 基线**
   3.1 创建 GitHub Actions：
       - `backend.yml`: gofmt 检查 → golangci-lint → go test ./... → `make build`.  
       - `frontend.yml`: npm ci → npm run lint → npm run test → npm run build。  
       - `e2e.yml`: docker-compose 启动依赖后运行 Playwright（端到端流程）。  
   3.2 配置 Dependabot / Renovate：监控 Go modules、npm、GitHub Actions。  
   3.3 在 PR 模板中要求列出变更类型、测试结果、影响服务。

## 二、后端基础设施
1. **配置与启动框架**
   - `internal/config`：写全所有配置结构体（App、Database、Redis、MQ、ObjectStorage、Providers、JWT、Observability、Billing、RAG）。实现 `Load()`：读取 `.env`、环境变量覆盖、提供默认值、校验必填字段。  
   - 在 `cmd/<service>/main.go` 创建统一 `bootstrap()`：logger.Init → config.Load → tracing.Init → database.InitPostgres → redis.Init → mq.Init → cache.Init → service.Init → HTTP server 启动，拦截 `SIGTERM/SIGINT` 优雅退出。  
   - 抽象 `internal/server/http.go`：封装 gin 初始化、通用中间件（RequestID、日志、Recover、CORS、限流）、健康检查路由、Graceful Shutdown。
2. **数据库与迁移**
   - 清点所有业务实体（user、session、message、channel、model_pricing、usage、billing、invoice、token、knowledge_document、chunk、embedding、async_task、audit_log 等）。为缺失表新建迁移（上/下行 SQL）。  
   - 在 Makefile 增加 `make migrate-up`, `make migrate-down`, `make migrate-create name=xxx`，并在 CI 中运行 `make migrate-up` 确认可执行。  
   - 为高频查询加索引（如 channel.enabled+type、usage.user_id+created_at、billing.user_id+period）；编写 docs/DB_SCHEMA.md 描述字段含义、约束。
3. **仓储与事务控制**
   - 完成 `repository/` 内所有方法：错误返回 `ErrNotFound/ErrConflict`，禁止返回 `nil, nil`。  
   - 提供 `database.WithTransaction(ctx, fn)`，在计费扣费、渠道上下线、批量写入场景使用；添加单元测试验证事务回滚。  
   - 将查询结果缓存到 Redis/内存（例如渠道列表），并设计缓存一致性策略（写后删缓存 + 异步刷新）。

## 三、认证、鉴权与缓存体系
1. **安全模块落地**
   - `security/encryption.go`：`HashPassword/VerifyPassword` 使用 bcrypt；`Encrypt/Decrypt` 使用 AES-256-GCM + 随机 nonce；写单测验证兼容性。  
   - `security/input_validation.go`：`isValidJSON` 改用 `json.Unmarshal`；扩展 XSS/SQL 检测规则；提供 `ValidateString/IntRange/Email/URL` 的统一错误格式。  
   - 引入 `pkg/rate`：可按 IP、Token、User 设定限流；在 gateway/relay/chat service 中用中间件挂载。
2. **CacheManager 实现**
   - L1: 使用 `sync.Map` + 读写锁；L2: Redis；布隆过滤器用于防止不存在用户穿透。  
   - `getFromDatabase`：查询 `UserRepository` + `TokenRepository` + `QuotaRepository`；使用 singleflight 避免并发回源；设置随机 TTL 防雪崩。  
   - 暴露缓存命中率、击穿次数等指标到 Prometheus。
3. **鉴权中间件**
   - 设计 `AuthMethod` 枚举（Bearer、APIKey、WebSocket、Claude/Gemini 特殊 header等），实现多个 extractor。  
   - 验证流程：提取 Token → 缓存查询 → 校验状态/过期/权限/配额 → 将 `user_id/token_id/role/quota` 写入 gin.Context。  
   - 请求结束时扣减配额、记录 usage；配额不足返回 402 并触发通知；所有认证事件写入 `audit_log`。

## 四、Relay 与多模型中转
1. **渠道选择器**
   - 定义 `ChannelSelector` 接口；`SelectRequest` 包含模型、场景、用户组、区域、排除列表。  
   - 实现策略：随机、轮询、加权轮询、最少错误、最低延迟、一致性哈希；策略可配置。  
   - 共用 ChannelCache（内存 + Redis）；定期与数据库同步；健康检查（Ping API）失败自动下线，成功后恢复。
2. **请求客户端与熔断**
   - `RequestClient.DoRequest`：支持 body 复用、Header merge、超时控制、重试（指数退避 + jitter）。  
   - Circuit Breaker：渠道连续失败 N 次自动熔断，过期后尝试半开；统计失败率/延迟。  
   - 流式请求：封装 SSE 解析，实时将 chunk 写回客户端；错误时自动切换备用渠道。  
   - 将请求成功率、延迟、token 使用、切换次数写入 metrics + Usage。
3. **适配器**
   - OpenAI：chat/stream/embedding/image/speech；Claude、Gemini、其他厂商；统一 request/response schema。  
   - 错误处理：转换成标准 AdapterError（code、message、retryable），用于上层策略判断。  
   - HealthCheck：调用供应商状态接口或尝试轻量请求；记录延迟、错率；定期执行并写入数据库。
4. **渠道管理**
   - 后端 API：渠道 CRUD、权重调整、黑名单、手动上下线、查看统计。  
   - 前端管理界面：展示渠道状态/健康/调用量；允许实时修改配置。  
   - 告警：渠道故障/恢复时发送邮件或 Webhook，并在控制台展示。

## 五、用量统计、定价与计费
1. **UsageLogger/UsageStore**
   - 设计 usage 表：request_id、user_id、channel_id、provider、model、request_tokens、response_tokens、total_tokens、cost、latency_ms、status、error_code、metadata。  
   - `UsageStore`：实现批量写入、查询、聚合、删除旧记录；提供 Mock 供单测。  
   - 在 Relay/Chat 完成后调用 `RecordUsage`；流式场景按段累计 tokens 并在结束时写入。  
   - 实时统计：使用 Redis Pub/Sub 或 Channel，将 usage 推送到前端 Dashboard；同时提供 REST API。
2. **Pricing & Quota**
   - 定价表：model、group、input_price、output_price、priority、group_ratio。  
   - 请求入口：根据 tokens 计算费用与配额扣减（原子更新）；写入 `quota_history`（条目包含原因、前后余额）。  
   - 当配额接近阈值时发送提醒；配额耗尽返回 402/429，并可触发自动充值逻辑。
3. **Billing Service**
   - 功能：充值（模拟/真实）、扣费、账单生成、发票导出。  
   - 定时任务：
       - 每日：汇总 usage，生成日度统计。  
       - 每月：生成账单、发送邮件、若逾期自动限制服务。  
   - 对接支付网关（Stripe/支付宝等），若暂未接入则提供模拟接口和统一抽象层。  
   - 管理端可查看账单、导出、手动调整余额；提供 API 和前端页面。

## 六、消息队列与异步任务
1. **RabbitMQ 接入**
   - 封装 `pkg/mq`：负责连接、重连、心跳、channel 管理；提供统一的 Publish/Consume API。  
   - 声明交换机/队列/绑定配置文件；支持延迟队列、死信队列；可通过 YAML/JSON 配置载入。  
   - 记录 MQ 指标（连接状态、message backlog、nack 次数）并暴露到 Prometheus。
2. **AsyncQueue & Worker**
   - `AsyncQueue.RegisterHandler` 存储 handler；`Submit`：写数据库（task 表）→ 发布到 MQ。  
   - Worker：按 TaskType 消费，执行 handler，记录执行日志、耗时；失败则按策略重试，超限进入 DLQ；提供重放接口。  
   - 管理端：展示任务状态、DLQ、手动重试、暂停/恢复消费。
3. **典型任务落地**
   - 计费汇总、账单生成；  
   - 知识库解析、向量化；  
   - 渠道健康检查、余额同步；  
   - 通知推送（邮件/Slack/Webhook）。

## 七、知识库与 RAG
1. **文档上传与处理**
   - 上传 API：验文件类型/大小，保存 MinIO/S3，并创建数据库记录。  
   - 触发解析任务：下载文件 → 使用 parser（PDF、DOCX、TXT、HTML）提取内容 → 存储原始文本及元数据。  
   - Chunker：按配置（长度、重叠、标题检测）切分；结果写入 `chunks` 表，附带文档 ID、排序、元数据。  
   - 状态机：pending → parsing → chunking → embedding → ready；前端可实时查询进度。
2. **向量化与缓存**
   - `EmbeddingService` 支持多模型（OpenAI、Azure、私有）；实现批量 API、重试、超时、速率限制，并记录 tokens 消耗。  
   - 缓存已向量化的文本（text+model 哈希）至 Redis/数据库，避免重复计费。  
   - 如果嵌入失败，记录错误并支持重试/跳过。
3. **检索与回答**
   - VectorStore：pgvector 或 Milvus/Weaviate；实现 Save/Search/Delete；支持按文档/用户过滤。  
   - BM25：使用 Postgres full-text（tsvector）或 Bleve；用于关键词召回。  
   - Hybrid：合并向量和 BM25 结果，按权重排序；`RerankingService` 使用 cross-encoder（可选）或词重叠。  
   - Chat 服务：
       1) 判断是否触发 RAG（基于查询长度/关键词/用户配置）。  
       2) 执行检索 → 重排 → 构造 Prompt（附引用标记）。  
       3) 调用模型 → 将引用插入回答 → 记录 usage。

## 八、Token 管理与安全
1. **TokenStore 与生命周期**
   - token 表字段：id、user_id、hash、type、scope、状态、创建时间、过期时间、最近使用、使用次数、标签。  
   - 实现 CRUD、Rotate、Expire、Revoke；对 API 提供分页查询、筛选、导出。  
   - Lifecycle Manager：定期扫描即将过期 token，发送提醒邮件/通知；可自动轮换（生成新 token、旧 token 标记为 rotated）。
2. **安全监控与告警**
   - `TokenSecurityManager`：记录使用模式（模型列表、IP 列表、平均请求大小）；检测异常（模型突变、IP 异常、连续失败、超速）。  
   - 泄露监控：可集成外部泄露数据库或自建名单；一旦发现，立即锁定 token、通知用户、写入审计。  
   - 告警通道：邮件、Slack、Webhook；在配置文件中设定阈值。
3. **审计日志**
   - `audit_log` 表：时间、用户、操作、对象、详情、结果、IP、user-agent。  
   - 在认证、配额调整、渠道配置、token 操作、计费操作中写入日志。  
   - 提供查询 API + 管理端界面，支持导出 CSV、按条件过滤。

## 九、可观测性与运维
1. **监控与追踪**
   - 集成 Prometheus 指标：HTTP/gRPC 请求数、延迟、状态码；数据库/Redis/MQ 操作统计；RAG 阶段用时；渠道选择器指标（成功率、切换次数）。  
   - 部署 Prometheus + Grafana；提供 Dashboard JSON（系统概览、Relay 状态、计费用量、知识库处理进度）。  
   - 集成 OpenTelemetry：trace 通过 Jaeger/Tempo 可视化；在 gateway 注入 TraceID 并向下游透传。
2. **日志体系**
   - 统一使用 Zap，输出 JSON 日志；包含 trace_id、span_id、request_id、user_id、service。  
   - gateway 为每个请求生成 trace_id；中间件自动写日志，捕获 panic 并返回标准错误。  
   - 关键事件（支付、渠道故障、安全告警）写结构化日志，方便 ELK/Loki 检索。
3. **健康检查与运维手册**
   - 每个服务提供 `/healthz`（进程存活）与 `/readyz`（依赖检查：DB/Redis/MQ/外部 API）。  
   - Kubernetes 部署配置 liveness/readiness probe；出现异常自动重启/下线。  
   - 编写运维手册：部署步骤、环境变量说明、监控指标、告警阈值、常见故障排查、数据备份/恢复流程。

## 十、前端与用户体验
1. **接口封装与鉴权**
   - `services/api.ts`：封装 axios，统一错误码（401→刷新/跳转、402/429→配额提示、5xx→重试/告警）；支持 SSE EventSource 的封装和自动重连。  
   - 页面：登录/注册/忘记密码、Dashboard、对话列表+详情（含流式展示、引用、工具调用）、知识库管理、渠道管理、账单/用量、API Key 管理。  
   - 组件：消息气泡、代码块、高亮、文档上传、统计图表等。
2. **状态管理与交互**
   - Zustand 管理全局用户/配置状态；React Query 管理远程数据，开启缓存、失效策略、刷新机制。  
   - SSE 渲染：将 streaming 响应增量写入 UI，处理断线重连、取消请求。  
   - 加入全局错误边界、Toast、Loading Skeleton、空状态；支持权限控制（普通用户 vs 管理员）。
3. **设计体系与国际化**
   - 使用 Tailwind + antd-style 建立一致的色板、排版、组件；支持亮/暗主题切换。  
   - Next.js i18n：中/英语言包，所有文案通过 `t()`；提供语言切换 UI。  
   - Storybook + Testing Library：为关键组件编写 stories 和单测，保障 UI 可维护性。

## 十一、测试、发布与部署
1. **测试矩阵**
   - Go：gofmt/golangci-lint 强制通过；对 config、repository、adapter、relay、usage、billing、cache、security、mq 编写单元测试（Mock+testcontainers），目标覆盖率 >70%。  
   - 集成测试：使用 docker-compose 启动依赖，运行 API 测试脚本（登录→建会话→触发 RAG→查看账单→扣费）。  
   - 前端：ESLint、Prettier、Jest 覆盖组件/Hook；Playwright 端到端测试（登录→聊天→查看用量→管理渠道）。
2. **CI/CD**
   - 后端、前端 pipelines 成功后，执行 Docker build（多阶段、非 root），推送至容器仓库（GHCR/ECR）。  
   - 集成 SAST（gosec、npm audit）与依赖扫描；不通过则阻断合并。  
   - 发布流程：打 tag → 自动生成 Release Notes → 部署到测试环境 → 回归 → Promote 到生产，同时记录版本号与变更摘要。
3. **生产部署与运维**
   - Dockerfile：多阶段构建（编译阶段 + 运行阶段），运行阶段使用 distroless/alpine；添加 `HEALTHCHECK`。  
   - Kubernetes/Helm：准备 Chart & values，包含 ConfigMap、Secret、Ingress、HPA、ServiceMonitor；在 values 中允许配置副本数、资源、自动扩缩。  
   - 运维手册：列出监控指标、告警规则、伸缩策略、日志查看方式、数据备份/恢复、紧急故障处理 SOP。

完成上述细化步骤后，项目的认证、渠道中转、知识库检索、计费结算、消息队列、可观测性、前端体验等模块都将具备可验证的实现，能够在生产环境稳定运行并对外提供服务。
