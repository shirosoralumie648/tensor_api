package relay

import (
	"bytes"
	"context"
	"encoding/json"
)

// ChatHandler Chat 处理器
type ChatHandler struct {
	*BaseRelayHandler
	client *RequestClient
}

// NewChatHandler 创建 Chat 处理器
func NewChatHandler(client *RequestClient) *ChatHandler {
	return &ChatHandler{
		BaseRelayHandler: NewBaseRelayHandler(RequestTypeChat, "chat", true),
		client:           client,
	}
}

// Handle 处理同步请求
func (ch *ChatHandler) Handle(ctx context.Context, req *HandlerRequest) (*HandlerResponse, error) {
	// 验证请求
	if err := ch.ValidateRequest(req); err != nil {
		ch.RecordFailure()
		return nil, err
	}

	// 构建请求头
	headers := make(map[string]string)
	for k, v := range req.Headers {
		headers[k] = v
	}
	headers["Content-Type"] = "application/json"

	// 使用 RequestClient 发送请求
	respBody, respHeaders, err := ch.client.DoRequest(
		ctx,
		"POST",
		req.Endpoint,
		bytes.NewReader(req.Body),
		headers,
	)

	if err != nil {
		ch.RecordFailure()
		return &HandlerResponse{
			StatusCode: 500,
			Error:      err.Error(),
		}, err
	}

	ch.RecordSuccess(int64(len(req.Body)))

	// 将 http.Header 转换为 map[string]string
	headerMap := make(map[string]string)
	for k, v := range respHeaders {
		if len(v) > 0 {
			headerMap[k] = v[0]
		}
	}

	// 验证响应
	resp := &HandlerResponse{
		StatusCode: 200, // 假设成功返回 200
		Body:       respBody,
		Headers:    headerMap,
	}

	if err := ch.ValidateResponse(resp); err != nil {
		ch.RecordFailure()
		return nil, err
	}

	return resp, nil
}

// HandleStream 处理流式请求
func (ch *ChatHandler) HandleStream(ctx context.Context, req *HandlerRequest, responseCh chan *HandlerResponse) error {
	// 验证请求
	if err := ch.ValidateRequest(req); err != nil {
		ch.RecordFailure()
		return err
	}

	// 这里应该调用真实的流式处理逻辑
	// 暂时返回模拟响应
	resp := &HandlerResponse{
		StatusCode: 200,
		Body:       []byte(`{"choices":[{"message":{"content":"Hello from stream"}}]}`),
		Headers:    make(map[string]string),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case responseCh <- resp:
	}

	close(responseCh)
	return nil
}

// EmbeddingHandler Embedding 处理器
type EmbeddingHandler struct {
	*BaseRelayHandler
	client *RequestClient
}

// NewEmbeddingHandler 创建 Embedding 处理器
func NewEmbeddingHandler(client *RequestClient) *EmbeddingHandler {
	return &EmbeddingHandler{
		BaseRelayHandler: NewBaseRelayHandler(RequestTypeEmbedding, "embedding", false),
		client:           client,
	}
}

// Handle 处理同步请求
func (eh *EmbeddingHandler) Handle(ctx context.Context, req *HandlerRequest) (*HandlerResponse, error) {
	// 验证请求
	if err := eh.ValidateRequest(req); err != nil {
		eh.RecordFailure()
		return nil, err
	}

	headers := make(map[string]string)
	for k, v := range req.Headers {
		headers[k] = v
	}
	headers["Content-Type"] = "application/json"

	respBody, respHeaders, err := eh.client.DoRequest(
		ctx,
		"POST",
		req.Endpoint,
		bytes.NewReader(req.Body),
		headers,
	)

	if err != nil {
		eh.RecordFailure()
		return &HandlerResponse{
			StatusCode: 500,
			Error:      err.Error(),
		}, err
	}

	eh.RecordSuccess(int64(len(req.Body)))

	// 将 http.Header 转换为 map[string]string
	headerMap := make(map[string]string)
	for k, v := range respHeaders {
		if len(v) > 0 {
			headerMap[k] = v[0]
		}
	}

	resp := &HandlerResponse{
		StatusCode: 200,
		Body:       respBody,
		Headers:    headerMap,
	}

	if err := eh.ValidateResponse(resp); err != nil {
		eh.RecordFailure()
		return nil, err
	}

	return resp, nil
}

// HandleStream 不支持流式
func (eh *EmbeddingHandler) HandleStream(ctx context.Context, req *HandlerRequest, responseCh chan *HandlerResponse) error {
	return ErrUnsupportedType{Type: RequestTypeEmbedding}
}

// ImageHandler Image 处理器
type ImageHandler struct {
	*BaseRelayHandler
	client *RequestClient
}

// NewImageHandler 创建 Image 处理器
func NewImageHandler(client *RequestClient) *ImageHandler {
	return &ImageHandler{
		BaseRelayHandler: NewBaseRelayHandler(RequestTypeImage, "image", false),
		client:           client,
	}
}

// Handle 处理同步请求
func (ih *ImageHandler) Handle(ctx context.Context, req *HandlerRequest) (*HandlerResponse, error) {
	// 验证请求
	if err := ih.ValidateRequest(req); err != nil {
		ih.RecordFailure()
		return nil, err
	}

	headers := make(map[string]string)
	for k, v := range req.Headers {
		headers[k] = v
	}
	headers["Content-Type"] = "application/json"

	respBody, respHeaders, err := ih.client.DoRequest(
		ctx,
		"POST",
		req.Endpoint,
		bytes.NewReader(req.Body),
		headers,
	)

	if err != nil {
		ih.RecordFailure()
		return &HandlerResponse{
			StatusCode: 500,
			Error:      err.Error(),
		}, err
	}

	ih.RecordSuccess(int64(len(req.Body)))

	// 将 http.Header 转换为 map[string]string
	headerMap := make(map[string]string)
	for k, v := range respHeaders {
		if len(v) > 0 {
			headerMap[k] = v[0]
		}
	}

	resp := &HandlerResponse{
		StatusCode: 200,
		Body:       respBody,
		Headers:    headerMap,
	}

	if err := ih.ValidateResponse(resp); err != nil {
		ih.RecordFailure()
		return nil, err
	}

	return resp, nil
}

// HandleStream 不支持流式
func (ih *ImageHandler) HandleStream(ctx context.Context, req *HandlerRequest, responseCh chan *HandlerResponse) error {
	return ErrUnsupportedType{Type: RequestTypeImage}
}

// AudioHandler Audio 处理器
type AudioHandler struct {
	*BaseRelayHandler
	client *RequestClient
}

// NewAudioHandler 创建 Audio 处理器
func NewAudioHandler(client *RequestClient) *AudioHandler {
	return &AudioHandler{
		BaseRelayHandler: NewBaseRelayHandler(RequestTypeAudio, "audio", true),
		client:           client,
	}
}

// Handle 处理同步请求
func (ah *AudioHandler) Handle(ctx context.Context, req *HandlerRequest) (*HandlerResponse, error) {
	// 验证请求
	if err := ah.ValidateRequest(req); err != nil {
		ah.RecordFailure()
		return nil, err
	}

	headers := make(map[string]string)
	for k, v := range req.Headers {
		headers[k] = v
	}
	// Audio 可能支持多种内容类型
	if headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/octet-stream"
	}

	respBody, respHeaders, err := ah.client.DoRequest(
		ctx,
		"POST",
		req.Endpoint,
		bytes.NewReader(req.Body),
		headers,
	)

	if err != nil {
		ah.RecordFailure()
		return &HandlerResponse{
			StatusCode: 500,
			Error:      err.Error(),
		}, err
	}

	ah.RecordSuccess(int64(len(req.Body)))

	// 将 http.Header 转换为 map[string]string
	headerMap := make(map[string]string)
	for k, v := range respHeaders {
		if len(v) > 0 {
			headerMap[k] = v[0]
		}
	}

	resp := &HandlerResponse{
		StatusCode: 200,
		Body:       respBody,
		Headers:    headerMap,
	}

	if err := ah.ValidateResponse(resp); err != nil {
		ah.RecordFailure()
		return nil, err
	}

	return resp, nil
}

// HandleStream 处理流式请求
func (ah *AudioHandler) HandleStream(ctx context.Context, req *HandlerRequest, responseCh chan *HandlerResponse) error {
	// 验证请求
	if err := ah.ValidateRequest(req); err != nil {
		ah.RecordFailure()
		return err
	}

	// 这里应该调用真实的流式处理逻辑
	resp := &HandlerResponse{
		StatusCode: 200,
		Body:       []byte("audio stream data"),
		Headers:    make(map[string]string),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case responseCh <- resp:
	}

	close(responseCh)
	return nil
}

// ValidateRequest 覆盖基础验证（Chat 特殊处理）
func (ch *ChatHandler) ValidateRequest(req *HandlerRequest) error {
	if err := ch.BaseRelayHandler.ValidateRequest(req); err != nil {
		return err
	}

	// Chat 特殊验证：检查必要的字段
	var chatReq map[string]interface{}
	if err := json.Unmarshal(req.Body, &chatReq); err != nil {
		return ErrInvalidRequest{Message: "invalid chat request JSON"}
	}

	if _, ok := chatReq["messages"]; !ok {
		return ErrInvalidRequest{Message: "messages field is required in chat request"}
	}

	return nil
}

// ValidateRequest 覆盖基础验证（Embedding 特殊处理）
func (eh *EmbeddingHandler) ValidateRequest(req *HandlerRequest) error {
	if err := eh.BaseRelayHandler.ValidateRequest(req); err != nil {
		return err
	}

	// Embedding 特殊验证
	var embReq map[string]interface{}
	if err := json.Unmarshal(req.Body, &embReq); err != nil {
		return ErrInvalidRequest{Message: "invalid embedding request JSON"}
	}

	if _, ok := embReq["input"]; !ok {
		return ErrInvalidRequest{Message: "input field is required in embedding request"}
	}

	return nil
}

// ValidateRequest 覆盖基础验证（Image 特殊处理）
func (ih *ImageHandler) ValidateRequest(req *HandlerRequest) error {
	if err := ih.BaseRelayHandler.ValidateRequest(req); err != nil {
		return err
	}

	// Image 特殊验证
	var imgReq map[string]interface{}
	if err := json.Unmarshal(req.Body, &imgReq); err != nil {
		return ErrInvalidRequest{Message: "invalid image request JSON"}
	}

	if _, ok := imgReq["prompt"]; !ok {
		return ErrInvalidRequest{Message: "prompt field is required in image request"}
	}

	return nil
}

// ValidateRequest 覆盖基础验证（Audio 特殊处理）
func (ah *AudioHandler) ValidateRequest(req *HandlerRequest) error {
	if err := ah.BaseRelayHandler.ValidateRequest(req); err != nil {
		return err
	}

	// Audio 可能没有特殊验证或需要检查音频格式
	// 这里暂时保留基础验证

	return nil
}
