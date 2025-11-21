package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/oblivious/backend/internal/model"
)

func TestRequirePermission(t *testing.T) {
	t.Run("with_permission", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Permissions: []model.PermissionDTO{
				{Name: "user.create", Resource: "user", Action: "create"},
			},
		})

		handler := RequirePermission("user.create")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.True(t, called)
	})

	t.Run("without_permission", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID:      1,
			Permissions: []model.PermissionDTO{},
		})

		handler := RequirePermission("user.delete")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("not_authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler := RequirePermission("user.create")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.False(t, called)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequirePermissions(t *testing.T) {
	t.Run("with_one_of_permissions", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Permissions: []model.PermissionDTO{
				{Name: "user.create", Resource: "user", Action: "create"},
			},
		})

		handler := RequirePermissions("user.create", "user.delete")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.True(t, called)
	})

	t.Run("without_any_permission", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID:      1,
			Permissions: []model.PermissionDTO{},
		})

		handler := RequirePermissions("user.create", "user.delete")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireAllPermissions(t *testing.T) {
	t.Run("with_all_permissions", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Permissions: []model.PermissionDTO{
				{Name: "user.create", Resource: "user", Action: "create"},
				{Name: "user.delete", Resource: "user", Action: "delete"},
			},
		})

		handler := RequireAllPermissions("user.create", "user.delete")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.True(t, called)
	})

	t.Run("without_all_permissions", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Permissions: []model.PermissionDTO{
				{Name: "user.create", Resource: "user", Action: "create"},
			},
		})

		handler := RequireAllPermissions("user.create", "user.delete")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireRole(t *testing.T) {
	t.Run("with_role", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Roles:  []string{"admin"},
		})

		handler := RequireRole("admin")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.True(t, called)
	})

	t.Run("without_role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Roles:  []string{"user"},
		})

		handler := RequireRole("admin")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireRoles(t *testing.T) {
	t.Run("with_one_of_roles", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Roles:  []string{"developer"},
		})

		handler := RequireRoles("admin", "developer")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.True(t, called)
	})

	t.Run("without_any_role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", 1)
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Roles:  []string{"user"},
		})

		handler := RequireRoles("admin", "developer")
		called := false

		handler(func(ctx *gin.Context) {
			called = true
		})(c)

		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestCheckResourceAccess(t *testing.T) {
	t.Run("has_permission", func(t *testing.T) {
		userPerms := &model.UserPermissions{
			UserID: 1,
			Permissions: []model.PermissionDTO{
				{Name: "user.create", Resource: "user", Action: "create"},
			},
		}

		hasAccess, reason := CheckResourceAccess(userPerms, "user", "create")
		assert.True(t, hasAccess)
		assert.Equal(t, "", reason)
	})

	t.Run("no_permission", func(t *testing.T) {
		userPerms := &model.UserPermissions{
			UserID:      1,
			Permissions: []model.PermissionDTO{},
		}

		hasAccess, reason := CheckResourceAccess(userPerms, "user", "create")
		assert.False(t, hasAccess)
		assert.NotEmpty(t, reason)
	})

	t.Run("nil_permissions", func(t *testing.T) {
		hasAccess, reason := CheckResourceAccess(nil, "user", "create")
		assert.False(t, hasAccess)
		assert.NotEmpty(t, reason)
	})
}

func TestGetUserRoleNames(t *testing.T) {
	t.Run("with_roles", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Roles:  []string{"admin", "developer"},
		})

		roles := GetUserRoleNames(c)
		assert.Equal(t, []string{"admin", "developer"}, roles)
	})

	t.Run("no_permissions_set", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		roles := GetUserRoleNames(c)
		assert.Equal(t, []string{}, roles)
	})
}

func TestGetUserPermissionNames(t *testing.T) {
	t.Run("with_permissions", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_permissions", &model.UserPermissions{
			UserID: 1,
			Permissions: []model.PermissionDTO{
				{Name: "user.create"},
				{Name: "user.delete"},
			},
		})

		perms := GetUserPermissionNames(c)
		assert.Equal(t, []string{"user.create", "user.delete"}, perms)
	})

	t.Run("no_permissions_set", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		perms := GetUserPermissionNames(c)
		assert.Equal(t, []string{}, perms)
	})
}

func TestUserRoleIsActive(t *testing.T) {
	t.Run("active_role", func(t *testing.T) {
		expireAt := time.Now().Add(24 * time.Hour)
		ur := &model.UserRole{
			UserID:   1,
			RoleID:   1,
			ExpireAt: &expireAt,
		}

		assert.True(t, ur.IsActive())
		assert.False(t, ur.IsExpired())
	})

	t.Run("expired_role", func(t *testing.T) {
		expireAt := time.Now().Add(-24 * time.Hour)
		ur := &model.UserRole{
			UserID:   1,
			RoleID:   1,
			ExpireAt: &expireAt,
		}

		assert.False(t, ur.IsActive())
		assert.True(t, ur.IsExpired())
	})

	t.Run("permanent_role", func(t *testing.T) {
		ur := &model.UserRole{
			UserID:   1,
			RoleID:   1,
			ExpireAt: nil,
		}

		assert.True(t, ur.IsActive())
		assert.False(t, ur.IsExpired())
	})
}

func TestRoleCanManage(t *testing.T) {
	adminRole := &model.Role{ID: 1, Name: "admin", Level: 10}
	userRole := &model.Role{ID: 2, Name: "user", Level: 100}

	t.Run("can_manage_lower_level_role", func(t *testing.T) {
		assert.True(t, adminRole.CanManage(userRole))
	})

	t.Run("cannot_manage_higher_level_role", func(t *testing.T) {
		assert.False(t, userRole.CanManage(adminRole))
	})

	t.Run("can_manage_same_level_role", func(t *testing.T) {
		assert.True(t, adminRole.CanManage(adminRole))
	})
}

func BenchmarkRequirePermission(b *testing.B) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", 1)
	c.Set("user_permissions", &model.UserPermissions{
		UserID: 1,
		Permissions: []model.PermissionDTO{
			{Name: "user.create", Resource: "user", Action: "create"},
		},
	})

	handler := RequirePermission("user.create")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		handler(func(ctx *gin.Context) {})(c)
	}
}

func BenchmarkCheckResourceAccess(b *testing.B) {
	userPerms := &model.UserPermissions{
		UserID: 1,
		Permissions: []model.PermissionDTO{
			{Name: "user.create", Resource: "user", Action: "create"},
			{Name: "user.delete", Resource: "user", Action: "delete"},
			{Name: "user.update", Resource: "user", Action: "update"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CheckResourceAccess(userPerms, "user", "create")
	}
}

