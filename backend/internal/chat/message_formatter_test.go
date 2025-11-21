package chat

import (
	"strings"
	"testing"
)

func TestLanguageHighlighter(t *testing.T) {
	lh := NewLanguageHighlighter()

	tests := []struct {
		language string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"javascript", true},
		{"unknown", false},
	}

	for _, test := range tests {
		result := lh.IsSupported(test.language)
		if result != test.expected {
			t.Errorf("IsSupported(%s) = %v, expected %v", test.language, result, test.expected)
		}
	}
}

func TestLanguageHighlighterRegister(t *testing.T) {
	lh := NewLanguageHighlighter()

	lh.RegisterLanguage("customlang")

	if !lh.IsSupported("customlang") {
		t.Errorf("Expected customlang to be supported after registration")
	}
}

func TestCodeExtractor(t *testing.T) {
	ce := NewCodeExtractor()

	content := "```go\nfunc main() {\n  fmt.Println(\"Hello\")\n}\n```"

	codeBlocks := ce.ExtractCodeBlocks(content)

	if len(codeBlocks) != 1 {
		t.Errorf("Expected 1 code block, got %d", len(codeBlocks))
	}

	if codeBlocks[0].Language != "go" {
		t.Errorf("Expected language go, got %s", codeBlocks[0].Language)
	}
}

func TestCodeExtractorMultiple(t *testing.T) {
	ce := NewCodeExtractor()

	content := "```go\ncode1\n```\n\n```python\ncode2\n```"

	codeBlocks := ce.ExtractCodeBlocks(content)

	if len(codeBlocks) != 2 {
		t.Errorf("Expected 2 code blocks, got %d", len(codeBlocks))
	}
}

func TestCodeExtractorInline(t *testing.T) {
	ce := NewCodeExtractor()

	content := "Use `variable` in your code and `function()` are useful."

	codes := ce.ExtractInlineCode(content)

	if len(codes) != 2 {
		t.Errorf("Expected 2 inline codes, got %d", len(codes))
	}
}

func TestFormulaDetector(t *testing.T) {
	fd := NewFormulaDetector()

	content := "The equation is $$x^2 + y^2 = z^2$$ which is important."

	if !fd.HasFormula(content) {
		t.Errorf("Expected formula detection")
	}
}

func TestFormulaDetectorInline(t *testing.T) {
	fd := NewFormulaDetector()

	content := "The value is $x = 5$ in the equation."

	if !fd.HasFormula(content) {
		t.Errorf("Expected inline formula detection")
	}
}

func TestFormulaExtract(t *testing.T) {
	fd := NewFormulaDetector()

	content := "Formula: $$E = mc^2$$ and $a = 5$"

	formulas := fd.ExtractFormulas(content)

	if len(formulas) < 1 {
		t.Errorf("Expected at least 1 formula")
	}
}

func TestMarkdownProcessorHTML(t *testing.T) {
	mp := NewMarkdownProcessor()

	content := "**bold** and *italic*"

	html := mp.ToHTML(content)

	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Errorf("Expected bold HTML, got %s", html)
	}

	if !strings.Contains(html, "<em>italic</em>") {
		t.Errorf("Expected italic HTML, got %s", html)
	}
}

func TestMarkdownProcessorLink(t *testing.T) {
	mp := NewMarkdownProcessor()

	content := "[Google](https://google.com)"

	html := mp.ToHTML(content)

	if !strings.Contains(html, `<a href="https://google.com">Google</a>`) {
		t.Errorf("Expected link HTML, got %s", html)
	}
}

func TestMessageFormatterFormat(t *testing.T) {
	mf := NewMessageFormatter()

	content := "Hello **world** with `code`"

	formatted, err := mf.Format(content, TypeHTML)

	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	if formatted == nil {
		t.Errorf("Expected formatted message")
	}

	if formatted.Type != TypeHTML {
		t.Errorf("Expected type TypeHTML")
	}
}

func TestMessageFormatterDetectLanguages(t *testing.T) {
	mf := NewMessageFormatter()

	content := "```go\ncode\n```\n\n```python\ncode\n```"

	languages := mf.DetectLanguages(content)

	if len(languages) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(languages))
	}
}

func TestMessageFormatterExtractCodeSnippets(t *testing.T) {
	mf := NewMessageFormatter()

	content := "```go\nfunc main() {}\n```"

	snippets := mf.ExtractCodeSnippets(content)

	if len(snippets) != 1 {
		t.Errorf("Expected 1 code snippet")
	}

	if snippets[0].Language != "go" {
		t.Errorf("Expected go language")
	}
}

func TestMessageFormatterRegisterCustomLanguage(t *testing.T) {
	mf := NewMessageFormatter()

	mf.RegisterCustomLanguage("mylang")

	if !mf.highlighter.IsSupported("mylang") {
		t.Errorf("Expected mylang to be supported")
	}
}

func TestMessageFormatterStatistics(t *testing.T) {
	mf := NewMessageFormatter()

	content := "Hello world"
	mf.Format(content, TypeHTML)

	stats := mf.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if totalFormatted, ok := stats["total_formatted"].(int64); !ok || totalFormatted != 1 {
		t.Errorf("Expected total_formatted to be 1")
	}
}

func TestBatchFormatter(t *testing.T) {
	bf := NewBatchFormatter()

	contents := []string{
		"**bold**",
		"*italic*",
		"[link](https://example.com)",
	}

	results := bf.FormatBatch(contents, TypeHTML)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Type != TypeHTML {
			t.Errorf("Expected TypeHTML")
		}
	}
}

func TestFormattedMessageWithFormula(t *testing.T) {
	mf := NewMessageFormatter()

	content := "The formula is $$x^2 + y^2 = z^2$$"

	formatted, _ := mf.Format(content, TypeMarkdown)

	if !formatted.HasFormula {
		t.Errorf("Expected formula detection")
	}
}

func BenchmarkFormat(b *testing.B) {
	mf := NewMessageFormatter()

	content := "Hello **world** with `code` and $$E = mc^2$$"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mf.Format(content, TypeHTML)
	}
}

func BenchmarkDetectLanguages(b *testing.B) {
	mf := NewMessageFormatter()

	content := "```go\ncode\n```\n\n```python\ncode\n```"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mf.DetectLanguages(content)
	}
}

func BenchmarkConvertToHTML(b *testing.B) {
	mf := NewMessageFormatter()

	content := "# Header\n**bold** and *italic* [link](https://example.com)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mf.ConvertToHTML(content)
	}
}

func BenchmarkBatchFormat(b *testing.B) {
	bf := NewBatchFormatter()

	contents := []string{
		"**bold**",
		"*italic*",
		"[link](https://example.com)",
		"```go\ncode\n```",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bf.FormatBatch(contents, TypeHTML)
	}
}

func TestCodeExtractorEmpty(t *testing.T) {
	ce := NewCodeExtractor()

	content := "No code blocks here"

	codeBlocks := ce.ExtractCodeBlocks(content)

	if len(codeBlocks) != 0 {
		t.Errorf("Expected 0 code blocks")
	}
}

func TestFormulaDetectorNoFormula(t *testing.T) {
	fd := NewFormulaDetector()

	content := "No formula here"

	if fd.HasFormula(content) {
		t.Errorf("Expected no formula")
	}
}

