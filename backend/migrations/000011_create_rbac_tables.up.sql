-- RBAC (Role-Based Access Control) 权限管理系统迁移

-- 创建 roles 表（角色）
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    
    -- 权限配置（JSON格式）
    permissions JSONB DEFAULT '{}',
    
    -- 权限等级（用于继承，数值越小权限越大）
    level INT DEFAULT 100,
    
    -- 是否为系统内置角色
    is_builtin BOOLEAN DEFAULT FALSE,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    -- 元数据
    metadata JSONB
);

-- 创建 user_roles 关联表（用户-角色映射）
CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    
    -- 赋予时间
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    assigned_by INT REFERENCES users(id),
    
    -- 过期时间（可选，为空表示永久有效）
    expire_at TIMESTAMP,
    
    -- 关联元数据
    metadata JSONB,
    
    -- 确保同一用户不会有重复的角色
    UNIQUE(user_id, role_id)
);

-- 创建 permissions 表（权限定义）
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200),
    description TEXT,
    
    -- 权限所属的资源或模块
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    
    -- 权限级别（用于排序）
    level INT DEFAULT 100,
    
    -- 是否为系统权限
    is_system BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    -- 唯一约束：同一资源的同一动作只能有一个权限
    UNIQUE(resource, action)
);

-- 创建 role_permissions 关联表（角色-权限映射）
CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    -- 同一角色不能有重复的权限
    UNIQUE(role_id, permission_id)
);

-- 创建 permission_audit_log 表（权限操作审计日志）
CREATE TABLE IF NOT EXISTS permission_audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    
    -- 操作类型：assign_role, revoke_role, grant_permission, revoke_permission
    operation VARCHAR(50) NOT NULL,
    
    -- 相关的角色或权限
    target_user_id INT,
    role_id INT REFERENCES roles(id) ON DELETE SET NULL,
    permission_id INT REFERENCES permissions(id) ON DELETE SET NULL,
    
    -- 变更详情
    details JSONB,
    
    -- 审计信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- 创建 role_hierarchy 表（角色继承关系）
CREATE TABLE IF NOT EXISTS role_hierarchy (
    id SERIAL PRIMARY KEY,
    parent_role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    child_role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    -- 同一子角色不能继承同一父角色
    UNIQUE(parent_role_id, child_role_id)
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_roles_level ON roles(level);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_expire_at ON user_roles(expire_at) 
    WHERE expire_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_permission_audit_log_user_id ON permission_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_permission_audit_log_created_at ON permission_audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_role_hierarchy_parent ON role_hierarchy(parent_role_id);
CREATE INDEX IF NOT EXISTS idx_role_hierarchy_child ON role_hierarchy(child_role_id);

-- 创建组合索引用于常见查询
CREATE INDEX IF NOT EXISTS idx_user_roles_user_active 
    ON user_roles(user_id) WHERE expire_at IS NULL OR expire_at > CURRENT_TIMESTAMP;

-- 创建触发器自动更新 updated_at
CREATE OR REPLACE FUNCTION update_roles_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_roles_timestamp ON roles;
CREATE TRIGGER update_roles_timestamp
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION update_roles_timestamp();

-- 创建函数：检查用户是否拥有特定权限
CREATE OR REPLACE FUNCTION user_has_permission(p_user_id INT, p_permission_name VARCHAR)
RETURNS BOOLEAN AS $$
BEGIN
    -- 检查用户是否有该权限（通过角色）
    RETURN EXISTS (
        SELECT 1
        FROM user_roles ur
        JOIN role_permissions rp ON ur.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE ur.user_id = p_user_id
          AND p.name = p_permission_name
          AND (ur.expire_at IS NULL OR ur.expire_at > CURRENT_TIMESTAMP)
    );
END;
$$ LANGUAGE plpgsql;

-- 创建函数：检查用户是否拥有特定角色
CREATE OR REPLACE FUNCTION user_has_role(p_user_id INT, p_role_name VARCHAR)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        WHERE ur.user_id = p_user_id
          AND r.name = p_role_name
          AND (ur.expire_at IS NULL OR ur.expire_at > CURRENT_TIMESTAMP)
    );
END;
$$ LANGUAGE plpgsql;

-- 创建函数：获取用户的所有权限
CREATE OR REPLACE FUNCTION get_user_permissions(p_user_id INT)
RETURNS TABLE(permission_id INT, permission_name VARCHAR, resource VARCHAR, action VARCHAR) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT
        p.id,
        p.name,
        p.resource,
        p.action
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = p_user_id
      AND (ur.expire_at IS NULL OR ur.expire_at > CURRENT_TIMESTAMP);
END;
$$ LANGUAGE plpgsql;

-- 创建函数：获取用户的所有角色
CREATE OR REPLACE FUNCTION get_user_roles(p_user_id INT)
RETURNS TABLE(role_id INT, role_name VARCHAR, display_name VARCHAR) AS $$
BEGIN
    RETURN QUERY
    SELECT
        r.id,
        r.name,
        r.display_name
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = p_user_id
      AND (ur.expire_at IS NULL OR ur.expire_at > CURRENT_TIMESTAMP);
END;
$$ LANGUAGE plpgsql;

-- 创建函数：记录权限操作审计日志
CREATE OR REPLACE FUNCTION log_permission_audit(
    p_user_id INT,
    p_operation VARCHAR,
    p_target_user_id INT,
    p_role_id INT,
    p_permission_id INT,
    p_details JSONB,
    p_ip_address VARCHAR,
    p_user_agent TEXT
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO permission_audit_log (
        user_id,
        operation,
        target_user_id,
        role_id,
        permission_id,
        details,
        ip_address,
        user_agent
    ) VALUES (
        p_user_id,
        p_operation,
        p_target_user_id,
        p_role_id,
        p_permission_id,
        p_details,
        p_ip_address,
        p_user_agent
    );
END;
$$ LANGUAGE plpgsql;

-- 插入系统内置角色
INSERT INTO roles (name, display_name, description, level, is_builtin) VALUES
    ('super_admin', '超级管理员', '拥有所有权限', 1, TRUE),
    ('admin', '管理员', '可以管理系统和用户', 10, TRUE),
    ('developer', '开发者', '可以使用API和创建应用', 50, TRUE),
    ('user', '普通用户', '基础用户权限', 100, TRUE)
ON CONFLICT (name) DO NOTHING;

-- 插入系统权限
INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    -- 用户管理权限
    ('user.create', '创建用户', '创建新用户', 'user', 'create', TRUE),
    ('user.read', '查看用户', '查看用户信息', 'user', 'read', TRUE),
    ('user.update', '更新用户', '更新用户信息', 'user', 'update', TRUE),
    ('user.delete', '删除用户', '删除用户', 'user', 'delete', TRUE),
    
    -- 角色管理权限
    ('role.create', '创建角色', '创建新角色', 'role', 'create', TRUE),
    ('role.read', '查看角色', '查看角色信息', 'role', 'read', TRUE),
    ('role.update', '更新角色', '更新角色信息', 'role', 'update', TRUE),
    ('role.delete', '删除角色', '删除角色', 'role', 'delete', TRUE),
    ('role.assign', '分配角色', '为用户分配角色', 'role', 'assign', TRUE),
    
    -- API权限
    ('api.access', '访问API', '使用API服务', 'api', 'access', TRUE),
    ('api.create', '创建API', '创建新的API密钥', 'api', 'create', TRUE),
    ('api.manage', '管理API', '管理API密钥和权限', 'api', 'manage', TRUE),
    
    -- 系统管理权限
    ('system.admin', '系统管理', '系统级别的管理权限', 'system', 'admin', TRUE),
    ('system.config', '系统配置', '修改系统配置', 'system', 'config', TRUE),
    ('system.logs', '查看日志', '查看系统日志', 'system', 'logs', TRUE)
ON CONFLICT (resource, action) DO NOTHING;

-- 为系统内置角色分配权限

-- 超级管理员 - 拥有所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'super_admin' AND p.is_system = TRUE
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 管理员 - 大部分管理权限（除了系统级权限）
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin' 
  AND p.is_system = TRUE
  AND p.resource IN ('user', 'role', 'api')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 开发者 - 仅限API相关权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'developer'
  AND p.is_system = TRUE
  AND (p.resource = 'api' OR (p.resource = 'user' AND p.action = 'read'))
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 普通用户 - 仅限基础访问权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'user'
  AND p.is_system = TRUE
  AND p.name = 'api.access'
ON CONFLICT (role_id, permission_id) DO NOTHING;

