package rag

import (
	"context"
	"testing"
)

func TestEmbeddingModel(t *testing.T) {
	models := []*EmbeddingModel{
		ModelAdaV2,
		Model3Small,
		Model3Large,
	}

	for _, model := range models {
		if model.Name == "" {
			t.Errorf("Model name should not be empty")
		}

		if model.Dimension <= 0 {
			t.Errorf("Model dimension should be positive")
		}

		if model.MaxTokens <= 0 {
			t.Errorf("Model max tokens should be positive")
		}
	}
}

func TestMockEmbeddingClient(t *testing.T) {
	client := &MockEmbeddingClient{model: ModelAdaV2}

	ctx := context.Background()
	text := "This is a test sentence"

	vec, tokens, err := client.Embed(ctx, text)
	if err != nil {
		t.Errorf("Embed failed: %v", err)
	}

	if len(vec) != ModelAdaV2.Dimension {
		t.Errorf("Expected dimension %d, got %d", ModelAdaV2.Dimension, len(vec))
	}

	if tokens <= 0 {
		t.Errorf("Expected positive token count")
	}
}

func TestMockEmbeddingClientBatch(t *testing.T) {
	client := &MockEmbeddingClient{model: ModelAdaV2}

	ctx := context.Background()
	texts := []string{
		"First sentence",
		"Second sentence",
		"Third sentence",
	}

	vectors, tokens, err := client.EmbedBatch(ctx, texts)
	if err != nil {
		t.Errorf("EmbedBatch failed: %v", err)
	}

	if len(vectors) != len(texts) {
		t.Errorf("Expected %d vectors, got %d", len(texts), len(vectors))
	}

	if tokens <= 0 {
		t.Errorf("Expected positive token count")
	}

	for i, vec := range vectors {
		if len(vec) != ModelAdaV2.Dimension {
			t.Errorf("Vector %d has wrong dimension", i)
		}
	}
}

func TestEmbeddingService(t *testing.T) {
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)

	ctx := context.Background()
	text := "This is a test sentence"

	embedding, err := service.Embed(ctx, text)
	if err != nil {
		t.Errorf("Embed failed: %v", err)
	}

	if len(embedding.Vector) != ModelAdaV2.Dimension {
		t.Errorf("Expected dimension %d", ModelAdaV2.Dimension)
	}

	if embedding.Model != ModelAdaV2.Name {
		t.Errorf("Expected model %s", ModelAdaV2.Name)
	}
}

func TestEmbeddingServiceBatch(t *testing.T) {
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)

	ctx := context.Background()
	texts := []string{
		"First sentence",
		"Second sentence",
	}

	embeddings, err := service.EmbedBatch(ctx, texts)
	if err != nil {
		t.Errorf("EmbedBatch failed: %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Errorf("Expected %d embeddings", len(texts))
	}
}

func TestEmbeddingServiceCache(t *testing.T) {
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)

	ctx := context.Background()
	text := "This is a test sentence"

	// 第一次调用
	embedding1, _ := service.Embed(ctx, text)

	// 第二次调用应该从缓存获取
	embedding2, _ := service.Embed(ctx, text)

	if embedding1.ID != embedding2.ID {
		t.Errorf("Expected same ID from cache")
	}
}

func TestEmbeddingServiceStatistics(t *testing.T) {
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)

	ctx := context.Background()
	service.Embed(ctx, "Test")

	stats := service.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if total, ok := stats["total_embeddings"].(int64); !ok || total <= 0 {
		t.Errorf("Expected positive total_embeddings")
	}
}

func TestEmbeddingCache(t *testing.T) {
	cache := NewEmbeddingCache(100)

	embedding := &Embedding{
		ID:     "test-1",
		Vector: make([]float32, 10),
	}

	cache.Set("text", "model", embedding)

	retrieved, exists := cache.Get("text", "model")
	if !exists {
		t.Errorf("Expected cached embedding")
	}

	if retrieved.ID != "test-1" {
		t.Errorf("Expected embedding with ID test-1")
	}
}

func TestEmbeddingCacheMiss(t *testing.T) {
	cache := NewEmbeddingCache(100)

	_, exists := cache.Get("nonexistent", "model")
	if exists {
		t.Errorf("Expected cache miss")
	}
}

func TestEmbeddingCacheLRU(t *testing.T) {
	cache := NewEmbeddingCache(2)

	e1 := &Embedding{ID: "1"}
	e2 := &Embedding{ID: "2"}
	e3 := &Embedding{ID: "3"}

	cache.Set("text1", "model", e1)
	cache.Set("text2", "model", e2)
	cache.Set("text3", "model", e3) // Should evict least recently used

	if cache.Size() > 2 {
		t.Errorf("Cache should not exceed max size")
	}
}

func TestEmbeddingCacheClear(t *testing.T) {
	cache := NewEmbeddingCache(100)

	cache.Set("text", "model", &Embedding{ID: "1"})
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected empty cache after clear")
	}
}

func TestInMemoryVectorStore(t *testing.T) {
	store := NewInMemoryVectorStore()

	ctx := context.Background()
	embedding := &Embedding{
		ID:      "emb-1",
		ChunkID: "chunk-1",
		Vector:  make([]float32, 10),
	}

	err := store.SaveEmbedding(ctx, embedding)
	if err != nil {
		t.Errorf("SaveEmbedding failed: %v", err)
	}

	retrieved, err := store.GetEmbedding(ctx, "emb-1")
	if err != nil {
		t.Errorf("GetEmbedding failed: %v", err)
	}

	if retrieved.ID != "emb-1" {
		t.Errorf("Expected embedding with ID emb-1")
	}
}

func TestInMemoryVectorStoreBatch(t *testing.T) {
	store := NewInMemoryVectorStore()

	ctx := context.Background()
	embeddings := []*Embedding{
		{ID: "emb-1", Vector: make([]float32, 10)},
		{ID: "emb-2", Vector: make([]float32, 10)},
	}

	err := store.SaveEmbeddings(ctx, embeddings)
	if err != nil {
		t.Errorf("SaveEmbeddings failed: %v", err)
	}
}

func TestInMemoryVectorStoreDelete(t *testing.T) {
	store := NewInMemoryVectorStore()

	ctx := context.Background()
	embedding := &Embedding{ID: "emb-1", Vector: make([]float32, 10)}

	store.SaveEmbedding(ctx, embedding)
	store.DeleteEmbedding(ctx, "emb-1")

	_, err := store.GetEmbedding(ctx, "emb-1")
	if err == nil {
		t.Errorf("Expected error after deletion")
	}
}

func TestInMemoryVectorStoreDeleteByChunkID(t *testing.T) {
	store := NewInMemoryVectorStore()

	ctx := context.Background()
	e1 := &Embedding{ID: "emb-1", ChunkID: "chunk-1", Vector: make([]float32, 10)}
	e2 := &Embedding{ID: "emb-2", ChunkID: "chunk-1", Vector: make([]float32, 10)}

	store.SaveEmbedding(ctx, e1)
	store.SaveEmbedding(ctx, e2)
	store.DeleteByChunkID(ctx, "chunk-1")

	_, err1 := store.GetEmbedding(ctx, "emb-1")
	_, err2 := store.GetEmbedding(ctx, "emb-2")

	if err1 == nil || err2 == nil {
		t.Errorf("Expected both embeddings to be deleted")
	}
}

func TestInMemoryVectorStoreSearch(t *testing.T) {
	store := NewInMemoryVectorStore()

	ctx := context.Background()

	// 创建测试向量
	vec1 := make([]float32, 3)
	vec1[0] = 1
	vec2 := make([]float32, 3)
	vec2[0] = 1
	vec3 := make([]float32, 3)
	vec3[0] = 0

	store.SaveEmbedding(ctx, &Embedding{ID: "1", Vector: vec1})
	store.SaveEmbedding(ctx, &Embedding{ID: "2", Vector: vec2})
	store.SaveEmbedding(ctx, &Embedding{ID: "3", Vector: vec3})

	query := make([]float32, 3)
	query[0] = 1

	results, err := store.Search(ctx, query, 2)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results")
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	c := []float32{0, 1, 0}

	sim1 := cosineSimilarity(a, b)
	sim2 := cosineSimilarity(a, c)

	if sim1 <= sim2 {
		t.Errorf("Same vectors should have higher similarity")
	}
}

func BenchmarkEmbedding(b *testing.B) {
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Embed(ctx, "This is a test sentence")
	}
}

func BenchmarkBatchEmbedding(b *testing.B) {
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)

	ctx := context.Background()
	texts := []string{
		"First sentence",
		"Second sentence",
		"Third sentence",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.EmbedBatch(ctx, texts)
	}
}

func BenchmarkStoreVectorSearch(b *testing.B) {
	store := NewInMemoryVectorStore()
	ctx := context.Background()

	// 预加载向量
	for i := 0; i < 1000; i++ {
		vec := make([]float32, 1536)
		vec[0] = float32(i)
		store.SaveEmbedding(ctx, &Embedding{
			ID:     string(rune(i)),
			Vector: vec,
		})
	}

	query := make([]float32, 1536)
	query[0] = 500

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Search(ctx, query, 10)
	}
}

