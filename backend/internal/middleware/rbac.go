package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/cache"
	"github.com/oblivious/backend/internal/model"
)

// RBACManager RBAC 管理器
type RBACManager struct {
	permissionCache cache.Cache // 权限缓存
	ttl             time.Duration
}

// NewRBACManager 创建新的 RBAC 管理器
func NewRBACManager(ttl time.Duration) *RBACManager {
	return &RBACManager{
		ttl: ttl,
	}
}

// SetPermissionCache 设置权限缓存
func (rm *RBACManager) SetPermissionCache(c cache.Cache) {
	rm.permissionCache = c
}

// GetUserPermissions 获取用户的所有权限
// 返回: (权限集合, 错误)
func (rm *RBACManager) GetUserPermissions(c *gin.Context, userID int) (*model.UserPermissions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user_permissions:%d", userID)
	if rm.permissionCache != nil {
		if cached, err := rm.permissionCache.Get(ctx, cacheKey); err == nil {
			return cached.(*model.UserPermissions), nil
		}
	}

	// 从数据库查询权限（这部分由 repository 层实现）
	// 这里返回一个占位符实现
	userPerms := &model.UserPermissions{
		UserID:      userID,
		Roles:       []string{},
		Permissions: []model.PermissionDTO{},
		CachedAt:    time.Now(),
		ExpireAt:    time.Now().Add(rm.ttl),
	}

	// 缓存权限
	if rm.permissionCache != nil {
		_ = rm.permissionCache.Set(ctx, cacheKey, userPerms, rm.ttl)
	}

	return userPerms, nil
}

// RequirePermission 验证用户是否拥有特定权限的中间件
// 使用方式: router.POST("/api/users", middleware.RequirePermission("user.create"), handler)
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _, _, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		// 从上下文获取用户权限
		userPerms, ok := c.Get("user_permissions")
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: user permissions not found",
			})
			c.Abort()
			return
		}

		userPermSet, ok := userPerms.(*model.UserPermissions)
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: invalid permission data",
			})
			c.Abort()
			return
		}

		// 检查是否拥有该权限
		hasPermission := false
		for _, perm := range userPermSet.Permissions {
			if perm.Name == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(403, gin.H{
				"error": fmt.Sprintf("Forbidden: permission '%s' required", permission),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermissions 验证用户是否拥有多个权限中的任意一个的中间件
// 使用方式: router.POST("/api/data", middleware.RequirePermissions("data.create", "data.admin"), handler)
func RequirePermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _, _, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		userPerms, ok := c.Get("user_permissions")
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: user permissions not found",
			})
			c.Abort()
			return
		}

		userPermSet, ok := userPerms.(*model.UserPermissions)
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: invalid permission data",
			})
			c.Abort()
			return
		}

		// 检查是否拥有任意一个权限
		permissionSet := make(map[string]bool)
		for _, perm := range permissions {
			permissionSet[perm] = false
		}

		for _, perm := range userPermSet.Permissions {
			if _, exists := permissionSet[perm.Name]; exists {
				permissionSet[perm.Name] = true
			}
		}

		hasAny := false
		for _, has := range permissionSet {
			if has {
				hasAny = true
				break
			}
		}

		if !hasAny {
			c.JSON(403, gin.H{
				"error": fmt.Sprintf("Forbidden: one of %v required", permissions),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions 验证用户是否拥有所有指定权限的中间件
// 使用方式: router.DELETE("/api/data/:id", middleware.RequireAllPermissions("data.delete", "data.verify"), handler)
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _, _, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		userPerms, ok := c.Get("user_permissions")
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: user permissions not found",
			})
			c.Abort()
			return
		}

		userPermSet, ok := userPerms.(*model.UserPermissions)
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: invalid permission data",
			})
			c.Abort()
			return
		}

		// 检查是否拥有所有权限
		permissionMap := make(map[string]bool)
		for _, perm := range permissions {
			permissionMap[perm] = false
		}

		for _, perm := range userPermSet.Permissions {
			if _, exists := permissionMap[perm.Name]; exists {
				permissionMap[perm.Name] = true
			}
		}

		for _, has := range permissionMap {
			if !has {
				c.JSON(403, gin.H{
					"error": fmt.Sprintf("Forbidden: all of %v required", permissions),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRole 验证用户是否拥有特定角色的中间件
// 使用方式: router.GET("/api/admin/settings", middleware.RequireRole("admin"), handler)
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _, _, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		userPerms, ok := c.Get("user_permissions")
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: user permissions not found",
			})
			c.Abort()
			return
		}

		userPermSet, ok := userPerms.(*model.UserPermissions)
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: invalid permission data",
			})
			c.Abort()
			return
		}

		// 检查是否拥有该角色
		hasRole := false
		for _, r := range userPermSet.Roles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(403, gin.H{
				"error": fmt.Sprintf("Forbidden: role '%s' required", role),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRoles 验证用户是否拥有多个角色中的任意一个的中间件
// 使用方式: router.POST("/api/audit", middleware.RequireRoles("admin", "auditor"), handler)
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _, _, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		userPerms, ok := c.Get("user_permissions")
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: user permissions not found",
			})
			c.Abort()
			return
		}

		userPermSet, ok := userPerms.(*model.UserPermissions)
		if !ok {
			c.JSON(403, gin.H{
				"error": "Forbidden: invalid permission data",
			})
			c.Abort()
			return
		}

		// 检查是否拥有任意一个角色
		roleSet := make(map[string]bool)
		for _, r := range roles {
			roleSet[r] = false
		}

		for _, r := range userPermSet.Roles {
			if _, exists := roleSet[r]; exists {
				roleSet[r] = true
			}
		}

		hasAny := false
		for _, has := range roleSet {
			if has {
				hasAny = true
				break
			}
		}

		if !hasAny {
			c.JSON(403, gin.H{
				"error": fmt.Sprintf("Forbidden: one of %v required", roles),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoadUserPermissions 加载用户权限到上下文的中间件
// 应该在认证中间件之后调用
func LoadUserPermissions(rbacManager *RBACManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _, _, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			// 如果没有经过认证，设置为空权限
			c.Set("user_permissions", &model.UserPermissions{
				UserID:      0,
				Roles:       []string{},
				Permissions: []model.PermissionDTO{},
				CachedAt:    time.Now(),
				ExpireAt:    time.Now(),
			})
			c.Next()
			return
		}

		// 获取用户权限
		userPerms, err := rbacManager.GetUserPermissions(c, userID)
		if err != nil {
			// 错误处理：设置为空权限，让后续的权限检查中间件处理拒绝
			userPerms = &model.UserPermissions{
				UserID:      userID,
				Roles:       []string{},
				Permissions: []model.PermissionDTO{},
				CachedAt:    time.Now(),
				ExpireAt:    time.Now(),
			}
		}

		c.Set("user_permissions", userPerms)
		c.Next()
	}
}

// CheckResourceAccess 检查用户是否可以访问特定资源
// 返回: (是否有权限, 原因)
func CheckResourceAccess(userPerms *model.UserPermissions, resource string, action string) (bool, string) {
	if userPerms == nil {
		return false, "user permissions not loaded"
	}

	for _, perm := range userPerms.Permissions {
		if perm.Resource == resource && perm.Action == action {
			return true, ""
		}
	}

	return false, fmt.Sprintf("permission %s:%s required", resource, action)
}

// GetUserRoleNames 从上下文获取用户的角色名称
func GetUserRoleNames(c *gin.Context) []string {
	userPerms, ok := c.Get("user_permissions")
	if !ok {
		return []string{}
	}

	userPermSet, ok := userPerms.(*model.UserPermissions)
	if !ok {
		return []string{}
	}

	return userPermSet.Roles
}

// GetUserPermissionNames 从上下文获取用户的权限名称
func GetUserPermissionNames(c *gin.Context) []string {
	userPerms, ok := c.Get("user_permissions")
	if !ok {
		return []string{}
	}

	userPermSet, ok := userPerms.(*model.UserPermissions)
	if !ok {
		return []string{}
	}

	result := make([]string, len(userPermSet.Permissions))
	for i, perm := range userPermSet.Permissions {
		result[i] = perm.Name
	}

	return result
}

