package chat

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// TableCell 表格单元格
type TableCell struct {
	Content string `json:"content"`
	Align   string `json:"align"` // left, center, right
}

// TableRow 表格行
type TableRow struct {
	Cells []*TableCell `json:"cells"`
}

// Table 表格
type Table struct {
	Header *TableRow  `json:"header"`
	Rows   []*TableRow `json:"rows"`
}

// ChecklistItem 检查清单项
type ChecklistItem struct {
	Content string `json:"content"`
	Checked bool   `json:"checked"`
}

// Checklist 检查清单
type Checklist struct {
	Items []*ChecklistItem `json:"items"`
}

// TaskList 任务列表
type TaskList struct {
	Checklists []*Checklist `json:"checklists"`
}

// TableExtractor 表格提取器
type TableExtractor struct {
	tableRegex *regexp.Regexp
}

// NewTableExtractor 创建表格提取器
func NewTableExtractor() *TableExtractor {
	// 匹配 markdown 表格
	tableRegex := regexp.MustCompile(`(?m)^\|(.+?)\|$`)

	return &TableExtractor{
		tableRegex: tableRegex,
	}
}

// ExtractTables 提取表格
func (te *TableExtractor) ExtractTables(content string) []*Table {
	var tables []*Table
	lines := strings.Split(content, "\n")

	i := 0
	for i < len(lines) {
		line := lines[i]

		// 检查是否是表格行
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			table := &Table{}

			// 解析表头
			headerCells := te.parseTableRow(line)
			table.Header = &TableRow{Cells: headerCells}

			// 解析分隔符行
			i++
			if i < len(lines) && strings.Contains(lines[i], "---") {
				i++
			}

			// 解析表格行
			for i < len(lines) && strings.HasPrefix(lines[i], "|") && strings.HasSuffix(lines[i], "|") {
				cells := te.parseTableRow(lines[i])
				table.Rows = append(table.Rows, &TableRow{Cells: cells})
				i++
			}

			tables = append(tables, table)
			continue
		}

		i++
	}

	return tables
}

// parseTableRow 解析表格行
func (te *TableExtractor) parseTableRow(line string) []*TableCell {
	// 移除前后的 |
	line = strings.Trim(line, "|")

	// 分割列
	parts := strings.Split(line, "|")
	var cells []*TableCell

	for _, part := range parts {
		part = strings.TrimSpace(part)
		cell := &TableCell{
			Content: part,
			Align:   "left",
		}
		cells = append(cells, cell)
	}

	return cells
}

// ChecklistExtractor 检查清单提取器
type ChecklistExtractor struct {
	checklistRegex *regexp.Regexp
}

// NewChecklistExtractor 创建检查清单提取器
func NewChecklistExtractor() *ChecklistExtractor {
	// 匹配检查清单：- [ ] item 或 - [x] item
	checklistRegex := regexp.MustCompile(`^\s*-\s+\[([ xX])\]\s+(.+)$`)

	return &ChecklistExtractor{
		checklistRegex: checklistRegex,
	}
}

// ExtractChecklists 提取检查清单
func (ce *ChecklistExtractor) ExtractChecklists(content string) *TaskList {
	taskList := &TaskList{}
	lines := strings.Split(content, "\n")

	var currentChecklist *Checklist

	for _, line := range lines {
		if ce.checklistRegex.MatchString(line) {
			matches := ce.checklistRegex.FindStringSubmatch(line)
			if len(matches) >= 3 {
				if currentChecklist == nil {
					currentChecklist = &Checklist{}
				}

				checked := matches[1] != " "
				item := &ChecklistItem{
					Content: matches[2],
					Checked: checked,
				}

				currentChecklist.Items = append(currentChecklist.Items, item)
			}
		} else if strings.TrimSpace(line) == "" {
			// 空行意味着新的清单开始
			if currentChecklist != nil && len(currentChecklist.Items) > 0 {
				taskList.Checklists = append(taskList.Checklists, currentChecklist)
				currentChecklist = nil
			}
		}
	}

	// 添加最后一个清单
	if currentChecklist != nil && len(currentChecklist.Items) > 0 {
		taskList.Checklists = append(taskList.Checklists, currentChecklist)
	}

	return taskList
}

// QuoteExtractor 引用提取器
type QuoteExtractor struct {
	blockquoteRegex *regexp.Regexp
}

// NewQuoteExtractor 创建引用提取器
func NewQuoteExtractor() *QuoteExtractor {
	// 匹配块引用：> quote
	blockquoteRegex := regexp.MustCompile(`^\s*>\s+(.+)$`)

	return &QuoteExtractor{
		blockquoteRegex: blockquoteRegex,
	}
}

// ExtractQuotes 提取引用
func (qe *QuoteExtractor) ExtractQuotes(content string) []string {
	var quotes []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if qe.blockquoteRegex.MatchString(line) {
			matches := qe.blockquoteRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				quotes = append(quotes, strings.TrimSpace(matches[1]))
			}
		}
	}

	return quotes
}

// SyntaxHighlighter 语法高亮器
type SyntaxHighlighter struct {
	keywords map[string][]string
	mu       sync.RWMutex
}

// NewSyntaxHighlighter 创建语法高亮器
func NewSyntaxHighlighter() *SyntaxHighlighter {
	sh := &SyntaxHighlighter{
		keywords: make(map[string][]string),
	}

	// 注册 Go 关键词
	sh.keywords["go"] = []string{
		"package", "import", "func", "type", "struct", "interface",
		"if", "else", "for", "switch", "case", "default",
		"return", "defer", "go", "select", "chan",
	}

	// 注册 Python 关键词
	sh.keywords["python"] = []string{
		"def", "class", "import", "from", "if", "else", "elif",
		"for", "while", "return", "try", "except", "finally",
		"with", "as", "lambda", "yield",
	}

	// 注册 JavaScript 关键词
	sh.keywords["javascript"] = []string{
		"function", "const", "let", "var", "if", "else",
		"for", "while", "return", "class", "extends",
		"import", "export", "async", "await", "try", "catch",
	}

	return sh
}

// RegisterKeywords 注册关键词
func (sh *SyntaxHighlighter) RegisterKeywords(language string, keywords []string) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.keywords[language] = keywords
}

// GetKeywords 获取关键词
func (sh *SyntaxHighlighter) GetKeywords(language string) []string {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	if keywords, ok := sh.keywords[language]; ok {
		return keywords
	}

	return []string{}
}

// HighlightCode 高亮代码
func (sh *SyntaxHighlighter) HighlightCode(code string, language string) string {
	keywords := sh.GetKeywords(language)
	if len(keywords) == 0 {
		return code
	}

	result := code
	for _, keyword := range keywords {
		// 简单的关键词替换（实际应用中使用更复杂的解析）
		pattern := fmt.Sprintf(`\b%s\b`, keyword)
		regex := regexp.MustCompile(pattern)
		result = regex.ReplaceAllString(result, fmt.Sprintf(`<span class="keyword">%s</span>`, keyword))
	}

	return result
}

// AdvancedMessageFormatter 高级消息格式化器
type AdvancedMessageFormatter struct {
	// 基础格式化器
	baseFormatter *MessageFormatter

	// 表格提取器
	tableExtractor *TableExtractor

	// 检查清单提取器
	checklistExtractor *ChecklistExtractor

	// 引用提取器
	quoteExtractor *QuoteExtractor

	// 语法高亮器
	syntaxHighlighter *SyntaxHighlighter

	// 统计信息
	totalFormatted int64
	processTime    atomic.Int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewAdvancedMessageFormatter 创建高级消息格式化器
func NewAdvancedMessageFormatter() *AdvancedMessageFormatter {
	return &AdvancedMessageFormatter{
		baseFormatter:      NewMessageFormatter(),
		tableExtractor:     NewTableExtractor(),
		checklistExtractor: NewChecklistExtractor(),
		quoteExtractor:     NewQuoteExtractor(),
		syntaxHighlighter:  NewSyntaxHighlighter(),
		logFunc:            defaultLogFunc,
	}
}

// FormatAdvanced 高级格式化
func (amf *AdvancedMessageFormatter) FormatAdvanced(content string, targetType MessageType) (*FormattedMessage, error) {
	start := time.Now()
	defer func() {
		amf.processTime.Store(int64(time.Since(start).Milliseconds()))
	}()

	// 使用基础格式化器
	formatted, err := amf.baseFormatter.Format(content, targetType)
	if err != nil {
		return nil, err
	}

	// 提取表格
	tables := amf.tableExtractor.ExtractTables(content)
	if len(tables) > 0 {
		amf.logFunc("debug", fmt.Sprintf("Found %d tables", len(tables)))
	}

	// 提取检查清单
	taskList := amf.checklistExtractor.ExtractChecklists(content)
	if len(taskList.Checklists) > 0 {
		amf.logFunc("debug", fmt.Sprintf("Found %d checklists", len(taskList.Checklists)))
	}

	// 提取引用
	quotes := amf.quoteExtractor.ExtractQuotes(content)
	if len(quotes) > 0 {
		amf.logFunc("debug", fmt.Sprintf("Found %d quotes", len(quotes)))
	}

	// 高亮代码
	for _, codeBlock := range formatted.CodeBlocks {
		codeBlock.Content = amf.syntaxHighlighter.HighlightCode(codeBlock.Content, codeBlock.Language)
	}

	atomic.AddInt64(&amf.totalFormatted, 1)

	return formatted, nil
}

// ExtractStructure 提取结构
func (amf *AdvancedMessageFormatter) ExtractStructure(content string) map[string]interface{} {
	structure := make(map[string]interface{})

	// 提取表格
	tables := amf.tableExtractor.ExtractTables(content)
	structure["tables"] = tables

	// 提取检查清单
	taskList := amf.checklistExtractor.ExtractChecklists(content)
	structure["task_lists"] = taskList.Checklists

	// 提取引用
	quotes := amf.quoteExtractor.ExtractQuotes(content)
	structure["quotes"] = quotes

	// 提取代码块
	codeBlocks := amf.baseFormatter.ExtractCodeSnippets(content)
	structure["code_blocks"] = codeBlocks

	// 检测语言
	languages := amf.baseFormatter.DetectLanguages(content)
	structure["languages"] = languages

	return structure
}

// GetStatistics 获取统计信息
func (amf *AdvancedMessageFormatter) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_formatted":     atomic.LoadInt64(&amf.totalFormatted),
		"last_process_time_ms": amf.processTime.Load(),
	}
}

