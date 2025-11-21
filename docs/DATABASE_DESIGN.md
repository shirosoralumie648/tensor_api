# 数据库设计文档

## 概述

Oblivious 使用 PostgreSQL 作为主数据库，利用其强大的扩展能力（如 pgvector）和事务支持。数据库设计遵循规范化原则，同时考虑性能优化和可扩展性。

## 技术选型

- **数据库**：PostgreSQL 15+
- **ORM**：GORM (Go)
- **迁移工具**：golang-migrate
- **扩展**：
  - `uuid-ossp` - UUID 生成
  - `pgvector` - 向量相似度搜索

## 数据库架构

### ER 图概览

```
users (用户表)
  ├─ 1:N ─> sessions (会话表)
  │          └─ 1:N ─> messages (消息表)
  ├─ 1:N ─> billing_logs (计费日志)
  ├─ 1:N ─> quota_logs (额度变更日志)
  ├─ 1:N ─> knowledge_bases (知识库)
  │          └─ 1:N ─> documents (文档)
  │                     └─ 1:N ─> chunks (文本块)
  └─ 1:1 ─> user_settings (用户设置)

channels (渠道表)
  └─ 模型配置

agents (助手表)
  └─ 助手模板

tokens (API令牌表)
  └─ API密钥管理
```

## 核心表设计

### 1. users - 用户表

存储用户基本信息和额度信息。

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    role INT DEFAULT 1,                    -- 用户角色：1=普通用户，10=管理员
    quota BIGINT DEFAULT 0,                -- 当前可用额度（单位：分）
    total_quota BIGINT DEFAULT 0,          -- 累计获得的总额度
    used_quota BIGINT DEFAULT 0,           -- 已使用的额度
    invite_code VARCHAR(20) UNIQUE,        -- 邀请码
    invited_by INT REFERENCES users(id),   -- 邀请人
    status INT DEFAULT 1,                  -- 状态：1=正常，2=禁用
    last_login_at TIMESTAMP,               -- 最后登录时间
    last_login_ip VARCHAR(50),             -- 最后登录IP
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL              -- 软删除
);

-- 索引
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_invite_code ON users(invite_code);
CREATE INDEX idx_users_status ON users(status, created_at DESC) WHERE deleted_at IS NULL;
```

**字段说明**：

- `quota`：用户当前可用额度，单位为分（1元=100分）
- `total_quota`：累计获得的总额度，用于统计
- `used_quota`：已使用的额度，用于统计
- `role`：角色标识，支持未来的 RBAC
- `deleted_at`：软删除标记，便于数据恢复

### 2. user_settings - 用户设置表

存储用户个性化设置。

```sql
CREATE TABLE user_settings (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) DEFAULT 'zh-CN',  -- 界面语言
    theme VARCHAR(20) DEFAULT 'auto',      -- 主题：auto/light/dark
    default_model VARCHAR(100),            -- 默认模型
    chat_settings JSONB,                   -- 对话设置（JSON格式）
    notification_settings JSONB,           -- 通知设置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_settings_user ON user_settings(user_id);
```

### 3. sessions - 会话表

存储对话会话信息。

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_id INT,                          -- 关联的助手ID
    group_id UUID,                         -- 会话分组
    title VARCHAR(200),                    -- 会话标题
    description TEXT,                      -- 会话描述
    pinned BOOLEAN DEFAULT FALSE,          -- 是否置顶
    archived BOOLEAN DEFAULT FALSE,        -- 是否归档
    model VARCHAR(100),                    -- 使用的模型
    temperature FLOAT DEFAULT 0.7,         -- 温度参数
    top_p FLOAT DEFAULT 1.0,               -- Top-P参数
    max_tokens INT,                        -- 最大token数
    system_role TEXT,                      -- 系统角色提示词
    context_length INT DEFAULT 4,          -- 上下文长度
    plugin_ids INT[],                      -- 启用的插件ID数组
    knowledge_base_ids INT[],              -- 关联的知识库ID数组
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 索引
CREATE INDEX idx_sessions_user ON sessions(user_id, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_pinned ON sessions(user_id, pinned DESC, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_archived ON sessions(user_id, archived, updated_at DESC) WHERE deleted_at IS NULL;
```

**设计说明**：

- 使用 UUID 作为主键，便于分布式部署
- `pinned` 和 `archived` 支持用户管理会话
- `plugin_ids` 和 `knowledge_base_ids` 使用数组类型，方便查询

### 4. messages - 消息表

存储对话消息内容和元数据。

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    topic_id UUID,                         -- 话题ID（支持分支对话）
    parent_id UUID,                        -- 父消息ID（树形结构）
    role VARCHAR(20) NOT NULL,             -- 角色：user/assistant/system
    content TEXT NOT NULL,                 -- 消息内容
    model VARCHAR(100),                    -- 使用的模型
    input_tokens INT DEFAULT 0,            -- 输入token数
    output_tokens INT DEFAULT 0,           -- 输出token数
    total_tokens INT DEFAULT 0,            -- 总token数
    cost BIGINT DEFAULT 0,                 -- 消耗的额度（分）
    metadata JSONB,                        -- 元数据（JSON格式）
    files JSONB,                           -- 附件列表
    tool_calls JSONB,                      -- 工具调用记录
    status INT DEFAULT 1,                  -- 状态：1=成功，2=失败，3=处理中
    error_message TEXT,                    -- 错误信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_messages_session ON messages(session_id, created_at DESC);
CREATE INDEX idx_messages_topic ON messages(topic_id, created_at DESC);
CREATE INDEX idx_messages_parent ON messages(parent_id);
CREATE INDEX idx_messages_role ON messages(role, created_at DESC);
CREATE INDEX idx_messages_status ON messages(status, created_at DESC);
```

**特点**：

- 支持树形对话结构（`parent_id`）
- 支持话题分组（`topic_id`）
- 记录详细的 token 使用情况
- `metadata` 和 `files` 使用 JSONB，灵活存储附加信息

### 5. billing_logs - 计费日志表

记录每次 AI 调用的计费详情。

```sql
CREATE TABLE billing_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID,
    message_id UUID,
    model VARCHAR(100) NOT NULL,
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    cost BIGINT DEFAULT 0,                 -- 消耗额度（分）
    cost_usd FLOAT8 DEFAULT 0,             -- 美元成本
    status INT DEFAULT 1,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 索引
CREATE INDEX idx_billing_logs_user ON billing_logs(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_logs_session ON billing_logs(session_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_logs_model ON billing_logs(model) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_logs_created ON billing_logs(created_at) WHERE deleted_at IS NULL;
```

### 6. quota_logs - 额度变更日志表

记录用户额度的所有变更操作。

```sql
CREATE TABLE quota_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation_type VARCHAR(50) NOT NULL,   -- 操作类型：recharge/consume/refund/gift
    amount BIGINT NOT NULL,                -- 变更金额（正数为增加，负数为减少）
    reason TEXT,                           -- 变更原因
    billing_log_id INT REFERENCES billing_logs(id),
    balance_before BIGINT NOT NULL,        -- 变更前余额
    balance_after BIGINT NOT NULL,         -- 变更后余额
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 索引
CREATE INDEX idx_quota_logs_user ON quota_logs(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_quota_logs_operation ON quota_logs(operation_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_quota_logs_created ON quota_logs(created_at) WHERE deleted_at IS NULL;
```

**用途**：

- 财务审计
- 用户消费明细查询
- 对账和数据一致性校验

### 7. channels - 渠道表

管理上游 AI 服务提供商的渠道配置。

```sql
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type INT NOT NULL,                     -- 类型：1=OpenAI, 2=Claude, 3=Gemini等
    base_url TEXT NOT NULL,
    api_key TEXT NOT NULL,
    models TEXT[],                         -- 支持的模型列表
    priority INT DEFAULT 0,                -- 优先级（数字越大优先级越高）
    weight INT DEFAULT 100,                -- 权重（负载均衡用）
    status INT DEFAULT 1,                  -- 状态：1=启用，2=禁用
    test_time TIMESTAMP,                   -- 最后测试时间
    response_time INT,                     -- 平均响应时间（毫秒）
    config JSONB,                          -- 额外配置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_channels_type ON channels(type, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_channels_status ON channels(status, priority DESC) WHERE deleted_at IS NULL;
```

### 8. knowledge_bases - 知识库表

RAG 功能的知识库管理。

```sql
CREATE TABLE knowledge_bases (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    embedding_model VARCHAR(100) DEFAULT 'text-embedding-3-small',
    chunk_size INT DEFAULT 512,            -- 文本块大小
    chunk_overlap INT DEFAULT 50,          -- 块重叠大小
    document_count INT DEFAULT 0,          -- 文档数量
    total_chunks INT DEFAULT 0,            -- 总文本块数
    status INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_kb_user ON knowledge_bases(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_kb_created ON knowledge_bases(created_at DESC);
```

### 9. documents - 文档表

知识库中的文档信息。

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kb_id INT NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    file_url TEXT,
    file_type VARCHAR(50),
    file_size BIGINT,
    status INT DEFAULT 1,                  -- 1=待处理，2=处理中，3=完成，4=失败
    chunk_count INT DEFAULT 0,
    error_message TEXT,
    processing_started_at TIMESTAMP,
    processing_completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_documents_kb ON documents(kb_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_status ON documents(status, created_at DESC);
```

### 10. chunks - 文本块表（向量存储）

**最重要的 RAG 表**，使用 pgvector 扩展存储文本向量。

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,                 -- 文本内容
    embedding vector(1536),                -- 向量嵌入（OpenAI: 1536维）
    metadata JSONB,                        -- 元数据（页码、章节等）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- HNSW 索引用于快速向量相似度搜索
CREATE INDEX idx_chunks_embedding ON chunks USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 200);

CREATE INDEX idx_chunks_document ON chunks(document_id);
```

**HNSW 索引参数说明**：

- `m = 16`：连接数，影响搜索质量和索引大小
- `ef_construction = 200`：构建时的搜索深度

**向量相似度查询示例**：

```sql
-- 查找最相似的 5 个文本块
SELECT id, content, 1 - (embedding <=> $1::vector) AS similarity
FROM chunks
WHERE document_id IN (SELECT id FROM documents WHERE kb_id = $2)
ORDER BY embedding <=> $1::vector
LIMIT 5;
```

### 11. agents - 助手表

AI 助手模板管理。

```sql
CREATE TABLE agents (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    avatar_url TEXT,
    author_id INT REFERENCES users(id),
    category VARCHAR(50),                  -- 分类：编程/写作/翻译等
    tags TEXT[],                           -- 标签
    system_prompt TEXT NOT NULL,           -- 系统提示词
    welcome_message TEXT,                  -- 欢迎语
    suggested_messages TEXT[],             -- 建议问题
    default_model VARCHAR(100),
    temperature FLOAT DEFAULT 0.7,
    is_public BOOLEAN DEFAULT FALSE,       -- 是否公开到市场
    is_official BOOLEAN DEFAULT FALSE,     -- 是否官方助手
    usage_count INT DEFAULT 0,             -- 使用次数
    like_count INT DEFAULT 0,              -- 点赞数
    status INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_agents_public ON agents(is_public, like_count DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_agents_category ON agents(category, usage_count DESC);
```

## 数据迁移

### 迁移文件命名规范

```
000001_create_users_table.up.sql          # 创建用户表
000001_create_users_table.down.sql        # 回滚用户表
000002_create_user_settings_table.up.sql
000002_create_user_settings_table.down.sql
...
```

### 迁移命令

```bash
# 应用所有未执行的迁移
migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/oblivious?sslmode=disable" up

# 回滚最后一次迁移
migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/oblivious?sslmode=disable" down 1

# 强制指定版本
migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/oblivious?sslmode=disable" force 10
```

## 性能优化

### 1. 索引策略

**主键索引**：
- 所有表都有主键索引（自动创建）

**外键索引**：
- 外键字段自动创建索引

**复合索引**：
```sql
-- 用户会话列表查询
CREATE INDEX idx_sessions_user_updated ON sessions(user_id, updated_at DESC) WHERE deleted_at IS NULL;

-- 消息查询
CREATE INDEX idx_messages_session_created ON messages(session_id, created_at DESC);
```

**部分索引**（Partial Index）：
```sql
-- 只索引未删除的记录
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
```

### 2. 查询优化

**使用 EXPLAIN ANALYZE**：
```sql
EXPLAIN ANALYZE
SELECT * FROM sessions
WHERE user_id = 123 AND deleted_at IS NULL
ORDER BY updated_at DESC
LIMIT 20;
```

**避免 N+1 查询**：
```go
// 不好的做法
sessions := []Session{}
db.Find(&sessions)
for _, s := range sessions {
    db.Model(&s).Related(&s.Messages)  // N+1 查询
}

// 好的做法
sessions := []Session{}
db.Preload("Messages").Find(&sessions)
```

### 3. 连接池配置

```go
// backend/internal/database/connection.go
db.DB().SetMaxOpenConns(100)
db.DB().SetMaxIdleConns(10)
db.DB().SetConnMaxLifetime(time.Hour)
```

### 4. 分区表（未来优化）

对于大表（如 `billing_logs`），可以按时间分区：

```sql
CREATE TABLE billing_logs (
    id SERIAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ...
) PARTITION BY RANGE (created_at);

CREATE TABLE billing_logs_2024_01 PARTITION OF billing_logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE billing_logs_2024_02 PARTITION OF billing_logs
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
```

## 数据一致性

### 1. 事务处理

```go
// 扣费操作必须在事务中
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 更新用户额度
if err := tx.Model(&user).Update("quota", user.Quota - cost).Error; err != nil {
    tx.Rollback()
    return err
}

// 记录扣费日志
billingLog := &BillingLog{...}
if err := tx.Create(billingLog).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

### 2. 外键约束

- `ON DELETE CASCADE`：删除用户时自动删除关联数据
- `ON DELETE SET NULL`：保留历史记录

### 3. 触发器

```sql
-- 自动更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

## 备份策略

### 1. 定期备份

```bash
# 每日全量备份
pg_dump -U postgres -d oblivious -F c -f backup_$(date +%Y%m%d).dump

# 压缩备份
pg_dump -U postgres -d oblivious | gzip > backup_$(date +%Y%m%d).sql.gz
```

### 2. 增量备份

使用 WAL 归档：

```sql
-- postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'cp %p /path/to/archive/%f'
```

### 3. 恢复

```bash
# 从备份恢复
pg_restore -U postgres -d oblivious_new backup_20240101.dump

# 从 SQL 文件恢复
gunzip < backup_20240101.sql.gz | psql -U postgres -d oblivious_new
```

## 监控指标

### 关键指标

- **连接数**：`SELECT count(*) FROM pg_stat_activity;`
- **慢查询**：启用 `log_min_duration_statement = 1000` (ms)
- **表大小**：`SELECT pg_size_pretty(pg_total_relation_size('table_name'));`
- **索引使用率**：查询 `pg_stat_user_indexes`
- **缓存命中率**：查询 `pg_stat_database`

## 相关文档

- [架构设计](ARCHITECTURE.md)
- [API 参考](API_REFERENCE.md)
- [快速开始](QUICK_START.md)
