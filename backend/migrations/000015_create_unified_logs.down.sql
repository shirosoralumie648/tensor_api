-- 回滚统一日志表
-- Version: 000015

BEGIN;

DROP INDEX IF EXISTS idx_logs_user_type_time;
DROP INDEX IF EXISTS idx_logs_type_time;
DROP INDEX IF EXISTS idx_logs_user_time;
DROP INDEX IF EXISTS idx_logs_request_id;
DROP INDEX IF EXISTS idx_logs_channel;
DROP INDEX IF EXISTS idx_logs_model;
DROP INDEX IF EXISTS idx_logs_created_at;
DROP INDEX IF EXISTS idx_logs_type;
DROP INDEX IF EXISTS idx_logs_user;

DROP TABLE IF EXISTS unified_logs;

COMMIT;
