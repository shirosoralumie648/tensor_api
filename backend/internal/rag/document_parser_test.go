package rag

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"
)

func TestTextParser(t *testing.T) {
	parser := &TextParser{}

	content := "Hello, this is a test document.\nLine 2\nLine 3"
	reader := strings.NewReader(content)

	parsed, err := parser.Parse(reader, "test.txt")
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if parsed.Content != content {
		t.Errorf("Expected content to match")
	}

	if parsed.Title != "test.txt" {
		t.Errorf("Expected title test.txt")
	}
}

func TestJSONParser(t *testing.T) {
	parser := &JSONParser{}

	content := `{"key": "value", "number": 42}`
	reader := strings.NewReader(content)

	parsed, err := parser.Parse(reader, "test.json")
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if parsed.Content != content {
		t.Errorf("Expected content to match")
	}
}

func TestCSVParser(t *testing.T) {
	parser := &CSVParser{}

	content := "name,age\nAlice,30\nBob,25"
	reader := strings.NewReader(content)

	parsed, err := parser.Parse(reader, "test.csv")
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if parsed.Content != content {
		t.Errorf("Expected content to match")
	}

	if parsed.Metadata["type"] != "csv" {
		t.Errorf("Expected csv type in metadata")
	}
}

func TestHTMLParser(t *testing.T) {
	parser := &HTMLParser{}

	content := "<html><body><p>Test content</p></body></html>"
	reader := strings.NewReader(content)

	parsed, err := parser.Parse(reader, "test.html")
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	// Should strip HTML tags
	if strings.Contains(parsed.Content, "<") {
		t.Errorf("Expected HTML tags to be stripped")
	}
}

func TestXMLParser(t *testing.T) {
	parser := &XMLParser{}

	content := `<root><item>test</item></root>`
	reader := strings.NewReader(content)

	parsed, err := parser.Parse(reader, "test.xml")
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if parsed.Content != content {
		t.Errorf("Expected content to match")
	}
}

func TestYAMLParser(t *testing.T) {
	parser := &YAMLParser{}

	content := "key: value\nnumber: 42"
	reader := strings.NewReader(content)

	parsed, err := parser.Parse(reader, "test.yaml")
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if parsed.Content != content {
		t.Errorf("Expected content to match")
	}
}

func TestDocumentParserRegistry(t *testing.T) {
	registry := NewDocumentParserRegistry()

	types := registry.GetSupportedTypes()
	if len(types) == 0 {
		t.Errorf("Expected supported types")
	}

	// Test parsing with registry
	content := "Test content"
	reader := strings.NewReader(content)

	parsed, err := registry.Parse(reader, "test.txt")
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if parsed.Content != content {
		t.Errorf("Expected content to match")
	}
}

func TestDocumentParserRegistryUnsupported(t *testing.T) {
	registry := NewDocumentParserRegistry()

	content := "Test content"
	reader := strings.NewReader(content)

	_, err := registry.Parse(reader, "test.unknown")
	if err == nil {
		t.Errorf("Expected error for unsupported type")
	}
}

func TestDocumentUploadManager(t *testing.T) {
	dum := NewDocumentUploadManager(1024 * 1024) // 1MB limit

	// Create a test file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Errorf("CreateFormFile failed: %v", err)
	}

	content := "Test document content"
	if _, err := io.WriteString(part, content); err != nil {
		t.Errorf("WriteString failed: %v", err)
	}

	writer.Close()

	// Parse the multipart body
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1024 * 1024)
	if err != nil {
		t.Errorf("ReadForm failed: %v", err)
	}

	if len(form.File["file"]) > 0 {
		fileHeader := form.File["file"][0]

		info, err := dum.UploadFile(fileHeader)
		if err != nil {
			t.Errorf("UploadFile failed: %v", err)
		}

		if info.Status != "completed" {
			t.Errorf("Expected status completed, got %s", info.Status)
		}

		if info.ParsedContent.Content != content {
			t.Errorf("Expected content to match")
		}
	}
}

func TestDocumentUploadManagerLargeFile(t *testing.T) {
	dum := NewDocumentUploadManager(100) // 100 byte limit

	// Create a large test file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Errorf("CreateFormFile failed: %v", err)
	}

	content := strings.Repeat("a", 200) // 200 bytes
	if _, err := io.WriteString(part, content); err != nil {
		t.Errorf("WriteString failed: %v", err)
	}

	writer.Close()

	// Parse the multipart body
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1024)
	if err != nil {
		// This is expected to fail
		return
	}

	if len(form.File["file"]) > 0 {
		fileHeader := form.File["file"][0]

		_, err := dum.UploadFile(fileHeader)
		if err == nil {
			t.Errorf("Expected error for file exceeding size limit")
		}
	}
}

func TestDocumentUploadHistory(t *testing.T) {
	dum := NewDocumentUploadManager(1024 * 1024)

	types := dum.GetSupportedTypes()
	if len(types) == 0 {
		t.Errorf("Expected supported types")
	}

	if !contains(types, ".txt") {
		t.Errorf("Expected .txt to be supported")
	}
}

func TestParserNames(t *testing.T) {
	parsers := []DocumentParser{
		&TextParser{},
		&JSONParser{},
		&CSVParser{},
		&HTMLParser{},
		&XMLParser{},
		&YAMLParser{},
	}

	for _, parser := range parsers {
		if parser.Name() == "" {
			t.Errorf("Parser name should not be empty")
		}

		if len(parser.SupportedTypes()) == 0 {
			t.Errorf("Parser should support at least one type")
		}
	}
}

func BenchmarkTextParser(b *testing.B) {
	parser := &TextParser{}
	content := strings.Repeat("Test line\n", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		_, _ = parser.Parse(reader, "test.txt")
	}
}

func BenchmarkJSONParser(b *testing.B) {
	parser := &JSONParser{}
	content := `{"key": "value", "number": 42}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		_, _ = parser.Parse(reader, "test.json")
	}
}

func BenchmarkDocumentParserRegistry(b *testing.B) {
	registry := NewDocumentParserRegistry()
	content := "Test content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		_, _ = registry.Parse(reader, "test.txt")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

