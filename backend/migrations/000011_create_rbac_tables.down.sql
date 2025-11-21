-- 回滚 RBAC 权限管理系统迁移

-- 删除函数
DROP FUNCTION IF EXISTS log_permission_audit(INT, VARCHAR, INT, INT, INT, JSONB, VARCHAR, TEXT);
DROP FUNCTION IF EXISTS get_user_roles(INT);
DROP FUNCTION IF EXISTS get_user_permissions(INT);
DROP FUNCTION IF EXISTS user_has_role(INT, VARCHAR);
DROP FUNCTION IF EXISTS user_has_permission(INT, VARCHAR);

-- 删除触发器
DROP TRIGGER IF EXISTS update_roles_timestamp ON roles;
DROP FUNCTION IF EXISTS update_roles_timestamp();

-- 删除表（注意删除顺序：先删除依赖的表）
DROP TABLE IF EXISTS role_hierarchy CASCADE;
DROP TABLE IF EXISTS permission_audit_log CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

