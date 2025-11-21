package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestChatHandler(t *testing.T) {
	// 创建 mock client
	client := NewRequestClient(30 * time.Second)

	handler := NewChatHandler(client)

	// 验证类型
	if handler.GetType() != RequestTypeChat {
		t.Errorf("Expected type %d, got %d", RequestTypeChat, handler.GetType())
	}

	// 验证名称
	if handler.GetName() != "chat" {
		t.Errorf("Expected name 'chat', got %s", handler.GetName())
	}

	// 验证流式支持
	if !handler.SupportsStreaming() {
		t.Errorf("Chat handler should support streaming")
	}
}

func TestEmbeddingHandler(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewEmbeddingHandler(client)

	if handler.GetType() != RequestTypeEmbedding {
		t.Errorf("Expected type %d, got %d", RequestTypeEmbedding, handler.GetType())
	}

	if handler.GetName() != "embedding" {
		t.Errorf("Expected name 'embedding', got %s", handler.GetName())
	}

	if handler.SupportsStreaming() {
		t.Errorf("Embedding handler should not support streaming")
	}
}

func TestImageHandler(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewImageHandler(client)

	if handler.GetType() != RequestTypeImage {
		t.Errorf("Expected type %d, got %d", RequestTypeImage, handler.GetType())
	}

	if handler.GetName() != "image" {
		t.Errorf("Expected name 'image', got %s", handler.GetName())
	}
}

func TestAudioHandler(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewAudioHandler(client)

	if handler.GetType() != RequestTypeAudio {
		t.Errorf("Expected type %d, got %d", RequestTypeAudio, handler.GetType())
	}

	if handler.GetName() != "audio" {
		t.Errorf("Expected name 'audio', got %s", handler.GetName())
	}

	if !handler.SupportsStreaming() {
		t.Errorf("Audio handler should support streaming")
	}
}

func TestHandlerValidation(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewChatHandler(client)

	// 测试无效请求
	tests := []struct {
		name    string
		req     *HandlerRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "missing model",
			req: &HandlerRequest{
				Type:     RequestTypeChat,
				Endpoint: "/v1/chat",
				Body:     []byte(`{"messages":[]}`),
			},
			wantErr: true,
		},
		{
			name: "missing endpoint",
			req: &HandlerRequest{
				Type:  RequestTypeChat,
				Model: "gpt-4",
				Body:  []byte(`{"messages":[]}`),
			},
			wantErr: true,
		},
		{
			name: "invalid request type",
			req: &HandlerRequest{
				Type:     RequestTypeEmbedding,
				Model:    "gpt-4",
				Endpoint: "/v1/chat",
				Body:     []byte(`{"messages":[]}`),
			},
			wantErr: true,
		},
		{
			name: "missing messages in chat",
			req: &HandlerRequest{
				Type:     RequestTypeChat,
				Model:    "gpt-4",
				Endpoint: "/v1/chat",
				Body:     []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "valid chat request",
			req: &HandlerRequest{
				Type:     RequestTypeChat,
				Model:    "gpt-4",
				Endpoint: "/v1/chat",
				Body:     []byte(`{"messages":[{"role":"user","content":"hello"}]}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandlerRegistry(t *testing.T) {
	registry := NewHandlerRegistry()
	client := NewRequestClient(30 * time.Second)

	// 注册处理器
	handlers := []RelayHandler{
		NewChatHandler(client),
		NewEmbeddingHandler(client),
		NewImageHandler(client),
		NewAudioHandler(client),
	}

	for _, handler := range handlers {
		if err := registry.RegisterHandler(handler); err != nil {
			t.Fatalf("Failed to register handler: %v", err)
		}
	}

	// 测试获取处理器
	for _, expectedHandler := range handlers {
		got, err := registry.GetHandler(expectedHandler.GetType())
		if err != nil {
			t.Errorf("Failed to get handler: %v", err)
		}
		if got.GetName() != expectedHandler.GetName() {
			t.Errorf("Expected handler %s, got %s", expectedHandler.GetName(), got.GetName())
		}
	}

	// 测试获取不存在的处理器
	_, err := registry.GetHandler(RequestType(999))
	if err == nil {
		t.Errorf("Expected error for non-existent handler")
	}

	// 测试重复注册
	err = registry.RegisterHandler(NewChatHandler(client))
	if err == nil {
		t.Errorf("Expected error when registering duplicate handler")
	}

	// 测试注销处理器
	if err := registry.UnregisterHandler(RequestTypeChat); err != nil {
		t.Errorf("Failed to unregister handler: %v", err)
	}

	// 验证已注销
	_, err = registry.GetHandler(RequestTypeChat)
	if err == nil {
		t.Errorf("Expected error for unregistered handler")
	}
}

func TestHandlerManager(t *testing.T) {
	manager := NewHandlerManager()
	client := NewRequestClient(30 * time.Second)

	// 初始化默认处理器
	if err := InitializeDefaultHandlers(manager, client); err != nil {
		t.Fatalf("Failed to initialize handlers: %v", err)
	}

	// 测试获取处理器
	handler, err := manager.GetHandler(RequestTypeChat)
	if err != nil {
		t.Errorf("Failed to get handler: %v", err)
	}
	if handler.GetName() != "chat" {
		t.Errorf("Expected 'chat', got %s", handler.GetName())
	}

	// 获取统计信息
	stats := manager.GetStatistics()
	if stats == nil {
		t.Errorf("Expected statistics, got nil")
	}

	handlers, ok := stats["handlers"].(map[string]map[string]interface{})
	if !ok {
		t.Errorf("Invalid statistics format")
	}

	if len(handlers) != 4 {
		t.Errorf("Expected 4 handlers in stats, got %d", len(handlers))
	}

	// 测试缓存
	handler1, _ := manager.GetHandler(RequestTypeChat)
	handler2, _ := manager.GetHandler(RequestTypeChat)
	if handler1 != handler2 {
		t.Errorf("Expected same handler from cache")
	}

	// 清除缓存
	manager.ClearRouteCache()

	// 统计重置
	manager.ResetStatistics()
	stats = manager.GetStatistics()
	if stats["total_handled_requests"].(int64) != 0 {
		t.Errorf("Expected 0 total handled requests after reset")
	}
}

func TestHandlerFactory(t *testing.T) {
	client := NewRequestClient(30 * time.Second)

	tests := []struct {
		name     string
		handler  RelayHandler
		wantType RequestType
	}{
		{"Chat Factory", NewChatHandler(client), RequestTypeChat},
		{"Embedding Factory", NewEmbeddingHandler(client), RequestTypeEmbedding},
		{"Image Factory", NewImageHandler(client), RequestTypeImage},
		{"Audio Factory", NewAudioHandler(client), RequestTypeAudio},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.handler.GetType() != tt.wantType {
				t.Errorf("Factory created wrong type: got %d, want %d", tt.handler.GetType(), tt.wantType)
			}
		})
	}
}

func TestEmbeddingValidation(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewEmbeddingHandler(client)

	tests := []struct {
		name    string
		req     *HandlerRequest
		wantErr bool
	}{
		{
			name: "missing input",
			req: &HandlerRequest{
				Type:     RequestTypeEmbedding,
				Model:    "text-embedding-3",
				Endpoint: "/v1/embeddings",
				Body:     []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "valid embedding request",
			req: &HandlerRequest{
				Type:     RequestTypeEmbedding,
				Model:    "text-embedding-3",
				Endpoint: "/v1/embeddings",
				Body:     []byte(`{"input":"hello world"}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageValidation(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewImageHandler(client)

	tests := []struct {
		name    string
		req     *HandlerRequest
		wantErr bool
	}{
		{
			name: "missing prompt",
			req: &HandlerRequest{
				Type:     RequestTypeImage,
				Model:    "dall-e-3",
				Endpoint: "/v1/images/generations",
				Body:     []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "valid image request",
			req: &HandlerRequest{
				Type:     RequestTypeImage,
				Model:    "dall-e-3",
				Endpoint: "/v1/images/generations",
				Body:     []byte(`{"prompt":"a cat"}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandlerStatistics(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewChatHandler(client)

	// 初始统计
	stats := handler.GetStatistics()
	if totalReqs, ok := stats["total_requests"].(int64); !ok || totalReqs != 0 {
		t.Errorf("Expected 0 initial requests")
	}

	// 模拟请求
	handler.RecordSuccess(1024)
	handler.RecordSuccess(2048)
	handler.RecordFailure()

	stats = handler.GetStatistics()
	if totalReqs, ok := stats["total_requests"].(int64); !ok || totalReqs != 3 {
		t.Errorf("Expected 3 total requests, got %v", totalReqs)
	}

	if successReqs, ok := stats["successful_requests"].(int64); !ok || successReqs != 2 {
		t.Errorf("Expected 2 successful requests, got %v", successReqs)
	}

	if failedReqs, ok := stats["failed_requests"].(int64); !ok || failedReqs != 1 {
		t.Errorf("Expected 1 failed request, got %v", failedReqs)
	}

	if totalBytes, ok := stats["total_bytes"].(int64); !ok || totalBytes != 3072 {
		t.Errorf("Expected 3072 total bytes, got %v", totalBytes)
	}

	// 重置统计
	handler.ResetStatistics()
	stats = handler.GetStatistics()
	if totalReqs, ok := stats["total_requests"].(int64); !ok || totalReqs != 0 {
		t.Errorf("Expected 0 requests after reset")
	}
}

func TestRequestTypeString(t *testing.T) {
	tests := []struct {
		name     string
		reqType  RequestType
		expected string
	}{
		{"Chat", RequestTypeChat, "chat"},
		{"Embedding", RequestTypeEmbedding, "embedding"},
		{"Image", RequestTypeImage, "image"},
		{"Audio", RequestTypeAudio, "audio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reqType.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.reqType.String())
			}
		})
	}
}

func TestHandlerStreamingCapability(t *testing.T) {
	client := NewRequestClient(30 * time.Second)

	tests := []struct {
		name              string
		handler           RelayHandler
		supportsStreaming bool
	}{
		{"Chat", NewChatHandler(client), true},
		{"Embedding", NewEmbeddingHandler(client), false},
		{"Image", NewImageHandler(client), false},
		{"Audio", NewAudioHandler(client), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.handler.SupportsStreaming() != tt.supportsStreaming {
				t.Errorf("Expected streaming=%v, got %v", tt.supportsStreaming, tt.handler.SupportsStreaming())
			}
		})
	}
}

func TestHandlerStreamingError(t *testing.T) {
	client := NewRequestClient(30 * time.Second)

	// Embedding 不支持流式
	handler := NewEmbeddingHandler(client)
	req := &HandlerRequest{
		Type:     RequestTypeEmbedding,
		Model:    "text-embedding-3",
		Endpoint: "/v1/embeddings",
		Body:     []byte(`{"input":"hello"}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	responseCh := make(chan *HandlerResponse, 10)
	err := handler.HandleStream(ctx, req, responseCh)

	if err == nil {
		t.Errorf("Expected error for unsupported streaming")
	}
}

func TestHandlerResponseValidation(t *testing.T) {
	client := NewRequestClient(30 * time.Second)
	handler := NewChatHandler(client)

	tests := []struct {
		name    string
		resp    *HandlerResponse
		wantErr bool
	}{
		{
			name:    "nil response",
			resp:    nil,
			wantErr: true,
		},
		{
			name: "negative status code",
			resp: &HandlerResponse{
				StatusCode: -1,
				Body:       []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "valid response",
			resp: &HandlerResponse{
				StatusCode: 200,
				Body:       []byte(`{"choices":[{"message":{"content":"hello"}}]}`),
				Headers:    make(map[string]string),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateResponse(tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkHandlerValidation(b *testing.B) {
	client := NewRequestClient(30 * time.Second)
	handler := NewChatHandler(client)

	req := &HandlerRequest{
		Type:     RequestTypeChat,
		Model:    "gpt-4",
		Endpoint: "/v1/chat",
		Body:     []byte(`{"messages":[{"role":"user","content":"hello"}]}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.ValidateRequest(req)
	}
}

func BenchmarkRegistryLookup(b *testing.B) {
	registry := NewHandlerRegistry()
	client := NewRequestClient(30 * time.Second)

	handlers := []RelayHandler{
		NewChatHandler(client),
		NewEmbeddingHandler(client),
		NewImageHandler(client),
		NewAudioHandler(client),
	}

	for _, handler := range handlers {
		_ = registry.RegisterHandler(handler)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetHandler(RequestTypeChat)
	}
}

