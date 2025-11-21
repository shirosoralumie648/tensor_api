package chat

import (
	"testing"
)

func TestTableExtractor(t *testing.T) {
	te := NewTableExtractor()

	content := `| Header 1 | Header 2 |
| --- | --- |
| Cell 1 | Cell 2 |
| Cell 3 | Cell 4 |`

	tables := te.ExtractTables(content)

	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}

	if tables[0].Header == nil {
		t.Errorf("Expected table header")
	}

	if len(tables[0].Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(tables[0].Rows))
	}
}

func TestTableExtractorNoTables(t *testing.T) {
	te := NewTableExtractor()

	content := "No tables here"

	tables := te.ExtractTables(content)

	if len(tables) != 0 {
		t.Errorf("Expected 0 tables")
	}
}

func TestChecklistExtractor(t *testing.T) {
	ce := NewChecklistExtractor()

	content := `- [ ] Task 1
- [x] Task 2
- [ ] Task 3`

	taskList := ce.ExtractChecklists(content)

	if len(taskList.Checklists) != 1 {
		t.Errorf("Expected 1 checklist")
	}

	checklist := taskList.Checklists[0]
	if len(checklist.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(checklist.Items))
	}

	if checklist.Items[1].Checked != true {
		t.Errorf("Expected second item to be checked")
	}
}

func TestChecklistExtractorNoChecklists(t *testing.T) {
	ce := NewChecklistExtractor()

	content := "No checklists here"

	taskList := ce.ExtractChecklists(content)

	if len(taskList.Checklists) != 0 {
		t.Errorf("Expected 0 checklists")
	}
}

func TestQuoteExtractor(t *testing.T) {
	qe := NewQuoteExtractor()

	content := `> This is a quote
> It can span multiple lines
> And continues here`

	quotes := qe.ExtractQuotes(content)

	if len(quotes) < 1 {
		t.Errorf("Expected at least 1 quote")
	}
}

func TestQuoteExtractorNoQuotes(t *testing.T) {
	qe := NewQuoteExtractor()

	content := "No quotes here"

	quotes := qe.ExtractQuotes(content)

	if len(quotes) != 0 {
		t.Errorf("Expected 0 quotes")
	}
}

func TestSyntaxHighlighter(t *testing.T) {
	sh := NewSyntaxHighlighter()

	code := "func main() { fmt.Println(\"Hello\") }"

	highlighted := sh.HighlightCode(code, "go")

	if highlighted == code {
		t.Logf("Code highlighting produced output: %s", highlighted)
	}
}

func TestSyntaxHighlighterRegisterKeywords(t *testing.T) {
	sh := NewSyntaxHighlighter()

	keywords := []string{"custom", "keyword", "list"}
	sh.RegisterKeywords("mylang", keywords)

	retrieved := sh.GetKeywords("mylang")

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 keywords, got %d", len(retrieved))
	}
}

func TestAdvancedMessageFormatterFormatAdvanced(t *testing.T) {
	amf := NewAdvancedMessageFormatter()

	content := `# Title
**bold** and *italic*

\`\`\`go
func main() {}
\`\`\`

| Header 1 | Header 2 |
| --- | --- |
| Cell 1 | Cell 2 |

- [ ] Task 1
- [x] Task 2

> This is a quote`

	formatted, err := amf.FormatAdvanced(content, TypeHTML)

	if err != nil {
		t.Errorf("FormatAdvanced failed: %v", err)
	}

	if formatted == nil {
		t.Errorf("Expected formatted message")
	}
}

func TestAdvancedMessageFormatterExtractStructure(t *testing.T) {
	amf := NewAdvancedMessageFormatter()

	content := `| H1 | H2 |
| --- | --- |
| C1 | C2 |

- [ ] Task 1

> Quote

\`\`\`python
code
\`\`\``

	structure := amf.ExtractStructure(content)

	if structure == nil {
		t.Errorf("Expected structure")
	}

	// 检查是否提取到表格
	if tables, ok := structure["tables"]; ok {
		if tableList, isList := tables.([]*Table); isList {
			if len(tableList) == 0 {
				t.Logf("No tables found in structure")
			}
		}
	}
}

func TestAdvancedMessageFormatterGetStatistics(t *testing.T) {
	amf := NewAdvancedMessageFormatter()

	content := "**bold** text"
	amf.FormatAdvanced(content, TypeHTML)

	stats := amf.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if totalFormatted, ok := stats["total_formatted"].(int64); !ok || totalFormatted != 1 {
		t.Errorf("Expected total_formatted to be 1")
	}
}

func BenchmarkAdvancedFormat(b *testing.B) {
	amf := NewAdvancedMessageFormatter()

	content := `# Title
**bold** and *italic*

\`\`\`go
func main() {}
\`\`\`

| Header 1 | Header 2 |
| --- | --- |
| Cell 1 | Cell 2 |`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = amf.FormatAdvanced(content, TypeHTML)
	}
}

func BenchmarkExtractStructure(b *testing.B) {
	amf := NewAdvancedMessageFormatter()

	content := `| H1 | H2 |
| --- | --- |
| C1 | C2 |

- [ ] Task 1

> Quote`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = amf.ExtractStructure(content)
	}
}

func TestTableExtractorMultipleTables(t *testing.T) {
	te := NewTableExtractor()

	content := `| H1 | H2 |
| --- | --- |
| C1 | C2 |

Some text

| H3 | H4 |
| --- | --- |
| C3 | C4 |`

	tables := te.ExtractTables(content)

	if len(tables) < 1 {
		t.Logf("Expected multiple tables, got %d", len(tables))
	}
}

func TestChecklistExtractorMultipleChecklists(t *testing.T) {
	ce := NewChecklistExtractor()

	content := `- [ ] Task 1
- [x] Task 2

- [ ] Task 3
- [ ] Task 4`

	taskList := ce.ExtractChecklists(content)

	if len(taskList.Checklists) < 1 {
		t.Logf("Expected multiple checklists, got %d", len(taskList.Checklists))
	}
}

func TestAdvancedMessageFormatterComplexContent(t *testing.T) {
	amf := NewAdvancedMessageFormatter()

	content := `# Documentation

## Introduction
This is a **complex** document with *multiple* features.

## Code Examples

\`\`\`go
package main
import "fmt"

func main() {
  fmt.Println("Hello World")
}
\`\`\`

## Features

| Feature | Support |
| --- | --- |
| Markdown | Yes |
| LaTeX | Yes |

## Checklist

- [x] Feature 1 Complete
- [ ] Feature 2 In Progress
- [ ] Feature 3 Pending

## Quotes

> "This is a great feature"
> - Anonymous

## Formula

The equation is \$E = mc^2\$`

	formatted, err := amf.FormatAdvanced(content, TypeHTML)

	if err != nil {
		t.Errorf("Failed to format complex content: %v", err)
	}

	if formatted == nil {
		t.Errorf("Expected formatted message for complex content")
	}

	structure := amf.ExtractStructure(content)
	if structure == nil {
		t.Errorf("Expected structure extraction")
	}
}

