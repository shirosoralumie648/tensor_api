package rag

import (
	"context"
	"testing"
)

func TestBM25Creation(t *testing.T) {
	bm25 := NewBM25()

	if bm25.K1 != 1.5 {
		t.Errorf("Expected K1=1.5")
	}

	if bm25.B != 0.75 {
		t.Errorf("Expected B=0.75")
	}
}

func TestBM25CalculateIDF(t *testing.T) {
	bm25 := NewBM25()

	idf := bm25.calculateIDF("test", 100, 10)

	if idf <= 0 {
		t.Errorf("Expected positive IDF")
	}
}

func TestRetrieverIndexChunk(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()
	chunk := &Chunk{
		ID:      "chunk-1",
		Content: "This is a test chunk",
	}

	err := retriever.IndexChunk(ctx, chunk)
	if err != nil {
		t.Errorf("IndexChunk failed: %v", err)
	}

	// Verify chunk is stored
	retriever.chunkMu.RLock()
	_, exists := retriever.chunkStore["chunk-1"]
	retriever.chunkMu.RUnlock()

	if !exists {
		t.Errorf("Expected chunk to be stored")
	}
}

func TestRetrieverIndexChunks(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()
	chunks := []*Chunk{
		{ID: "chunk-1", Content: "First chunk"},
		{ID: "chunk-2", Content: "Second chunk"},
	}

	err := retriever.IndexChunks(ctx, chunks)
	if err != nil {
		t.Errorf("IndexChunks failed: %v", err)
	}

	retriever.chunkMu.RLock()
	if len(retriever.chunkStore) != 2 {
		t.Errorf("Expected 2 chunks stored")
	}
	retriever.chunkMu.RUnlock()
}

func TestRetrieverVectorSearch(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	// Index chunks
	chunks := []*Chunk{
		{ID: "chunk-1", Content: "Machine learning is great"},
		{ID: "chunk-2", Content: "Deep learning algorithms"},
		{ID: "chunk-3", Content: "Cat and dog"},
	}
	retriever.IndexChunks(ctx, chunks)

	// Search
	results, err := retriever.VectorSearch(ctx, "machine learning", 2)
	if err != nil {
		t.Errorf("VectorSearch failed: %v", err)
	}

	if len(results) == 0 {
		t.Errorf("Expected search results")
	}

	for _, result := range results {
		if result.Method != "vector" {
			t.Errorf("Expected method=vector")
		}
	}
}

func TestRetrieverBM25Search(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	// Index chunks
	chunks := []*Chunk{
		{ID: "chunk-1", Content: "Machine learning is great"},
		{ID: "chunk-2", Content: "Deep learning algorithms"},
	}
	retriever.IndexChunks(ctx, chunks)

	// Search
	results, err := retriever.BM25Search(ctx, "machine learning", 5)
	if err != nil {
		t.Errorf("BM25Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Errorf("Expected search results")
	}

	for _, result := range results {
		if result.Method != "bm25" {
			t.Errorf("Expected method=bm25")
		}
	}
}

func TestRetrieverHybridSearch(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	// Index chunks
	chunks := []*Chunk{
		{ID: "chunk-1", Content: "Machine learning is great"},
		{ID: "chunk-2", Content: "Deep learning algorithms"},
	}
	retriever.IndexChunks(ctx, chunks)

	// Hybrid search
	results, err := retriever.HybridSearch(ctx, "machine learning", 5, 0.7)
	if err != nil {
		t.Errorf("HybridSearch failed: %v", err)
	}

	if len(results) == 0 {
		t.Errorf("Expected search results")
	}

	for _, result := range results {
		if result.Method != "hybrid" {
			t.Errorf("Expected method=hybrid")
		}
	}
}

func TestRetrieverDeleteChunk(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	chunk := &Chunk{ID: "chunk-1", Content: "Test"}
	retriever.IndexChunk(ctx, chunk)
	retriever.DeleteChunk(ctx, "chunk-1")

	retriever.chunkMu.RLock()
	_, exists := retriever.chunkStore["chunk-1"]
	retriever.chunkMu.RUnlock()

	if exists {
		t.Errorf("Expected chunk to be deleted")
	}
}

func TestRetrieverStatistics(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()
	chunk := &Chunk{ID: "chunk-1", Content: "Test"}
	retriever.IndexChunk(ctx, chunk)

	stats := retriever.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if indexed, ok := stats["indexed_chunks"].(int); !ok || indexed <= 0 {
		t.Errorf("Expected positive indexed_chunks")
	}
}

func TestRerankingService(t *testing.T) {
	service := NewRerankingService("test-model")

	results := []*SearchResult{
		{ChunkID: "1", Content: "Machine learning is great", Score: 0.8},
		{ChunkID: "2", Content: "Deep learning", Score: 0.7},
		{ChunkID: "3", Content: "Cat and dog", Score: 0.5},
	}

	reranked := service.Rerank("machine learning", results, 2)

	if len(reranked) != 2 {
		t.Errorf("Expected 2 reranked results")
	}

	if reranked[0].Rank != 1 {
		t.Errorf("Expected first result to have rank 1")
	}
}

func TestRetrieverMultipleSearches(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	// Index many chunks
	for i := 0; i < 10; i++ {
		chunk := &Chunk{
			ID:      string(rune(i)),
			Content: "Test content chunk",
		}
		retriever.IndexChunk(ctx, chunk)
	}

	// Perform multiple searches
	for i := 0; i < 5; i++ {
		_, err := retriever.VectorSearch(ctx, "test", 3)
		if err != nil {
			t.Errorf("VectorSearch failed: %v", err)
		}

		_, err = retriever.BM25Search(ctx, "test", 3)
		if err != nil {
			t.Errorf("BM25Search failed: %v", err)
		}
	}

	stats := retriever.GetStatistics()
	if searches, ok := stats["total_searches"].(int64); !ok || searches < 10 {
		t.Errorf("Expected at least 10 searches")
	}
}

func TestSearchResultScoring(t *testing.T) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	chunks := []*Chunk{
		{ID: "1", Content: "query query query"},
		{ID: "2", Content: "query other"},
		{ID: "3", Content: "unrelated content"},
	}
	retriever.IndexChunks(ctx, chunks)

	results, _ := retriever.VectorSearch(ctx, "query", 3)

	if len(results) > 0 && results[0].Score > 0.5 {
		t.Errorf("Expected reasonable scores")
	}
}

func BenchmarkRetrieverVectorSearch(b *testing.B) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	// Index chunks
	for i := 0; i < 100; i++ {
		chunk := &Chunk{
			ID:      string(rune(i)),
			Content: "Test content for chunk",
		}
		retriever.IndexChunk(ctx, chunk)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = retriever.VectorSearch(ctx, "test", 10)
	}
}

func BenchmarkBM25Search(b *testing.B) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	// Index chunks
	for i := 0; i < 100; i++ {
		chunk := &Chunk{
			ID:      string(rune(i)),
			Content: "Test content for chunk",
		}
		retriever.IndexChunk(ctx, chunk)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = retriever.BM25Search(ctx, "test", 10)
	}
}

func BenchmarkHybridSearch(b *testing.B) {
	store := NewInMemoryVectorStore()
	client := &MockEmbeddingClient{model: ModelAdaV2}
	service := NewEmbeddingService(ModelAdaV2, client)
	retriever := NewRetriever(store, service)

	ctx := context.Background()

	// Index chunks
	for i := 0; i < 100; i++ {
		chunk := &Chunk{
			ID:      string(rune(i)),
			Content: "Test content for chunk",
		}
		retriever.IndexChunk(ctx, chunk)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = retriever.HybridSearch(ctx, "test", 10, 0.7)
	}
}

