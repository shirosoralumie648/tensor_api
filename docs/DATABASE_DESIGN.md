# 数据库设计文档

## 概述

Oblivious 使用 **PostgreSQL 15** 作为主数据库，采用主从复制实现读写分离，当数据量增长到一定规模后可按用户 ID 分片。

## 数据库命名规范

- 表名：小写下划线分隔，复数形式（如 `users`, `sessions`）
- 字段名：小写下划线分隔（如 `created_at`, `user_id`）
- 主键：统一命名为 `id`
- 外键：`<关联表单数>_id`（如 `user_id`, `session_id`）
- 时间戳字段：`created_at`, `updated_at`, `deleted_at`（软删除）

---

## 核心表结构

### 1. 用户表 (users)

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    role INT DEFAULT 1,  -- 1: 普通用户, 10: VIP, 100: 管理员
    quota BIGINT DEFAULT 0,  -- 剩余额度（单位：分，1元=100分）
    total_quota BIGINT DEFAULT 0,  -- 累计充值额度
    used_quota BIGINT DEFAULT 0,  -- 已使用额度
    invite_code VARCHAR(20) UNIQUE,  -- 邀请码
    invited_by INT REFERENCES users(id),  -- 邀请人
    status INT DEFAULT 1,  -- 1: 正常, 2: 禁用, 3: 注销
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_invite_code ON users(invite_code);
CREATE INDEX idx_users_invited_by ON users(invited_by);
```

### 2. 用户设置表 (user_settings)

```sql
CREATE TABLE user_settings (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) DEFAULT 'zh-CN',
    theme VARCHAR(20) DEFAULT 'auto',  -- auto, light, dark
    font_size INT DEFAULT 14,
    tts_enabled BOOLEAN DEFAULT FALSE,
    tts_voice VARCHAR(50),
    tts_speed FLOAT DEFAULT 1.0,
    stt_enabled BOOLEAN DEFAULT FALSE,
    send_key VARCHAR(20) DEFAULT 'Enter',  -- Enter, Ctrl+Enter
    avatar_style VARCHAR(20) DEFAULT 'circle',
    message_display VARCHAR(20) DEFAULT 'bubble',
    custom_config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3. 额度日志表 (quota_logs)

```sql
CREATE TABLE quota_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    type INT NOT NULL,  -- 1: 充值, 2: 消费, 3: 退款, 4: 赠送, 5: 邀请奖励
    amount BIGINT NOT NULL,  -- 变动金额（正数为增加，负数为减少）
    balance BIGINT NOT NULL,  -- 变更后余额
    description TEXT,
    related_order_id VARCHAR(100),  -- 关联订单号
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_quota_logs_user ON quota_logs(user_id, created_at DESC);
CREATE INDEX idx_quota_logs_type ON quota_logs(type, created_at DESC);
```

### 4. 会话表 (sessions)

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT NOT NULL REFERENCES users(id),
    agent_id INT REFERENCES agents(id),  -- 关联的助手
    group_id UUID REFERENCES session_groups(id),  -- 所属分组
    title VARCHAR(200),
    description TEXT,
    pinned BOOLEAN DEFAULT FALSE,
    archived BOOLEAN DEFAULT FALSE,
    model VARCHAR(100),  -- 默认模型
    temperature FLOAT DEFAULT 0.7,
    top_p FLOAT DEFAULT 1.0,
    max_tokens INT,
    system_role TEXT,  -- 自定义系统提示词
    context_length INT DEFAULT 4,  -- 上下文消息数
    plugin_ids INT[],  -- 启用的插件 ID 列表
    knowledge_base_ids INT[],  -- 关联的知识库 ID 列表
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_sessions_user ON sessions(user_id, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_agent ON sessions(agent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_group ON sessions(group_id);
```

### 5. 会话分组表 (session_groups)

```sql
CREATE TABLE session_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_session_groups_user ON session_groups(user_id, sort_order);
```

### 6. 消息表 (messages)

**分区策略**：按月分区（降低单表数据量，提升查询性能）

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    topic_id UUID REFERENCES topics(id),
    parent_id UUID REFERENCES messages(id),  -- 父消息（用于树状结构）
    role VARCHAR(20) NOT NULL,  -- 'user' | 'assistant' | 'system' | 'tool'
    content TEXT NOT NULL,
    model VARCHAR(100),
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    cost BIGINT DEFAULT 0,  -- 本次对话花费（单位：分）
    metadata JSONB,  -- 扩展字段
    files JSONB,  -- 附件列表 [{file_id, file_name, file_type, file_url}]
    tool_calls JSONB,  -- 工具调用记录
    status INT DEFAULT 1,  -- 1: 正常, 2: 错误, 3: 已删除
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- 创建分区表（按月）
CREATE TABLE messages_2024_11 PARTITION OF messages
    FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
CREATE TABLE messages_2024_12 PARTITION OF messages
    FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');
-- 后续月份由定时任务自动创建

CREATE INDEX idx_messages_session ON messages(session_id, created_at DESC);
CREATE INDEX idx_messages_topic ON messages(topic_id, created_at DESC);
CREATE INDEX idx_messages_parent ON messages(parent_id);
```

### 7. 话题表 (topics)

```sql
CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    title VARCHAR(200),
    summary TEXT,
    message_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_topics_session ON topics(session_id, created_at DESC);
```

### 8. 助手表 (agents)

```sql
CREATE TABLE agents (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),  -- NULL 表示系统内置助手
    identifier VARCHAR(100) UNIQUE NOT NULL,  -- 唯一标识符
    name VARCHAR(100) NOT NULL,
    avatar VARCHAR(255),
    description TEXT,
    category VARCHAR(50),  -- 分类：写作、编程、翻译等
    system_role TEXT,  -- System Prompt
    model VARCHAR(100),
    temperature FLOAT DEFAULT 0.7,
    top_p FLOAT DEFAULT 1.0,
    max_tokens INT,
    tools JSONB,  -- 启用的工具配置
    plugins INT[],  -- 启用的插件 ID 列表
    knowledge_bases INT[],  -- 关联的知识库
    is_public BOOLEAN DEFAULT FALSE,  -- 是否在市场公开
    is_featured BOOLEAN DEFAULT FALSE,  -- 是否精选
    views INT DEFAULT 0,
    likes INT DEFAULT 0,
    forks INT DEFAULT 0,  -- 被复制次数
    status INT DEFAULT 1,  -- 1: 正常, 2: 下架
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_agents_user ON agents(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_agents_public ON agents(is_public, is_featured, likes DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_agents_identifier ON agents(identifier);
```

### 9. 助手标签表 (agent_tags)

```sql
CREATE TABLE agent_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE agent_tag_relations (
    agent_id INT REFERENCES agents(id) ON DELETE CASCADE,
    tag_id INT REFERENCES agent_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (agent_id, tag_id)
);

CREATE INDEX idx_agent_tag_relations_tag ON agent_tag_relations(tag_id);
```

### 10. 知识库表 (knowledge_bases)

```sql
CREATE TABLE knowledge_bases (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    embedding_model VARCHAR(100) DEFAULT 'text-embedding-3-small',
    chunk_size INT DEFAULT 512,
    chunk_overlap INT DEFAULT 50,
    document_count INT DEFAULT 0,
    total_chunks INT DEFAULT 0,
    status INT DEFAULT 1,  -- 1: 正常, 2: 处理中, 3: 已归档
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_kb_user ON knowledge_bases(user_id) WHERE deleted_at IS NULL;
```

### 11. 文档表 (documents)

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kb_id INT NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    file_url TEXT,
    file_type VARCHAR(50),  -- pdf, docx, txt, html, markdown
    file_size BIGINT,  -- 字节数
    status INT DEFAULT 1,  -- 1: 待处理, 2: 处理中, 3: 已完成, 4: 失败
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

### 12. 文档片段表 (chunks)

**使用 pgvector 扩展存储向量**

```sql
-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    embedding VECTOR(1536),  -- OpenAI text-embedding-3-small 维度为 1536
    metadata JSONB,  -- {page_number, section_title, ...}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 使用 IVFFlat 索引加速向量检索
CREATE INDEX idx_chunks_embedding ON chunks USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX idx_chunks_document ON chunks(document_id);
```

### 13. 渠道表 (channels)

```sql
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type INT NOT NULL,  -- 1: OpenAI, 2: Claude, 3: Gemini, 4: Azure, ...
    base_url TEXT,
    key TEXT NOT NULL,  -- 加密存储的 API Key
    other_config JSONB,  -- 其他配置（如 Azure 的 deployment_name）
    models TEXT[],  -- 支持的模型列表
    priority INT DEFAULT 0,  -- 优先级（数字越大越优先）
    weight INT DEFAULT 100,  -- 负载均衡权重
    status INT DEFAULT 1,  -- 1: 启用, 2: 禁用, 3: 自动禁用（测试失败）
    test_enabled BOOLEAN DEFAULT TRUE,
    test_time TIMESTAMP,  -- 上次测试时间
    test_result TEXT,
    response_time INT,  -- 平均响应时间(ms)
    success_rate FLOAT,  -- 成功率
    balance FLOAT,  -- 余额（如适用）
    balance_updated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_channels_status ON channels(status, priority DESC, weight DESC);
CREATE INDEX idx_channels_type ON channels(type);
```

### 14. 模型价格表 (pricing)

```sql
CREATE TABLE pricing (
    id SERIAL PRIMARY KEY,
    model VARCHAR(100) UNIQUE NOT NULL,
    provider VARCHAR(50),
    input_price DECIMAL(10, 6) NOT NULL,  -- 单位：元/1K tokens
    output_price DECIMAL(10, 6) NOT NULL,
    currency VARCHAR(10) DEFAULT 'CNY',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 示例数据
INSERT INTO pricing (model, provider, input_price, output_price) VALUES
    ('gpt-4', 'OpenAI', 0.210, 0.420),
    ('gpt-4-turbo', 'OpenAI', 0.070, 0.210),
    ('gpt-3.5-turbo', 'OpenAI', 0.0035, 0.0070),
    ('claude-3-5-sonnet', 'Anthropic', 0.021, 0.105);
```

### 15. 计费日志表 (billing_logs)

**分区策略**：按月分区

```sql
CREATE TABLE billing_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    session_id UUID REFERENCES sessions(id),
    message_id UUID REFERENCES messages(id),
    model VARCHAR(100) NOT NULL,
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    cost BIGINT NOT NULL,  -- 花费（单位：分）
    channel_id INT REFERENCES channels(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- 创建分区表
CREATE TABLE billing_logs_2024_11 PARTITION OF billing_logs
    FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');

CREATE INDEX idx_billing_user ON billing_logs(user_id, created_at DESC);
CREATE INDEX idx_billing_session ON billing_logs(session_id);
```

### 16. 插件表 (plugins)

```sql
CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    identifier VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    author VARCHAR(100),
    version VARCHAR(20) DEFAULT '1.0.0',
    manifest JSONB NOT NULL,  -- 插件的完整定义（参考 LobeChat 插件规范）
    api_endpoint TEXT,
    icon_url TEXT,
    category VARCHAR(50),
    is_builtin BOOLEAN DEFAULT FALSE,  -- 内置插件
    is_public BOOLEAN DEFAULT TRUE,
    installs INT DEFAULT 0,
    rating FLOAT DEFAULT 0,
    status INT DEFAULT 1,  -- 1: 正常, 2: 下架
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plugins_public ON plugins(is_public, rating DESC);
CREATE INDEX idx_plugins_identifier ON plugins(identifier);
```

### 17. 用户插件配置表 (user_plugins)

```sql
CREATE TABLE user_plugins (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    plugin_id INT REFERENCES plugins(id) ON DELETE CASCADE,
    config JSONB,  -- 用户的个性化配置（如 API Key）
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, plugin_id)
);

CREATE INDEX idx_user_plugins_user ON user_plugins(user_id) WHERE enabled = TRUE;
```

### 18. 文件表 (files)

```sql
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,  -- 字节数
    mime_type VARCHAR(100),
    storage_path TEXT NOT NULL,  -- MinIO 中的路径
    hash VARCHAR(64),  -- SHA256 哈希（用于去重）
    metadata JSONB,  -- {width, height, duration, ...}
    related_type VARCHAR(50),  -- 关联类型：message, document, avatar
    related_id VARCHAR(100),  -- 关联 ID
    status INT DEFAULT 1,  -- 1: 正常, 2: 已删除
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_files_user ON files(user_id, created_at DESC) WHERE status = 1;
CREATE INDEX idx_files_hash ON files(hash);
CREATE INDEX idx_files_related ON files(related_type, related_id);
```

### 19. 系统配置表 (system_options)

```sql
CREATE TABLE system_options (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 示例配置
INSERT INTO system_options (key, value, description) VALUES
    ('site_name', 'Oblivious', '站点名称'),
    ('default_quota', '5000', '新用户默认额度（分）'),
    ('enable_registration', 'true', '是否开放注册'),
    ('max_message_length', '10000', '单条消息最大字符数'),
    ('file_upload_max_size', '20971520', '文件上传最大大小（字节）');
```

---

## 数据库优化策略

### 1. 索引优化

- 为高频查询字段添加索引
- 复合索引遵循"最左前缀"原则
- 使用部分索引（`WHERE` 条件）减少索引大小

### 2. 分区表

对于高增长表（`messages`, `billing_logs`）采用时间分区：

```sql
-- 自动创建下月分区（定时任务）
CREATE OR REPLACE FUNCTION create_next_month_partition()
RETURNS void AS $$
DECLARE
    next_month DATE := date_trunc('month', CURRENT_DATE + INTERVAL '1 month');
    partition_name TEXT;
BEGIN
    partition_name := 'messages_' || to_char(next_month, 'YYYY_MM');
    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF messages FOR VALUES FROM (%L) TO (%L)',
                   partition_name, next_month, next_month + INTERVAL '1 month');
END;
$$ LANGUAGE plpgsql;
```

### 3. 读写分离

- **主库**：处理所有写操作
- **从库**：处理所有只读查询（会话列表、消息历史等）

### 4. 数据归档

对于超过 1 年的历史数据，迁移到归档库：

```sql
-- 迁移旧数据到归档表
INSERT INTO messages_archive SELECT * FROM messages WHERE created_at < CURRENT_DATE - INTERVAL '1 year';
DELETE FROM messages WHERE created_at < CURRENT_DATE - INTERVAL '1 year';
```

### 5. 连接池

```go
db.SetMaxOpenConns(100)      // 最大连接数
db.SetMaxIdleConns(10)       // 最大空闲连接数
db.SetConnMaxLifetime(time.Hour)  // 连接最大存活时间
```

---

## 数据迁移管理

使用 **golang-migrate** 管理数据库版本：

```bash
migrate create -ext sql -dir migrations -seq create_users_table
```

```sql
-- migrations/000001_create_users_table.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    ...
);

-- migrations/000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

---

## 备份策略

1. **全量备份**：每天凌晨 3 点执行
2. **增量备份**：每小时执行 WAL 归档
3. **备份保留**：保留最近 30 天的备份
4. **异地备份**：备份文件同步到另一个地域的对象存储

```bash
# pg_dump 全量备份
pg_dump -U postgres -d oblivious -F c -f /backups/oblivious_$(date +%Y%m%d).dump

# 恢复
pg_restore -U postgres -d oblivious -c /backups/oblivious_20241120.dump
```

---

## 总结

数据库设计遵循范式化原则，同时针对高并发场景进行了优化（索引、分区、缓存）。通过合理的表结构设计和分片策略，可以支撑千万级用户和亿级消息数据。

