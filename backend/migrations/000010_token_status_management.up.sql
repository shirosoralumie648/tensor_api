-- Token 状态管理系统迁移

-- 创建 token 表（如果尚未存在）
-- 这里假设 token 表已存在，我们只添加新的列

-- 添加状态相关的列到现有的 tokens 表
-- 注意: 这些 ALTER 语句应该根据实际的现有表结构调整
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_status INT DEFAULT 1;
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_expire_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '30 days');
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_deleted_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_renewed_at TIMESTAMP;

-- 创建 tokens 表（用于存储 API tokens）
CREATE TABLE IF NOT EXISTS tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100),
    description TEXT,
    
    -- 状态: 1=正常, 2=已耗尽, 3=已禁用, 4=已过期, 5=已删除(软删除)
    status INT DEFAULT 1 NOT NULL,
    
    -- 配额相关
    quota_limit BIGINT,
    quota_used BIGINT DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expire_at TIMESTAMP,
    renewed_at TIMESTAMP,
    deleted_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    -- 元数据
    last_used_at TIMESTAMP,
    ip_whitelist TEXT[], -- IP 白名单列表
    model_whitelist TEXT[], -- 模型白名单列表
    metadata JSONB
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_status ON tokens(status);
CREATE INDEX IF NOT EXISTS idx_tokens_expire_at ON tokens(expire_at);
CREATE INDEX IF NOT EXISTS idx_tokens_deleted_at ON tokens(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tokens_token_hash ON tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_tokens_user_status ON tokens(user_id, status);

-- 创建 token_audit_log 表（审计日志）
CREATE TABLE IF NOT EXISTS token_audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id INT NOT NULL REFERENCES tokens(id) ON DELETE CASCADE,
    
    -- 操作类型: create, update, delete, renew, disable, enable, expire
    operation VARCHAR(50) NOT NULL,
    
    -- 状态变化
    old_status INT,
    new_status INT,
    
    -- 变更信息
    details JSONB,
    
    -- 审计信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON token_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_token_id ON token_audit_log(token_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_operation ON token_audit_log(operation);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON token_audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_created ON token_audit_log(user_id, created_at);

-- 创建 token_renewal_log 表（续期日志）
CREATE TABLE IF NOT EXISTS token_renewal_log (
    id BIGSERIAL PRIMARY KEY,
    token_id INT NOT NULL REFERENCES tokens(id) ON DELETE CASCADE,
    
    -- 续期前后的过期时间
    old_expire_at TIMESTAMP,
    new_expire_at TIMESTAMP,
    
    -- 续期信息
    renewal_reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_renewal_log_token_id ON token_renewal_log(token_id);
CREATE INDEX IF NOT EXISTS idx_renewal_log_created_at ON token_renewal_log(created_at);

-- 创建 token_quota_threshold 表（配额预警阈值）
CREATE TABLE IF NOT EXISTS token_quota_threshold (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id INT NOT NULL REFERENCES tokens(id) ON DELETE CASCADE,
    
    -- 预警阈值百分比 (0-100)
    threshold_percent INT DEFAULT 50,
    
    -- 是否已发送预警
    is_warned BOOLEAN DEFAULT FALSE,
    warned_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_threshold_user_id ON token_quota_threshold(user_id);
CREATE INDEX IF NOT EXISTS idx_threshold_token_id ON token_quota_threshold(token_id);

-- 创建视图：当前有效的 token
CREATE OR REPLACE VIEW valid_tokens AS
SELECT 
    t.id,
    t.user_id,
    t.token_hash,
    t.name,
    t.status,
    t.quota_limit,
    t.quota_used,
    t.created_at,
    t.expire_at,
    t.last_used_at,
    CASE 
        WHEN t.status != 1 THEN FALSE
        WHEN t.expire_at IS NOT NULL AND t.expire_at < CURRENT_TIMESTAMP THEN FALSE
        WHEN t.deleted_at IS NOT NULL THEN FALSE
        ELSE TRUE
    END AS is_valid
FROM tokens t
WHERE t.status = 1 
    AND (t.expire_at IS NULL OR t.expire_at > CURRENT_TIMESTAMP)
    AND t.deleted_at IS NULL;

-- 创建函数：更新 token 状态
CREATE OR REPLACE FUNCTION update_token_status()
RETURNS VOID AS $$
BEGIN
    -- 标记已过期的 token
    UPDATE tokens 
    SET status = 4 
    WHERE status != 4 
        AND status != 5 
        AND expire_at IS NOT NULL 
        AND expire_at < CURRENT_TIMESTAMP;
    
    -- 标记配额已耗尽的 token
    UPDATE tokens 
    SET status = 2 
    WHERE status = 1 
        AND quota_limit IS NOT NULL 
        AND quota_used >= quota_limit;
END;
$$ LANGUAGE plpgsql;

-- 创建函数：记录 token 审计日志
CREATE OR REPLACE FUNCTION log_token_audit(
    p_user_id INT,
    p_token_id INT,
    p_operation VARCHAR(50),
    p_old_status INT,
    p_new_status INT,
    p_details JSONB,
    p_ip_address VARCHAR(45),
    p_user_agent TEXT
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO token_audit_log (
        user_id,
        token_id,
        operation,
        old_status,
        new_status,
        details,
        ip_address,
        user_agent
    ) VALUES (
        p_user_id,
        p_token_id,
        p_operation,
        p_old_status,
        p_new_status,
        p_details,
        p_ip_address,
        p_user_agent
    );
END;
$$ LANGUAGE plpgsql;

-- 创建触发器：自动更新 updated_at
CREATE OR REPLACE FUNCTION update_tokens_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_tokens_timestamp ON tokens;
CREATE TRIGGER update_tokens_timestamp
BEFORE UPDATE ON tokens
FOR EACH ROW
EXECUTE FUNCTION update_tokens_timestamp();

-- 创建定期检查过期 token 的函数
CREATE OR REPLACE FUNCTION check_and_update_expired_tokens()
RETURNS TABLE(updated_count INT) AS $$
DECLARE
    v_updated_count INT;
BEGIN
    -- 更新过期的 token 状态
    UPDATE tokens 
    SET status = 4
    WHERE status != 4 
        AND status != 5 
        AND expire_at IS NOT NULL 
        AND expire_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS v_updated_count = ROW_COUNT;
    
    RETURN QUERY SELECT v_updated_count;
END;
$$ LANGUAGE plpgsql;

-- 为查询性能添加组合索引
CREATE INDEX IF NOT EXISTS idx_tokens_user_status_valid 
ON tokens(user_id, status, expire_at) 
WHERE status != 5;

-- 为审计日志的时间范围查询优化
CREATE INDEX IF NOT EXISTS idx_audit_log_created_range 
ON token_audit_log(created_at DESC);

