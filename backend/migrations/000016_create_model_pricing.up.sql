-- 创建模型定价表
-- Version: 000016
-- Description: 支持两种计费模式（按次/按量）和用户分组倍率

BEGIN;

CREATE TABLE IF NOT EXISTS model_pricing (
    id SERIAL PRIMARY KEY,
    model VARCHAR(100) NOT NULL,
    "group" VARCHAR(64) DEFAULT 'default',
    quota_type INT DEFAULT 0, -- 0:按量计费 1:按次计费
    model_price FLOAT8, -- 按次价格（quota_type=1）
    model_ratio FLOAT8, -- Token倍率（quota_type=0）
    completion_ratio FLOAT8 DEFAULT 1.0, -- 输出Token额外倍率
    group_ratio FLOAT8 DEFAULT 1.0, -- 用户分组倍率
    vendor_id VARCHAR(50), -- 供应商ID（openai、claude等）
    enabled BOOLEAN DEFAULT true,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_pricing_model ON model_pricing(model) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pricing_group ON model_pricing("group") WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pricing_enabled ON model_pricing(enabled) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pricing_vendor ON model_pricing(vendor_id) WHERE deleted_at IS NULL;

-- 唯一约束（每个模型在每个分组下只有一个定价）
-- 注意：为了支持 ON CONFLICT，需要创建不带 WHERE 条件的唯一索引
ALTER TABLE model_pricing ADD CONSTRAINT uq_pricing_model_group UNIQUE (model, "group");

-- 复合索引（优化查询）
CREATE INDEX IF NOT EXISTS idx_pricing_query 
ON model_pricing(model, "group", enabled) WHERE deleted_at IS NULL;

-- 插入常用模型的默认定价（示例）
INSERT INTO model_pricing (model, "group", quota_type, model_ratio, completion_ratio, vendor_id, description)
VALUES 
    -- OpenAI 模型
    ('gpt-4o', 'default', 0, 15.0, 2.0, 'openai', 'GPT-4o 最新模型'),
    ('gpt-4o-mini', 'default', 0, 0.6, 2.0, 'openai', 'GPT-4o Mini 经济模型'),
    ('gpt-4-turbo', 'default', 0, 30.0, 2.0, 'openai', 'GPT-4 Turbo'),
    ('gpt-3.5-turbo', 'default', 0, 1.5, 2.0, 'openai', 'GPT-3.5 Turbo'),
    
    -- Claude 模型
    ('claude-3.5-sonnet-20241022', 'default', 0, 15.0, 5.0, 'claude', 'Claude 3.5 Sonnet'),
    ('claude-3-opus-20240229', 'default', 0, 75.0, 5.0, 'claude', 'Claude 3 Opus'),
    ('claude-3-sonnet-20240229', 'default', 0, 15.0, 5.0, 'claude', 'Claude 3 Sonnet'),
    ('claude-3-haiku-20240307', 'default', 0, 1.25, 5.0, 'claude', 'Claude 3 Haiku'),
    
    -- Gemini 模型
    ('gemini-2.0-flash-exp', 'default', 0, 0.0, 1.0, 'gemini', 'Gemini 2.0 Flash（免费实验）'),
    ('gemini-1.5-pro', 'default', 0, 7.0, 2.0, 'gemini', 'Gemini 1.5 Pro'),
    ('gemini-1.5-flash', 'default', 0, 0.35, 2.0, 'gemini', 'Gemini 1.5 Flash'),
    
    -- 国产模型
    ('qwen-max', 'default', 0, 8.0, 2.0, 'qwen', '通义千问 Max'),
    ('qwen-plus', 'default', 0, 2.0, 2.0, 'qwen', '通义千问 Plus'),
    ('deepseek-chat', 'default', 0, 1.0, 2.0, 'deepseek', 'DeepSeek Chat'),
    ('glm-4', 'default', 0, 5.0, 2.0, 'zhipu', '智谱 GLM-4')
ON CONFLICT (model, "group") DO NOTHING;

-- 添加注释
COMMENT ON TABLE model_pricing IS '模型定价表 - 支持按次/按量计费和用户分组倍率';
COMMENT ON COLUMN model_pricing.quota_type IS '计费类型（0:按量计费 1:按次计费）';
COMMENT ON COLUMN model_pricing.model_price IS '按次价格（仅当 quota_type=1 时有效）';
COMMENT ON COLUMN model_pricing.model_ratio IS 'Token 倍率（仅当 quota_type=0 时有效）';
COMMENT ON COLUMN model_pricing.completion_ratio IS '输出 Token 额外倍率';
COMMENT ON COLUMN model_pricing.group_ratio IS '用户分组倍率（用于不同等级用户差异化定价）';

COMMIT;
