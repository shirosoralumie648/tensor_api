# 后端代码审计规划 (Backend Audit Plan)

本规划旨在对 `backend` 目录下的所有代码进行全面审计，识别潜在Bug、安全漏洞、冗余代码及架构改进点。

## 1. 审计目标
- **代码质量**：检查代码规范、错误处理、资源释放等。
- **安全性**：检查SQL注入、XSS、权限控制、敏感信息泄漏等。
- **功能完整性**：验证逻辑是否符合业务需求，是否存在死代码。
- **性能**：识别潜在的性能瓶颈（如N+1查询、锁竞争等）。

## 2. 审计阶段规划

### 第一阶段：核心基础设施 (Infrastructure)
**目标**：确保基础组件的稳定性与安全性。
**涉及文件**：
- [ ] **数据库与Schema**
    - `backend/internal/database/redis.go`
    - `backend/internal/database/schema.go`
- [ ] **中间件 (Middleware)**
    - `backend/internal/middleware/auth.go`
    - `backend/internal/middleware/auth_cached.go`
    - `backend/internal/middleware/auth_factory.go`
    - `backend/internal/middleware/auth_handler.go`
    - `backend/internal/middleware/cors.go`
    - `backend/internal/middleware/logger.go`
    - `backend/internal/middleware/rate_limit.go`
    - `backend/internal/middleware/rbac.go`
    - `backend/internal/middleware/request_id.go`
    - `backend/internal/middleware/security.go`
- [ ] **安全与工具 (Security & Utils)**
    - `backend/internal/security/encryption.go`
    - `backend/internal/security/input_validation.go`
    - `backend/internal/utils/crypto.go`
    - `backend/internal/utils/jwt.go`
    - `backend/internal/utils/response.go`
- [ ] **日志 (Logging)**
    - `backend/internal/logging/logger.go`
    - `backend/pkg/logger/logger.go`

### 第二阶段：数据层 (Data Layer)
**目标**：检查数据模型一致性及数据库操作的正确性。
**涉及文件**：
- [ ] **数据模型 (Models)**
    - `backend/internal/model/` 下所有文件
- [ ] **仓储层 (Repositories)**
    - `backend/internal/repository/` 下所有文件
- [ ] **数据库迁移 (Migrations)**
    - `backend/migrations/` 下所有 `.sql` 文件 (检查索引、约束)

### 第三阶段：业务服务层 (Service Layer)
**目标**：审计核心业务逻辑，关注事务一致性与业务规则。
**涉及文件**：
- [ ] **通用服务**
    - `backend/internal/service/user_service.go`
    - `backend/internal/service/agent_service.go`
    - `backend/internal/service/channel_service.go`
    - `backend/internal/service/channel_ability_service.go`
- [ ] **计费与配额 (Billing & Quota)**
    - `backend/internal/service/billing_service.go`
    - `backend/internal/service/billing_advanced.go`
    - `backend/internal/service/pricing_service.go`
    - `backend/internal/service/token_service.go`
    - `backend/internal/quota/` 下所有文件
    - `backend/internal/ratelimit/` 下所有文件
- [ ] **队列与异步 (Queue)**
    - `backend/internal/queue/` 下所有文件

### 第四阶段：复杂子系统 (Subsystems)
**目标**：深度审计高复杂度、高风险模块。
**涉及文件**：
- [ ] **Relay 系统 (核心转发逻辑)**
    - `backend/internal/relay/` 目录下的所有文件 (分为：Handler, Cache, LoadBalancer, Monitor等子模块审计)
- [ ] **RAG 系统 (检索增强生成)**
    - `backend/internal/rag/` 下所有文件
    - `backend/internal/service/rag_service.go`
- [ ] **工具与插件 (Tools & Plugins)**
    - `backend/internal/tools/` 下所有文件
- [ ] **分词器 (Tokenizer)**
    - `backend/internal/tokenizer/` 下所有文件
- [ ] **选择器 (Selector)**
    - `backend/internal/selector/` 下所有文件
- [ ] **监控 (Monitoring)**
    - `backend/internal/monitoring/` 下所有文件

### 第五阶段：接口层 (Interface Layer)
**目标**：检查API入口参数校验、权限控制及错误返回。
**涉及文件**：
- [ ] **HTTP Handlers**
    - `backend/internal/handler/` 下所有文件
- [ ] **入口文件 (Cmd)**
    - `backend/cmd/` 下所有 `main.go` 及相关启动代码

### 第六阶段：全链路与综合检查 (End-to-End & Final Review)
**目标**：串联流程，查漏补缺。
1. **全链路审计**：
    - **请求处理链路**：Middleware -> Handler -> Service -> Repo -> DB
    - **Relay链路**：接收请求 -> 鉴权 -> 路由选择 -> 转发 -> 响应处理 -> 计费扣除
2. **遗漏检查**：
    - 扫描是否有未列入上述计划的文件。
    - 检查是否有定义的函数从未被调用（Dead Code）。
    - 检查TODO/FIXME标记。

## 3. 审计结果记录
所有审计发现将实时记录在 `backend/BACKEND_AUDIT_REPORT.md` 文件中，格式如下：
- **文件路径**: `path/to/file.go`
- **问题级别**: [高/中/低]
- **问题描述**: ...
- **改进建议**: ...
