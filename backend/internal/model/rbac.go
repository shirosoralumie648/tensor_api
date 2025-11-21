package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          int            `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:50" json:"name"`
	DisplayName string         `gorm:"size:100" json:"display_name"`
	Description string         `json:"description"`
	Level       int            `json:"level" default:"100"` // 权限等级，数值越小权限越大
	Permissions datatypes.JSON `gorm:"type:jsonb" json:"permissions"`
	IsBuiltin   bool           `json:"is_builtin" default:"false"`

	// 关联
	UserRoles       []UserRole       `gorm:"foreignKey:RoleID" json:"user_roles,omitempty"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID" json:"role_permissions,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex:idx_resource_action;size:100" json:"name"`
	DisplayName string    `gorm:"size:200" json:"display_name"`
	Description string    `json:"description"`
	Resource    string    `gorm:"uniqueIndex:idx_resource_action;size:100" json:"resource"`
	Action      string    `gorm:"uniqueIndex:idx_resource_action;size:100" json:"action"`
	Level       int       `json:"level" default:"100"`
	IsSystem    bool      `json:"is_system" default:"false"`

	// 关联
	RolePermissions []RolePermission `gorm:"foreignKey:PermissionID" json:"role_permissions,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// UserRole 用户-角色关联模型
type UserRole struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"uniqueIndex:idx_user_role;index" json:"user_id"`
	RoleID    int       `gorm:"uniqueIndex:idx_user_role;index" json:"role_id"`
	AssignedAt time.Time `json:"assigned_at"`
	AssignedBy *int      `json:"assigned_by"`
	ExpireAt  *time.Time `json:"expire_at"` // 角色过期时间

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`

	Metadata datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
}

// RolePermission 角色-权限关联模型
type RolePermission struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	RoleID       int       `gorm:"uniqueIndex:idx_role_permission;index" json:"role_id"`
	PermissionID int       `gorm:"uniqueIndex:idx_role_permission;index" json:"permission_id"`

	// 关联
	Role       *Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// RoleHierarchy 角色继承关系模型
type RoleHierarchy struct {
	ID           int   `gorm:"primaryKey" json:"id"`
	ParentRoleID int   `gorm:"uniqueIndex:idx_hierarchy;index" json:"parent_role_id"`
	ChildRoleID  int   `gorm:"uniqueIndex:idx_hierarchy;index" json:"child_role_id"`

	// 关联
	ParentRole *Role `gorm:"foreignKey:ParentRoleID" json:"parent_role,omitempty"`
	ChildRole  *Role `gorm:"foreignKey:ChildRoleID" json:"child_role,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// PermissionAuditLog 权限审计日志模型
type PermissionAuditLog struct {
	ID            int64          `gorm:"primaryKey" json:"id"`
	UserID        int            `gorm:"index" json:"user_id"`
	Operation     string         `json:"operation"` // assign_role, revoke_role, grant_permission, revoke_permission
	TargetUserID  *int           `json:"target_user_id"`
	RoleID        *int           `json:"role_id"`
	PermissionID  *int           `json:"permission_id"`
	Details       datatypes.JSON `gorm:"type:jsonb" json:"details"`
	IPAddress     string         `json:"ip_address"`
	UserAgent     string         `json:"user_agent"`

	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (RoleHierarchy) TableName() string {
	return "role_hierarchy"
}

func (PermissionAuditLog) TableName() string {
	return "permission_audit_log"
}

// RoleDTO 角色数据传输对象
type RoleDTO struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	Level       int            `json:"level"`
	Permissions []PermissionDTO `json:"permissions,omitempty"`
	IsBuiltin   bool           `json:"is_builtin"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// PermissionDTO 权限数据传输对象
type PermissionDTO struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Level       int       `json:"level"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRoleDTO 用户角色数据传输对象
type UserRoleDTO struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Role      RoleDTO    `json:"role"`
	AssignedAt time.Time `json:"assigned_at"`
	ExpireAt  *time.Time `json:"expire_at"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description"`
	Level       *int   `json:"level"`
	Permissions []int  `json:"permissions"` // permission IDs
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
	Permissions *[]int  `json:"permissions"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID   int        `json:"user_id" binding:"required"`
	RoleID   int        `json:"role_id" binding:"required"`
	ExpireAt *time.Time `json:"expire_at"`
}

// RevokeRoleRequest 撤销角色请求
type RevokeRoleRequest struct {
	UserID int `json:"user_id" binding:"required"`
	RoleID int `json:"role_id" binding:"required"`
}

// PermissionCheck 权限检查结果
type PermissionCheck struct {
	HasPermission bool   `json:"has_permission"`
	Reason        string `json:"reason,omitempty"`
	AllowedResources []string `json:"allowed_resources,omitempty"`
}

// UserPermissions 用户权限集合
type UserPermissions struct {
	UserID      int             `json:"user_id"`
	Roles       []string        `json:"roles"`
	Permissions []PermissionDTO `json:"permissions"`
	CachedAt    time.Time       `json:"cached_at"`
	ExpireAt    time.Time       `json:"expire_at"`
}

// Scan 实现 sql.Scanner 接口
func (ur *UserRole) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	return json.Unmarshal(bytes, &ur)
}

// Value 实现 driver.Valuer 接口
func (ur UserRole) Value() (driver.Value, error) {
	return json.Marshal(ur)
}

// ToDTO 将 Role 转换为 RoleDTO
func (r *Role) ToDTO() RoleDTO {
	return RoleDTO{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		Level:       r.Level,
		IsBuiltin:   r.IsBuiltin,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// ToDTO 将 Permission 转换为 PermissionDTO
func (p *Permission) ToDTO() PermissionDTO {
	return PermissionDTO{
		ID:          p.ID,
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Description: p.Description,
		Resource:    p.Resource,
		Action:      p.Action,
		Level:       p.Level,
		IsSystem:    p.IsSystem,
		CreatedAt:   p.CreatedAt,
	}
}

// ToDTO 将 UserRole 转换为 UserRoleDTO
func (ur *UserRole) ToDTO() UserRoleDTO {
	roleDTO := RoleDTO{}
	if ur.Role != nil {
		roleDTO = ur.Role.ToDTO()
	}

	return UserRoleDTO{
		ID:        ur.ID,
		UserID:    ur.UserID,
		Role:      roleDTO,
		AssignedAt: ur.AssignedAt,
		ExpireAt:  ur.ExpireAt,
	}
}

// IsActive 检查角色是否仍然活跃（未过期）
func (ur *UserRole) IsActive() bool {
	if ur.ExpireAt == nil {
		return true
	}
	return ur.ExpireAt.After(time.Now())
}

// IsExpired 检查角色是否已过期
func (ur *UserRole) IsExpired() bool {
	if ur.ExpireAt == nil {
		return false
	}
	return ur.ExpireAt.Before(time.Now())
}

// SetExpireAt 设置过期时间
func (ur *UserRole) SetExpireAt(duration time.Duration) {
	expireAt := time.Now().Add(duration)
	ur.ExpireAt = &expireAt
}

// CanCreate 检查是否可以创建角色
func (r *Role) CanCreate(targetLevel int) bool {
	// 角色等级必须小于目标等级（权限更大）
	return r.Level < targetLevel
}

// CanManage 检查是否可以管理目标角色
func (r *Role) CanManage(targetRole *Role) bool {
	// 只有权限大于等于目标角色的角色才能管理
	return r.Level <= targetRole.Level
}

// CanAssign 检查是否可以分配角色
func (r *Role) CanAssign(targetRole *Role) bool {
	// 只能分配权限小于等于自己的角色
	return r.Level <= targetRole.Level
}

// Validate 验证角色
func (r *Role) Validate() error {
	if r.Name == "" {
		return gorm.ErrInvalidData
	}
	if r.Level < 0 {
		return gorm.ErrInvalidData
	}
	return nil
}

// Validate 验证权限
func (p *Permission) Validate() error {
	if p.Name == "" || p.Resource == "" || p.Action == "" {
		return gorm.ErrInvalidData
	}
	return nil
}

