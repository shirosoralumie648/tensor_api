-- 回滚渠道能力表
-- Version: 000014

BEGIN;

DROP INDEX IF EXISTS idx_abilities_query;
DROP INDEX IF EXISTS idx_abilities_unique;
DROP INDEX IF EXISTS idx_abilities_enabled;
DROP INDEX IF EXISTS idx_abilities_group;
DROP INDEX IF EXISTS idx_abilities_model;
DROP INDEX IF EXISTS idx_abilities_channel;

DROP TABLE IF EXISTS channel_abilities;

COMMIT;
