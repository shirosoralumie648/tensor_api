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
**开发周期**: 2周
**优先级**: P0 (最高)

#### 任务分解

**Week 1-2: Token缓存与状态管理**

##### 任务1.1.1: Token多级缓存机制（3天）
- **文件**: `backend/internal/middleware/auth.go`
- **实现内容**:
  ```go
  // 实现用户缓存结构
  type UserCache struct {
    UserID    int
    Username  string
    Group     string
    Quota     int64
    Role      int
    Status    int
    ExpireAt  time.Time
  }
  
  // 实现GetUserCache方法
  func GetUserCache(userId int) (*UserCache, error)
  func SetUserCache(userId int, cache *UserCache) error
  func InvalidateUserCache(userId int) error
  ```
- **技术要求**:
  - L1缓存：sync.Map本地内存缓存（过期时间5分钟）
  - L2缓存：Redis缓存（过期时间30分钟）
  - 缓存穿透防护：布隆过滤器
  - 缓存击穿防护：singleflight模式
- **验收标准**:
  - [ ] 认证QPS提升至5000+/秒
  - [ ] 数据库查询减少95%以上
  - [ ] 缓存命中率达到90%以上
  - [ ] 单元测试覆盖率>80%

##### 任务1.1.2: Token状态管理系统（2天）
- **数据库迁移**: `backend/migrations/000008_add_token_status.up.sql`
- **实现内容**:
  ```sql
  ALTER TABLE tokens ADD COLUMN status INT DEFAULT 1;
  ALTER TABLE tokens ADD COLUMN expire_at TIMESTAMP;
  ALTER TABLE tokens ADD COLUMN deleted_at TIMESTAMP;
  CREATE INDEX idx_tokens_status ON tokens(status);
  ```
- **状态枚举定义**:
  - 1: 正常（Normal）
  - 2: 已耗尽（Exhausted）
  - 3: 已禁用（Disabled）
  - 4: 已过期（Expired）
  - 5: 已删除（Deleted - 软删除）
- **业务逻辑**:
  - Token续期：临期前7天可续期
  - 自动标记：配额耗尽自动标记状态2
  - 定时任务：每小时检查过期Token
- **验收标准**:
  - [ ] 所有状态转换有审计日志
  - [ ] 软删除Token可恢复
  - [ ] 过期Token无法使用
  - [ ] 配额耗尽自动禁用

##### 任务1.1.3: 多种认证方式支持（2天）
- **文件**: `backend/internal/middleware/auth.go`
- **实现内容**:
  ```go
  // WebSocket认证
  func extractWebSocketToken(c *gin.Context) string
  
  // Claude SDK认证
  func extractClaudeToken(c *gin.Context) string
  
  // Gemini API认证
  func extractGeminiToken(c *gin.Context) string
  ```
- **支持格式**:
  - 标准：`Authorization: Bearer sk-xxx`
  - WebSocket：URL参数 `?token=sk-xxx`
  - Claude：`x-api-key: sk-xxx`
  - Gemini：`x-goog-api-key: sk-xxx`
- **验收标准**:
  - [ ] 支持至少4种认证格式
  - [ ] 认证失败返回统一错误格式
  - [ ] 集成测试覆盖所有格式

##### 任务1.1.4: RBAC权限控制（3天）
- **数据库表**: `backend/migrations/000009_create_rbac_tables.up.sql`
- **表结构设计**:
  ```sql
  CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB
  );
  
  CREATE TABLE user_roles (
    user_id INT REFERENCES users(id),
    role_id INT REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
  );
  ```
- **权限模型**:
  - 角色：超级管理员/管理员/普通用户/开发者
  - 权限粒度：API端点级别
  - 权限继承：支持角色继承
- **中间件实现**:
  ```go
  func RequirePermission(permission string) gin.HandlerFunc
  func RequireRole(role string) gin.HandlerFunc
  ```
- **验收标准**:
  - [ ] 支持动态权限配置
  - [ ] 权限检查耗时<1ms
  - [ ] 有完整的权限审计日志
  - [ ] 前端展示权限可见性控制

### 1.2 API中转与路由系统
**当前状态**: 基础的Relay Service
**开发周期**: 2.5周
**优先级**: P0 (最高)

#### 任务分解

**Week 3-4: 流式响应与请求处理**

##### 任务1.2.1: SSE流式响应优化（4天）
- **文件**: `backend/cmd/relay/main.go`, `backend/internal/relay/stream.go`
- **实现内容**:
  ```go
  type StreamHandler struct {
    writer    http.ResponseWriter
    flusher   http.Flusher
    buffer    *bytes.Buffer
    done      chan struct{}
    heartbeat *time.Ticker
  }
  
  func (h *StreamHandler) WriteSSE(data []byte) error
  func (h *StreamHandler) StartHeartbeat() 
  func (h *StreamHandler) HandleError(err error)
  ```
- **技术要求**:
  - SSE格式规范：`data: {json}\n\n`
  - 心跳间隔：15秒发送一次注释保持连接
  - 超时控制：客户端90秒无响应自动断开
  - 错误流式传输：错误也以SSE格式返回
  - 支持中途取消：监听context.Done()
- **验收标准**:
  - [ ] 支持10000+并发流式连接
  - [ ] 流式延迟<50ms（首字节时间）
  - [ ] 连接异常自动清理，无泄漏
  - [ ] 压力测试：持续1小时无异常

##### 任务1.2.2: 智能请求重试机制（3天）
- **文件**: `backend/internal/relay/retry.go`
- **实现内容**:
  ```go
  type RetryConfig struct {
    MaxRetries    int
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    Multiplier    float64
  }
  
  func ShouldRetry(err error, statusCode int) bool
  func GetRetryDelay(attempt int, config RetryConfig) time.Duration
  func RetryWithBackoff(fn func() error, config RetryConfig) error
  ```
- **重试策略**:
  - 可重试错误：5xx、超时、网络错误、429限流
  - 不可重试：4xx（除429）、认证失败、参数错误
  - 指数退避：1s → 2s → 4s → 8s（最大30s）
  - 最大重试次数：3次
  - 渠道切换：单渠道失败2次自动切换
- **验收标准**:
  - [ ] 重试成功率提升至98%+
  - [ ] 单次重试耗时不超过配置的最大延迟
  - [ ] 所有重试有详细日志
  - [ ] 重试统计数据可查询

##### 任务1.2.3: 请求体缓存与恢复（2天）
- **文件**: `backend/internal/middleware/body_cache.go`
- **实现内容**:
  ```go
  type CachedBody struct {
    body   []byte
    reader *bytes.Reader
    mu     sync.Mutex
  }
  
  func BodyCacheMiddleware() gin.HandlerFunc
  func GetCachedBody(c *gin.Context) ([]byte, error)
  func RestoreBody(c *gin.Context) error
  ```
- **技术要求**:
  - 小请求（<1MB）：全量缓存到内存
  - 大请求（>1MB）：使用临时文件
  - 自动清理：请求结束后删除临时文件
  - 并发安全：使用读写锁保护
- **验收标准**:
  - [ ] 支持任意大小的请求体
  - [ ] 内存占用合理（<100MB总缓存）
  - [ ] 无临时文件泄漏
  - [ ] 性能损耗<5%

##### 任务1.2.4: 中继处理器抽象层（3天）
- **文件**: `backend/internal/relay/handler.go`
- **接口设计**:
  ```go
  type RelayHandler interface {
    ValidateRequest(req interface{}) error
    PrepareUpstreamRequest(req interface{}) (*http.Request, error)
    HandleResponse(resp *http.Response) (interface{}, error)
    HandleStreamResponse(resp *http.Response, writer StreamWriter) error
    ExtractUsage(resp interface{}) (*Usage, error)
  }
  
  type ChatCompletionHandler struct{}
  type EmbeddingHandler struct{}
  type ImageGenerationHandler struct{}
  type AudioHandler struct{}
  ```
- **处理器实现**:
  - Chat Completions：支持流式和非流式
  - Embeddings：批量embedding处理
  - Image Generation：异步生成+轮询
  - Audio：语音识别和合成
- **验收标准**:
  - [ ] 每种处理器有完整单元测试
  - [ ] 支持轻松扩展新的API类型
  - [ ] 错误处理统一且详细
  - [ ] 响应格式符合OpenAI标准

**Week 5 前半周: 集成测试与文档**

##### 任务1.2.5: 集成测试与性能优化（2天）
- **测试场景**:
  - 并发10000请求压力测试
  - 流式响应长时间稳定性测试
  - 重试机制正确性测试
  - 大请求（100MB）处理测试
- **性能指标**:
  - P99延迟：<500ms（非流式）
  - P99延迟：<100ms（流式首字节）
  - 吞吐量：>5000 QPS
  - 错误率：<0.1%
- **验收标准**:
  - [ ] 所有性能指标达标
  - [ ] 压力测试无内存泄漏
  - [ ] 有完整的API文档
  - [ ] 有监控指标和告警

### 1.3 渠道管理与负载均衡
**当前状态**: 简单的渠道选择
**开发周期**: 3周
**优先级**: P0 (最高)

#### 任务分解

**Week 5-6: 渠道缓存与选择算法**

##### 任务1.3.1: 渠道多级缓存系统（3天）
- **文件**: `backend/internal/model/channel.go`, `backend/internal/cache/channel_cache.go`
- **数据结构**:
  ```go
  type Channel struct {
    ID           int
    Name         string
    Type         int
    BaseURL      string
    Keys         []string
    Models       []string
    Group        string
    Status       int
    Priority     int
    Weight       int
    SuccessCount int64
    FailCount    int64
    LastCheck    time.Time
  }
  
  type ChannelCache struct {
    channels      map[int]*Channel
    groupChannels map[string][]*Channel
    modelChannels map[string][]*Channel
    mu            sync.RWMutex
  }
  ```
- **缓存策略**:
  - 全量加载：启动时加载所有渠道到内存
  - 增量更新：每30秒检查数据库变更
  - 热更新：配置变更立即生效，不中断服务
  - 多维索引：支持按ID、分组、模型查询
- **验收标准**:
  - [ ] 渠道查询耗时<1ms
  - [ ] 支持10000+渠道配置
  - [ ] 配置变更3秒内生效
  - [ ] 缓存内存占用<50MB

##### 任务1.3.2: 智能渠道选择算法（4天）
- **文件**: `backend/internal/service/channel_selector.go`
- **实现内容**:
  ```go
  type ChannelSelector struct {
    cache    *ChannelCache
    strategy LoadBalanceStrategy
  }
  
  func (s *ChannelSelector) SelectChannel(
    group string, 
    model string, 
    excludeIDs []int,
  ) (*Channel, error)
  
  func (s *ChannelSelector) SelectChannelWithRetry(
    group string, 
    model string, 
    maxRetries int,
  ) (*Channel, error)
  ```
- **选择流程**:
  1. 按用户分组过滤
  2. 按模型名称匹配（支持通配符）
  3. 过滤已失败渠道
  4. 检查渠道状态和配额
  5. 根据策略选择（权重/轮询/最少连接）
  6. 记录选择历史
- **模型匹配规则**:
  - 精确匹配：`gpt-4`
  - 前缀匹配：`gpt-4*` 匹配 `gpt-4-turbo`
  - 正则匹配：支持复杂规则
  - 模型映射：支持别名映射
- **验收标准**:
  - [ ] 选择耗时<5ms
  - [ ] 负载分布均匀（方差<15%）
  - [ ] 支持动态调整权重
  - [ ] 失败渠道自动跳过

##### 任务1.3.3: 多密钥轮询系统（2天）
- **文件**: `backend/internal/model/channel_key.go`
- **实现内容**:
  ```go
  type KeyManager struct {
    keys       []string
    mode       KeySelectionMode  // Random | RoundRobin
    index      int
    failCounts map[string]int
    mu         sync.Mutex
  }
  
  func (km *KeyManager) GetNextKey() (string, error)
  func (km *KeyManager) MarkKeyFailed(key string)
  func (km *KeyManager) ResetKeyStatus(key string)
  ```
- **轮询模式**:
  - Random：随机选择可用密钥
  - RoundRobin：按顺序轮询
  - FailureAware：跳过最近失败的密钥
- **失败处理**:
  - 密钥失败3次→标记为临时不可用（5分钟）
  - 所有密钥失败→渠道标记为不可用
  - 定时恢复：每5分钟重置失败计数
- **验收标准**:
  - [ ] 支持100+密钥/渠道
  - [ ] 密钥切换无感知
  - [ ] 失败密钥自动隔离
  - [ ] 有密钥使用统计

##### 任务1.3.4: 渠道健康检查（3天）
- **文件**: `backend/cmd/healthcheck/main.go`
- **实现内容**:
  ```go
  type HealthChecker struct {
    channels  []*Channel
    interval  time.Duration
    timeout   time.Duration
    results   map[int]*HealthResult
  }
  
  type HealthResult struct {
    ChannelID    int
    Status       HealthStatus
    Latency      time.Duration
    ErrorMessage string
    CheckedAt    time.Time
  }
  
  func (hc *HealthChecker) CheckChannel(channel *Channel) *HealthResult
  func (hc *HealthChecker) AutoDisable(channelID int)
  ```
- **检查策略**:
  - 检查频率：每5分钟
  - 检查方法：发送测试请求（/v1/models）
  - 超时设置：10秒
  - 失败阈值：连续3次失败→自动禁用
  - 自动恢复：禁用后30分钟重新检查
- **监控指标**:
  - 渠道可用率
  - 平均响应时间
  - 失败率统计
  - 配额使用情况
- **验收标准**:
  - [ ] 故障渠道5分钟内自动禁用
  - [ ] 恢复渠道30分钟内自动启用
  - [ ] 健康检查不影响正常业务
  - [ ] 有完整的健康历史记录

**Week 7: 负载均衡与能力管理**

##### 任务1.3.5: 负载均衡策略（2天）
- **文件**: `backend/internal/lb/strategy.go`
- **策略实现**:
  ```go
  type LoadBalanceStrategy interface {
    Select(channels []*Channel) *Channel
    UpdateMetrics(channelID int, latency time.Duration)
  }
  
  type WeightedRoundRobinStrategy struct{}  // 加权轮询
  type RandomStrategy struct{}               // 随机选择
  type LeastConnectionStrategy struct{}      // 最少连接
  type LeastLatencyStrategy struct{}         // 最低延迟
  ```
- **策略详情**:
  - 加权轮询：根据权重分配流量
  - 随机：完全随机，权重作为概率
  - 最少连接：选择当前连接数最少的渠道
  - 最低延迟：选择P99延迟最低的渠道
- **验收标准**:
  - [ ] 每种策略有单元测试
  - [ ] 支持动态切换策略
  - [ ] 策略选择耗时<1ms
  - [ ] 负载分布符合预期

##### 任务1.3.6: 渠道能力管理（2天）
- **数据库表**: `backend/migrations/000010_channel_abilities.up.sql`
- **表结构**:
  ```sql
  CREATE TABLE channel_abilities (
    id SERIAL PRIMARY KEY,
    channel_id INT REFERENCES channels(id),
    model_name VARCHAR(100) NOT NULL,
    support_stream BOOLEAN DEFAULT true,
    support_function_call BOOLEAN DEFAULT false,
    support_vision BOOLEAN DEFAULT false,
    max_tokens INT DEFAULT 4096,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(channel_id, model_name)
  );
  
  CREATE INDEX idx_abilities_model ON channel_abilities(model_name);
  ```
- **能力查询接口**:
  ```go
  func GetChannelAbilities(channelID int) ([]*Ability, error)
  func CheckChannelSupport(channelID int, model string, features []string) bool
  func UpdateChannelAbilities(channelID int, abilities []*Ability) error
  ```
- **验收标准**:
  - [ ] 支持动态添加新能力字段
  - [ ] 能力查询有缓存
  - [ ] 不支持的特性返回明确错误
  - [ ] 有能力版本管理

### 1.4 计费与配额系统
**当前状态**: 基础的Billing Service
**开发周期**: 2.5周
**优先级**: P0 (最高)

#### 任务分解

**Week 8-9: Token计数与计费核心**

##### 任务1.4.1: Token精准计数系统（4天）
- **文件**: `backend/internal/service/token_counter.go`
- **依赖库**: `github.com/pkoukk/tiktoken-go`
- **核心实现**:
  ```go
  type TokenCounter struct {
    encoders map[string]*tiktoken.Tiktoken
    mu       sync.RWMutex
  }
  
  func (tc *TokenCounter) CountTokens(model string, messages []Message) (int, error)
  func (tc *TokenCounter) CountMultiModal(text string, images []Image) (*Tokens, error)
  ```
- **支持模型编码**:
  - GPT-4/GPT-4-turbo: cl100k_base
  - GPT-3.5: cl100k_base
  - Claude: 按字符估算 (1 token ≈ 4 chars)
  - 通用模型: p50k_base
- **验收标准**:
  - [ ] 与OpenAI官方计数误差<2%
  - [ ] 计数性能>10000次/秒
  - [ ] 支持流式Token计数
  - [ ] 单元测试覆盖率>90%

##### 任务1.4.2: 预扣费与后扣费机制（3天）
- **文件**: `backend/internal/service/quota_service.go`
- **数据库表**: `backend/migrations/000011_quota_records.up.sql`
- **表设计**:
  ```sql
  CREATE TABLE quota_records (
    id VARCHAR(36) PRIMARY KEY,
    user_id INT NOT NULL,
    token_id INT NOT NULL,
    estimated_quota BIGINT,
    actual_quota BIGINT,
    status INT DEFAULT 1, -- 1:预扣 2:已确认 3:已退款
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP,
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_status (status, created_at)
  );
  ```
- **核心流程**:
  ```go
  // 预扣费
  func PreConsumeQuota(userID, tokenID int, quota int64) (recordID string, err error)
  // 后扣费（补差）
  func PostConsumeQuota(recordID string, actualQuota int64) error
  // 退款
  func RefundQuota(recordID string) error
  ```
- **验收标准**:
  - [ ] 并发TPS>1000，无冲突
  - [ ] 退款准确率100%
  - [ ] 定时任务清理超时记录
  - [ ] 有完整审计日志

##### 任务1.4.3: 模型定价系统（2天）
- **文件**: `backend/internal/model/pricing.go`
- **数据表**: `backend/migrations/000012_model_pricing.up.sql`
- **定价配置**:
  ```go
  type ModelPrice struct {
    Model          string
    InputPrice     float64  // 每1K tokens价格(美元)
    OutputPrice    float64
    ImagePrice     float64  // 每张图片
    AudioPrice     float64  // 每秒音频
    PriceType      int      // 1:按量 2:按次
    GroupMultiplier map[string]float64 // 分组倍率
  }
  ```
- **验收标准**:
  - [ ] 支持热更新价格
  - [ ] 历史价格可追溯
  - [ ] 分组倍率灵活配置
  - [ ] 价格缓存命中率>95%

**Week 10 前半周: 配额管理与异步记账**

##### 任务1.4.4: 配额管理与预警（2天）
- **实现功能**:
  ```go
  // 配额查询
  func GetUserQuota(userID int) (*Quota, error)
  // 配额预警
  func CheckQuotaAlert(userID int, threshold float64) bool
  // 配额充值
  func RechargeQuota(userID int, amount int64, expireAt time.Time) error
  ```
- **预警规则**:
  - 剩余10%时发送邮件/站内信
  - 剩余5%时限制部分功能
  - 耗尽时禁用服务
- **验收标准**:
  - [ ] 预警延迟<30秒
  - [ ] 充值实时生效
  - [ ] 支持配额有效期
  - [ ] 过期配额自动清零

##### 任务1.4.5: 异步记账系统（2天）
- **消息队列**: RabbitMQ
- **实现内容**:
  ```go
  type BillingEvent struct {
    UserID      int
    TokenID     int
    Model       string
    InputTokens int
    OutputTokens int
    TotalQuota  int64
    RequestID   string
    Timestamp   time.Time
  }
  
  func PublishBillingEvent(event *BillingEvent) error
  func ConsumeBillingEvents() error
  ```
- **消费者处理**:
  1. 从队列获取计费事件
  2. 写入usage_logs表
  3. 更新统计数据
  4. 触发预警检查
- **验收标准**:
  - [ ] 消息处理延迟<5秒
  - [ ] 失败消息自动重试3次
  - [ ] 死信队列处理
  - [ ] 消息不丢失不重复

---

## 二、用户服务功能 (User-Facing Features)

### 2.1 AI对话系统优化
**当前状态**: Chat Service处理对话
**开发周期**: 2周
**优先级**: P1 (高)

#### 任务分解

**Week 11-12: 对话核心与上下文管理**

##### 任务2.1.1: 流式对话优化（3天）
- **文件**: `backend/cmd/chat/main.go`, `backend/internal/service/chat_service.go`
- **核心实现**:
  ```go
  type ChatService struct {
    relayClient *relay.Client
    msgRepo     *MessageRepository
  }
  
  func (cs *ChatService) StreamChat(
    sessionID string,
    messages []Message,
    stream chan<- string,
  ) error
  ```
- **流式处理流程**:
  1. 接收用户消息，保存到数据库
  2. 调用Relay Service获取流式响应
  3. 逐chunk转发给前端
  4. 完成后保存完整AI响应
  5. 更新会话更新时间
- **验收标准**:
  - [ ] 首字节延迟<200ms
  - [ ] 支持中途取消
  - [ ] 连接断开自动重试
  - [ ] 流式数据完整性保证

##### 任务2.1.2: 上下文管理系统（4天）
- **文件**: `backend/internal/service/context_manager.go`
- **实现内容**:
  ```go
  type ContextManager struct {
    maxTokens     int
    summaryPrompt string
  }
  
  // 智能截断
  func (cm *ContextManager) TruncateMessages(
    messages []Message, 
    maxTokens int,
  ) []Message
  
  // 上下文总结
  func (cm *ContextManager) SummarizeContext(
    messages []Message,
  ) (summary string, err error)
  ```
- **压缩策略**:
  - 保留最近N条消息（默认10条）
  - 超出部分进行总结
  - 系统消息始终保留
  - 重要消息标记保留
- **窗口管理**:
  - 动态计算可用token数
  - 根据模型调整窗口大小
  - 支持手动设置窗口
- **验收标准**:
  - [ ] Token计数误差<5%
  - [ ] 总结保留关键信息>90%
  - [ ] 压缩后对话质量不下降
  - [ ] 支持100+轮对话

##### 任务2.1.3: 消息格式化系统（2天）
- **文件**: `backend/internal/formatter/message_formatter.go`
- **支持格式**:
  ```go
  type MessageFormatter struct{}
  
  // Markdown处理
  func (mf *MessageFormatter) FormatMarkdown(text string) string
  // 代码块提取
  func (mf *MessageFormatter) ExtractCodeBlocks(text string) []CodeBlock
  // LaTeX公式识别
  func (mf *MessageFormatter) ParseLatex(text string) []LatexBlock
  ```
- **处理内容**:
  - Markdown: 标题、列表、引用、链接
  - 代码: 语言识别、语法高亮标记
  - LaTeX: 公式提取和标记
  - 多媒体: URL识别和预览生成
- **验收标准**:
  - [ ] Markdown渲染准确率>99%
  - [ ] 支持30+编程语言识别
  - [ ] LaTeX公式正确解析
  - [ ] 格式化耗时<10ms

##### 任务2.1.4: AI Agent系统（3天）
- **文件**: `backend/internal/agent/general_chat_agent.go`
- **参考**: LobeChat GeneralChatAgent.ts
- **核心实现**:
  ```go
  type GeneralChatAgent struct {
    systemPrompt string
    model        string
    temperature  float64
    maxTokens    int
    tools        []Tool
  }
  
  func (agent *GeneralChatAgent) Execute(
    sessionID string,
    userMessage string,
  ) (*Response, error)
  
  func (agent *GeneralChatAgent) ExecuteWithTools(
    sessionID string,
    userMessage string,
  ) (*Response, error)
  ```
- **Agent配置**:
  - 系统提示词模板
  - 模型参数配置
  - 工具函数注册
  - 能力扩展接口
- **验收标准**:
  - [ ] 支持自定义系统提示词
  - [ ] 支持工具调用
  - [ ] Agent可热更新配置
  - [ ] 有完整的执行日志

### 2.2 会话管理
**当前状态**: 基础的对话记录
**开发周期**: 1.5周
**优先级**: P1 (高)

#### 任务分解

**Week 13: 会话核心功能**

##### 任务2.2.1: 会话CRUD与生命周期（3天）
- **数据库完善**: `backend/migrations/000013_enhance_sessions.up.sql`
- **表结构增强**:
  ```sql
  ALTER TABLE sessions ADD COLUMN config JSONB;
  ALTER TABLE sessions ADD COLUMN tags TEXT[];
  ALTER TABLE sessions ADD COLUMN is_archived BOOLEAN DEFAULT false;
  ALTER TABLE sessions ADD COLUMN archived_at TIMESTAMP;
  CREATE INDEX idx_sessions_archived ON sessions(is_archived, updated_at);
  CREATE INDEX idx_sessions_tags ON sessions USING GIN(tags);
  ```
- **API实现**:
  ```go
  func CreateSession(userID int, title string, config *SessionConfig) (*Session, error)
  func UpdateSession(sessionID string, updates map[string]interface{}) error
  func ArchiveSession(sessionID string) error
  func DeleteSession(sessionID string) error
  func SearchSessions(userID int, query string, filters *Filters) ([]*Session, error)
  ```
- **验收标准**:
  - [ ] CRUD操作耗时<50ms
  - [ ] 支持软删除和恢复
  - [ ] 全文搜索准确率>95%
  - [ ] 支持批量操作

##### 任务2.2.2: 会话配置系统（2天）
- **文件**: `backend/internal/model/session_config.go`
- **配置结构**:
  ```go
  type SessionConfig struct {
    AgentConfig  AgentConfig  `json:"agent_config"`
    ModelConfig  ModelConfig  `json:"model_config"`
    PluginIDs    []int        `json:"plugin_ids"`
    KnowledgeIDs []int        `json:"knowledge_ids"`
  }
  
  type ModelConfig struct {
    Model       string  `json:"model"`
    Temperature float64 `json:"temperature"`
    MaxTokens   int     `json:"max_tokens"`
    TopP        float64 `json:"top_p"`
  }
  ```
- **功能实现**:
  - 配置模板管理
  - 配置继承和覆盖
  - 配置版本控制
  - 配置导入导出
- **验收标准**:
  - [ ] 配置变更实时生效
  - [ ] 支持配置回滚
  - [ ] 配置验证完整
  - [ ] 有配置变更历史

##### 任务2.2.3: 消息高级管理（3天）
- **数据库增强**: `backend/migrations/000014_enhance_messages.up.sql`
- **表结构**:
  ```sql
  ALTER TABLE messages ADD COLUMN parent_id UUID REFERENCES messages(id);
  ALTER TABLE messages ADD COLUMN branch_id VARCHAR(36);
  ALTER TABLE messages ADD COLUMN edited_at TIMESTAMP;
  ALTER TABLE messages ADD COLUMN attachments JSONB;
  CREATE INDEX idx_messages_branch ON messages(session_id, branch_id);
  ```
- **功能实现**:
  ```go
  // 消息分支
  func CreateBranch(messageID string) (branchID string, error)
  func SwitchBranch(sessionID, branchID string) error
  // 消息编辑
  func EditMessage(messageID string, newContent string) error
  // 消息引用
  func QuoteMessage(sessionID, messageID string) error
  // 消息导出
  func ExportSession(sessionID string, format string) ([]byte, error)
  ```
- **验收标准**:
  - [ ] 支持无限分支
  - [ ] 分支切换无延迟
  - [ ] 导出格式完整
  - [ ] 支持Markdown/JSON/PDF导出

**Week 14 前半周: 会话共享与模板**

##### 任务2.2.4: 会话共享系统（2天）
- **数据库表**: `backend/migrations/000015_session_shares.up.sql`
- **表结构**:
  ```sql
  CREATE TABLE session_shares (
    id VARCHAR(36) PRIMARY KEY,
    session_id UUID REFERENCES sessions(id),
    user_id INT REFERENCES users(id),
    share_mode INT DEFAULT 1, -- 1:只读 2:可评论 3:可编辑
    expire_at TIMESTAMP,
    view_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
- **功能实现**:
  ```go
  func ShareSession(sessionID string, mode int, expireAt time.Time) (shareID string, error)
  func GetSharedSession(shareID string) (*Session, error)
  func RevokeShare(shareID string) error
  ```
- **验收标准**:
  - [ ] 分享链接支持密码保护
  - [ ] 支持访问统计
  - [ ] 过期自动失效
  - [ ] 有权限控制

### 2.3 知识库与RAG系统
**当前状态**: 未实现
**开发周期**: 3周
**优先级**: P1 (高)

#### 任务分解

**Week 14-15: 文档处理与向量化**

##### 任务2.3.1: 文档上传与解析服务（4天）
- **新建服务**: `backend/cmd/knowledge/main.go`
- **数据库表**: `backend/migrations/000016_knowledge_base.up.sql`
- **表结构**:
  ```sql
  CREATE TABLE knowledge_bases (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    embedding_model VARCHAR(50),
    chunk_size INT DEFAULT 500,
    chunk_overlap INT DEFAULT 50,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    kb_id INT REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    filename VARCHAR(255),
    file_type VARCHAR(50),
    file_size BIGINT,
    file_url TEXT,
    status INT DEFAULT 1, -- 1:上传中 2:处理中 3:完成 4:失败
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
- **文件解析器**:
  ```go
  type DocumentParser interface {
    Parse(file io.Reader) (*ParsedContent, error)
  }
  
  type PDFParser struct{}
  type WordParser struct{}
  type ExcelParser struct{}
  type TextParser struct{}
  ```
- **支持格式**:
  - PDF: 使用pdfcpu库
  - Word: 使用docx库
  - Excel: 使用excelize库
  - TXT/Markdown: 直接读取
  - HTML: 提取正文
- **验收标准**:
  - [ ] 支持10+种文件格式
  - [ ] 单文件解析<30秒
  - [ ] 解析准确率>95%
  - [ ] 支持100MB+大文件

##### 任务2.3.2: 智能文本分块系统（3天）
- **文件**: `backend/internal/rag/chunker.go`
- **实现内容**:
  ```go
  type Chunker struct {
    chunkSize    int
    chunkOverlap int
    separators   []string
  }
  
  func (c *Chunker) ChunkText(text string) ([]*Chunk, error)
  func (c *Chunker) ChunkWithMetadata(doc *Document) ([]*Chunk, error)
  ```
- **分块策略**:
  - 按段落分块（优先）
  - 按句子分块
  - 按固定长度分块
  - 保留重叠部分（默认50 tokens）
- **元数据保留**:
  - 文档来源
  - 页码/章节
  - 标题层级
  - 创建时间
- **验收标准**:
  - [ ] 分块大小均匀（方差<20%）
  - [ ] 保留文档结构
  - [ ] 分块耗时<5秒/MB
  - [ ] 支持中英文

##### 任务2.3.3: 向量化与存储系统（4天）
- **数据库表**: `backend/migrations/000017_embeddings.up.sql`
- **表结构**:
  ```sql
  CREATE EXTENSION IF NOT EXISTS vector;
  
  CREATE TABLE chunks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    chunk_index INT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE TABLE embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chunk_id UUID REFERENCES chunks(id) ON DELETE CASCADE,
    embedding vector(1536), -- OpenAI ada-002维度
    model VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE INDEX ON embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
  ```
- **Embedding实现**:
  ```go
  type EmbeddingService struct {
    model string
    apiKey string
  }
  
  func (es *EmbeddingService) Embed(texts []string) ([][]float32, error)
  func (es *EmbeddingService) EmbedBatch(texts []string, batchSize int) error
  ```
- **支持模型**:
  - OpenAI text-embedding-ada-002
  - OpenAI text-embedding-3-small/large
  - 本地模型（可选）
- **验收标准**:
  - [ ] 批量embedding速度>100条/秒
  - [ ] 向量存储成功率100%
  - [ ] 索引构建时间<10分钟/100万条
  - [ ] 支持增量更新

**Week 16: 语义检索与RAG集成**

##### 任务2.3.4: 语义检索引擎（3天）
- **文件**: `backend/internal/rag/retriever.go`
- **核心实现**:
  ```go
  type Retriever struct {
    db            *gorm.DB
    embeddingSvc  *EmbeddingService
    rerankModel   string
  }
  
  func (r *Retriever) Search(
    query string,
    kbID int,
    topK int,
  ) ([]*SearchResult, error)
  
  func (r *Retriever) HybridSearch(
    query string,
    kbID int,
    topK int,
    alpha float64, // 向量权重
  ) ([]*SearchResult, error)
  ```
- **检索策略**:
  - 纯向量检索：余弦相似度
  - 混合检索：向量(70%) + BM25(30%)
  - 重排序：使用cross-encoder
  - 过滤：基于元数据过滤
- **验收标准**:
  - [ ] 检索延迟<200ms
  - [ ] Top-10准确率>85%
  - [ ] 支持百万级文档检索
  - [ ] 混合检索效果提升>15%

##### 任务2.3.5: RAG流程集成（2天）
- **文件**: `backend/internal/service/rag_service.go`
- **集成流程**:
  ```go
  func (rs *RAGService) EnhancePrompt(
    sessionID string,
    userQuery string,
  ) (*EnhancedPrompt, error) {
    // 1. 获取会话关联的知识库
    // 2. 执行语义检索
    // 3. 构建增强提示词
    // 4. 返回带引用的提示
  }
  ```
- **提示词模板**:
  ```
  根据以下参考信息回答用户问题：
  
  === 参考信息 ===
  [1] {chunk1_content} (来源: {doc1_name}, 页码: {page})
  [2] {chunk2_content} (来源: {doc2_name}, 页码: {page})
  
  === 用户问题 ===
  {user_query}
  
  请基于参考信息回答，并标注引用来源。
  ```
- **验收标准**:
  - [ ] RAG回答质量提升>30%
  - [ ] 引用来源准确率>95%
  - [ ] 支持自动触发/手动触发
  - [ ] 有检索质量评估

### 2.4 插件与工具调用
**当前状态**: 未实现
**开发周期**: 2周
**优先级**: P2 (中)

#### 任务分解

**Week 17: Function Calling核心**

##### 任务2.4.1: Function Calling引擎（4天）
- **文件**: `backend/internal/tools/function_engine.go`
- **参考**: LobeChat createToolEngine.ts
- **核心实现**:
  ```go
  type FunctionEngine struct {
    tools map[string]Tool
  }
  
  type Tool struct {
    Name        string
    Description string
    Parameters  JSONSchema
    Handler     func(args map[string]interface{}) (interface{}, error)
  }
  
  func (fe *FunctionEngine) RegisterTool(tool Tool) error
  func (fe *FunctionEngine) ExecuteTool(name string, args map[string]interface{}) (interface{}, error)
  func (fe *FunctionEngine) ConvertToOpenAIFormat() []OpenAIFunction
  ```
- **工具调用流程**:
  1. AI模型返回function_call
  2. 解析函数名和参数
  3. 验证参数schema
  4. 执行工具函数
  5. 将结果返回给AI
  6. 获取最终回复
- **验收标准**:
  - [ ] 支持OpenAI Function Calling格式
  - [ ] 参数验证准确率100%
  - [ ] 工具执行超时控制
  - [ ] 有详细的执行日志

##### 任务2.4.2: 内置工具实现（3天）
- **文件**: `backend/internal/tools/builtin/`
- **工具列表**:
  ```go
  // 网络搜索
  type WebSearchTool struct {
    apiKey string
    engine string // google/bing/duckduckgo
  }
  
  // 代码执行（沙箱）
  type CodeExecutorTool struct {
    timeout   time.Duration
    languages []string
  }
  
  // 文件操作
  type FileOperationTool struct {
    allowedPaths []string
  }
  
  // HTTP请求
  type HTTPRequestTool struct {
    allowedDomains []string
  }
  ```
- **安全措施**:
  - 代码执行：Docker沙箱隔离
  - 文件操作：路径白名单限制
  - HTTP请求：域名白名单
  - 执行时间：统一超时30秒
- **验收标准**:
  - [ ] 至少实现4个内置工具
  - [ ] 沙箱隔离无法逃逸
  - [ ] 工具执行成功率>95%
  - [ ] 有完整的错误处理

**Week 18: 插件系统与管理**

##### 任务2.4.3: MCP插件系统（4天）
- **新建服务**: `backend/cmd/plugin/main.go`
- **数据库表**: `backend/migrations/000018_plugins.up.sql`
- **表结构**:
  ```sql
  CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    version VARCHAR(20),
    type INT, -- 1:stdio 2:http 3:builtin
    config JSONB,
    status INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE TABLE user_plugins (
    user_id INT REFERENCES users(id),
    plugin_id INT REFERENCES plugins(id),
    enabled BOOLEAN DEFAULT true,
    config JSONB,
    PRIMARY KEY (user_id, plugin_id)
  );
  ```
- **MCP协议实现**:
  ```go
  type MCPClient struct {
    pluginType string // stdio or http
    conn       net.Conn
  }
  
  func (mc *MCPClient) CallTool(toolName string, args interface{}) (interface{}, error)
  func (mc *MCPClient) ListTools() ([]Tool, error)
  func (mc *MCPClient) GetToolSchema(toolName string) (*JSONSchema, error)
  ```
- **验收标准**:
  - [ ] 支持stdio和http两种模式
  - [ ] 插件崩溃不影响主服务
  - [ ] 插件调用延迟<500ms
  - [ ] 有插件健康检查

##### 任务2.4.4: 插件管理界面（3天）
- **API实现**:
  ```go
  func InstallPlugin(name, version string) error
  func EnablePlugin(userID, pluginID int) error
  func DisablePlugin(userID, pluginID int) error
  func ConfigurePlugin(userID, pluginID int, config map[string]interface{}) error
  func ListAvailablePlugins() ([]*Plugin, error)
  ```
- **功能实现**:
  - 插件市场浏览
  - 插件搜索过滤
  - 插件安装/卸载
  - 插件配置管理
  - 插件权限控制
- **验收标准**:
  - [ ] 支持热插拔
  - [ ] 配置变更实时生效
  - [ ] 有插件依赖检查
  - [ ] 有权限审计日志

---

## 三、开发者服务功能 (Developer-Facing Features)

### 3.1 多模型适配器
**当前状态**: 基础转发
**开发周期**: 2周
**优先级**: P0 (最高)

#### 任务分解

**Week 19: 适配器核心框架**

##### 任务3.1.1: 适配器接口设计（2天）
- **文件**: `backend/internal/adapter/adapter.go`
- **参考**: New API adapter.go
- **接口定义**:
  ```go
  type Adaptor interface {
    // 请求转换
    ConvertRequest(req *OpenAIRequest) (interface{}, error)
    // 发送请求
    DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error)
    // 响应解析
    ParseResponse(resp *http.Response) (*OpenAIResponse, error)
    // 流式响应
    ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error)
    // 提取使用量
    ExtractUsage(resp interface{}) (*Usage, error)
    // 获取支持的模型
    GetSupportedModels() []string
  }
  ```
- **验收标准**:
  - [ ] 接口设计清晰易扩展
  - [ ] 有完整的接口文档
  - [ ] 有接口使用示例
  - [ ] 单元测试覆盖率>90%

##### 任务3.1.2: 核心提供商适配器实现（5天）
- **文件**: `backend/internal/adapter/providers/`
- **实现提供商**:
  ```go
  // OpenAI (基准)
  type OpenAIAdaptor struct{}
  
  // Anthropic Claude
  type ClaudeAdaptor struct{}
  
  // Google Gemini
  type GeminiAdaptor struct{}
  
  // 百度文心一言
  type BaiduAdaptor struct{}
  
  // 阿里通义千问
  type QwenAdaptor struct{}
  ```
- **转换重点**:
  - 消息格式差异处理
  - 参数名称映射
  - 多模态内容转换
  - 工具调用格式转换
  - 错误码映射
- **验收标准**:
  - [ ] 每个提供商有完整测试
  - [ ] 格式转换准确率100%
  - [ ] 支持流式和非流式
  - [ ] 错误处理完整

**Week 20: 扩展与优化**

##### 任务3.1.3: 批量适配器实现（3天）
- **扩展提供商**（每个0.5天）:
  - DeepSeek
  - Moonshot（月之暗面）
  - MiniMax
  - 智谱AI
  - 讯飞星火
  - Cohere
- **通用适配器**:
  ```go
  type GenericAdaptor struct {
    mapping ConfigMapping // 配置化映射
  }
  ```
- **验收标准**:
  - [ ] 至少支持10个提供商
  - [ ] 通用适配器可配置化
  - [ ] 新增提供商<1天

##### 任务3.1.4: 适配器注册与管理（2天）
- **文件**: `backend/internal/adapter/registry.go`
- **实现内容**:
  ```go
  type AdaptorRegistry struct {
    adaptors map[int]*AdaptorInfo
  }
  
  func (ar *AdaptorRegistry) Register(apiType int, adaptor Adaptor) error
  func (ar *AdaptorRegistry) Get(apiType int) (Adaptor, error)
  func (ar *AdaptorRegistry) List() []AdaptorInfo
  ```
- **验收标准**:
  - [ ] 支持动态注册
  - [ ] 并发安全
  - [ ] 有适配器版本管理
  - [ ] 支持适配器热更新

### 3.2 API密钥管理
**当前状态**: 简单的token验证
**开发周期**: 1周
**优先级**: P1 (高)

#### 任务分解 (Week 21)

##### 任务3.2.1: Token生命周期完善（2天）
- 增强表字段（已在1.1.2完成部分）
- Token分类管理（开发/生产/测试）
- Token自动轮换机制
- 验收：支持1000+ tokens/用户

##### 任务3.2.2: Token权限系统（2天）
- 模型白名单/黑名单配置
- IP白名单管理
- 请求来源验证
- 验收：权限检查<1ms

##### 任务3.2.3: Token安全加固（1天）
- AES加密存储
- 异常行为检测（流量突增、异地登录）
- Token泄露监控
- 验收：安全事件检测率>95%

### 3.3 使用量统计与分析
**当前状态**: 基础记录
**开发周期**: 1.5周
**优先级**: P1 (高)

#### 任务分解 (Week 22-23前半)

##### 任务3.3.1: 详细用量记录（2天）
- 表设计: `000019_usage_analytics.up.sql`
- 记录字段：时间、用户、Token、模型、输入/输出tokens、费用、渠道、延迟、状态
- 分区表按月分区
- 验收：写入TPS>5000

##### 任务3.3.2: 实时统计引擎（3天）
- Redis + ClickHouse/TimescaleDB
- 实时QPS/Token使用量/费用统计
- 滑动窗口聚合（5分钟/小时/天）
- 验收：统计延迟<5秒

##### 任务3.3.3: 数据分析API（2天）
- RESTful API接口
- 多维度查询（时间/用户/模型/Token）
- 数据可视化支持
- 验收：复杂查询<1秒

### 3.4 速率限制与配额控制
**当前状态**: 基础配额扣除
**开发周期**: 1周
**优先级**: P1 (高)

#### 任务分解 (Week 23后半 - Week 24)

##### 任务3.4.1: 多级限流中间件（2天）
- 文件: `backend/internal/middleware/ratelimit.go`
- 用户/Token/IP/模型四级限流
- Redis Lua脚本原子操作
- 验收：限流判断<3ms

##### 任务3.4.2: 限流算法实现（2天）
- 滑动窗口算法（精确限流）
- Token Bucket（流量整形）
- 分布式限流（Redis）
- 验收：算法准确率100%

##### 任务3.4.3: 配额控制增强（1天）
- 日/月配额管理
- 并发数限制
- 超限降级策略
- 验收：429响应带Retry-After

---

## 四、数据层优化 (Data Layer)
**总开发周期**: 2周 (Week 25-26)
**优先级**: P1

### 4.1 数据库Schema完善 (1周)
#### 任务4.1.1: Schema审查与优化 (3天)
- 补充所有缺失表和字段
- 索引优化（高频查询添加复合索引）
- 外键约束和级联规则
- 验收：查询性能提升>50%

#### 任务4.1.2: 数据迁移脚本 (2天)
- 编写20+个迁移脚本
- 迁移脚本回滚测试
- 生产环境迁移方案
- 验收：零停机迁移

### 4.2 Redis缓存系统 (0.5周)
#### 任务4.2.1: Redis集成 (2天)
- 初始化Redis客户端（支持集群）
- 实现缓存中间件
- 热数据缓存策略
- 验收：缓存命中率>90%

#### 任务4.2.2: 缓存同步 (1天)
- 数据变更时缓存失效
- 分布式缓存一致性
- 缓存预热机制
- 验收：缓存延迟<10ms

### 4.3 消息队列 (0.5周)
#### 任务4.3.1: RabbitMQ部署 (1天)
- MQ集群部署
- 交换机和队列配置
- 消费者实现
- 验收：消息不丢失

#### 任务4.3.2: 异步任务队列实现 (2天)
- **文件**: `backend/pkg/queue/consumer.go`
- **队列定义**:
  ```go
  const (
    QueueBilling      = "billing"      // 计费队列
    QueueNotification = "notification" // 通知队列
    QueueAnalytics    = "analytics"    // 分析队列
    QueueExport       = "export"       // 导出队列
  )
  ```
- **消费者实现**:
  - 计费消费者：处理usage_logs写入
  - 通知消费者：发送邮件/站内信
  - 分析消费者：统计数据聚合
  - 导出消费者：生成CSV/Excel
- **错误处理**:
  - 失败重试3次（指数退避）
  - 死信队列（DLQ）处理
  - 消息持久化
- **验收标准**:
  - [ ] 消息处理延迟<5秒
  - [ ] 消息不丢失不重复
  - [ ] 有完整的监控指标
  - [ ] 支持优雅关闭

- ✅ **队列监控**:
  - 消息积压监控
  - 消费速率统计
  - 失败消息告警

---

## 五、前端应用 (Frontend)

### 5.1 用户界面
**当前状态**: 基础React应用
**开发周期**: 1.5周 (Week 27-28前半)
**优先级**: P1 (高)

#### 任务分解

##### 任务5.1.1: Next.js框架搭建（2天）
- **项目初始化**:
  ```bash
  npx create-next-app@latest oblivious-frontend --typescript --tailwind --app
  ```
- **配置清单**:
  - TailwindCSS配置（`tailwind.config.ts`）
  - shadcn/ui安装（`npx shadcn-ui@latest init`）
  - 路由结构设计（App Router）
  - 环境变量配置
- **基础布局**:
  - 响应式导航栏
  - 侧边栏组件
  - 主内容区域
- **验收标准**:
  - [ ] 项目能正常启动
  - [ ] TailwindCSS生效
  - [ ] 响应式布局完整
  - [ ] 首屏加载<2秒

##### 任务5.1.2: 对话界面实现（3天）
- **文件**: `frontend/src/app/chat/page.tsx`, `frontend/src/components/ChatPanel.tsx`
- **核心组件**:
  ```typescript
  interface ChatPanelProps {
    sessionId: string;
    messages: Message[];
    onSendMessage: (content: string) => void;
  }
  ```
- **功能实现**:
  - 流式打字效果（逐字显示）
  - Markdown渲染（react-markdown）
  - 代码高亮（prism-react-renderer）
  - 多模态消息（图片预览、文件下载）
  - 消息操作（复制、重新生成、编辑）
- **验收标准**:
  - [ ] 流式显示流畅
  - [ ] Markdown渲染正确
  - [ ] 代码高亮支持20+语言
  - [ ] 图片加载有占位符

##### 任务5.1.3: 会话管理界面（2天）
- **文件**: `frontend/src/components/SessionSidebar.tsx`
- **功能实现**:
  ```typescript
  - 会话列表（分页加载）
  - 新建会话按钮
  - 会话搜索（实时过滤）
  - 会话分组显示
  - 拖拽排序
  - 右键菜单（重命名/删除/归档）
  ```
- **验收标准**:
  - [ ] 支持1000+会话流畅滚动
  - [ ] 搜索响应<100ms
  - [ ] 拖拽顺滑无卡顿
  - [ ] 操作有动画反馈

### 5.2 开发者控制台
**当前状态**: 未实现
**开发周期**: 1周 (Week 28后半-29前半)
**优先级**: P1 (高)

#### 任务分解

##### 任务5.2.1: 控制台框架（2天）
- **文件**: `frontend/src/app/developer/page.tsx`
- **布局设计**:
  - 左侧导航菜单
  - 顶部面包屑导航
  - 主内容区域
- **仪表盘实现**:
  ```typescript
  - 实时数据卡片（今日请求、费用、成功率）
  - ECharts折线图（请求趋势）
  - 饼图（模型分布）
  - 表格（最近API调用）
  ```
- **验收标准**:
  - [ ] 仪表盘数据实时更新（5秒刷新）
  - [ ] 图表交互流畅
  - [ ] 支持日期范围筛选

##### 任务5.2.2: API密钥管理（2天）
- **文件**: `frontend/src/app/developer/tokens/page.tsx`
- **功能实现**:
  - 密钥列表展示（表格）
  - 创建密钥对话框
  - 密钥权限配置
  - 复制/删除/禁用操作
  - 使用量统计
- **验收标准**:
  - [ ] 创建后显示一次完整密钥
  - [ ] 操作有确认提示
  - [ ] 支持批量操作

##### 任务5.2.3: 使用统计与账单（1天）
- **文件**: `frontend/src/app/developer/usage/page.tsx`
- **功能实现**:
  - 使用量图表（ECharts折线图/柱状图）
  - 时间范围选择器
  - 模型筛选器
  - 账单明细表格
  - 数据导出（CSV）
- **验收标准**:
  - [ ] 图表加载<1秒
  - [ ] 支持导出10000+条记录
  - [ ] 实时数据更新

### 5.3 实时通信
**当前状态**: 基础HTTP
**开发周期**: 0.5周 (Week 29后半)
**优先级**: P0 (最高)

#### 任务分解

##### 任务5.3.1: SSE流式客户端（2天）
- **文件**: `frontend/src/services/chat.ts`
- **核心实现**:
  ```typescript
  export class ChatService {
    private eventSource: EventSource | null = null;
    
    streamChat(sessionId: string, message: string, onChunk: (text: string) => void) {
      const url = `/api/chat/stream?session=${sessionId}`;
      this.eventSource = new EventSource(url);
      
      this.eventSource.onmessage = (event) => {
        const data = JSON.parse(event.data);
        onChunk(data.content);
      };
    }
  }
  ```
- **功能要点**:
  - 自动重连机制（3次重试）
  - 错误处理
  - 连接超时检测
  - 停止生成功能
- **验收标准**:
  - [ ] 流式延迟<100ms
  - [ ] 断线自动重连
  - [ ] 无内存泄漏

##### 任务5.3.2: 状态管理（1天）
- **文件**: `frontend/src/store/chatStore.ts`
- **使用Zustand**:
  ```typescript
  interface ChatStore {
    sessions: Session[];
    currentSession: Session | null;
    addMessage: (message: Message) => void;
    updateMessage: (id: string, content: string) => void;
  }
  ```
- **验收标准**:
  - [ ] 状态更新无闪烁
  - [ ] 支持离线缓存
  - [ ] 乐观更新

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

## 八、开发优先级与时间规划（更新版）

> 基于详细任务分解，总开发周期调整为 **32周（8个月）**

### Phase 1: 核心功能完善 (10周)
**目标**: 构建稳定可靠的API中转核心

#### Week 1-2: 认证授权系统 (1.1节)
- Token多级缓存机制（L1+L2）
- Token状态管理（5种状态）
- 多种认证方式（4种格式）
- RBAC权限控制

#### Week 3-5: API中转与路由 (1.2节)
- SSE流式响应优化（支持10000+并发）
- 智能请求重试（指数退避）
- 请求体缓存与恢复
- 中继处理器抽象（4种API类型）
- 集成测试

#### Week 5-7: 渠道管理与负载均衡 (1.3节)
- 渠道多级缓存（支持10000+渠道）
- 智能选择算法（通配符匹配）
- 多密钥轮询（3种模式）
- 健康检查（自动禁用/恢复）
- 负载均衡策略（4种）
- 渠道能力管理

#### Week 8-10: 计费与配额系统 (1.4节)
- Token精准计数（tiktoken）
- 预扣费/后扣费机制
- 模型定价系统
- 配额管理与预警
- RabbitMQ异步记账

**交付成果**:
- API中转服务（QPS>5000，P99<500ms）
- 支持10+AI提供商
- 完善的计费系统（TPS>1000）
- 基础监控指标

### Phase 2: 用户服务增强 (8.5周)
**目标**: 打造优秀的用户对话体验

#### Week 11-12: AI对话系统 (2.1节)
- 流式对话优化（首字节<200ms）
- 上下文管理（支持100+轮对话）
- 消息格式化（Markdown/LaTeX/代码）
- AI Agent系统

#### Week 13: 会话管理 (2.2节)
- 会话CRUD与生命周期
- 会话配置系统（JSONB）
- 消息高级管理（分支/导出）
- 会话共享系统

#### Week 14-16: 知识库与RAG (2.3节)
- 文档上传与解析（10+格式）
- 智能文本分块
- 向量化与存储（pgvector）
- 语义检索引擎（延迟<200ms）
- RAG流程集成

#### Week 17-18: 插件与工具调用 (2.4节)
- Function Calling引擎
- 内置工具实现（4个）
- MCP插件系统（stdio/http）
- 插件管理界面

**交付成果**:
- 完整的对话系统（首字节<200ms）
- 会话管理（支持分支/导出）
- RAG系统（Top-10准确率>85%）
- 插件系统（支持热插拔）

### Phase 3: 开发者服务增强 (5.5周)
**目标**: 提供完善的开发者工具和文档

#### Week 19-20: 多模型适配器 (3.1节)
- 适配器接口设计
- 核心提供商实现（5个）
- 批量适配器（6个）
- 适配器注册与管理

#### Week 21: API密钥管理 (3.2节)
- Token生命周期完善
- Token权限系统
- Token安全加固

#### Week 22-23: 使用统计与限流 (3.3-3.4节)
- 详细用量记录（分区表）
- 实时统计引擎（Redis+ClickHouse）
- 数据分析API
- 多级限流中间件
- 限流算法实现

**交付成果**:
- 支持10+AI提供商适配器
- 完整的API密钥管理
- 实时统计仪表盘（延迟<5秒）
- 多级限流系统（判断<3ms）

### Phase 4: 数据层优化 (2周)
**目标**: 完善数据存储和异步处理

#### Week 25-26: 数据层 (第四部分)
- Schema审查与索引优化
- 数据迁移脚本（20+个）
- Redis缓存集成（命中率>90%）
- RabbitMQ消息队列部署
- 异步任务队列实现

**交付成果**:
- 完整的数据库Schema
- Redis缓存系统（延迟<10ms）
- 消息队列（延迟<5秒）

### Phase 5: 前端应用 (3周)
**目标**: 打造现代化的用户界面

#### Week 27-29: 前端开发 (第五部分)
- Next.js框架搭建
- 对话界面实现（流式打字）
- 会话管理界面
- 开发者控制台（仪表盘+密钥管理）
- SSE流式客户端
- Zustand状态管理

**交付成果**:
- 响应式用户界面（首屏<2秒）
- 开发者控制台
- 实时流式通信（延迟<100ms）

### Phase 6: 运维与监控 (2周)
**目标**: 建立完善的监控体系

#### Week 30-31: DevOps (第六部分)
- Prometheus指标采集
- Grafana仪表盘（5个）
- 告警规则配置
- Zap结构化日志
- Loki日志聚合
- Dockerfile优化（镜像<500MB）
- Helm Chart编写
- GitHub Actions CI/CD

**交付成果**:
- 监控告警系统
- 日志查询界面（检索<1秒）
- 自动化部署流程（部署<5分钟）

### Phase 7: 安全与性能 (1周)
**目标**: 生产就绪

#### Week 32: 最终优化 (第七部分)
- 输入验证中间件
- 敏感数据加密（AES-256）
- HTTPS/TLS配置
- 数据库查询优化
- Goroutine池管理
- 性能压测（QPS>5000）

**交付成果**:
- 安全加固（无高危漏洞）
- 性能优化（P99<500ms）
- 生产环境就绪

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

## 九、质量保证体系 (Quality Assurance)

### 9.1 测试策略与覆盖
**目标**: 达到 85%+ 整体覆盖率，关键路径 >95%
**贯穿周期**: Week 1-32 (全阶段)

#### 9.1.1 单元测试 (Unit Tests)
**覆盖范围**: 业务逻辑层 (Service/Handler/Util)
**目标覆盖率**: >85%

- **测试框架**: `testify + gomock`
- **运行频率**: 每次代码提交 (Git Pre-commit Hook)
- **执行时间**: 应<5分钟
- **关键模块**:
  ```go
  // Token计数系统
  TestTokenCounter_CountTokens_Accuracy()
  TestTokenCounter_MultiModal_Handling()
  
  // 计费系统
  TestQuotaService_PreConsume_Concurrent()
  TestQuotaService_Refund_Correctness()
  TestBillingService_RateCalculation_Accuracy()
  
  // 渠道选择
  TestChannelSelector_LoadBalance_Distribution()
  TestChannelSelector_Retry_Logic()
  TestChannelSelector_FailoverSwitch()
  
  // RAG系统
  TestRetriever_SemanticSearch_Accuracy()
  TestChunker_TextSplitting_Quality()
  TestEmbedding_Batch_Processing()
  ```

#### 9.1.2 集成测试 (Integration Tests)
**覆盖范围**: 跨服务通信、数据库操作、外部API调用
**目标覆盖率**: >70%

- **测试框架**: `testcontainers-go + docker-compose`
- **运行频率**: 每日 (Nightly Build) + PR提交
- **执行时间**: 应<30分钟
- **关键场景**:
  ```go
  // API中转流程
  TestRelayService_EndToEnd_ChatCompletion()
  TestRelayService_EndToEnd_Streaming()
  TestRelayService_EndToEnd_MultiChannel_Fallback()
  
  // 用户认证
  TestAuth_Login_Register_Logout_Flow()
  TestAuth_TokenRefresh_Expiry()
  TestAuth_Permission_Check()
  
  // 知识库
  TestKnowledge_Upload_Parse_Embed_Flow()
  TestKnowledge_Search_Ranking_Quality()
  
  // 计费
  TestBilling_Request_to_ChargeRecord_Flow()
  TestBilling_Concurrent_Quota_Deduction()
  ```

#### 9.1.3 端到端测试 (E2E Tests)
**覆盖范围**: 完整用户场景 (前端 + 后端)
**目标覆盖率**: >60%（关键路径 >80%）

- **测试框架**: `Playwright / Cypress`
- **运行环境**: Staging 环境
- **运行频率**: 每日两次 (晨/晚) + 发版前
- **执行时间**: 应<20分钟
- **关键用户路径**:
  ```
  [登录] → [创建对话] → [流式聊天] → [查看历史] → [导出对话]
  [注册] → [绑定API密钥] → [调用API] → [查看统计] → [支付]
  [上传文档] → [知识库检索] → [RAG对话] → [参考引用]
  [创建工作流] → [配置Agent] → [执行任务] → [查看日志]
  ```

#### 9.1.4 性能测试 (Performance Testing)
**工具**: Apache JMeter / k6
**运行频率**: 每周 (周五下午)
**基准指标**:

| 指标 | 目标值 | P50 | P95 | P99 |
|------|------|-----|-----|-----|
| 非流式API延迟 | <500ms | 100ms | 300ms | 500ms |
| 流式首字节延迟 | <200ms | 50ms | 100ms | 200ms |
| 并发用户支持 | 10000+ | - | - | - |
| 吞吐量(QPS) | >5000 | - | - | - |
| 错误率 | <0.1% | - | - | - |
| CPU使用率 | <70% | - | - | - |
| 内存使用率 | <80% | - | - | - |

#### 9.1.5 安全测试 (Security Testing)
**工具**: OWASP ZAP / Burp Suite
**运行频率**: 每两周 + 发版前

- **漏洞扫描**: 自动化 SAST (SonarQube)
- **依赖扫描**: Dependabot / Snyk
- **渗透测试**: 月度专项
- **检查清单**:
  - [ ] SQL 注入防护
  - [ ] XSS 防护
  - [ ] CSRF Token 验证
  - [ ] 敏感数据加密
  - [ ] 认证绕过检查
  - [ ] 权限提升检查
  - [ ] 密钥泄露扫描

#### 9.1.6 容错性测试 (Chaos Engineering)
**工具**: Chaos Mesh / Gremlin
**运行频率**: 每月一次

```yaml
测试场景:
  1. 数据库宕机 → 服务自动降级
  2. Redis故障 → 缓存击穿防护
  3. 某渠道不可用 → 自动切换
  4. 网络延迟增加 → 超时处理
  5. 内存泄漏 → 告警触发
```

### 9.2 代码审查流程 (Code Review)
**周期**: 持续集成
**审查标准**: 
- [ ] 所有PR必须有 2+ 人审核
- [ ] 关键模块 (计费/认证/AI适配) 需架构师审核
- [ ] 必须通过自动化检查 (测试/Lint/安全扫描)
- [ ] 需要更新文档和测试用例

**评审清单**:
```markdown
## 功能完整性
- [ ] 需求理解正确
- [ ] 边界情况处理
- [ ] 错误处理完善

## 代码质量
- [ ] 符合编码规范
- [ ] 无代码重复（DRY原则）
- [ ] 复杂度在可接受范围
- [ ] 有必要的注释和文档

## 性能和安全
- [ ] 无N+1查询
- [ ] 缓存策略合理
- [ ] 无安全漏洞
- [ ] 日志包含RequestID

## 测试覆盖
- [ ] 单元测试>80%
- [ ] 集成测试覆盖新功能
- [ ] E2E测试涵盖关键路径
```

### 9.3 质量度量指标
**监控维度**: 整个开发周期

```
关键指标:
├─ 测试覆盖率
│  ├─ 单元测试: >85%
│  ├─ 集成测试: >70%
│  └─ E2E测试: >60%
├─ 代码质量
│  ├─ 技术债指数: <5%
│  ├─ Bug密度: <2/1000LOC
│  ├─ 代码复杂度: <10
│  └─ 重复代码率: <5%
├─ 缺陷追踪
│  ├─ P0缺陷: 0
│  ├─ P1缺陷解决率: >95%
│  ├─ 平均修复时间: <24h
│  └─ 重开缺陷率: <5%
├─ 性能指标
│  ├─ API延迟P99: <500ms
│  ├─ 流式首字节: <200ms
│  ├─ 错误率: <0.1%
│  └─ 可用性: >99.9%
└─ 安全指标
   ├─ 已知漏洞: 0
   ├─ 扫描覆盖: 100%
   ├─ 修复时间: <72h
   └─ 审计事件覆盖: 100%
```

---

## 十、风险管理与应急预案 (Risk Management)

### 10.1 主要风险识别

| 序号 | 风险 | 概率 | 影响 | 应对措施 | 负责人 |
|------|------|------|------|---------|--------|
| R1 | 外部AI提供商API变更 | 中 | 高 | 建立适配器抽象层；监控API更新；提前集成测试 | 技术leader |
| R2 | 渠道密钥泄露 | 低 | 高 | 密钥加密存储；审计日志；定期轮换；告警检测 | 安全 |
| R3 | 计费系统精度问题 | 低 | 高 | 精确的Token计数；预审计；双账户验证；定期对账 | PM |
| R4 | 数据库性能瓶颈 | 中 | 中 | 提前优化索引；分库分表；读写分离；监控告警 | DBA |
| R5 | 并发竞态条件 | 中 | 中 | 充分的单元/集成测试；压力测试；代码审查 | QA |
| R6 | 知识库向量搜索精度低 | 中 | 中 | A/B测试；用户反馈收集；模型迭代；混合检索 | AI |
| R7 | 前端SEO不友好 | 低 | 低 | SSR配置；Meta标签；Sitemap；结构化数据 | 前端 |
| R8 | 部署失败导致宕机 | 低 | 高 | 蓝绿部署；灾难恢复；自动回滚；烟雾测试 | DevOps |
| R9 | 用户隐私数据泄露 | 极低 | 极高 | 加密存储；访问控制；审计日志；定期安全审计 | 安全 |
| R10 | 竞品功能快速迭代 | 中 | 中 | 敏捷开发；快速迭代；用户反馈驱动 | PM |

### 10.2 应急预案

#### 10.2.1 API提供商故障
```
触发条件: 单渠道失败率 >50% 持续 >5分钟
应急步骤:
  1. 自动切换到备用渠道 (<10秒)
  2. 告警通知 (Slack/钉钉)
  3. 降级策略：限流用户或返回缓存响应
  4. RCA分析，更新故障处理文档
  
预防措施:
  - 配置 3+ 个备用渠道
  - 每小时健康检查
  - 自动切换无感知
  - 有完整的故障历史记录
```

#### 10.2.2 数据库故障
```
触发条件: DB连接失败 >100个 or 查询超时率 >5%
应急步骤:
  1. 立即切换到只读副本 (<5秒)
  2. 触发自动故障转移
  3. 告警通知DBA
  4. 非关键写操作进入队列
  5. 恢复后补偿处理
  
预防措施:
  - 主从复制 + 自动故障转移
  - 连接池配置合理
  - 查询超时控制
  - 定期备份验证
  - RTO: 60秒, RPO: 1秒
```

#### 10.2.3 缓存雪崩
```
触发条件: 缓存命中率突降 <10% or Redis宕机
应急步骤:
  1. 启用本地内存缓存 (10分钟数据)
  2. 限流削减 50% 流量
  3. 返回降级响应
  4. 异步恢复缓存
  
预防措施:
  - 布隆过滤器防穿透
  - 缓存随机过期时间
  - Redis集群部署
  - 监控缓存命中率
```

#### 10.2.4 计费系统异常
```
触发条件: 计费记录数量 <预期 50% or 用户投诉
应急步骤:
  1. 立即暂停消费记账
  2. 触发告警，人工介入
  3. 校验计费逻辑
  4. 补偿用户账户
  5. 进行账单对账
  
预防措施:
  - 预扣费保证可靠性
  - 定期账单对账 (日级)
  - 严格的单元测试
  - 计费链路完整日志
```

#### 10.2.5 安全事件 (DDoS/泄露)
```
触发条件: 流量异常增加 >500% or 敏感数据泄露
应急步骤:
  1. 启用WAF规则，限制请求速率
  2. 切换到CDN高防
  3. 立即通知安全团队
  4. 冻结异常账户
  5. 用户通知和补偿
  
预防措施:
  - CloudFlare/Akamai 高防
  - 速率限制中间件
  - 异常行为检测
  - 密钥定期轮换
  - 定期渗透测试
```

### 10.3 故障恢复指标 (RTO/RPO)

| 故障类型 | RTO | RPO | 验证方法 |
|---------|------|-----|---------|
| 单API提供商故障 | <10s | 0 | 自动切换测试 |
| 数据库主从故障 | <60s | 1s | 每月演练 |
| Redis缓存故障 | <30s | 10m | 月度测试 |
| 多AZ故障 | <5m | 1m | 季度演练 |
| 数据中心故障 | <30m | 5m | 年度演练 |

---

## 十一、文档与知识管理 (Documentation)

### 11.1 文档清单与维护

**文档体系** (更新至Week 32):

```
docs/
├─ README.md (项目概览)
├─ QUICK_START.md (5分钟快速开始)
├─ ARCHITECTURE.md (架构总览)
├─ API_REFERENCE.md (完整API文档)
├─ DEPLOYMENT_GUIDE.md (部署指南)
├─ SECURITY_HARDENING.md (安全加固指南)
├─ OPERATIONS.md (运维手册)
├─ TROUBLESHOOTING.md (故障排查)
├─ PERFORMANCE_TUNING.md (性能调优)
├─ CONTRIBUTING.md (贡献指南)
├─ CHANGELOG.md (版本日志)
│
├─ admin/
│  ├─ RBAC_GUIDE.md (权限管理)
│  ├─ CHANNEL_MANAGEMENT.md (渠道管理)
│  ├─ BILLING_ADMIN.md (账单管理)
│  └─ MONITORING_ADMIN.md (监控管理)
│
├─ developer/
│  ├─ API_GUIDE.md (开发指南)
│  ├─ SDK_REFERENCE.md (SDK文档)
│  ├─ ADAPTER_DEVELOPMENT.md (适配器开发)
│  ├─ PLUGIN_DEVELOPMENT.md (插件开发)
│  └─ EXAMPLES.md (代码示例)
│
├─ user/
│  ├─ USER_GUIDE.md (用户手册)
│  ├─ FAQ.md (常见问题)
│  ├─ KNOWLEDGE_BASE.md (知识库使用)
│  └─ PLUGIN_LIBRARY.md (插件库)
│
└─ internal/
   ├─ DESIGN_DECISIONS.md (设计决策)
   ├─ DATABASE_SCHEMA.md (数据库设计)
   ├─ CODE_STYLE.md (代码规范)
   └─ TESTING_STRATEGY.md (测试策略)
```

### 11.2 文档更新频率

| 文档类型 | 更新频率 | 责任人 |
|---------|---------|--------|
| API文档 | 每次API变更时 | 开发者 + API负责人 |
| 部署指南 | 每个发版 | DevOps |
| 故障排查 | 每次生产事件 | 技术负责人 |
| 性能调优 | 每周 | DBA + 后端负责人 |
| 变更日志 | 每个发版 | PM |
| 设计文档 | 设计完成时 | 架构师 |

### 11.3 知识库建设

**内部Wiki** (Confluence/Notion):
- 团队工作流程
- 常见问题解决方案
- 性能调优经验
- 故障案例分析
- 技术分享记录

**外部帮助中心** (Front/Zendesk):
- FAQ (前20个常见问题)
- 教程视频 (5-10个关键功能)
- 社区论坛支持
- 用户反馈收集

---

## 十二、可观测性增强 (Observability)

### 12.1 三支柱增强

#### 12.1.1 指标 (Metrics) 补充
```yaml
应用级指标:
  - 业务指标
    - 日活用户 (DAU)
    - 日API调用数 (DAC)
    - 月度营收 (MRR)
    - 续费率 (Retention)
  
  - 性能指标
    - API端点粒度的P50/P95/P99延迟
    - 按模型的成功率
    - 按用户的配额消耗趋势
    - 按渠道的可用率

基础设施指标:
  - CPU利用率 (按服务/Pod)
  - 内存占用 (按服务/Pod)
  - 网络吞吐 (按服务/方向)
  - 磁盘I/O (按数据库/缓存)
```

#### 12.1.2 日志 (Logs) 增强
```yaml
结构化日志字段:
  - 基础字段
    - timestamp: RFC3339格式
    - level: DEBUG/INFO/WARN/ERROR
    - service: 服务名
    - pod_id: Pod标识
  
  - 追踪字段
    - request_id: 贯穿整个链路
    - user_id: 用户标识
    - trace_id: 分布式追踪ID
    - span_id: 子操作ID
  
  - 业务字段
    - operation: 操作类型 (chat/token_count/billing)
    - resource: 资源标识
    - status: 成功/失败
    - error_code: 错误代码
  
  - 性能字段
    - duration_ms: 操作耗时
    - db_query_count: 数据库查询数
    - cache_hits: 缓存命中数
    - external_calls: 外部服务调用数
```

#### 12.1.3 追踪 (Traces) 实现
```yaml
关键链路追踪:
  1. API请求链路
     请求入口 → 认证 → 速率限制 → 业务逻辑 → 数据库 → 外部API → 响应
  
  2. 计费记账链路
     API调用 → Token计数 → 预扣费 → 消息队列 → 异步记账 → 账单生成
  
  3. RAG检索链路
     用户查询 → 向量化 → 语义搜索 → 结果排序 → 提示词增强 → AI调用
  
  4. 渠道切换链路
     渠道选择 → 健康检查 → 故障检测 → 自动切换 → 重试 → 成功/失败告警

采样策略:
  - 错误请求: 100% 采样
  - 慢请求 (>500ms): 10% 采样
  - 正常请求: 1% 采样
```

### 12.2 告警规则补充

**关键业务告警**:
```yaml
计费异常:
  - 告警名: BillingRecordStaleAlert
    条件: 最近10分钟无新的计费记录
    阈值: 5分钟
    严重级别: P1
    通知: 财务+技术

用户服务异常:
  - 告警名: HighUserAPIErrorRate
    条件: 用户API错误率 > 1%
    阈值: 5分钟
    严重级别: P1
    通知: 用户支持+技术

渠道可用性:
  - 告警名: ChannelHealthCheckFailure
    条件: 渠道连续失败 >3次
    阈值: 15分钟
    严重级别: P2
    通知: 技术

资源告急:
  - 告警名: DatabaseConnectionPoolExhausted
    条件: 连接池使用率 >90%
    阈值: 2分钟
    严重级别: P1
    通知: DBA

安全事件:
  - 告警名: AbnormalLoginAttempts
    条件: 单用户10分钟内失败 >5次
    阈值: 1分钟
    严重级别: P1
    通知: 安全+用户
```

---

## 十三、性能基准与SLA (Performance Baselines & SLA)

### 13.1 服务级别协议 (SLA)

```
=== 用户服务 (User-Facing) ===
可用性目标:     99.95% (月度可接受宕机时间 ≤ 21.6分钟)
平均响应时间:   <500ms (P99)
错误率:         <0.1%
首字节延迟:     <200ms (流式API)

=== 开发者API服务 ===
可用性目标:     99.9% (月度可接受宕机时间 ≤ 43.2分钟)
平均响应时间:   <1s (P99, 包括后端处理)
错误率:         <0.5%
速率限制准确性: ≥99.9%

=== 后端关键服务 ===
计费系统:
  可用性:       99.99% (年度 ≤ 52.6分钟宕机)
  精确度:       100% (零容差)
  处理延迟:     <5秒

渠道管理:
  选择耗时:     <5ms (P99)
  故障切换:     <10秒
  缓存命中率:   >95%

知识库检索:
  搜索延迟:     <200ms (P99)
  准确率:       >85% (Top-10)
  容量:         百万级文档
```

### 13.2 性能基准 (Benchmarks)

```
后端性能基准 (通过压测验证):

吞吐量指标:
  - Chat Completions API: 5000+ QPS
  - Embeddings API: 10000+ QPS
  - Token Count API: 50000+ QPS
  - Admin API: 1000+ QPS

延迟指标 (单位: ms):
  指标                 P50   P95   P99   Max
  ────────────────────────────────────────
  认证检查             <1    <2    <5    <10
  Token计数            <5    <20   <50   <100
  预扣费               <10   <30   <100  <200
  渠道选择             <2    <5    <10   <20
  知识库搜索           <50   <150  <200  <500
  AI响应(非流式)       <500  <1000 <2000 <5000

资源利用率基准:
  服务                 CPU   内存   磁盘
  ───────────────────────────────────
  Gateway              <50%  <1GB   -
  Chat Service         <40%  <2GB   <10GB
  Billing Service      <20%  <500MB <50GB
  Knowledge Service    <30%  <3GB   <100GB
  
前端性能基准:
  指标                        基准值   目标值
  ──────────────────────────────────────
  首屏加载时间 (First Paint)  <2s     <1s
  可交互时间 (TTI)            <3s     <2s
  首输入延迟 (FID)            <100ms  <50ms
  累积布局偏移 (CLS)          <0.1    <0.05
  最大内容绘制 (LCP)          <2.5s   <1.5s
```

### 13.3 SLA监控与报告

```
月度SLA报告:
  - 整体可用性百分比
  - 按服务的可用性分解
  - 故障事件统计 (MTBF/MTTR)
  - 性能指标分布 (P50/P95/P99)
  - 问题根因分析
  - 改进行动项

季度技术审查:
  - 性能趋势分析
  - 容量规划评估
  - 架构改进评议
  - 技术债状态
```

---

## 十四、生产发布清单 (Production Readiness)

### 14.1 发版前检查清单

**功能完整性** (Week 32前):
- [ ] 所有核心功能已实现
- [ ] 所有计划的API端点已上线
- [ ] UI/UX已完成并通过可用性测试
- [ ] 文档已更新
- [ ] 变更日志已记录

**质量保证** (Week 31-32):
- [ ] 单元测试覆盖率 ≥85%
- [ ] 集成测试覆盖率 ≥70%
- [ ] E2E测试关键路径 ≥80%
- [ ] 代码审查通过率 100%
- [ ] 安全扫描零高危漏洞
- [ ] 性能压测通过 (P99 <500ms)

**基础设施准备** (Week 31):
- [ ] 容器镜像构建完成
- [ ] Kubernetes manifest准备就绪
- [ ] 数据库迁移脚本验证通过
- [ ] 备份/恢复演练通过
- [ ] 灾难恢复流程验证
- [ ] 监控告警规则已配置

**运维准备** (Week 30-31):
- [ ] 部署手册编写完成
- [ ] 故障排查指南完成
- [ ] 团队培训完成
- [ ] 值班表安排妥当
- [ ] 支持流程文档化
- [ ] 沟通渠道已建立

**用户传播** (Week 29-31):
- [ ] 发版公告已准备
- [ ] 教程视频已录制
- [ ] 用户指南已发布
- [ ] 社区告知已提前通知
- [ ] 反馈渠道已建立

### 14.2 灰度发版计划 (Canary Deployment)

```
Phase 1: 金丝雀阶段 (Week 32, Day 1-2)
  - 版本: 1.0-canary
  - 流量: 5% (内部用户 + 付费用户)
  - 监控: 每5分钟一次性能和错误率检查
  - 决策: 如果P99延迟<500ms && 错误率<0.1% → 进入Phase 2

Phase 2: 小范围发版 (Week 32, Day 2-3)
  - 版本: 1.0-rc1
  - 流量: 25% (地区分布)
  - 监控: 每10分钟一次检查
  - 决策: 如果持续稳定24小时 → 进入Phase 3

Phase 3: 扩大发版 (Week 32, Day 3-4)
  - 版本: 1.0
  - 流量: 50% (按地区均衡)
  - 监控: 每30分钟一次检查
  - 决策: 如果无P1事件 → 进入Phase 4

Phase 4: 全量发版 (Week 32, Day 4-5)
  - 版本: 1.0
  - 流量: 100%
  - 监控: 每小时一次检查
  - 决策: 等待一周，确保稳定性

回滚条件 (任意阶段):
  - P99延迟 >1000ms
  - 错误率 >0.5%
  - 计费系统异常
  - 数据一致性问题
  - 安全漏洞发现
```

### 14.3 发版后操作 (Post-Release)

```
Day 1: 发版日
  - 08:00 灰度发版 (Phase 1)
  - 09:00 首次检查
  - 持续监控 (30分钟间隔)
  - 19:00 更新到Phase 2

Day 2-3: 稳定期
  - 每小时检查一次
  - 收集用户反馈
  - 记录关键指标

Day 4-5: 全量期
  - 每天两次检查
  - 完整的SLA监控
  - 性能基准验证

Week 2: 回顾期
  - 问题总结
  - 改进行动项
  - 经验萃取
```

---

## 总结 (Enhanced)

本开发计划经过完善，现已成为一份**企业级产品开发指南**，包含：

### ✅ 核心交付物
1. **功能规划**: 7大模块 50+ 功能点
2. **质量体系**: 测试策略 + 代码审查 + 质量度量
3. **风险管理**: 10大风险识别 + 应急预案
4. **文档体系**: 30+ 份文档，分层级管理
5. **可观测性**: 指标 + 日志 + 追踪，三支柱完整
6. **性能基准**: 详细的SLA和性能目标
7. **发版流程**: 灰度发版 + 回滚机制

### 🎯 关键成功因素
| 因素 | 行动项 | 负责人 |
|------|--------|--------|
| 质量 | 每周QA同步，缺陷率<2/1000 LOC | QA Lead |
| 性能 | 周度性能测试，P99<500ms | 后端Lead |
| 安全 | 双周扫描，零高危漏洞 | 安全 |
| 交付 | 双周发版评审 | PM |
| 学习 | 月度复盘，持续改进 | 全体 |

### 📊 预期成果
- **代码质量**: 85%+ 测试覆盖，0 已知高危漏洞
- **用户体验**: API P99<500ms，流式首字节<200ms
- **业务指标**: 日活>10000，API调用>100M/月
- **团队效能**: 发版周期<2周，MTTR<1小时

**建议开发顺序**: 核心功能 → 用户服务 → 开发者服务 → 质量保证 → 运维优化 → 安全性能

**预计完成时间**: 32周（8个月）| **关键路径**: 核心功能 + 计费系统（10周不能延长）
