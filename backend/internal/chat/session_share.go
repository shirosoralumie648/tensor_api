package chat

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// SharePermission 分享权限
type SharePermission string

const (
	PermissionView     SharePermission = "view"      // 仅查看
	PermissionComment  SharePermission = "comment"   // 查看+评论
	PermissionEdit     SharePermission = "edit"      // 编辑
	PermissionAdmin    SharePermission = "admin"     // 管理员
)

// ShareLink 分享链接
type ShareLink struct {
	// 链接 ID
	ID string `json:"id"`

	// 链接码
	LinkCode string `json:"link_code"`

	// 会话 ID
	SessionID string `json:"session_id"`

	// 分享者 ID
	SharerID int64 `json:"sharer_id"`

	// 分享名称
	Name string `json:"name"`

	// 分享描述
	Description string `json:"description"`

	// 权限级别
	Permission SharePermission `json:"permission"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 过期时间（nil 表示永不过期）
	ExpiresAt *time.Time `json:"expires_at"`

	// 访问次数
	AccessCount int64 `json:"access_count"`

	// 启用状态
	Enabled bool `json:"enabled"`

	// 互斥锁
	mu sync.RWMutex
}

// IsExpired 检查是否过期
func (sl *ShareLink) IsExpired() bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if sl.ExpiresAt == nil {
		return false
	}

	return time.Now().After(*sl.ExpiresAt)
}

// IsAccessible 检查是否可访问
func (sl *ShareLink) IsAccessible() bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	return sl.Enabled && !sl.IsExpired()
}

// IncrementAccessCount 增加访问次数
func (sl *ShareLink) IncrementAccessCount() {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.AccessCount++
}

// GetAccessCount 获取访问次数
func (sl *ShareLink) GetAccessCount() int64 {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	return sl.AccessCount
}

// Disable 禁用分享链接
func (sl *ShareLink) Disable() {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.Enabled = false
}

// ShareAccess 分享访问记录
type ShareAccess struct {
	// 访问 ID
	ID string `json:"id"`

	// 分享链接 ID
	LinkID string `json:"link_id"`

	// 访问者 IP
	VisitorIP string `json:"visitor_ip"`

	// 访问者代理
	UserAgent string `json:"user_agent"`

	// 访问时间
	AccessedAt time.Time `json:"accessed_at"`

	// 访问者用户 ID（可选，若已登录）
	UserID *int64 `json:"user_id"`
}

// ShareManager 分享管理器
type ShareManager struct {
	// 分享链接存储
	links map[string]*ShareLink
	linksMu sync.RWMutex

	// 会话分享索引
	sessionShares map[string][]string // sessionID -> linkIDs
	sessionSharesMu sync.RWMutex

	// 分享访问记录
	accesses map[string][]*ShareAccess // linkID -> accesses
	accessesMu sync.RWMutex

	// 统计信息
	totalShares int64
	totalAccess int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewShareManager 创建分享管理器
func NewShareManager() *ShareManager {
	return &ShareManager{
		links:           make(map[string]*ShareLink),
		sessionShares:   make(map[string][]string),
		accesses:        make(map[string][]*ShareAccess),
		logFunc:         defaultLogFunc,
	}
}

// generateLinkCode 生成分享码
func (sm *ShareManager) generateLinkCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// CreateShareLink 创建分享链接
func (sm *ShareManager) CreateShareLink(sessionID string, sharerID int64, name, description string, permission SharePermission, expiresIn *time.Duration) (*ShareLink, error) {
	linkCode, err := sm.generateLinkCode()
	if err != nil {
		return nil, err
	}

	sm.linksMu.Lock()
	defer sm.linksMu.Unlock()

	linkID := fmt.Sprintf("share-%d", time.Now().UnixNano())

	var expiresAt *time.Time
	if expiresIn != nil {
		expTime := time.Now().Add(*expiresIn)
		expiresAt = &expTime
	}

	link := &ShareLink{
		ID:          linkID,
		LinkCode:    linkCode,
		SessionID:   sessionID,
		SharerID:    sharerID,
		Name:        name,
		Description: description,
		Permission:  permission,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Enabled:     true,
	}

	sm.links[linkID] = link

	// 更新会话分享索引
	sm.sessionSharesMu.Lock()
	sm.sessionShares[sessionID] = append(sm.sessionShares[sessionID], linkID)
	sm.sessionSharesMu.Unlock()

	atomic.AddInt64(&sm.totalShares, 1)

	sm.logFunc("info", fmt.Sprintf("Created share link %s for session %s", linkID, sessionID))

	return link, nil
}

// GetShareLink 获取分享链接
func (sm *ShareManager) GetShareLink(linkID string) (*ShareLink, error) {
	sm.linksMu.RLock()
	defer sm.linksMu.RUnlock()

	link, exists := sm.links[linkID]
	if !exists {
		return nil, fmt.Errorf("share link %s not found", linkID)
	}

	return link, nil
}

// GetShareLinkByCode 按分享码获取分享链接
func (sm *ShareManager) GetShareLinkByCode(linkCode string) (*ShareLink, error) {
	sm.linksMu.RLock()
	defer sm.linksMu.RUnlock()

	for _, link := range sm.links {
		if link.LinkCode == linkCode {
			return link, nil
		}
	}

	return nil, fmt.Errorf("share link with code %s not found", linkCode)
}

// AccessShareLink 访问分享链接
func (sm *ShareManager) AccessShareLink(linkID, visitorIP, userAgent string, userID *int64) (*ShareAccess, error) {
	link, err := sm.GetShareLink(linkID)
	if err != nil {
		return nil, err
	}

	if !link.IsAccessible() {
		return nil, fmt.Errorf("share link is not accessible")
	}

	link.IncrementAccessCount()

	sm.accessesMu.Lock()
	defer sm.accessesMu.Unlock()

	accessID := fmt.Sprintf("access-%d", time.Now().UnixNano())

	access := &ShareAccess{
		ID:         accessID,
		LinkID:     linkID,
		VisitorIP:  visitorIP,
		UserAgent:  userAgent,
		AccessedAt: time.Now(),
		UserID:     userID,
	}

	sm.accesses[linkID] = append(sm.accesses[linkID], access)

	atomic.AddInt64(&sm.totalAccess, 1)

	sm.logFunc("debug", fmt.Sprintf("Accessed share link %s from %s", linkID, visitorIP))

	return access, nil
}

// GetSessionShares 获取会话的所有分享链接
func (sm *ShareManager) GetSessionShares(sessionID string) []*ShareLink {
	sm.sessionSharesMu.RLock()
	linkIDs := sm.sessionShares[sessionID]
	sm.sessionSharesMu.RUnlock()

	var links []*ShareLink

	sm.linksMu.RLock()
	defer sm.linksMu.RUnlock()

	for _, id := range linkIDs {
		if link, exists := sm.links[id]; exists {
			links = append(links, link)
		}
	}

	return links
}

// DisableShareLink 禁用分享链接
func (sm *ShareManager) DisableShareLink(linkID string) error {
	link, err := sm.GetShareLink(linkID)
	if err != nil {
		return err
	}

	link.Disable()

	sm.logFunc("info", fmt.Sprintf("Disabled share link %s", linkID))

	return nil
}

// DeleteShareLink 删除分享链接
func (sm *ShareManager) DeleteShareLink(linkID string) error {
	sm.linksMu.Lock()
	defer sm.linksMu.Unlock()

	link, exists := sm.links[linkID]
	if !exists {
		return fmt.Errorf("share link %s not found", linkID)
	}

	// 从会话索引中移除
	sm.sessionSharesMu.Lock()
	sessionID := link.SessionID
	shares := sm.sessionShares[sessionID]
	for i, id := range shares {
		if id == linkID {
			sm.sessionShares[sessionID] = append(shares[:i], shares[i+1:]...)
			break
		}
	}
	sm.sessionSharesMu.Unlock()

	// 删除访问记录
	sm.accessesMu.Lock()
	delete(sm.accesses, linkID)
	sm.accessesMu.Unlock()

	delete(sm.links, linkID)

	sm.logFunc("info", fmt.Sprintf("Deleted share link %s", linkID))

	return nil
}

// GetShareAccesses 获取分享链接的访问记录
func (sm *ShareManager) GetShareAccesses(linkID string) []*ShareAccess {
	sm.accessesMu.RLock()
	defer sm.accessesMu.RUnlock()

	accesses, exists := sm.accesses[linkID]
	if !exists {
		return make([]*ShareAccess, 0)
	}

	result := make([]*ShareAccess, len(accesses))
	copy(result, accesses)

	return result
}

// GetStatistics 获取统计信息
func (sm *ShareManager) GetStatistics() map[string]interface{} {
	sm.linksMu.RLock()
	activeShares := 0
	disabledShares := 0

	for _, link := range sm.links {
		if link.Enabled {
			activeShares++
		} else {
			disabledShares++
		}
	}
	sm.linksMu.RUnlock()

	return map[string]interface{}{
		"total_shares":     atomic.LoadInt64(&sm.totalShares),
		"total_access":     atomic.LoadInt64(&sm.totalAccess),
		"active_shares":    activeShares,
		"disabled_shares":  disabledShares,
	}
}

// PurgeExpiredShares 清理已过期的分享链接（可选）
func (sm *ShareManager) PurgeExpiredShares() int {
	sm.linksMu.Lock()
	defer sm.linksMu.Unlock()

	count := 0

	for linkID, link := range sm.links {
		if link.IsExpired() {
			// 从会话索引中移除
			sm.sessionSharesMu.Lock()
			sessionID := link.SessionID
			shares := sm.sessionShares[sessionID]
			for i, id := range shares {
				if id == linkID {
					sm.sessionShares[sessionID] = append(shares[:i], shares[i+1:]...)
					break
				}
			}
			sm.sessionSharesMu.Unlock()

			// 删除访问记录
			sm.accessesMu.Lock()
			delete(sm.accesses, linkID)
			sm.accessesMu.Unlock()

			delete(sm.links, linkID)
			count++
		}
	}

	sm.logFunc("info", fmt.Sprintf("Purged %d expired share links", count))

	return count
}

// PermissionValidator 权限验证器
type PermissionValidator struct {
	// 权限等级映射
	levels map[SharePermission]int
}

// NewPermissionValidator 创建权限验证器
func NewPermissionValidator() *PermissionValidator {
	return &PermissionValidator{
		levels: map[SharePermission]int{
			PermissionView:    1,
			PermissionComment: 2,
			PermissionEdit:    3,
			PermissionAdmin:   4,
		},
	}
}

// CanPerform 检查是否可以执行操作
func (pv *PermissionValidator) CanPerform(permission SharePermission, action string) bool {
	level, exists := pv.levels[permission]
	if !exists {
		return false
	}

	switch action {
	case "view":
		return level >= 1
	case "comment":
		return level >= 2
	case "edit":
		return level >= 3
	case "admin":
		return level >= 4
	default:
		return false
	}
}

// GetRequiredPermission 获取操作所需的权限
func (pv *PermissionValidator) GetRequiredPermission(action string) SharePermission {
	switch action {
	case "view":
		return PermissionView
	case "comment":
		return PermissionComment
	case "edit":
		return PermissionEdit
	case "admin":
		return PermissionAdmin
	default:
		return PermissionView
	}
}

