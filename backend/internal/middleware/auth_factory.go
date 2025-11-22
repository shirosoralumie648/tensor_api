package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/shirosoralumie648/Oblivious/backend/internal/cache"
)

// AuthMethod 认证方法类型
type AuthMethod string

const (
	AuthMethodBearer    AuthMethod = "bearer"
	AuthMethodClaude    AuthMethod = "claude"
	AuthMethodGemini    AuthMethod = "gemini"
	AuthMethodWebSocket AuthMethod = "websocket"
)

// TokenExtractor Token 提取器接口
type TokenExtractor interface {
	// Extract 从请求中提取 Token
	Extract(c *gin.Context) (string, error)
	// Name 返回提取器的名称
	Name() string
	// Priority 返回优先级（数值越小优先级越高）
	Priority() int
}

// BearerExtractor Bearer Token 提取器
type BearerExtractor struct{}

func (e *BearerExtractor) Extract(c *gin.Context) (string, error) {
	return ExtractTokenFromBearer(c)
}

func (e *BearerExtractor) Name() string {
	return string(AuthMethodBearer)
}

func (e *BearerExtractor) Priority() int {
	return 1
}

// ClaudeExtractor Claude API 风格的提取器
type ClaudeExtractor struct{}

func (e *ClaudeExtractor) Extract(c *gin.Context) (string, error) {
	return ExtractTokenFromClaudeHeader(c)
}

func (e *ClaudeExtractor) Name() string {
	return string(AuthMethodClaude)
}

func (e *ClaudeExtractor) Priority() int {
	return 2
}

// GeminiExtractor Gemini API 风格的提取器
type GeminiExtractor struct{}

func (e *GeminiExtractor) Extract(c *gin.Context) (string, error) {
	return ExtractTokenFromGeminiHeader(c)
}

func (e *GeminiExtractor) Name() string {
	return string(AuthMethodGemini)
}

func (e *GeminiExtractor) Priority() int {
	return 3
}

// WebSocketExtractor WebSocket 连接参数提取器
type WebSocketExtractor struct{}

func (e *WebSocketExtractor) Extract(c *gin.Context) (string, error) {
	return ExtractTokenFromWebSocket(c)
}

func (e *WebSocketExtractor) Name() string {
	return string(AuthMethodWebSocket)
}

func (e *WebSocketExtractor) Priority() int {
	return 4
}

// AuthExtractorFactory 认证提取器工厂
type AuthExtractorFactory struct {
	extractors map[AuthMethod]TokenExtractor
}

// NewAuthExtractorFactory 创建新的工厂
func NewAuthExtractorFactory() *AuthExtractorFactory {
	return &AuthExtractorFactory{
		extractors: make(map[AuthMethod]TokenExtractor),
	}
}

// RegisterExtractor 注册提取器
func (f *AuthExtractorFactory) RegisterExtractor(method AuthMethod, extractor TokenExtractor) {
	f.extractors[method] = extractor
}

// ExtractToken 使用注册的提取器提取 Token
func (f *AuthExtractorFactory) ExtractToken(c *gin.Context) (string, AuthMethod, error) {
	// 收集所有可用的提取器
	var extractors []TokenExtractor
	for _, extractor := range f.extractors {
		extractors = append(extractors, extractor)
	}

	// 按优先级排序
	for i := 0; i < len(extractors); i++ {
		for j := i + 1; j < len(extractors); j++ {
			if extractors[j].Priority() < extractors[i].Priority() {
				extractors[i], extractors[j] = extractors[j], extractors[i]
			}
		}
	}

	// 尝试每个提取器
	var lastErr error
	for _, extractor := range extractors {
		token, err := extractor.Extract(c)
		if err == nil && token != "" {
			return token, AuthMethod(extractor.Name()), nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", "", lastErr
	}
	return "", "", fmt.Errorf("no valid token found using any auth method")
}

// GetDefaultFactory 获取默认的工厂实例（包含所有标准提取器）
func GetDefaultFactory() *AuthExtractorFactory {
	factory := NewAuthExtractorFactory()

	// 注册所有标准提取器
	factory.RegisterExtractor(AuthMethodBearer, &BearerExtractor{})
	factory.RegisterExtractor(AuthMethodClaude, &ClaudeExtractor{})
	factory.RegisterExtractor(AuthMethodGemini, &GeminiExtractor{})
	factory.RegisterExtractor(AuthMethodWebSocket, &WebSocketExtractor{})

	return factory
}

// AuthExtractorRegistry 全局提取器注册表
var globalExtractorFactory *AuthExtractorFactory

// init 初始化全局工厂
func init() {
	globalExtractorFactory = GetDefaultFactory()
}

// RegisterGlobalExtractor 在全局工厂中注册提取器
func RegisterGlobalExtractor(method AuthMethod, extractor TokenExtractor) {
	globalExtractorFactory.RegisterExtractor(method, extractor)
}

// ExtractTokenGlobal 使用全局工厂提取 Token
func ExtractTokenGlobal(c *gin.Context) (string, AuthMethod, error) {
	return globalExtractorFactory.ExtractToken(c)
}

// AuthFactory 认证中间件工厂
type AuthFactory struct {
	jwtSecret []byte
	cache     cache.Cache
}

// NewAuthFactory 创建认证工厂
func NewAuthFactory(jwtSecret []byte, cache cache.Cache) *AuthFactory {
	return &AuthFactory{
		jwtSecret: jwtSecret,
		cache:     cache,
	}
}

// JWT 创建 JWT 认证中间件
func (af *AuthFactory) JWT() gin.HandlerFunc {
	return AuthMiddleware(af.jwtSecret)
}

// JWTWithCache 创建带缓存的 JWT 认证中间件
func (af *AuthFactory) JWTWithCache() gin.HandlerFunc {
	if af.cache != nil {
		return CachedAuthMiddleware(af.jwtSecret, nil) // TODO: 传入 CacheManager
	}
	return AuthMiddleware(af.jwtSecret)
}

// APIKey 创建 API 密钥认证中间件
func (af *AuthFactory) APIKey() gin.HandlerFunc {
	return APIKeyAuthMiddleware()
}

// Optional 创建可选认证中间件（认证失败不阻止请求）
func (af *AuthFactory) Optional() gin.HandlerFunc {
	return OptionalAuthMiddleware(af.jwtSecret)
}
