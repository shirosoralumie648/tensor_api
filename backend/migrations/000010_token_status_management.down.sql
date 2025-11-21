-- Token 状态管理系统回滚

-- 删除触发器
DROP TRIGGER IF EXISTS update_tokens_timestamp ON tokens;

-- 删除函数
DROP FUNCTION IF EXISTS update_tokens_timestamp();
DROP FUNCTION IF EXISTS check_and_update_expired_tokens();
DROP FUNCTION IF EXISTS log_token_audit(INT, INT, VARCHAR, INT, INT, JSONB, VARCHAR, TEXT);
DROP FUNCTION IF EXISTS update_token_status();

-- 删除视图
DROP VIEW IF EXISTS valid_tokens;

-- 删除索引
DROP INDEX IF EXISTS idx_tokens_user_id;
DROP INDEX IF EXISTS idx_tokens_status;
DROP INDEX IF EXISTS idx_tokens_expire_at;
DROP INDEX IF EXISTS idx_tokens_deleted_at;
DROP INDEX IF EXISTS idx_tokens_token_hash;
DROP INDEX IF EXISTS idx_tokens_user_status;
DROP INDEX IF EXISTS idx_audit_log_user_id;
DROP INDEX IF EXISTS idx_audit_log_token_id;
DROP INDEX IF EXISTS idx_audit_log_operation;
DROP INDEX IF EXISTS idx_audit_log_created_at;
DROP INDEX IF EXISTS idx_audit_log_user_created;
DROP INDEX IF EXISTS idx_renewal_log_token_id;
DROP INDEX IF EXISTS idx_renewal_log_created_at;
DROP INDEX IF EXISTS idx_threshold_user_id;
DROP INDEX IF EXISTS idx_threshold_token_id;
DROP INDEX IF EXISTS idx_tokens_user_status_valid;
DROP INDEX IF EXISTS idx_audit_log_created_range;

-- 删除表
DROP TABLE IF EXISTS token_quota_threshold;
DROP TABLE IF EXISTS token_renewal_log;
DROP TABLE IF EXISTS token_audit_log;
DROP TABLE IF EXISTS tokens;

-- 删除添加到 users 表的列
ALTER TABLE users DROP COLUMN IF EXISTS token_status;
ALTER TABLE users DROP COLUMN IF EXISTS token_expire_at;
ALTER TABLE users DROP COLUMN IF EXISTS token_deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS token_renewed_at;

