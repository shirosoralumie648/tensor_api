-- 回滚 channels 表增强
-- Version: 000012

BEGIN;

-- 删除新增的索引
DROP INDEX IF EXISTS idx_channels_group_status;
DROP INDEX IF EXISTS idx_channels_status_priority;
DROP INDEX IF EXISTS idx_channels_weight;
DROP INDEX IF EXISTS idx_channels_priority;
DROP INDEX IF EXISTS idx_channels_tag;
DROP INDEX IF EXISTS idx_channels_group;
DROP INDEX IF EXISTS idx_channels_status;

-- 删除新增的字段
ALTER TABLE channels DROP COLUMN IF EXISTS remark;
ALTER TABLE channels DROP COLUMN IF EXISTS other_settings;
ALTER TABLE channels DROP COLUMN IF EXISTS other_info;
ALTER TABLE channels DROP COLUMN IF EXISTS auto_ban;
ALTER TABLE channels DROP COLUMN IF EXISTS header_override;
ALTER TABLE channels DROP COLUMN IF EXISTS param_override;
ALTER TABLE channels DROP COLUMN IF EXISTS status_code_mapping;
ALTER TABLE channels DROP COLUMN IF EXISTS balance_updated_time;
ALTER TABLE channels DROP COLUMN IF EXISTS balance;
ALTER TABLE channels DROP COLUMN IF EXISTS used_quota;
ALTER TABLE channels DROP COLUMN IF EXISTS status;
ALTER TABLE channels DROP COLUMN IF EXISTS test_time;
ALTER TABLE channels DROP COLUMN IF EXISTS response_time;
ALTER TABLE channels DROP COLUMN IF EXISTS channel_info;
ALTER TABLE channels DROP COLUMN IF EXISTS tag;
ALTER TABLE channels DROP COLUMN IF EXISTS "group";
ALTER TABLE channels DROP COLUMN IF EXISTS priority;

-- 从备份恢复数据（可选）
-- 如果需要完全恢复，可以执行：
-- DROP TABLE IF EXISTS channels;
-- ALTER TABLE channels_backup_20251121 RENAME TO channels;

COMMIT;
