package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/internal/repository"
)

// TokenService Token 服务
type TokenService struct {
	tokenRepo repository.TokenRepository
}

// NewTokenService 创建新的 Token 服务
func NewTokenService(tokenRepo repository.TokenRepository) *TokenService {
	return &TokenService{
		tokenRepo: tokenRepo,
	}
}

// CreateToken 创建新的 Token
func (ts *TokenService) CreateToken(
	ctx context.Context,
	userID int,
	name string,
	description string,
	quotaLimit int64,
	expireDays int,
) (*model.Token, error) {
	// 生成 Token 哈希
	tokenHash := ts.generateTokenHash()

	// 设置过期时间
	expireAt := time.Now().AddDate(0, 0, expireDays)

	token := &model.Token{
		UserID:      userID,
		TokenHash:   tokenHash,
		Name:        name,
		Status:      model.TokenStatusNormal,
		QuotaUsed:   0,
		CreatedAt:   time.Now(),
		ExpireAt:    toNullTime(expireAt),
		IPWhitelist: []string{},
		ModelWhitelist: []string{},
		Metadata:    make(map[string]interface{}),
	}

	if description != "" {
		token.Description = toNullString(description)
	}

	if quotaLimit > 0 {
		token.QuotaLimit = toNullInt64(quotaLimit)
	}

	// 保存到数据库
	createdToken, err := ts.tokenRepo.Create(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// 记录审计日志
	_ = ts.logAudit(ctx, userID, createdToken.ID, model.TokenOpCreate, nil, &model.TokenStatusNormal, nil, "", "")

	return createdToken, nil
}

// GetTokenByHash 通过 Hash 获取 Token
func (ts *TokenService) GetTokenByHash(ctx context.Context, tokenHash string) (*model.Token, error) {
	token, err := ts.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// 检查状态并更新过期状态
	if token.Status == model.TokenStatusNormal && token.ExpireAt.Valid && token.ExpireAt.Time.Before(time.Now()) {
		token.Status = model.TokenStatusExpired
		_ = ts.updateTokenStatus(ctx, token, model.TokenStatusExpired)
	}

	return token, nil
}

// ValidateToken 验证 Token
func (ts *TokenService) ValidateToken(
	ctx context.Context,
	tokenHash string,
	ipAddress string,
	model string,
) (*model.Token, error) {
	token, err := ts.GetTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// 检查 Token 是否有效
	if !token.IsValid() {
		return nil, fmt.Errorf("token is not valid, status: %s", token.Status.String())
	}

	// 检查 IP 白名单
	if !token.ValidateIPAddress(ipAddress) {
		return nil, fmt.Errorf("ip address not in whitelist: %s", ipAddress)
	}

	// 检查模型白名单
	if !token.ValidateModel(model) {
		return nil, fmt.Errorf("model not in whitelist: %s", model)
	}

	// 更新最后使用时间
	_ = ts.updateLastUsedAt(ctx, token.ID)

	return token, nil
}

// RenewToken 续期 Token
func (ts *TokenService) RenewToken(ctx context.Context, tokenID int, extendDays int) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// 检查是否可以续期
	if !token.CanRenew() {
		return fmt.Errorf("token cannot be renewed, status: %s", token.Status.String())
	}

	// 记录续期日志
	oldExpireAt := token.ExpireAt
	newExpireAt := token.ExpireAt.Time.AddDate(0, 0, extendDays)

	token.ExpireAt = toNullTime(newExpireAt)
	token.RenewedAt = toNullTime(time.Now())

	// 更新数据库
	_, err = ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	// 记录续期日志
	_ = ts.logRenewal(ctx, tokenID, oldExpireAt, toNullTime(newExpireAt), "manual_renewal")

	// 记录审计日志
	oldStatus := token.Status
	_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpRenew, &oldStatus, &model.TokenStatusNormal, nil, "", "")

	return nil
}

// DisableToken 禁用 Token
func (ts *TokenService) DisableToken(ctx context.Context, tokenID int, reason string) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	oldStatus := token.Status
	newStatus := model.TokenStatusDisabled

	token.Status = newStatus

	// 更新数据库
	_, err = ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to disable token: %w", err)
	}

	// 记录审计日志
	details := map[string]interface{}{"reason": reason}
	_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpDisable, &oldStatus, &newStatus, details, "", "")

	return nil
}

// EnableToken 启用 Token
func (ts *TokenService) EnableToken(ctx context.Context, tokenID int) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if token.Status != model.TokenStatusDisabled {
		return fmt.Errorf("token is not disabled")
	}

	oldStatus := token.Status
	newStatus := model.TokenStatusNormal

	token.Status = newStatus

	// 更新数据库
	_, err = ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to enable token: %w", err)
	}

	// 记录审计日志
	_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpEnable, &oldStatus, &newStatus, nil, "", "")

	return nil
}

// SoftDeleteToken 软删除 Token
func (ts *TokenService) SoftDeleteToken(ctx context.Context, tokenID int) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	oldStatus := token.Status
	newStatus := model.TokenStatusDeleted

	token.Status = newStatus
	token.DeletedAt = toNullTime(time.Now())

	// 更新数据库
	_, err = ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	// 记录审计日志
	_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpDelete, &oldStatus, &newStatus, nil, "", "")

	return nil
}

// RestoreToken 恢复软删除的 Token
func (ts *TokenService) RestoreToken(ctx context.Context, tokenID int) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if token.Status != model.TokenStatusDeleted {
		return fmt.Errorf("token is not deleted")
	}

	oldStatus := token.Status
	newStatus := model.TokenStatusNormal

	token.Status = newStatus
	token.DeletedAt = toNullTime(time.Time{})

	// 更新数据库
	_, err = ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to restore token: %w", err)
	}

	// 记录审计日志
	_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpUpdate, &oldStatus, &newStatus, nil, "", "")

	return nil
}

// UseQuota 使用配额
func (ts *TokenService) UseQuota(ctx context.Context, tokenID int, amount int64) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// 检查是否有足够的配额
	if token.QuotaLimit.Valid && token.QuotaUsed+amount > token.QuotaLimit.Int64 {
		// 更新状态为已耗尽
		newStatus := model.TokenStatusExhausted
		oldStatus := token.Status
		token.Status = newStatus

		_, _ = ts.tokenRepo.Update(ctx, token)
		_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpUseQuota, &oldStatus, &newStatus, nil, "", "")

		return fmt.Errorf("quota exceeded for token %d", tokenID)
	}

	// 更新配额
	token.QuotaUsed += amount

	// 如果接近预警阈值，检查是否需要发送预警
	if token.QuotaLimit.Valid {
		percentage := (float64(token.QuotaUsed) / float64(token.QuotaLimit.Int64)) * 100
		if percentage >= 80 {
			// TODO: 发送预警通知
		}
	}

	_, err = ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	// 记录审计日志
	details := map[string]interface{}{"amount": amount, "new_used": token.QuotaUsed}
	_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpUseQuota, &token.Status, &token.Status, details, "", "")

	return nil
}

// RefundQuota 退款配额
func (ts *TokenService) RefundQuota(ctx context.Context, tokenID int, amount int64) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if token.QuotaUsed < amount {
		return fmt.Errorf("refund amount exceeds used quota")
	}

	token.QuotaUsed -= amount

	_, err = ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to refund quota: %w", err)
	}

	// 记录审计日志
	details := map[string]interface{}{"amount": amount, "new_used": token.QuotaUsed}
	_ = ts.logAudit(ctx, token.UserID, tokenID, model.TokenOpUseQuota, &token.Status, &token.Status, details, "", "")

	return nil
}

// ListTokens 列出用户的所有 Token
func (ts *TokenService) ListTokens(ctx context.Context, userID int, includeDeleted bool) ([]*model.Token, error) {
	tokens, err := ts.tokenRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %w", err)
	}

	// 过滤已删除的 Token（如果需要）
	if !includeDeleted {
		var activeTokens []*model.Token
		for _, token := range tokens {
			if token.Status != model.TokenStatusDeleted {
				activeTokens = append(activeTokens, token)
			}
		}
		return activeTokens, nil
	}

	return tokens, nil
}

// CheckAndUpdateExpiredTokens 检查并更新过期的 Token
func (ts *TokenService) CheckAndUpdateExpiredTokens(ctx context.Context) (int, error) {
	// 调用数据库函数
	count, err := ts.tokenRepo.CheckAndUpdateExpiredTokens(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to check expired tokens: %w", err)
	}
	return count, nil
}

// 私有方法

// generateTokenHash 生成 Token 哈希
func (ts *TokenService) generateTokenHash() string {
	data := []byte(fmt.Sprintf("%d-%d-%d", time.Now().UnixNano(), time.Now().Unix(), time.Now().Nanosecond()))
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// updateTokenStatus 更新 Token 状态
func (ts *TokenService) updateTokenStatus(ctx context.Context, token *model.Token, newStatus model.TokenStatus) error {
	oldStatus := token.Status
	token.Status = newStatus

	_, err := ts.tokenRepo.Update(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to update token status: %w", err)
	}

	// 记录审计日志
	_ = ts.logAudit(ctx, token.UserID, token.ID, model.TokenOpUpdate, &oldStatus, &newStatus, nil, "", "")

	return nil
}

// updateLastUsedAt 更新最后使用时间
func (ts *TokenService) updateLastUsedAt(ctx context.Context, tokenID int) error {
	token, err := ts.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return err
	}

	token.LastUsedAt = toNullTime(time.Now())
	_, err = ts.tokenRepo.Update(ctx, token)
	return err
}

// logAudit 记录审计日志
func (ts *TokenService) logAudit(
	ctx context.Context,
	userID int,
	tokenID int,
	operation model.TokenOperationType,
	oldStatus *model.TokenStatus,
	newStatus *model.TokenStatus,
	details map[string]interface{},
	ipAddress string,
	userAgent string,
) error {
	auditLog := &model.TokenAuditLog{
		UserID:    userID,
		TokenID:   tokenID,
		Operation: string(operation),
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Details:   details,
		CreatedAt: time.Now(),
	}

	if ipAddress != "" {
		auditLog.IPAddress = toNullString(ipAddress)
	}

	if userAgent != "" {
		auditLog.UserAgent = toNullString(userAgent)
	}

	return ts.tokenRepo.LogAudit(ctx, auditLog)
}

// logRenewal 记录续期日志
func (ts *TokenService) logRenewal(
	ctx context.Context,
	tokenID int,
	oldExpireAt interface{},
	newExpireAt interface{},
	reason string,
) error {
	renewalLog := &model.TokenRenewalLog{
		TokenID:         tokenID,
		RenewalReason:   reason,
		CreatedAt:       time.Now(),
	}

	// 处理旧过期时间
	if oldExpireAt != nil {
		if nullTime, ok := oldExpireAt.(time.Time); ok {
			renewalLog.OldExpireAt = toNullTime(nullTime)
		}
	}

	// 处理新过期时间
	if newExpireAt != nil {
		if nt, ok := newExpireAt.(interface{}).(interface{ Time() time.Time; Valid() bool }); ok {
			renewalLog.NewExpireAt = toNullTime(nt.Time())
		}
	}

	return ts.tokenRepo.LogRenewal(ctx, renewalLog)
}

// 辅助函数

func toNullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func toNullInt64(i int64) interface{} {
	if i == 0 {
		return nil
	}
	return i
}

func toNullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

// GetTokenDetailsJSON 获取 Token 详情的 JSON 格式
func (ts *TokenService) GetTokenDetailsJSON(token *model.Token) (string, error) {
	details := map[string]interface{}{
		"id":              token.ID,
		"name":            token.Name,
		"status":          token.Status.String(),
		"quota_limit":     token.QuotaLimit,
		"quota_used":      token.QuotaUsed,
		"quota_remaining": token.GetRemainingQuota(),
		"quota_percent":   token.GetQuotaPercentage(),
		"created_at":      token.CreatedAt,
		"expire_at":       token.ExpireAt,
		"expiring_soon":   token.IsExpiringSoon(),
	}

	jsonData, err := json.Marshal(details)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token details: %w", err)
	}

	return string(jsonData), nil
}

