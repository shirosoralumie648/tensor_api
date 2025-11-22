package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/relay"
	"github.com/shirosoralumie648/Oblivious/backend/internal/repository"
	"github.com/shirosoralumie648/Oblivious/backend/pkg/logger"
	"go.uber.org/zap"
)

type ChatService struct {
	sessionRepo    *repository.SessionRepository
	messageRepo    *repository.MessageRepository
	relayService   *RelayService
	billingService *BillingService
}

func NewChatService() *ChatService {
	return &ChatService{
		sessionRepo:    repository.NewSessionRepository(),
		messageRepo:    repository.NewMessageRepository(),
		relayService:   NewRelayService(),
		billingService: NewBillingService(),
	}
}

type CreateSessionRequest struct {
	Title         string  `json:"title" binding:"required"`
	Model         string  `json:"model" binding:"required"`
	Temperature   float64 `json:"temperature"`
	SystemRole    string  `json:"system_role"`
	ContextLength int     `json:"context_length"`
}

type SendMessageRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
	Content   string    `json:"content" binding:"required"`
}

// CreateSession 创建会话
func (s *ChatService) CreateSession(ctx context.Context, userID int, req *CreateSessionRequest) (*model.Session, error) {
	session := &model.Session{
		UserID:        userID,
		Title:         req.Title,
		Model:         req.Model,
		Temperature:   req.Temperature,
		SystemRole:    req.SystemRole,
		ContextLength: req.ContextLength,
	}

	// 设置默认值
	if session.Temperature == 0 {
		session.Temperature = 0.7
	}
	if session.ContextLength == 0 {
		session.ContextLength = 4
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetUserSessions 获取用户的会话列表
func (s *ChatService) GetUserSessions(ctx context.Context, userID int, page, pageSize int) ([]*model.Session, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.sessionRepo.FindByUserID(ctx, userID, page, pageSize)
}

// GetSessionByID 获取会话详情
func (s *ChatService) GetSessionByID(ctx context.Context, sessionID uuid.UUID, userID int) (*model.Session, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// 检查权限
	if session.UserID != userID {
		return nil, fmt.Errorf("permission denied")
	}

	return session, nil
}

// GetSessionMessages 获取会话的消息列表
func (s *ChatService) GetSessionMessages(ctx context.Context, sessionID uuid.UUID, userID int, page, pageSize int) ([]*model.Message, int64, error) {
	// 先检查会话权限
	_, err := s.GetSessionByID(ctx, sessionID, userID)
	if err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	return s.messageRepo.FindBySessionID(ctx, sessionID, page, pageSize)
}

// SendMessage 发送消息（调用中转服务获取 AI 响应）
func (s *ChatService) SendMessage(ctx context.Context, userID int, req *SendMessageRequest) (*model.Message, error) {
	// 1. 查询会话并检查权限
	session, err := s.GetSessionByID(ctx, req.SessionID, userID)
	if err != nil {
		return nil, err
	}

	// 2. 创建用户消息
	userMsg := &model.Message{
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Content,
		Model:     session.Model,
		Metadata:  "{}",
		Files:     "[]",
		ToolCalls: "[]",
	}
	if err := s.messageRepo.Create(ctx, userMsg); err != nil {
		return nil, err
	}

	// 3. 获取上下文消息（为后续调用 AI 准备）
	contextMessages, err := s.messageRepo.GetContextMessages(ctx, req.SessionID, session.ContextLength*2)
	if err != nil {
		contextMessages = []*model.Message{}
	}

	// 4. 构建上下文消息列表
	// 添加系统角色
	messages := make([]interface{}, 0)

	if session.SystemRole != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": session.SystemRole,
		})
	}

	// 添加上下文消息
	for _, msg := range contextMessages {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 添加当前用户消息
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": req.Content,
	})

	// 5. 调用中转服务获取 AI 响应
	// 构建 Relay 请求
	relayMessages := make([]relay.ChatMessage, 0, len(messages))
	for _, msg := range messages {
		if m, ok := msg.(map[string]interface{}); ok {
			relayMessages = append(relayMessages, relay.ChatMessage{
				Role:    m["role"].(string),
				Content: m["content"].(string),
			})
		}
	}

	relayReq := &relay.ChatCompletionRequest{
		Model:       session.Model,
		Messages:    relayMessages,
		Temperature: session.Temperature,
		Stream:      false,
	}

	// 注入上下文长度限制（如果模型支持）
	if session.MaxTokens != nil {
		relayReq.MaxTokens = *session.MaxTokens
	}

	// 调用 Relay Service
	relayResp, err := s.relayService.RelayChatCompletion(ctx, relayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// 提取响应内容
	aiContent := ""
	if len(relayResp.Choices) > 0 {
		aiContent = relayResp.Choices[0].Message.Content
	}

	// 提取 Token 使用量
	inputTokens := relayResp.Usage.PromptTokens
	outputTokens := relayResp.Usage.CompletionTokens

	// 6. 创建 AI 消息
	aiMsg := &model.Message{
		SessionID:    req.SessionID,
		Role:         "assistant",
		Content:      aiContent,
		Model:        session.Model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		Metadata:     "{}",
		Files:        "[]",
		ToolCalls:    "[]",
	}
	if err := s.messageRepo.Create(ctx, aiMsg); err != nil {
		return nil, err
	}

	// 7. 处理计费（如果有 Token 使用）
	if inputTokens > 0 || outputTokens > 0 {
		_, err := s.billingService.Charge(ctx, userID, req.SessionID, aiMsg.ID, session.Model, inputTokens, outputTokens)
		if err != nil {
			// 计费失败不影响消息的返回，仅记录日志
			fmt.Printf("计费失败: %v\n", err)
		}
	}

	// 8. 更新会话的 updated_at
	session.UpdatedAt = aiMsg.CreatedAt
	_ = s.sessionRepo.Update(ctx, session)

	return aiMsg, nil
}

// UpdateSession 更新会话
func (s *ChatService) UpdateSession(ctx context.Context, userID int, sessionID uuid.UUID, title string) (*model.Session, error) {
	session, err := s.GetSessionByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	if title != "" {
		session.Title = title
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// DeleteSession 删除会话
func (s *ChatService) DeleteSession(ctx context.Context, userID int, sessionID uuid.UUID) error {
	// 检查权限
	_, err := s.GetSessionByID(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	return s.sessionRepo.Delete(ctx, sessionID)
}

// SendMessageStream 流式发送消息（SSE）
func (s *ChatService) SendMessageStream(ctx context.Context, userID int, req *SendMessageRequest, writer io.Writer) error {
	// 1. 查询会话并检查权限
	session, err := s.GetSessionByID(ctx, req.SessionID, userID)
	if err != nil {
		return err
	}

	// 2. 创建用户消息
	userMsg := &model.Message{
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Content,
		Model:     session.Model,
		Metadata:  "{}",
		Files:     "[]",
		ToolCalls: "[]",
	}
	if err := s.messageRepo.Create(ctx, userMsg); err != nil {
		return err
	}

	// 3. 获取上下文消息
	contextMessages, err := s.messageRepo.GetContextMessages(ctx, req.SessionID, session.ContextLength*2)
	if err != nil {
		contextMessages = []*model.Message{}
	}

	// 4. 构建对话消息列表（relay 格式）
	relayMessages := make([]relay.ChatMessage, 0)

	// 添加系统角色
	if session.SystemRole != "" {
		relayMessages = append(relayMessages, relay.ChatMessage{
			Role:    "system",
			Content: session.SystemRole,
		})
	}

	// 添加上下文消息
	for _, msg := range contextMessages {
		relayMessages = append(relayMessages, relay.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前用户消息
	relayMessages = append(relayMessages, relay.ChatMessage{
		Role:    "user",
		Content: req.Content,
	})

	// 5. 调用 Relay 服务的流式端点
	maxTokens := 0
	if session.MaxTokens != nil {
		maxTokens = *session.MaxTokens
	}

	relayReq := &relay.ChatCompletionRequest{
		Model:       session.Model,
		Messages:    relayMessages,
		Temperature: session.Temperature,
		TopP:        1.0,
		Stream:      true,
		MaxTokens:   maxTokens,
	}

	// 6. 收集流式响应
	fullContent := ""
	totalInputTokens := 0
	totalOutputTokens := 0

	// 通过流式处理函数接收 Relay 响应
	err = s.relayService.StreamChatCompletion(ctx, relayReq, func(chunk *relay.ChatCompletionResponse) error {
		// 提取流式数据
		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]

			// 累积内容
			if choice.Delta.Content != "" {
				fullContent += choice.Delta.Content
			}

			// 发送给客户端
			data := map[string]interface{}{
				"type":    "chunk",
				"content": choice.Delta.Content,
				"model":   chunk.Model,
			}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(writer, "data: %s\n\n", string(jsonData))

			// 检查是否完成
			if choice.FinishReason == "stop" {
				// 更新 token 统计
				if chunk.Usage.PromptTokens > 0 {
					totalInputTokens = chunk.Usage.PromptTokens
				}
				if chunk.Usage.CompletionTokens > 0 {
					totalOutputTokens = chunk.Usage.CompletionTokens
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Error("relay stream error", zap.Error(err))
		return err
	}

	// 7. 创建 AI 消息记录
	aiMsg := &model.Message{
		SessionID:    req.SessionID,
		Role:         "assistant",
		Content:      fullContent,
		Model:        session.Model,
		InputTokens:  totalInputTokens,
		OutputTokens: totalOutputTokens,
		TotalTokens:  totalInputTokens + totalOutputTokens,
		Metadata:     "{}",
		Files:        "[]",
		ToolCalls:    "[]",
	}
	if err := s.messageRepo.Create(ctx, aiMsg); err != nil {
		logger.Error("failed to create message", zap.Error(err))
		return err
	}

	// 8. 处理计费
	if totalInputTokens > 0 || totalOutputTokens > 0 {
		_, err := s.billingService.Charge(ctx, userID, req.SessionID, aiMsg.ID, session.Model, totalInputTokens, totalOutputTokens)
		if err != nil {
			logger.Error("billing error", zap.Error(err))
			// 计费失败不影响消息返回
		}
	}

	// 9. 更新会话时间
	session.UpdatedAt = aiMsg.CreatedAt
	_ = s.sessionRepo.Update(ctx, session)

	// 10. 发送最终消息事件
	finalMsg := map[string]interface{}{
		"type":          "complete",
		"message_id":    aiMsg.ID.String(),
		"content":       fullContent,
		"input_tokens":  totalInputTokens,
		"output_tokens": totalOutputTokens,
		"total_tokens":  totalInputTokens + totalOutputTokens,
	}
	jsonData, _ := json.Marshal(finalMsg)
	fmt.Fprintf(writer, "data: %s\n\n", string(jsonData))

	return nil
}
