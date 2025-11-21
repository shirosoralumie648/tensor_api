package chat

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

// MessageType 消息类型
type MessageType int

const (
	TypePlainText MessageType = iota
	TypeMarkdown
	TypeHTML
	TypeLaTeX
)

// FormattedMessage 格式化消息
type FormattedMessage struct {
	// 原始内容
	Original string `json:"original"`

	// 格式化后的内容
	Formatted string `json:"formatted"`

	// 消息类型
	Type MessageType `json:"type"`

	// 检测到的代码块语言
	Languages []string `json:"languages"`

	// 是否包含公式
	HasFormula bool `json:"has_formula"`

	// 提取的代码块
	CodeBlocks []*CodeBlock `json:"code_blocks"`
}

// CodeBlock 代码块
type CodeBlock struct {
	// 语言
	Language string `json:"language"`

	// 代码内容
	Content string `json:"content"`

	// 行号
	StartLine int `json:"start_line"`

	// 高亮的行
	HighlightLines []int `json:"highlight_lines"`
}

// SupportedLanguages 支持的语言列表
var SupportedLanguages = []string{
	// 编程语言
	"go", "python", "javascript", "typescript", "java", "c", "cpp", "csharp",
	"rust", "php", "ruby", "swift", "kotlin", "scala", "r", "matlab",
	"perl", "lua", "erlang", "elixir", "clojure", "haskell", "f#",
	
	// 脚本语言
	"bash", "shell", "zsh", "powershell", "cmd",
	
	// 标记语言
	"html", "xml", "css", "scss", "sass", "less",
	
	// 数据格式
	"json", "yaml", "toml", "ini", "csv",
	
	// SQL
	"sql", "mysql", "postgresql", "sqlite", "mongodb",
	
	// 其他
	"diff", "patch", "dockerfile", "makefile", "cmake",
}

// LanguageHighlighter 语言高亮器
type LanguageHighlighter struct {
	// 支持的语言
	languages map[string]bool
	mu        sync.RWMutex
}

// NewLanguageHighlighter 创建语言高亮器
func NewLanguageHighlighter() *LanguageHighlighter {
	lh := &LanguageHighlighter{
		languages: make(map[string]bool),
	}

	for _, lang := range SupportedLanguages {
		lh.languages[strings.ToLower(lang)] = true
	}

	return lh
}

// IsSupported 检查语言是否支持
func (lh *LanguageHighlighter) IsSupported(language string) bool {
	lh.mu.RLock()
	defer lh.mu.RUnlock()

	_, supported := lh.languages[strings.ToLower(language)]
	return supported
}

// RegisterLanguage 注册语言
func (lh *LanguageHighlighter) RegisterLanguage(language string) {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	lh.languages[strings.ToLower(language)] = true
}

// GetSupportedLanguages 获取支持的语言
func (lh *LanguageHighlighter) GetSupportedLanguages() []string {
	lh.mu.RLock()
	defer lh.mu.RUnlock()

	languages := make([]string, 0, len(lh.languages))
	for lang := range lh.languages {
		languages = append(languages, lang)
	}

	return languages
}

// CodeExtractor 代码提取器
type CodeExtractor struct {
	codeBlockRegex *regexp.Regexp
	inlineCodeRegex *regexp.Regexp
}

// NewCodeExtractor 创建代码提取器
func NewCodeExtractor() *CodeExtractor {
	// 匹配 markdown 代码块：```language\ncode\n```
	codeBlockRegex := regexp.MustCompile("(?s)```([\\w-]*)\\n?([^`]+)```")
	// 匹配内联代码：`code`
	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")

	return &CodeExtractor{
		codeBlockRegex: codeBlockRegex,
		inlineCodeRegex: inlineCodeRegex,
	}
}

// ExtractCodeBlocks 提取代码块
func (ce *CodeExtractor) ExtractCodeBlocks(content string) []*CodeBlock {
	matches := ce.codeBlockRegex.FindAllStringSubmatchIndex(content, -1)
	var codeBlocks []*CodeBlock

	for _, match := range matches {
		if len(match) >= 6 {
			language := content[match[2]:match[3]]
			code := content[match[4]:match[5]]

			codeBlock := &CodeBlock{
				Language:  strings.ToLower(strings.TrimSpace(language)),
				Content:   strings.TrimSpace(code),
				StartLine: 0,
			}

			codeBlocks = append(codeBlocks, codeBlock)
		}
	}

	return codeBlocks
}

// ExtractInlineCode 提取内联代码
func (ce *CodeExtractor) ExtractInlineCode(content string) []string {
	matches := ce.inlineCodeRegex.FindAllString(content, -1)
	var codes []string

	for _, match := range matches {
		// 移除反引号
		code := strings.Trim(match, "`")
		codes = append(codes, code)
	}

	return codes
}

// FormulaDetector 公式检测器
type FormulaDetector struct {
	displayFormulaRegex *regexp.Regexp
	inlineFormulaRegex  *regexp.Regexp
}

// NewFormulaDetector 创建公式检测器
func NewFormulaDetector() *FormulaDetector {
	// 匹配 display 公式：$$formula$$
	displayFormulaRegex := regexp.MustCompile(`\$\$([^\$]+)\$\$`)
	// 匹配 inline 公式：$formula$
	inlineFormulaRegex := regexp.MustCompile(`\$([^\$]+)\$`)

	return &FormulaDetector{
		displayFormulaRegex: displayFormulaRegex,
		inlineFormulaRegex:  inlineFormulaRegex,
	}
}

// HasFormula 检测是否包含公式
func (fd *FormulaDetector) HasFormula(content string) bool {
	return fd.displayFormulaRegex.MatchString(content) ||
		fd.inlineFormulaRegex.MatchString(content)
}

// ExtractFormulas 提取公式
func (fd *FormulaDetector) ExtractFormulas(content string) []string {
	var formulas []string

	// 提取 display 公式
	displayMatches := fd.displayFormulaRegex.FindAllString(content, -1)
	formulas = append(formulas, displayMatches...)

	// 提取 inline 公式
	inlineMatches := fd.inlineFormulaRegex.FindAllString(content, -1)
	formulas = append(formulas, inlineMatches...)

	return formulas
}

// MarkdownProcessor Markdown 处理器
type MarkdownProcessor struct {
	headingRegex   *regexp.Regexp
	boldRegex      *regexp.Regexp
	italicRegex    *regexp.Regexp
	linkRegex      *regexp.Regexp
	listItemRegex  *regexp.Regexp
}

// NewMarkdownProcessor 创建 Markdown 处理器
func NewMarkdownProcessor() *MarkdownProcessor {
	return &MarkdownProcessor{
		headingRegex:  regexp.MustCompile(`^(#+)\s+(.+)$`),
		boldRegex:     regexp.MustCompile(`\*\*(.+?)\*\*`),
		italicRegex:   regexp.MustCompile(`\*(.+?)\*`),
		linkRegex:     regexp.MustCompile(`\[(.+?)\]\((.+?)\)`),
		listItemRegex: regexp.MustCompile(`^(\s*)[-*+]\s+(.+)$`),
	}
}

// ToHTML 转换为 HTML
func (mp *MarkdownProcessor) ToHTML(content string) string {
	html := content

	// 处理标题
	html = mp.headingRegex.ReplaceAllString(html, `<h$1>$2</h$1>`)

	// 处理粗体
	html = mp.boldRegex.ReplaceAllString(html, `<strong>$1</strong>`)

	// 处理斜体
	html = mp.italicRegex.ReplaceAllString(html, `<em>$1</em>`)

	// 处理链接
	html = mp.linkRegex.ReplaceAllString(html, `<a href="$2">$1</a>`)

	// 处理列表项（简单实现）
	lines := strings.Split(html, "\n")
	var result []string
	inList := false

	for _, line := range lines {
		if mp.listItemRegex.MatchString(line) {
			if !inList {
				result = append(result, "<ul>")
				inList = true
			}
			matches := mp.listItemRegex.FindStringSubmatch(line)
			if len(matches) >= 3 {
				result = append(result, fmt.Sprintf("<li>%s</li>", strings.TrimSpace(matches[2])))
			}
		} else {
			if inList {
				result = append(result, "</ul>")
				inList = false
			}
			if strings.TrimSpace(line) != "" {
				result = append(result, fmt.Sprintf("<p>%s</p>", line))
			}
		}
	}

	if inList {
		result = append(result, "</ul>")
	}

	return strings.Join(result, "\n")
}

// MessageFormatter 消息格式化器
type MessageFormatter struct {
	// 语言高亮器
	highlighter *LanguageHighlighter

	// 代码提取器
	extractor *CodeExtractor

	// 公式检测器
	detector *FormulaDetector

	// Markdown 处理器
	processor *MarkdownProcessor

	// 统计信息
	totalFormatted int64
	totalHTML      int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewMessageFormatter 创建消息格式化器
func NewMessageFormatter() *MessageFormatter {
	return &MessageFormatter{
		highlighter: NewLanguageHighlighter(),
		extractor:   NewCodeExtractor(),
		detector:    NewFormulaDetector(),
		processor:   NewMarkdownProcessor(),
		logFunc:     defaultLogFunc,
	}
}

// Format 格式化消息
func (mf *MessageFormatter) Format(content string, targetType MessageType) (*FormattedMessage, error) {
	formatted := &FormattedMessage{
		Original:   content,
		Type:       targetType,
		Languages:  []string{},
		CodeBlocks: []*CodeBlock{},
	}

	// 提取代码块
	codeBlocks := mf.extractor.ExtractCodeBlocks(content)
	formatted.CodeBlocks = codeBlocks

	// 收集语言
	languageMap := make(map[string]bool)
	for _, block := range codeBlocks {
		if block.Language != "" && mf.highlighter.IsSupported(block.Language) {
			languageMap[block.Language] = true
		}
	}

	for lang := range languageMap {
		formatted.Languages = append(formatted.Languages, lang)
	}

	// 检测公式
	formatted.HasFormula = mf.detector.HasFormula(content)

	// 格式化内容
	switch targetType {
	case TypeHTML:
		formatted.Formatted = mf.processor.ToHTML(content)
		atomic.AddInt64(&mf.totalHTML, 1)
	case TypeMarkdown:
		formatted.Formatted = content
	default:
		formatted.Formatted = content
	}

	atomic.AddInt64(&mf.totalFormatted, 1)

	mf.logFunc("debug", fmt.Sprintf("Formatted message with type %d", targetType))

	return formatted, nil
}

// DetectLanguages 检测消息中的编程语言
func (mf *MessageFormatter) DetectLanguages(content string) []string {
	codeBlocks := mf.extractor.ExtractCodeBlocks(content)
	languageMap := make(map[string]bool)

	for _, block := range codeBlocks {
		if block.Language != "" {
			languageMap[block.Language] = true
		}
	}

	languages := make([]string, 0, len(languageMap))
	for lang := range languageMap {
		languages = append(languages, lang)
	}

	return languages
}

// ExtractCodeSnippets 提取代码片段
func (mf *MessageFormatter) ExtractCodeSnippets(content string) []*CodeBlock {
	return mf.extractor.ExtractCodeBlocks(content)
}

// ConvertToHTML 转换为 HTML
func (mf *MessageFormatter) ConvertToHTML(content string) string {
	return mf.processor.ToHTML(content)
}

// RegisterCustomLanguage 注册自定义语言
func (mf *MessageFormatter) RegisterCustomLanguage(language string) {
	mf.highlighter.RegisterLanguage(language)
}

// GetStatistics 获取统计信息
func (mf *MessageFormatter) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_formatted": atomic.LoadInt64(&mf.totalFormatted),
		"total_html":      atomic.LoadInt64(&mf.totalHTML),
		"supported_languages": len(mf.highlighter.GetSupportedLanguages()),
	}
}

// BatchFormatter 批量格式化器
type BatchFormatter struct {
	formatter *MessageFormatter
	mu        sync.RWMutex
}

// NewBatchFormatter 创建批量格式化器
func NewBatchFormatter() *BatchFormatter {
	return &BatchFormatter{
		formatter: NewMessageFormatter(),
	}
}

// FormatBatch 批量格式化
func (bf *BatchFormatter) FormatBatch(contents []string, targetType MessageType) []*FormattedMessage {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	results := make([]*FormattedMessage, len(contents))
	for i, content := range contents {
		formatted, _ := bf.formatter.Format(content, targetType)
		results[i] = formatted
	}

	return results
}

// GetFormatter 获取格式化器
func (bf *BatchFormatter) GetFormatter() *MessageFormatter {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	return bf.formatter
}

