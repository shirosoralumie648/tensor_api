package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/config"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/repository"
	"github.com/shirosoralumie648/Oblivious/backend/internal/utils"
)

type UserService struct {
	userRepo *repository.UserRepository
	jwtCfg   *config.JWTConfig
}

func NewUserService(jwtCfg *config.JWTConfig) *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
		jwtCfg:   jwtCfg,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         *model.User `json:"user"`
}

// Register 注册用户
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	existingUser, err = s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 生成邀请码
	inviteCode := generateInviteCode()

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.Username,
		InviteCode:   inviteCode,
		Quota:        500000, // 新用户默认 5000 元（500000 分）
		TotalQuota:   500000,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 查询用户
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// 生成 Token
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Role,
		s.jwtCfg.ExpireHours,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		s.jwtCfg.RefreshExpireDays,
	)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtCfg.ExpireHours * 3600,
		User:         user,
	}, nil
}

// GetUserByID 获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshAccessToken 刷新 Access Token
func (s *UserService) RefreshAccessToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// 解析 Refresh Token
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 查询用户
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// 生成新的 Token
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Role,
		s.jwtCfg.ExpireHours,
	)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		s.jwtCfg.RefreshExpireDays,
	)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.jwtCfg.ExpireHours * 3600,
		User:         user,
	}, nil
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID int, req *UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// 生成邀请码（6 位随机字符）
func generateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}


