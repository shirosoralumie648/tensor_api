-- 回滚模型定价表
-- Version: 000016

BEGIN;

DROP INDEX IF EXISTS idx_pricing_query;
DROP INDEX IF EXISTS idx_pricing_model_group;
DROP INDEX IF EXISTS idx_pricing_vendor;
DROP INDEX IF EXISTS idx_pricing_enabled;
DROP INDEX IF EXISTS idx_pricing_group;
DROP INDEX IF EXISTS idx_pricing_model;

DROP TABLE IF EXISTS model_pricing;

COMMIT;
