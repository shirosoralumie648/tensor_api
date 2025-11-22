-- 创建渠道能力表
-- Version: 000014
-- Description: 优化渠道模型查询性能，替代 CSV 字符串解析

BEGIN;

CREATE TABLE IF NOT EXISTS channel_abilities (
    id SERIAL PRIMARY KEY,
    channel_id INT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    model VARCHAR(100) NOT NULL,
    "group" VARCHAR(64) DEFAULT 'default',
    enabled BOOLEAN DEFAULT true,
    priority BIGINT DEFAULT 0,
    weight INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 单列索引
CREATE INDEX IF NOT EXISTS idx_abilities_channel ON channel_abilities(channel_id);
CREATE INDEX IF NOT EXISTS idx_abilities_model ON channel_abilities(model);
CREATE INDEX IF NOT EXISTS idx_abilities_group ON channel_abilities("group");
CREATE INDEX IF NOT EXISTS idx_abilities_enabled ON channel_abilities(enabled);

-- 唯一约束（防止重复）
CREATE UNIQUE INDEX IF NOT EXISTS idx_abilities_unique 
ON channel_abilities(channel_id, model, "group");

-- 复合索引（优化查询：根据模型和分组查询可用渠道）
CREATE INDEX IF NOT EXISTS idx_abilities_query 
ON channel_abilities(model, "group", enabled, priority DESC, weight DESC);

-- 从现有 channels 表同步数据
INSERT INTO channel_abilities (channel_id, model, "group", enabled, priority, weight)
SELECT 
    c.id,
    TRIM(m.model),
    COALESCE(c."group", 'default'),
    (c.status = 1 OR c.enabled = true),
    COALESCE(c.priority, 0),
    c.weight
FROM 
    channels c,
    LATERAL unnest(string_to_array(c.support_models, ',')) AS m(model)
WHERE 
    c.deleted_at IS NULL
    AND c.support_models IS NOT NULL
    AND c.support_models != ''
ON CONFLICT (channel_id, model, "group") DO UPDATE
SET 
    enabled = EXCLUDED.enabled,
    priority = EXCLUDED.priority,
    weight = EXCLUDED.weight,
    updated_at = CURRENT_TIMESTAMP;

-- 添加注释
COMMENT ON TABLE channel_abilities IS '渠道能力表 - 存储每个渠道支持的模型，优化查询性能';
COMMENT ON COLUMN channel_abilities.channel_id IS '渠道 ID（外键关联 channels 表）';
COMMENT ON COLUMN channel_abilities.model IS '模型名称';
COMMENT ON COLUMN channel_abilities."group" IS '用户分组';
COMMENT ON COLUMN channel_abilities.enabled IS '是否启用';
COMMENT ON COLUMN channel_abilities.priority IS '优先级（继承自 channel）';
COMMENT ON COLUMN channel_abilities.weight IS '权重（继承自 channel）';

COMMIT;
