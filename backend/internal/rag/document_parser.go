package rag

import (
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"sync"
	"time"
)

// ParsedContent 解析后的内容
type ParsedContent struct {
	// 文档标题
	Title string `json:"title"`

	// 原始内容
	Content string `json:"content"`

	// 内容摘要
	Summary string `json:"summary"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata"`

	// 解析耗时
	ParseDuration time.Duration `json:"parse_duration"`

	// 页码（如果有）
	Pages int `json:"pages"`

	// 语言检测
	Language string `json:"language"`
}

// DocumentParser 文档解析器接口
type DocumentParser interface {
	// 支持的文件类型
	SupportedTypes() []string

	// 解析文件
	Parse(file io.Reader, filename string) (*ParsedContent, error)

	// 解析器名称
	Name() string
}

// TextParser 文本文件解析器
type TextParser struct{}

func (tp *TextParser) Name() string {
	return "TextParser"
}

func (tp *TextParser) SupportedTypes() []string {
	return []string{".txt", ".md", ".markdown"}
}

func (tp *TextParser) Parse(file io.Reader, filename string) (*ParsedContent, error) {
	start := time.Now()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := string(data)

	return &ParsedContent{
		Title:         filename,
		Content:       content,
		Summary:       tp.summarize(content),
		Metadata:      map[string]interface{}{"type": "text"},
		ParseDuration: time.Since(start),
		Language:      "unknown",
	}, nil
}

func (tp *TextParser) summarize(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) > 3 {
		return strings.Join(lines[:3], "\n") + "..."
	}
	return content
}

// JSONParser JSON文件解析器
type JSONParser struct{}

func (jp *JSONParser) Name() string {
	return "JSONParser"
}

func (jp *JSONParser) SupportedTypes() []string {
	return []string{".json"}
}

func (jp *JSONParser) Parse(file io.Reader, filename string) (*ParsedContent, error) {
	start := time.Now()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := string(data)

	return &ParsedContent{
		Title:         filename,
		Content:       content,
		Summary:       content[:min(len(content), 200)],
		Metadata:      map[string]interface{}{"type": "json"},
		ParseDuration: time.Since(start),
		Language:      "unknown",
	}, nil
}

// CSVParser CSV文件解析器
type CSVParser struct{}

func (cp *CSVParser) Name() string {
	return "CSVParser"
}

func (cp *CSVParser) SupportedTypes() []string {
	return []string{".csv"}
}

func (cp *CSVParser) Parse(file io.Reader, filename string) (*ParsedContent, error) {
	start := time.Now()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	return &ParsedContent{
		Title:         filename,
		Content:       content,
		Summary:       fmt.Sprintf("CSV file with %d rows", len(lines)),
		Metadata:      map[string]interface{}{"type": "csv", "rows": len(lines)},
		ParseDuration: time.Since(start),
		Language:      "unknown",
	}, nil
}

// HTMLParser HTML文件解析器
type HTMLParser struct{}

func (hp *HTMLParser) Name() string {
	return "HTMLParser"
}

func (hp *HTMLParser) SupportedTypes() []string {
	return []string{".html", ".htm"}
}

func (hp *HTMLParser) Parse(file io.Reader, filename string) (*ParsedContent, error) {
	start := time.Now()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := string(data)
	// 简单提取文本内容（去除HTML标签）
	text := hp.stripHTMLTags(content)

	return &ParsedContent{
		Title:         filename,
		Content:       text,
		Summary:       text[:min(len(text), 200)],
		Metadata:      map[string]interface{}{"type": "html"},
		ParseDuration: time.Since(start),
		Language:      "unknown",
	}, nil
}

func (hp *HTMLParser) stripHTMLTags(html string) string {
	// 简单的HTML标签移除（生产环境应使用htmlquery）
	result := ""
	inTag := false

	for _, c := range html {
		if c == '<' {
			inTag = true
		} else if c == '>' {
			inTag = false
			result += " "
		} else if !inTag {
			result += string(c)
		}
	}

	return result
}

// XMLParser XML文件解析器
type XMLParser struct{}

func (xp *XMLParser) Name() string {
	return "XMLParser"
}

func (xp *XMLParser) SupportedTypes() []string {
	return []string{".xml"}
}

func (xp *XMLParser) Parse(file io.Reader, filename string) (*ParsedContent, error) {
	start := time.Now()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := string(data)

	return &ParsedContent{
		Title:         filename,
		Content:       content,
		Summary:       content[:min(len(content), 200)],
		Metadata:      map[string]interface{}{"type": "xml"},
		ParseDuration: time.Since(start),
		Language:      "unknown",
	}, nil
}

// YAMLParser YAML文件解析器
type YAMLParser struct{}

func (yp *YAMLParser) Name() string {
	return "YAMLParser"
}

func (yp *YAMLParser) SupportedTypes() []string {
	return []string{".yaml", ".yml"}
}

func (yp *YAMLParser) Parse(file io.Reader, filename string) (*ParsedContent, error) {
	start := time.Now()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := string(data)

	return &ParsedContent{
		Title:         filename,
		Content:       content,
		Summary:       content[:min(len(content), 200)],
		Metadata:      map[string]interface{}{"type": "yaml"},
		ParseDuration: time.Since(start),
		Language:      "unknown",
	}, nil
}

// DocumentParserRegistry 文档解析器注册表
type DocumentParserRegistry struct {
	parsers map[string]DocumentParser
	mu      sync.RWMutex
}

// NewDocumentParserRegistry 创建解析器注册表
func NewDocumentParserRegistry() *DocumentParserRegistry {
	registry := &DocumentParserRegistry{
		parsers: make(map[string]DocumentParser),
	}

	// 注册默认解析器
	registry.Register(&TextParser{})
	registry.Register(&JSONParser{})
	registry.Register(&CSVParser{})
	registry.Register(&HTMLParser{})
	registry.Register(&XMLParser{})
	registry.Register(&YAMLParser{})

	return registry
}

// Register 注册解析器
func (dpr *DocumentParserRegistry) Register(parser DocumentParser) {
	dpr.mu.Lock()
	defer dpr.mu.Unlock()

	for _, ext := range parser.SupportedTypes() {
		dpr.parsers[ext] = parser
	}
}

// Parse 解析文件
func (dpr *DocumentParserRegistry) Parse(file io.Reader, filename string) (*ParsedContent, error) {
	// 获取文件扩展名
	ext := ""
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		ext = filename[idx:]
	}

	dpr.mu.RLock()
	parser, exists := dpr.parsers[strings.ToLower(ext)]
	dpr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	return parser.Parse(file, filename)
}

// GetSupportedTypes 获取所有支持的文件类型
func (dpr *DocumentParserRegistry) GetSupportedTypes() []string {
	dpr.mu.RLock()
	defer dpr.mu.RUnlock()

	types := make([]string, 0)
	seen := make(map[string]bool)

	for ext := range dpr.parsers {
		if !seen[ext] {
			types = append(types, ext)
			seen[ext] = true
		}
	}

	return types
}

// DocumentUploadManager 文档上传管理器
type DocumentUploadManager struct {
	// 解析器注册表
	parserRegistry *DocumentParserRegistry

	// 最大文件大小（字节）
	maxFileSize int64

	// 上传历史
	uploadHistory map[string]*DocumentUploadInfo
	historyMu     sync.RWMutex
}

// DocumentUploadInfo 文档上传信息
type DocumentUploadInfo struct {
	// 文件名
	Filename string `json:"filename"`

	// 文件大小
	FileSize int64 `json:"file_size"`

	// 文件类型
	FileType string `json:"file_type"`

	// 解析状态
	Status string `json:"status"` // uploading, processing, completed, failed

	// 解析结果
	ParsedContent *ParsedContent `json:"parsed_content"`

	// 错误信息
	ErrorMessage string `json:"error_message"`

	// 上传时间
	UploadedAt time.Time `json:"uploaded_at"`

	// 完成时间
	CompletedAt *time.Time `json:"completed_at"`

	// 解析耗时
	ParseDuration time.Duration `json:"parse_duration"`
}

// NewDocumentUploadManager 创建上传管理器
func NewDocumentUploadManager(maxFileSize int64) *DocumentUploadManager {
	return &DocumentUploadManager{
		parserRegistry: NewDocumentParserRegistry(),
		maxFileSize:    maxFileSize,
		uploadHistory:  make(map[string]*DocumentUploadInfo),
	}
}

// UploadFile 上传文件
func (dum *DocumentUploadManager) UploadFile(fileHeader *multipart.FileHeader) (*DocumentUploadInfo, error) {
	// 检查文件大小
	if fileHeader.Size > dum.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit: %d > %d", fileHeader.Size, dum.maxFileSize)
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建上传信息
	uploadID := fmt.Sprintf("upload-%d", time.Now().UnixNano())
	info := &DocumentUploadInfo{
		Filename:   fileHeader.Filename,
		FileSize:   fileHeader.Size,
		FileType:   fileHeader.Header.Get("Content-Type"),
		Status:     "processing",
		UploadedAt: time.Now(),
	}

	// 解析文件
	start := time.Now()
	parsed, err := dum.parserRegistry.Parse(file, fileHeader.Filename)

	if err != nil {
		info.Status = "failed"
		info.ErrorMessage = err.Error()
	} else {
		info.Status = "completed"
		info.ParsedContent = parsed
		info.ParseDuration = time.Since(start)
		completedAt := time.Now()
		info.CompletedAt = &completedAt
	}

	// 保存到历史
	dum.historyMu.Lock()
	dum.uploadHistory[uploadID] = info
	dum.historyMu.Unlock()

	return info, nil
}

// GetUploadHistory 获取上传历史
func (dum *DocumentUploadManager) GetUploadHistory() []*DocumentUploadInfo {
	dum.historyMu.RLock()
	defer dum.historyMu.RUnlock()

	history := make([]*DocumentUploadInfo, 0)
	for _, info := range dum.uploadHistory {
		history = append(history, info)
	}

	return history
}

// GetSupportedTypes 获取支持的文件类型
func (dum *DocumentUploadManager) GetSupportedTypes() []string {
	return dum.parserRegistry.GetSupportedTypes()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

