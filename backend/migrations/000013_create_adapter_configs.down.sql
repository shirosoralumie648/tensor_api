-- 回滚适配器配置表
-- Version: 000013

BEGIN;

DROP INDEX IF EXISTS idx_adapter_configs_enabled;
DROP INDEX IF EXISTS idx_adapter_configs_type;
DROP INDEX IF EXISTS idx_adapter_configs_name;

DROP TABLE IF EXISTS adapter_configs;

COMMIT;
