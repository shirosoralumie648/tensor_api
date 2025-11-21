package rag

import (
	"strings"
	"testing"
)

func TestChunkerEstimateTokens(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategyParagraph)

	tests := []struct {
		text     string
		minTokens int
	}{
		{"Hello world", 1},
		{"This is a longer text with multiple words", 5},
	}

	for _, test := range tests {
		tokens := chunker.estimateTokens(test.text)
		if tokens < test.minTokens {
			t.Errorf("EstimateTokens(%q) = %d, expected >= %d", test.text, tokens, test.minTokens)
		}
	}
}

func TestChunkerSplitByParagraph(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategyParagraph)

	text := "Paragraph 1\n\nParagraph 2\n\nParagraph 3"
	parts := chunker.splitByParagraph(text)

	if len(parts) != 3 {
		t.Errorf("Expected 3 paragraphs, got %d", len(parts))
	}
}

func TestChunkerSplitBySentence(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategySentence)

	text := "Sentence 1. Sentence 2. Sentence 3."
	parts := chunker.splitBySentence(text)

	if len(parts) < 2 {
		t.Errorf("Expected at least 2 sentences, got %d", len(parts))
	}
}

func TestChunkerSplitByFixed(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategyFixed)

	text := strings.Repeat("a", 5000)
	parts := chunker.splitByFixed(text)

	if len(parts) < 2 {
		t.Errorf("Expected multiple fixed chunks")
	}
}

func TestChunkerMergeChunks(t *testing.T) {
	chunker := NewChunker(500, 100, StrategyParagraph)

	parts := []string{
		"Short paragraph 1",
		"Short paragraph 2",
		"Short paragraph 3",
	}

	merged := chunker.mergeChunks(parts)

	if len(merged) == 0 {
		t.Errorf("Expected merged chunks")
	}
}

func TestChunkerChunkText(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategyParagraph)

	text := "Paragraph 1\n\nParagraph 2\n\nParagraph 3"
	chunks, err := chunker.ChunkText(text)

	if err != nil {
		t.Errorf("ChunkText failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Errorf("Expected at least 1 chunk")
	}

	for i, chunk := range chunks {
		if chunk.ChunkIndex != i {
			t.Errorf("Expected chunk index %d, got %d", i, chunk.ChunkIndex)
		}

		if chunk.TokenCount <= 0 {
			t.Errorf("Expected positive token count")
		}
	}
}

func TestChunkerChunkDocument(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategyParagraph)

	text := "Paragraph 1\n\nParagraph 2"
	chunks, err := chunker.ChunkDocument("doc-1", "Test Document", text, map[string]interface{}{"source": "test"})

	if err != nil {
		t.Errorf("ChunkDocument failed: %v", err)
	}

	for _, chunk := range chunks {
		if chunk.DocumentID != "doc-1" {
			t.Errorf("Expected document ID doc-1")
		}

		if chunk.Metadata["title"] != "Test Document" {
			t.Errorf("Expected title in metadata")
		}
	}
}

func TestChunkerGetStatistics(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategyParagraph)

	text := "Test paragraph"
	_, _ = chunker.ChunkText(text)

	stats := chunker.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if totalChunked, ok := stats["total_chunked"].(int64); !ok || totalChunked <= 0 {
		t.Errorf("Expected positive total_chunked")
	}
}

func TestChunkerWithDifferentStrategies(t *testing.T) {
	text := "Sentence 1. Sentence 2. Paragraph 2\n\nParagraph 3"

	strategies := []ChunkingStrategy{
		StrategyParagraph,
		StrategySentence,
		StrategyFixed,
		StrategyHybrid,
	}

	for _, strategy := range strategies {
		chunker := NewChunker(500, 100, strategy)
		chunks, err := chunker.ChunkText(text)

		if err != nil {
			t.Errorf("ChunkText with strategy %s failed: %v", strategy, err)
		}

		if len(chunks) == 0 {
			t.Errorf("Expected chunks with strategy %s", strategy)
		}
	}
}

func TestAdvancedChunker(t *testing.T) {
	chunker := NewAdvancedChunker(1000, 100)

	text := "# Title\n\nContent paragraph\n\n```code\nprint('hello')\n```"
	chunks, err := chunker.ChunkWithStructure(text)

	if err != nil {
		t.Errorf("ChunkWithStructure failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Errorf("Expected chunks with structure")
	}
}

func TestAdvancedChunkerDetectStructure(t *testing.T) {
	chunker := NewAdvancedChunker(1000, 100)

	// Test title detection
	titleChunk := &Chunk{Content: "# This is a title"}
	chunker.detectStructure(titleChunk, 0, "")

	if isTitle, ok := titleChunk.Metadata["is_title"].(bool); !ok || !isTitle {
		t.Errorf("Expected is_title to be true")
	}

	// Test code detection
	codeChunk := &Chunk{Content: "```python\ncode here\n```"}
	chunker.detectStructure(codeChunk, 0, "")

	if isCode, ok := codeChunk.Metadata["is_code"].(bool); !ok || !isCode {
		t.Errorf("Expected is_code to be true")
	}

	// Test table detection
	tableChunk := &Chunk{Content: "| Column1 | Column2 |\n| --- | --- |"}
	chunker.detectStructure(tableChunk, 0, "")

	if isTable, ok := tableChunk.Metadata["is_table"].(bool); !ok || !isTable {
		t.Errorf("Expected is_table to be true")
	}
}

func TestTokenCounter(t *testing.T) {
	counter := NewTokenCounter()

	tests := []struct {
		text     string
		minTokens int
	}{
		{"hello", 1},
		{"hello world", 2},
		{"This is a test", 3},
	}

	for _, test := range tests {
		tokens := counter.Count(test.text)
		if tokens < test.minTokens {
			t.Errorf("Count(%q) = %d, expected >= %d", test.text, tokens, test.minTokens)
		}
	}
}

func TestChunkerEmptyText(t *testing.T) {
	chunker := NewChunker(1000, 100, StrategyParagraph)

	chunks, err := chunker.ChunkText("")
	if err != nil {
		t.Errorf("ChunkText with empty text failed: %v", err)
	}

	if len(chunks) > 0 {
		t.Errorf("Expected no chunks for empty text")
	}
}

func TestChunkerLargeText(t *testing.T) {
	chunker := NewChunker(500, 100, StrategyParagraph)

	// Create a large text
	text := strings.Repeat("This is a test paragraph.\n\n", 100)
	chunks, err := chunker.ChunkText(text)

	if err != nil {
		t.Errorf("ChunkText with large text failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Errorf("Expected multiple chunks for large text")
	}
}

func BenchmarkChunkerEstimateTokens(b *testing.B) {
	chunker := NewChunker(1000, 100, StrategyParagraph)
	text := "This is a test paragraph with multiple words"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chunker.estimateTokens(text)
	}
}

func BenchmarkChunkerChunkText(b *testing.B) {
	chunker := NewChunker(1000, 100, StrategyParagraph)
	text := strings.Repeat("Paragraph\n\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chunker.ChunkText(text)
	}
}

func BenchmarkChunkerChunkDocument(b *testing.B) {
	chunker := NewChunker(1000, 100, StrategyParagraph)
	text := strings.Repeat("Paragraph\n\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chunker.ChunkDocument("doc-1", "Title", text, nil)
	}
}

