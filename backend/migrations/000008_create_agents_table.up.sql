-- 创建 agents 表（AI 助手）
CREATE TABLE IF NOT EXISTS agents (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    identifier VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    avatar TEXT,
    description TEXT,
    category VARCHAR(50),
    system_role TEXT,
    model VARCHAR(100) NOT NULL,
    temperature FLOAT DEFAULT 0.7,
    top_p FLOAT DEFAULT 1.0,
    max_tokens INT,
    tools JSONB,
    plugin_ids INTEGER[],
    knowledge_base_ids INTEGER[],
    is_public BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    views INT DEFAULT 0,
    likes INT DEFAULT 0,
    forks INT DEFAULT 0,
    status INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 创建索引
CREATE INDEX idx_agents_user ON agents(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_agents_public ON agents(is_public, is_featured, likes DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_agents_identifier ON agents(identifier);
CREATE INDEX idx_agents_model ON agents(model);
CREATE INDEX idx_agents_created ON agents(created_at DESC);

-- 创建 agent_forks 表（助手 Fork 记录）
CREATE TABLE IF NOT EXISTS agent_forks (
    id SERIAL PRIMARY KEY,
    original_id INT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fork_name VARCHAR(100),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 创建索引
CREATE INDEX idx_agent_forks_original ON agent_forks(original_id);
CREATE INDEX idx_agent_forks_user ON agent_forks(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_agent_forks_created ON agent_forks(created_at DESC);

-- 创建 agent_usages 表（助手使用统计）
CREATE TABLE IF NOT EXISTS agent_usages (
    id SERIAL PRIMARY KEY,
    agent_id INT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(255),
    message_count INT DEFAULT 0,
    token_count INT DEFAULT 0,
    cost FLOAT8 DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_agent_usages_agent ON agent_usages(agent_id);
CREATE INDEX idx_agent_usages_user ON agent_usages(user_id);
CREATE INDEX idx_agent_usages_created ON agent_usages(created_at DESC);

