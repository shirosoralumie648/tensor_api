package tokenizer

import (
	"context"
	"fmt"
)

// ExampleUsage Token计数器使用示例
func ExampleUsage() {
	ctx := context.Background()

	// 1. 创建工厂
	factory, err := NewTokenizerFactory()
	if err != nil {
		panic(err)
	}
	defer factory.Close()

	// 2. 计算OpenAI模型的Token
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: "Hello, how are you?",
		},
	}

	tokenizer, _ := factory.GetTokenizer("gpt-4o")
	count, _ := tokenizer.CountMessages(ctx, messages, "gpt-4o")
	fmt.Printf("GPT-4o token count: %d\n", count)

	// 3. 计算Claude模型的Token
	claudeTokenizer, _ := factory.GetTokenizer("claude-3.5-sonnet")
	claudeCount, _ := claudeTokenizer.CountMessages(ctx, messages, "claude-3.5-sonnet")
	fmt.Printf("Claude token count: %d\n", claudeCount)

	// 4. 使用完整的请求计数
	req := &TokenCountRequest{
		Model:    "gpt-4o",
		Messages: messages,
		ImageDetails: []ImageDetail{
			{
				Width:  1024,
				Height: 1024,
				Detail: "high",
			},
		},
	}

	result, _ := tokenizer.CountTokens(ctx, req)
	fmt.Printf("Total tokens: %d (prompt: %d)\n", result.TotalTokens, result.PromptTokens)

	// 5. 流式Token计数
	streamCounter, _ := factory.CreateStreamCounter("gpt-4o")
	streamCounter.AddChunk("Hello")
	streamCounter.AddChunk(" world")
	streamCounter.AddChunk("!")

	finalCount := streamCounter.Finalize()
	fmt.Printf("Stream token count: %d\n", finalCount)

	// 6. 批量流式计数
	batchCounter := factory.CreateBatchStreamCounter()
	batchCounter.CreateCounter("user1", "gpt-4o")
	batchCounter.CreateCounter("user2", "gpt-4o")

	batchCounter.AddChunk("user1", "Hello from user 1")
	batchCounter.AddChunk("user2", "Hello from user 2")

	count1 := batchCounter.Finalize("user1")
	count2 := batchCounter.Finalize("user2")

	fmt.Printf("User1: %d tokens, User2: %d tokens\n", count1, count2)

	// 7. 快速计数（使用全局工厂）
	quickCount, _ := CountTokensQuick("gpt-4o", messages)
	fmt.Printf("Quick count: %d tokens\n", quickCount)
}

// ExampleMultimodal 多模态Token计数示例
func ExampleMultimodal() {
	ctx := context.Background()
	factory, _ := NewTokenizerFactory()
	defer factory.Close()

	// 包含图片和文本的消息
	messages := []Message{
		{
			Role: "user",
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "What's in this image?",
				},
				map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url":    "https://example.com/image.jpg",
						"detail": "high",
					},
				},
			},
		},
	}

	req := &TokenCountRequest{
		Model:    "gpt-4o",
		Messages: messages,
		ImageDetails: []ImageDetail{
			{
				Width:  2048,
				Height: 1536,
				Detail: "high",
			},
		},
	}

	tokenizer, _ := factory.GetTokenizer("gpt-4o")
	result, _ := tokenizer.CountTokens(ctx, req)

	fmt.Printf("Multimodal message tokens: %d\n", result.TotalTokens)
}

// ExampleChineseModels 中文模型Token计数示例
func ExampleChineseModels() {
	ctx := context.Background()
	factory, _ := NewTokenizerFactory()
	defer factory.Close()

	messages := []Message{
		{
			Role:    "user",
			Content: "你好，请帮我写一首诗",
		},
	}

	// 通义千问
	qwenTokenizer, _ := factory.GetTokenizer("qwen-max")
	qwenCount, _ := qwenTokenizer.CountMessages(ctx, messages, "qwen-max")
	fmt.Printf("Qwen token count: %d\n", qwenCount)

	// 智谱GLM
	glmTokenizer, _ := factory.GetTokenizer("glm-4")
	glmCount, _ := glmTokenizer.CountMessages(ctx, messages, "glm-4")
	fmt.Printf("GLM token count: %d\n", glmCount)

	// DeepSeek
	deepseekTokenizer, _ := factory.GetTokenizer("deepseek-chat")
	deepseekCount, _ := deepseekTokenizer.CountMessages(ctx, messages, "deepseek-chat")
	fmt.Printf("DeepSeek token count: %d\n", deepseekCount)
}
