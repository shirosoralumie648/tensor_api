-- 增强 channels 表 - 添加 New API 的 19 个新字段
-- Version: 000012
-- Description: 将 channels 表升级到与 New API 一致的功能级别

BEGIN;

-- 1. 备份现有数据（创建备份表）
CREATE TABLE IF NOT EXISTS channels_backup_20251121 AS 
SELECT * FROM channels;

-- 2. 添加负载均衡和优先级字段
ALTER TABLE channels ADD COLUMN IF NOT EXISTS priority BIGINT DEFAULT 0;

-- 3. 添加分组和标签字段
ALTER TABLE channels ADD COLUMN IF NOT EXISTS "group" VARCHAR(64) DEFAULT 'default';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS tag VARCHAR(100);

-- 4. 添加多密钥配置
ALTER TABLE channels ADD COLUMN IF NOT EXISTS channel_info JSONB DEFAULT '{}';

-- 5. 添加监控和状态字段
ALTER TABLE channels ADD COLUMN IF NOT EXISTS response_time INT DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS test_time BIGINT DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS status INT;

-- 6. 添加配额和余额字段
ALTER TABLE channels ADD COLUMN IF NOT EXISTS used_quota BIGINT DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS balance FLOAT8 DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS balance_updated_time BIGINT DEFAULT 0;

-- 7. 添加高级配置字段
ALTER TABLE channels ADD COLUMN IF NOT EXISTS status_code_mapping VARCHAR(1024);
ALTER TABLE channels ADD COLUMN IF NOT EXISTS param_override TEXT;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS header_override TEXT;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS auto_ban INT DEFAULT 1;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS other_info TEXT DEFAULT '{}';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS other_settings TEXT DEFAULT '{}';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS remark VARCHAR(255);

-- 8. 迁移 enabled 字段到 status（保留 enabled 以便回滚）
-- status: 1=启用 2=手动禁用 3=自动禁用
UPDATE channels 
SET status = CASE 
    WHEN enabled THEN 1 
    ELSE 2 
END 
WHERE status IS NULL;

ALTER TABLE channels ALTER COLUMN status SET DEFAULT 1;

-- 9. 创建索引
CREATE INDEX IF NOT EXISTS idx_channels_status ON channels(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_group ON channels("group") WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_tag ON channels(tag) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_priority ON channels(priority DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_weight ON channels(weight DESC) WHERE deleted_at IS NULL;

-- 10. 复合索引（优化渠道选择查询）
CREATE INDEX IF NOT EXISTS idx_channels_status_priority 
ON channels(status, priority DESC, weight DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_channels_group_status 
ON channels("group", status) 
WHERE deleted_at IS NULL;

-- 11. 添加注释
COMMENT ON COLUMN channels.priority IS '优先级（数值越大优先级越高）';
COMMENT ON COLUMN channels."group" IS '用户分组（用于定价倍率）';
COMMENT ON COLUMN channels.tag IS '标签（用于批量操作）';
COMMENT ON COLUMN channels.channel_info IS '多密钥配置（JSON格式）';
COMMENT ON COLUMN channels.response_time IS '平均响应时间（毫秒）';
COMMENT ON COLUMN channels.test_time IS '最后测试时间（Unix时间戳）';
COMMENT ON COLUMN channels.status IS '状态（1:启用 2:手动禁用 3:自动禁用）';
COMMENT ON COLUMN channels.used_quota IS '已使用配额';
COMMENT ON COLUMN channels.balance IS '余额（美元）';
COMMENT ON COLUMN channels.auto_ban IS '自动禁用开关（1:开启 2:关闭）';

COMMIT;
