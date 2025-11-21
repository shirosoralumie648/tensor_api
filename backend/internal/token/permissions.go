package token

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Permission Token 权限
type Permission struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

// TokenPermissions Token 权限配置
type TokenPermissions struct {
	ID               string        `json:"id"`
	TokenID          string        `json:"token_id"`
	ModelWhitelist   []string      `json:"model_whitelist"`   // 模型白名单
	ModelBlacklist   []string      `json:"model_blacklist"`   // 模型黑名单
	IPWhitelist      []string      `json:"ip_whitelist"`      // IP 白名单
	IPBlacklist      []string      `json:"ip_blacklist"`      // IP 黑名单
	Permissions      []*Permission `json:"permissions"`       // 操作权限
	RateLimit        int64         `json:"rate_limit"`        // 请求数/分钟
	DailyQuota       int64         `json:"daily_quota"`       // 每日额度
	MonthlyQuota     int64         `json:"monthly_quota"`     // 每月额度
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// PermissionManager Token 权限管理器
type PermissionManager struct {
	mu         sync.RWMutex
	permissions map[string]*TokenPermissions
	store      PermissionStore
}

// PermissionStore 权限存储接口
type PermissionStore interface {
	Create(permissions *TokenPermissions) error
	Update(permissions *TokenPermissions) error
	Get(id string) (*TokenPermissions, error)
	GetByToken(tokenID string) (*TokenPermissions, error)
	Delete(id string) error
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager(store PermissionStore) *PermissionManager {
	return &PermissionManager{
		permissions: make(map[string]*TokenPermissions),
		store:       store,
	}
}

// CreatePermissions 创建 Token 权限
func (pm *PermissionManager) CreatePermissions(tokenID string) (*TokenPermissions, error) {
	permissions := &TokenPermissions{
		ID:           fmt.Sprintf("perm_%d", time.Now().UnixNano()),
		TokenID:      tokenID,
		ModelWhitelist: []string{},
		ModelBlacklist: []string{},
		IPWhitelist:    []string{},
		IPBlacklist:    []string{},
		Permissions:    []*Permission{},
		RateLimit:      1000, // 默认 1000 请求/分钟
		DailyQuota:     1000000,
		MonthlyQuota:   30000000,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := pm.store.Create(permissions); err != nil {
		return nil, err
	}

	pm.mu.Lock()
	pm.permissions[tokenID] = permissions
	pm.mu.Unlock()

	return permissions, nil
}

// GetPermissions 获取 Token 权限
func (pm *PermissionManager) GetPermissions(tokenID string) (*TokenPermissions, error) {
	pm.mu.RLock()
	perms, exists := pm.permissions[tokenID]
	pm.mu.RUnlock()

	if exists {
		return perms, nil
	}

	perms, err := pm.store.GetByToken(tokenID)
	if err != nil {
		return nil, err
	}

	pm.mu.Lock()
	pm.permissions[tokenID] = perms
	pm.mu.Unlock()

	return perms, nil
}

// SetModelWhitelist 设置模型白名单
func (pm *PermissionManager) SetModelWhitelist(tokenID string, models []string) error {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return err
	}

	perms.ModelWhitelist = models
	perms.UpdatedAt = time.Now()

	pm.mu.Lock()
	pm.permissions[tokenID] = perms
	pm.mu.Unlock()

	return pm.store.Update(perms)
}

// SetModelBlacklist 设置模型黑名单
func (pm *PermissionManager) SetModelBlacklist(tokenID string, models []string) error {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return err
	}

	perms.ModelBlacklist = models
	perms.UpdatedAt = time.Now()

	pm.mu.Lock()
	pm.permissions[tokenID] = perms
	pm.mu.Unlock()

	return pm.store.Update(perms)
}

// SetIPWhitelist 设置 IP 白名单
func (pm *PermissionManager) SetIPWhitelist(tokenID string, ips []string) error {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return err
	}

	perms.IPWhitelist = ips
	perms.UpdatedAt = time.Now()

	pm.mu.Lock()
	pm.permissions[tokenID] = perms
	pm.mu.Unlock()

	return pm.store.Update(perms)
}

// SetIPBlacklist 设置 IP 黑名单
func (pm *PermissionManager) SetIPBlacklist(tokenID string, ips []string) error {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return err
	}

	perms.IPBlacklist = ips
	perms.UpdatedAt = time.Now()

	pm.mu.Lock()
	pm.permissions[tokenID] = perms
	pm.mu.Unlock()

	return pm.store.Update(perms)
}

// CheckModelPermission 检查模型权限
func (pm *PermissionManager) CheckModelPermission(tokenID, model string) (bool, error) {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return false, err
	}

	// 检查黑名单
	for _, m := range perms.ModelBlacklist {
		if m == model {
			return false, nil
		}
	}

	// 检查白名单
	if len(perms.ModelWhitelist) > 0 {
		for _, m := range perms.ModelWhitelist {
			if m == model {
				return true, nil
			}
		}
		return false, nil
	}

	return true, nil
}

// CheckIPPermission 检查 IP 权限
func (pm *PermissionManager) CheckIPPermission(tokenID, clientIP string) (bool, error) {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return false, err
	}

	// 解析 IP
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false, fmt.Errorf("invalid ip address: %s", clientIP)
	}

	// 检查黑名单
	if pm.ipInList(ip, perms.IPBlacklist) {
		return false, nil
	}

	// 检查白名单
	if len(perms.IPWhitelist) > 0 {
		return pm.ipInList(ip, perms.IPWhitelist), nil
	}

	return true, nil
}

// CheckPermission 检查操作权限
func (pm *PermissionManager) CheckPermission(tokenID, resource, action string) (bool, error) {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return false, err
	}

	for _, perm := range perms.Permissions {
		if perm.Resource == resource {
			for _, act := range perm.Actions {
				if act == action {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// SetRateLimit 设置限流
func (pm *PermissionManager) SetRateLimit(tokenID string, requestsPerMinute int64) error {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return err
	}

	perms.RateLimit = requestsPerMinute
	perms.UpdatedAt = time.Now()

	pm.mu.Lock()
	pm.permissions[tokenID] = perms
	pm.mu.Unlock()

	return pm.store.Update(perms)
}

// SetQuota 设置配额
func (pm *PermissionManager) SetQuota(tokenID string, dailyQuota, monthlyQuota int64) error {
	perms, err := pm.GetPermissions(tokenID)
	if err != nil {
		return err
	}

	perms.DailyQuota = dailyQuota
	perms.MonthlyQuota = monthlyQuota
	perms.UpdatedAt = time.Now()

	pm.mu.Lock()
	pm.permissions[tokenID] = perms
	pm.mu.Unlock()

	return pm.store.Update(perms)
}

// ipInList 检查 IP 是否在列表中
func (pm *PermissionManager) ipInList(ip net.IP, list []string) bool {
	for _, item := range list {
		if strings.Contains(item, "/") {
			// CIDR 表示法
			_, network, err := net.ParseCIDR(item)
			if err != nil {
				continue
			}
			if network.Contains(ip) {
				return true
			}
		} else {
			// 单个 IP
			if ip.String() == item {
				return true
			}
		}
	}
	return false
}


