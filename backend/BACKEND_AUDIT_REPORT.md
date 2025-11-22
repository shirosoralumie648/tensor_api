# 后端代码审计报告 (Backend Audit Report)

## 1. 核心结论：系统处于不可用状态 (Critical Assessment)

**总体评价**: 本项目目前属于**高度未完成的骨架工程**，存在大量的"僵尸代码"和"伪实现"。虽然项目目录结构看起来非常专业且功能丰富（包含监控、分析、RAG、MCP插件等），但**绝大多数子系统未被主流程引用**。核心的对话链路是断裂的，前端与后端无法通信，安全机制形同虚设。

**严正警告**: 严禁将当前代码部署到任何环境。

---

## 2. 致命架构断层 (Fatal Architectural Disconnects)

### 2.1 核心链路是假的 (Fake Core Logic)
- **流式传输欺诈**: `backend/internal/relay/handler_impl.go` 中的 `HandleStream` 方法直接返回硬编码的 `{"content":"Hello from stream"}`，未对上游进行流式转发。
- **适配器空壳**: `backend/internal/adapter/providers.go` 中实现了 OpenAI/Claude/Gemini 等多个适配器，但：
    -   `HealthCheck` 全部直接返回 `nil`（未连接上游）。
    -   `ParseStreamResponse` 对于非 OpenAI 模型是空的或未实现。
    -   错误处理逻辑缺失。
- **网关阻塞**: `backend/cmd/gateway/main.go` 使用 `io.ReadAll` 读取全量请求和响应体。这不仅会导致内存溢出，更会导致**流式响应变成阻塞式响应**，彻底破坏用户体验。
- **SSE 协议错误**: 网关手动设置 `Transfer-Encoding: chunked` 且在接收完整响应后才开始发送，违反了 SSE 的实时性原则且可能导致协议解析错误。

### 2.2 前后端完全失联 (Frontend-Backend Disconnect)
- **API 路径不匹配**:
    -   前端 (`frontend/src/services/api.ts`) 假设 API 路径为 `/api/chat/completions`。
    -   后端 Gateway (`backend/cmd/gateway/main.go`) 注册的路径是 `/api/v1/...`。
    -   **结果**: 前端所有请求都会 404。
- **接口缺失**: 前端调用的 `/api/developer/*`, `/api/knowledge/*`, `/api/auth/*` 等接口在后端网关中**根本未注册**。
- **浏览器兼容性错误**: 前端 `useChat.ts` 使用 `axios` 的 `responseType: 'stream'` 并试图调用 `getReader()`。这是 Fetch API 的用法，Axios 在浏览器端不支持流式读取，导致前端聊天功能完全不可用。

### 2.3 僵尸子系统 (Zombie Subsystems)
以下模块在代码库中存在，但**未被任何启动入口或主业务流程引用**（Dead Code）：
-   `internal/analytics` (分析系统)
-   `internal/monitoring` (监控系统)
-   `internal/performance` (性能分析)
-   `internal/tools` (MCP 插件系统)
-   `internal/token` (Token 管理)
-   `internal/quota` (配额管理)
-   `internal/ratelimit` (高级限流)
-   `internal/queue` (队列系统)

---

## 3. 安全与合规灾难 (Security & Compliance Disasters)

### 3.1 伪安全实现
- **加密欺诈**: `internal/security/encryption.go` 中的 `HashPassword` 和 `NewEncryptionManager` 即使声称使用了 SHA-256，实际上可能只是简单的字符串操作或明文拷贝，**严重误导**。
- **无效校验**: `input_validation.go` 中的 `isValidJSON` 函数实现为 `return nil == nil` (恒为 true)，完全没有校验作用。
- **伪随机**: `KeyManager` 使用原子自增取模来模拟“随机”选择，导致密钥轮询完全可预测。

### 3.2 访问控制失效
- **CORS 全开**: 中间件默认允许所有 Origin 且允许凭证 (`AllowCredentials: true`)，这是严重的安全漏洞。
- **Body 破坏**: 签名验证中间件读取 Request Body 后未重置 `Request.Body`，导致后续 Handler 读取到 EOF 报错。
- **RBAC 硬编码**: 权限检查逻辑未连接数据库，而是返回硬编码的空权限或测试数据。

---

## 4. 业务逻辑缺陷 (Business Logic Flaws)

### 4.1 计费系统
- **事务缺失**: `BillingService` 的扣费流程（记录日志 -> 扣减余额）没有事务包裹，失败会导致数据不一致。
- **接口伪造**: `BillingHandler` 中的大量接口（如获取发票）直接返回硬编码的 JSON 数据。

### 4.2 渠道管理
- **重试失效**: `ChannelSelector.SelectWithRetry` 在重试时未将失败渠道加入排除列表，导致一直重试同一个坏渠道。
- **状态易失**: 渠道的健康状态和失败统计仅存在于内存中，重启即丢失。

---

## 5. 修复路线图 (Remediation Roadmap)

### 第一阶段：打通脉络 (Phase 1: Connectivity)
1.  **修复网关**: 移除 `io.ReadAll`，改用 `httputil.ReverseProxy` 实现真正的流式透传。
2.  **对齐接口**: 修正前后端 API 路径（统一为 `/api/v1`），并在网关注册所有缺失的路由。
3.  **前端重构**: 弃用 Axios 处理流，改用 `fetch` API 或专门的 SSE 库 (`@microsoft/fetch-event-source`)。

### 第二阶段：填充血肉 (Phase 2: Implementation)
1.  **实现适配器**: 补全 `providers.go` 中各厂商的真实调用和流式解析逻辑。
2.  **激活子系统**: 将 `internal/token`, `internal/quota` 等僵尸模块接入 `RelayService` 和 `Gateway`。
3.  **修复核心逻辑**: 修正 `KeyManager` 随机算法和 `ChannelSelector` 重试逻辑。

### 第三阶段：加固安全 (Phase 3: Security)
1.  **替换伪加密**: 使用 `golang.org/x/crypto/bcrypt` 和 `crypto/rand` 替换所有伪实现。
2.  **完善校验**: 实现真正的 JSON 校验和请求体复原逻辑。
3.  **数据库事务**: 为计费和关键状态变更添加事务支持。

### 总结
这就是一个"只有门面没有装修"的项目。建议先暂停所有新功能开发，全力集中于**让主流程真正跑通**。
