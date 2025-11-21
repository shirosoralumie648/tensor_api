package rag

import (
	"context"
	"strings"
	"testing"
)

func TestDefaultRAGConfig(t *testing.T) {
	config := DefaultRAGConfig()

	if !config.Enabled {
		t.Errorf("Expected RAG to be enabled by default")
	}

	if config.TopK != 5 {
		t.Errorf("Expected TopK=5")
	}

	if config.VectorWeight != 0.7 {
		t.Errorf("Expected VectorWeight=0.7")
	}
}

func TestRAGServiceCreation(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	if ragService.config == nil {
		t.Errorf("Expected config to be initialized")
	}

	if !ragService.config.Enabled {
		t.Errorf("Expected RAG to be enabled")
	}
}

func TestRAGEnhancePrompt(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	ctx := context.Background()

	// Index some chunks
	chunks := []*Chunk{
		{ID: "1", Content: "Machine learning is a subset of artificial intelligence"},
		{ID: "2", Content: "Deep learning uses neural networks"},
	}
	retriever.IndexChunks(ctx, chunks)

	// Test enhancement
	enhanced, err := ragService.EnhancePrompt(ctx, "What is machine learning?")
	if err != nil {
		t.Errorf("EnhancePrompt failed: %v", err)
	}

	if enhanced == nil {
		t.Errorf("Expected enhanced prompt")
	}

	if enhanced.EnhancedPrompt == "" {
		t.Errorf("Expected non-empty enhanced prompt")
	}

	if len(enhanced.Citations) == 0 {
		t.Errorf("Expected citations")
	}
}

func TestRAGShouldUseRAG(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	// Test with minimum query length
	if !ragService.shouldUseRAG("This is a query that is long enough") {
		t.Errorf("Expected RAG to be used for long query")
	}

	if ragService.shouldUseRAG("short") {
		t.Errorf("Expected RAG not to be used for short query")
	}
}

func TestRAGFilterResults(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)
	ragService.config.MinRelevance = 0.5

	results := []*SearchResult{
		{ChunkID: "1", Content: "Test 1", Score: 0.8},
		{ChunkID: "2", Content: "Test 2", Score: 0.3},
		{ChunkID: "3", Content: "Test 3", Score: 0.6},
	}

	filtered := ragService.filterResults(results)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered results, got %d", len(filtered))
	}
}

func TestRAGBuildCitations(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	results := []*SearchResult{
		{
			ChunkID:  "1",
			Content:  "Test content",
			Score:    0.8,
			Metadata: map[string]interface{}{"title": "Doc1", "page": 1},
		},
	}

	citations := ragService.buildCitations(results)

	if len(citations) != 1 {
		t.Errorf("Expected 1 citation")
	}

	if citations[0].SourceName != "Doc1" {
		t.Errorf("Expected citation source name")
	}
}

func TestRAGCalculateQualityScore(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	// Empty results
	score := ragService.calculateQualityScore([]*SearchResult{})
	if score != 0 {
		t.Errorf("Expected 0 score for empty results")
	}

	// With results
	results := []*SearchResult{
		{Score: 0.8},
		{Score: 0.9},
	}
	score = ragService.calculateQualityScore(results)

	if score <= 0 || score > 1 {
		t.Errorf("Expected score between 0 and 1, got %f", score)
	}
}

func TestRAGGetStatistics(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	ctx := context.Background()

	// Index chunks
	chunks := []*Chunk{
		{ID: "1", Content: "Machine learning is a subset of artificial intelligence"},
	}
	retriever.IndexChunks(ctx, chunks)

	// Enhance prompt
	ragService.EnhancePrompt(ctx, "What is machine learning?")

	stats := ragService.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if totalRAGs, ok := stats["total_rags"].(int64); !ok || totalRAGs <= 0 {
		t.Errorf("Expected positive total_rags")
	}
}

func TestRAGEnabledChat(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)
	chat := NewRAGEnabledChat(ragService)

	ctx := context.Background()

	// Index chunks
	chunks := []*Chunk{
		{ID: "1", Content: "Machine learning is great"},
	}
	retriever.IndexChunks(ctx, chunks)

	// Process query
	result, err := chat.ProcessQuery(ctx, "Tell me about machine learning", true)
	if err != nil {
		t.Errorf("ProcessQuery failed: %v", err)
	}

	if result.Query != "Tell me about machine learning" {
		t.Errorf("Expected correct query")
	}
}

func TestRAGQualityVerifier(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)
	verifier := NewRAGQualityVerifier(ragService, 0.5)

	// Test high quality
	enhanced := &EnhancedPrompt{
		QualityScore: 0.8,
	}

	if !verifier.Verify(enhanced) {
		t.Errorf("Expected high quality to pass")
	}

	// Test low quality
	enhanced.QualityScore = 0.3

	if verifier.Verify(enhanced) {
		t.Errorf("Expected low quality to fail")
	}

	// Check stats
	stats := verifier.GetStats()

	if passRate, ok := stats["pass_rate"].(float32); !ok || passRate <= 0 {
		t.Errorf("Expected positive pass rate")
	}
}

func TestRAGSetConfig(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	newConfig := &RAGConfig{
		Enabled:   false,
		TopK:      10,
		RetrievalMethod: "vector",
	}

	ragService.SetConfig(newConfig)

	if ragService.config.TopK != 10 {
		t.Errorf("Expected config to be updated")
	}
}

func TestRAGGenerateEnhancedPrompt(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	results := []*SearchResult{
		{
			Content:  "Machine learning basics",
			Score:    0.9,
			Metadata: map[string]interface{}{"title": "ML Guide"},
		},
	}

	enhanced := ragService.generateEnhancedPrompt("What is ML?", results)

	if !strings.Contains(enhanced, "Machine learning basics") {
		t.Errorf("Expected enhanced prompt to contain content")
	}

	if !strings.Contains(enhanced, "What is ML?") {
		t.Errorf("Expected enhanced prompt to contain query")
	}
}

func TestRAGMultipleQueries(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	ctx := context.Background()

	// Index chunks
	for i := 0; i < 5; i++ {
		chunk := &Chunk{
			ID:      string(rune(i)),
			Content: "Test content for query",
		}
		retriever.IndexChunk(ctx, chunk)
	}

	// Process multiple queries
	for i := 0; i < 10; i++ {
		ragService.EnhancePrompt(ctx, "Test query that is long enough")
	}

	stats := ragService.GetStatistics()

	if totalRAGs, ok := stats["total_rags"].(int64); !ok || totalRAGs != 10 {
		t.Errorf("Expected 10 total RAGs")
	}
}

func BenchmarkRAGEnhancePrompt(b *testing.B) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)

	ctx := context.Background()

	// Index chunks
	for i := 0; i < 100; i++ {
		chunk := &Chunk{
			ID:      string(rune(i)),
			Content: "Machine learning and artificial intelligence concepts",
		}
		retriever.IndexChunk(ctx, chunk)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ragService.EnhancePrompt(ctx, "What is machine learning and how does it work?")
	}
}

func BenchmarkRAGQualityVerify(b *testing.B) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)
	ragService := NewRAGService(retriever, nil)
	verifier := NewRAGQualityVerifier(ragService, 0.5)

	enhanced := &EnhancedPrompt{QualityScore: 0.8}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifier.Verify(enhanced)
	}
}

