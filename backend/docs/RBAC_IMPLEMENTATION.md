# RBAC 权限控制系统实现指南

## 概述

本文档描述了 Oblivious 平台的 RBAC (Role-Based Access Control) 权限控制系统实现。该系统提供：
- **角色管理**: 预定义的系统角色（超级管理员、管理员、开发者、普通用户）
- **权限管理**: 粒度到 API 端点级别的权限控制
- **权限检查**: <1ms 的高效权限验证
- **审计日志**: 完整的权限操作追踪
- **权限继承**: 支持角色间的权限继承

## 架构设计

### 数据模型

```
┌─────────────┐
│   User      │
└──────┬──────┘
       │
       │ (1:N)
       │
       ▼
┌─────────────────┐      ┌──────────────┐
│   UserRole      │─────▶│    Role      │
│ (user_id,       │      │ (角色定义)   │
│  role_id)       │      └──────┬───────┘
└─────────────────┘             │
                                │ (1:N)
                                │
                                ▼
                        ┌──────────────────┐
                        │ RolePermission   │
                        │ (role_id,        │
                        │  permission_id)  │
                        └────────┬─────────┘
                                 │
                                 │
                                 ▼
                        ┌──────────────────┐
                        │  Permission      │
                        │ (权限定义)       │
                        │ resource:action  │
                        └──────────────────┘
```

### 权限检查流程

```
请求到达
  ↓
认证 (获取 user_id)
  ↓
加载用户权限
  ├─ 从缓存查询
  └─ 或从数据库查询 (user_roles → roles → permissions)
  ↓
权限检查中间件
  ├─ RequirePermission (单个权限)
  ├─ RequirePermissions (任意权限)
  ├─ RequireAllPermissions (所有权限)
  ├─ RequireRole (单个角色)
  └─ RequireRoles (任意角色)
  ↓
允许/拒绝请求
```

## 系统内置角色

### 1. 超级管理员 (super_admin)
- **等级**: 1
- **权限**: 所有权限
- **能做**: 完全的系统管理

### 2. 管理员 (admin)
- **等级**: 10
- **权限**: 用户、角色、API 管理
- **能做**: 管理用户、分配角色、配置 API

### 3. 开发者 (developer)
- **等级**: 50
- **权限**: API 访问、创建、管理
- **能做**: 使用 API、创建应用、查看用户信息

### 4. 普通用户 (user)
- **等级**: 100
- **权限**: API 基础访问
- **能做**: 使用 API 服务

## 系统内置权限

### 用户管理 (user)
- `user.create` - 创建用户
- `user.read` - 查看用户
- `user.update` - 更新用户
- `user.delete` - 删除用户

### 角色管理 (role)
- `role.create` - 创建角色
- `role.read` - 查看角色
- `role.update` - 更新角色
- `role.delete` - 删除角色
- `role.assign` - 分配角色

### API 管理 (api)
- `api.access` - 访问 API
- `api.create` - 创建 API
- `api.manage` - 管理 API

### 系统管理 (system)
- `system.admin` - 系统管理
- `system.config` - 系统配置
- `system.logs` - 查看日志

## 使用方法

### 基础设置

```go
import (
    "github.com/oblivious/backend/internal/middleware"
    "time"
)

// 初始化 RBAC 管理器
rbacManager := middleware.NewRBACManager(30 * time.Minute)

// 应用中间件
router.Use(authHandler.HandleAuth())                      // 认证
router.Use(middleware.LoadUserPermissions(rbacManager))   // 加载权限
```

### 权限检查示例

```go
// 1. 单个权限检查
router.POST("/api/users", 
    middleware.RequirePermission("user.create"),
    createUserHandler)

// 2. 多个权限中的任意一个
router.GET("/api/reports", 
    middleware.RequirePermissions("report.read", "report.admin"),
    getReportsHandler)

// 3. 需要所有权限
router.DELETE("/api/users/:id", 
    middleware.RequireAllPermissions("user.delete", "user.verify"),
    deleteUserHandler)

// 4. 角色检查
router.GET("/api/admin/settings", 
    middleware.RequireRole("admin"),
    getAdminSettingsHandler)

// 5. 多个角色中的任意一个
router.POST("/api/audit", 
    middleware.RequireRoles("admin", "auditor"),
    createAuditHandler)
```

### 在处理器中使用权限信息

```go
func getUsersHandler(c *gin.Context) {
    // 获取用户的角色
    roles := middleware.GetUserRoleNames(c)
    fmt.Printf("User roles: %v\n", roles)
    
    // 获取用户的权限
    permissions := middleware.GetUserPermissionNames(c)
    fmt.Printf("User permissions: %v\n", permissions)
    
    // 检查特定资源访问
    if hasAccess, reason := middleware.CheckResourceAccess(
        c.MustGet("user_permissions").(*model.UserPermissions),
        "user",
        "read",
    ); !hasAccess {
        c.JSON(403, gin.H{"error": reason})
        return
    }
    
    // 处理请求...
}
```

## 权限验证性能

### 性能指标

| 操作 | 耗时 |
|------|------|
| 权限检查 (缓存命中) | <1ms |
| 权限检查 (DB查询) | <50ms |
| 角色检查 | <1ms |
| 权限加载 | <10ms (缓存) / <100ms (DB) |

### 性能优化

1. **多级缓存**
   - L1: 本地内存缓存 (30分钟 TTL)
   - L2: Redis 缓存

2. **权限批量加载**
   - 一次加载所有用户权限
   - 在上下文中复用

3. **权限预热**
   - 登录时预加载权限
   - Token 续期时刷新权限

## 权限管理 API

### 创建角色

```bash
POST /api/admin/roles
{
  "name": "custom_role",
  "display_name": "自定义角色",
  "description": "描述",
  "permissions": [1, 2, 3]  # 权限 ID 列表
}
```

### 分配角色给用户

```bash
POST /api/admin/users/:user_id/roles
{
  "role_id": 1,
  "expire_at": "2024-12-31T23:59:59Z"  # 可选，未指定表示永久
}
```

### 撤销角色

```bash
DELETE /api/admin/users/:user_id/roles/:role_id
```

### 查看用户权限

```bash
GET /api/admin/users/:user_id/permissions
```

### 查看用户角色

```bash
GET /api/admin/users/:user_id/roles
```

## 权限审计

### 审计日志记录

所有权限操作都会被记录到 `permission_audit_log` 表：

- 用户 ID
- 操作类型 (assign_role, revoke_role, grant_permission, revoke_permission)
- 目标用户 ID
- 相关角色/权限 ID
- 操作详情
- IP 地址
- User-Agent
- 操作时间戳

### 查询审计日志

```bash
GET /api/admin/audit-logs?user_id=123&operation=assign_role&limit=100
```

## 权限继承

### 角色继承

角色可以继承其他角色的权限：

```sql
-- 创建继承关系
INSERT INTO role_hierarchy (parent_role_id, child_role_id)
VALUES (admin_role_id, moderator_role_id);
```

### 权限继承规则

1. 子角色自动拥有父角色的所有权限
2. 权限不会逆向继承
3. 支持多层继承

## 权限缓存策略

### 缓存键格式

```
user_permissions:{user_id}
```

### 缓存失效场景

以下操作会导致权限缓存失效：

1. 角色分配/撤销
2. 权限添加/删除
3. 用户被删除
4. TTL 过期（30分钟）

### 手动刷新权限缓存

```go
// 在处理器中
rbacManager.InvalidateUserPermissionCache(c, userID)
```

## 最佳实践

### 1. 权限命名规范

```
resource.action

示例:
- user.create
- user.read
- user.update
- user.delete
- report.export
- api.manage
```

### 2. 使用组合权限

```go
// ❌ 不推荐：每个操作都定义一个权限
requir Permissions:
  - data.view_public
  - data.view_private
  - data.view_admin

// ✅ 推荐：使用角色组合
roles:
  - public_viewer (data.view_public)
  - private_viewer (data.view_private)
  - admin (all permissions)
```

### 3. 最小权限原则

```go
// ✅ 好的实践
router.POST("/api/delete",
    middleware.RequireAllPermissions("data.delete", "data.verify"),
    handler)

// ❌ 不好的实践
router.POST("/api/delete",
    middleware.RequirePermission("admin"),  // 权限过大
    handler)
```

### 4. 权限检查顺序

```go
// ✅ 最优顺序
router.Use(middleware.AuthMiddleware())
router.Use(middleware.LoadUserPermissions(rbacManager))

router.POST("/api/admin/users",
    middleware.RequireRole("admin"),              // 快速失败
    middleware.RequirePermission("user.create"),  // 细粒度控制
    handler)
```

### 5. 定期审计权限

```bash
# 查看哪些用户有特定权限
SELECT DISTINCT u.id, u.username
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN role_permissions rp ON ur.role_id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE p.name = 'user.delete'
  AND (ur.expire_at IS NULL OR ur.expire_at > NOW());

# 查看权限修改历史
SELECT * FROM permission_audit_log
ORDER BY created_at DESC
LIMIT 100;
```

## 常见问题

### Q: 权限检查需要多长时间？
A: 通常 <1ms，因为权限在每次请求时都会加载到内存中。首次加载时可能需要 10-50ms，之后从缓存读取。

### Q: 如何处理权限变更的实时性？
A: 
1. 短期内 (30分钟内) 依赖缓存
2. 关键操作后立即刷新缓存
3. 可以设置更短的 TTL (如 5 分钟) 提高实时性

### Q: 角色继承支持多少层？
A: 理论上无限层，但建议不超过 3 层以保持性能。

### Q: 如何支持动态权限？
A: 
1. 在 permissions 表中添加记录
2. 赋予角色该权限
3. 清除相关用户的缓存
4. 新请求时会加载新权限

### Q: 权限能否根据条件动态判断？
A: 需要在处理器中额外实现条件检查，例如：

```go
// 权限检查
middleware.RequirePermission("data.view")(c)

// 处理器中的条件检查
func getDataHandler(c *gin.Context) {
    userID := c.MustGet("user_id").(int)
    dataID := c.Param("data_id")
    
    // 检查用户是否拥有该数据
    data, err := repo.GetData(dataID)
    if err != nil || data.OwnerID != userID {
        c.JSON(403, gin.H{"error": "forbidden"})
        return
    }
    
    // 继续处理...
}
```

## 参考文档

- [认证系统](MULTI_AUTH_IMPLEMENTATION.md)
- [Token 多级缓存](TOKEN_CACHE_IMPLEMENTATION.md)
- [Token 状态管理](../model/token.go)

