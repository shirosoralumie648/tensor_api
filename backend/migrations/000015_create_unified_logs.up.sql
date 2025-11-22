-- 创建统一日志表
-- Version: 000015
-- Description: 整合计费日志和配额日志，提供完整的调用链路追踪

BEGIN;

CREATE TABLE IF NOT EXISTS unified_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    username VARCHAR(100),
    token_id INT,
    token_name VARCHAR(100),
    channel_id INT,
    channel_name VARCHAR(100),
    log_type INT NOT NULL, -- 1:充值 2:消费 3:管理 4:系统 5:错误 6:退款
    model_name VARCHAR(100),
    content TEXT,
    quota INT DEFAULT 0,
    prompt_tokens INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    use_time INT DEFAULT 0, -- 请求耗时（毫秒）
    is_stream BOOLEAN DEFAULT false,
    "group" VARCHAR(64),
    ip VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    other JSONB, -- 额外信息（错误详情、请求参数等）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 单列索引
CREATE INDEX IF NOT EXISTS idx_logs_user ON unified_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_logs_type ON unified_logs(log_type);
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON unified_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_model ON unified_logs(model_name);
CREATE INDEX IF NOT EXISTS idx_logs_channel ON unified_logs(channel_id);
CREATE INDEX IF NOT EXISTS idx_logs_request_id ON unified_logs(request_id);

-- 复合索引（优化常用查询）
CREATE INDEX IF NOT EXISTS idx_logs_user_time 
ON unified_logs(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_logs_type_time 
ON unified_logs(log_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_logs_user_type_time 
ON unified_logs(user_id, log_type, created_at DESC);

-- 分区表准备（可选，数据量大时启用）
-- 按月分区可以提高查询性能和数据维护效率
-- 示例：
-- ALTER TABLE unified_logs PARTITION BY RANGE (created_at);
-- CREATE TABLE unified_logs_2025_11 PARTITION OF unified_logs
-- FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

-- 添加注释
COMMENT ON TABLE unified_logs IS '统一日志表 - 记录所有操作和调用日志';
COMMENT ON COLUMN unified_logs.log_type IS '日志类型（1:充值 2:消费 3:管理 4:系统 5:错误 6:退款）';
COMMENT ON COLUMN unified_logs.quota IS '配额变化量（正数表示增加，负数表示减少）';
COMMENT ON COLUMN unified_logs.use_time IS '请求耗时（毫秒）';
COMMENT ON COLUMN unified_logs.request_id IS '请求追踪 ID（用于关联请求链路）';
COMMENT ON COLUMN unified_logs.other IS 'JSON 格式的额外信息（错误详情、请求参数等）';

COMMIT;
