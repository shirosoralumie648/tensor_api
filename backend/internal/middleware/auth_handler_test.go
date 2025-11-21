package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBearerExtractor(t *testing.T) {
	extractor := &BearerExtractor{}

	t.Run("valid_bearer_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer test-token-123")

		token, err := extractor.Extract(c)
		assert.NoError(t, err)
		assert.Equal(t, "test-token-123", token)
	})

	t.Run("missing_bearer_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)

		token, err := extractor.Extract(c)
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})

	t.Run("invalid_bearer_format", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat test-token")

		token, err := extractor.Extract(c)
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}

func TestClaudeExtractor(t *testing.T) {
	extractor := &ClaudeExtractor{}

	t.Run("valid_claude_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.Header.Set("x-api-key", "claude-key-123")

		token, err := extractor.Extract(c)
		assert.NoError(t, err)
		assert.Equal(t, "claude-key-123", token)
	})

	t.Run("missing_claude_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)

		token, err := extractor.Extract(c)
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}

func TestGeminiExtractor(t *testing.T) {
	extractor := &GeminiExtractor{}

	t.Run("valid_gemini_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.Header.Set("x-goog-api-key", "gemini-key-123")

		token, err := extractor.Extract(c)
		assert.NoError(t, err)
		assert.Equal(t, "gemini-key-123", token)
	})

	t.Run("missing_gemini_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)

		token, err := extractor.Extract(c)
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}

func TestWebSocketExtractor(t *testing.T) {
	extractor := &WebSocketExtractor{}

	t.Run("valid_websocket_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/?token=ws-token-123", nil)

		token, err := extractor.Extract(c)
		assert.NoError(t, err)
		assert.Equal(t, "ws-token-123", token)
	})

	t.Run("missing_websocket_token", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)

		token, err := extractor.Extract(c)
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}

func TestAuthExtractorFactory(t *testing.T) {
	factory := NewAuthExtractorFactory()

	t.Run("register_and_extract", func(t *testing.T) {
		factory.RegisterExtractor(AuthMethodBearer, &BearerExtractor{})

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer test-token")

		token, method, err := factory.ExtractToken(c)
		assert.NoError(t, err)
		assert.Equal(t, "test-token", token)
		assert.Equal(t, AuthMethodBearer, method)
	})

	t.Run("priority_order", func(t *testing.T) {
		factory := NewAuthExtractorFactory()

		// 注册多个提取器
		factory.RegisterExtractor(AuthMethodBearer, &BearerExtractor{})
		factory.RegisterExtractor(AuthMethodClaude, &ClaudeExtractor{})
		factory.RegisterExtractor(AuthMethodWebSocket, &WebSocketExtractor{})

		// Bearer 的优先级最高（Priority() = 1）
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/?token=ws-token", nil)
		c.Request.Header.Set("Authorization", "Bearer bearer-token")

		token, method, err := factory.ExtractToken(c)
		assert.NoError(t, err)
		assert.Equal(t, "bearer-token", token)
		assert.Equal(t, AuthMethodBearer, method)
	})
}

func TestGetDefaultFactory(t *testing.T) {
	factory := GetDefaultFactory()

	t.Run("all_extractors_registered", func(t *testing.T) {
		// 所有标准提取器都应该被注册
		assert.NotNil(t, factory.extractors[AuthMethodBearer])
		assert.NotNil(t, factory.extractors[AuthMethodClaude])
		assert.NotNil(t, factory.extractors[AuthMethodGemini])
		assert.NotNil(t, factory.extractors[AuthMethodWebSocket])
	})
}

func TestAuthMethod_String(t *testing.T) {
	tests := []struct {
		method   AuthMethod
		expected string
	}{
		{AuthMethodBearer, "bearer"},
		{AuthMethodClaude, "claude"},
		{AuthMethodGemini, "gemini"},
		{AuthMethodWebSocket, "websocket"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.method.String())
		})
	}
}

func TestExtractorPriorities(t *testing.T) {
	// 验证优先级顺序
	extractors := []TokenExtractor{
		&BearerExtractor{},
		&ClaudeExtractor{},
		&GeminiExtractor{},
		&WebSocketExtractor{},
	}

	priorities := make([]int, len(extractors))
	for i, e := range extractors {
		priorities[i] = e.Priority()
	}

	// 应该是递增的
	for i := 1; i < len(priorities); i++ {
		assert.Less(t, priorities[i-1], priorities[i],
			"priorities should be in ascending order")
	}
}

func BenchmarkBearerExtractor(b *testing.B) {
	extractor := &BearerExtractor{}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer test-token")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = extractor.Extract(c)
	}
}

func BenchmarkAuthExtractorFactory(b *testing.B) {
	factory := GetDefaultFactory()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer test-token")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = factory.ExtractToken(c)
	}
}

