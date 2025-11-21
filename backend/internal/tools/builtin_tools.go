package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WebSearchTool 网络搜索工具
type WebSearchTool struct {
	// 搜索引擎
	engine string // google, bing, duckduckgo

	// API 密钥
	apiKey string

	// 最大结果数
	maxResults int

	// 超时时间
	timeout time.Duration
}

// NewWebSearchTool 创建网络搜索工具
func NewWebSearchTool(engine string, apiKey string) *WebSearchTool {
	return &WebSearchTool{
		engine:     engine,
		apiKey:     apiKey,
		maxResults: 10,
		timeout:    10 * time.Second,
	}
}

// SearchResult 搜索结果
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Execute 执行搜索
func (wst *WebSearchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// 模拟搜索结果
	results := []SearchResult{
		{
			Title:       "Search Result 1: " + query,
			URL:         "https://example.com/result1",
			Description: "This is the first search result for: " + query,
		},
		{
			Title:       "Search Result 2: " + query,
			URL:         "https://example.com/result2",
			Description: "This is the second search result for: " + query,
		},
		{
			Title:       "Search Result 3: " + query,
			URL:         "https://example.com/result3",
			Description: "This is the third search result for: " + query,
		},
	}

	return results, nil
}

// GetTool 获取工具定义
func (wst *WebSearchTool) GetTool() *Tool {
	return &Tool{
		Name:        "web_search",
		Description: "Search the web for information",
		Parameters: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"query": {
					Type:        "string",
					Description: "Search query",
				},
				"max_results": {
					Type:        "integer",
					Description: "Maximum number of results (default: 10)",
				},
			},
			Required: []string{"query"},
		},
		Handler: wst.Execute,
		Timeout: 10 * time.Second,
	}
}

// CodeExecutorTool 代码执行工具
type CodeExecutorTool struct {
	// 支持的语言
	languages []string

	// 执行超时
	timeout time.Duration

	// 是否启用沙箱
	sandbox bool
}

// NewCodeExecutorTool 创建代码执行工具
func NewCodeExecutorTool() *CodeExecutorTool {
	return &CodeExecutorTool{
		languages: []string{"python", "javascript", "bash"},
		timeout:   30 * time.Second,
		sandbox:   true,
	}
}

// Execute 执行代码
func (cet *CodeExecutorTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	language, ok := args["language"].(string)
	if !ok || language == "" {
		return nil, fmt.Errorf("language parameter is required")
	}

	code, ok := args["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("code parameter is required")
	}

	// 检查支持的语言
	supported := false
	for _, lang := range cet.languages {
		if lang == language {
			supported = true
			break
		}
	}

	if !supported {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// 模拟代码执行结果
	result := map[string]interface{}{
		"language": language,
		"status":   "success",
		"output":   "Code executed successfully",
		"duration": 100, // 毫秒
	}

	return result, nil
}

// GetTool 获取工具定义
func (cet *CodeExecutorTool) GetTool() *Tool {
	return &Tool{
		Name:        "code_executor",
		Description: "Execute code in a sandboxed environment",
		Parameters: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"language": {
					Type:        "string",
					Description: "Programming language (python, javascript, bash)",
					Enum: []interface{}{"python", "javascript", "bash"},
				},
				"code": {
					Type:        "string",
					Description: "Code to execute",
				},
			},
			Required: []string{"language", "code"},
		},
		Handler: cet.Execute,
		Timeout: 30 * time.Second,
	}
}

// FileOperationTool 文件操作工具
type FileOperationTool struct {
	// 允许的路径
	allowedPaths []string

	// 基础路径
	basePath string
}

// NewFileOperationTool 创建文件操作工具
func NewFileOperationTool(basePath string) *FileOperationTool {
	return &FileOperationTool{
		basePath:     basePath,
		allowedPaths: []string{basePath},
	}
}

// Execute 执行文件操作
func (fot *FileOperationTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	operation, ok := args["operation"].(string)
	if !ok || operation == "" {
		return nil, fmt.Errorf("operation parameter is required")
	}

	path, ok := args["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("path parameter is required")
	}

	// 检查路径安全
	fullPath := filepath.Join(fot.basePath, path)
	if !strings.HasPrefix(fullPath, fot.basePath) {
		return nil, fmt.Errorf("access denied: path outside allowed directory")
	}

	switch operation {
	case "read":
		return fot.readFile(fullPath)
	case "write":
		content, ok := args["content"].(string)
		if !ok {
			return nil, fmt.Errorf("content parameter is required for write operation")
		}
		return fot.writeFile(fullPath, content)
	case "delete":
		return fot.deleteFile(fullPath)
	case "list":
		return fot.listDirectory(fullPath)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// readFile 读取文件
func (fot *FileOperationTool) readFile(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return map[string]interface{}{
		"operation": "read",
		"size":      len(data),
		"content":   string(data[:min(len(data), 1000)]), // 限制为 1000 字符
	}, nil
}

// writeFile 写入文件
func (fot *FileOperationTool) writeFile(path string, content string) (interface{}, error) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %v", err)
	}

	return map[string]interface{}{
		"operation": "write",
		"size":      len(content),
		"status":    "success",
	}, nil
}

// deleteFile 删除文件
func (fot *FileOperationTool) deleteFile(path string) (interface{}, error) {
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("failed to delete file: %v", err)
	}

	return map[string]interface{}{
		"operation": "delete",
		"status":    "success",
	}, nil
}

// listDirectory 列出目录
func (fot *FileOperationTool) listDirectory(path string) (interface{}, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %v", err)
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, map[string]interface{}{
			"name":      entry.Name(),
			"is_dir":    entry.IsDir(),
			"size":      info.Size(),
			"modified":  info.ModTime(),
		})
	}

	return map[string]interface{}{
		"operation": "list",
		"count":     len(files),
		"files":     files,
	}, nil
}

// GetTool 获取工具定义
func (fot *FileOperationTool) GetTool() *Tool {
	return &Tool{
		Name:        "file_operations",
		Description: "Perform file operations (read, write, delete, list)",
		Parameters: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"operation": {
					Type:        "string",
					Description: "Operation type (read, write, delete, list)",
					Enum: []interface{}{"read", "write", "delete", "list"},
				},
				"path": {
					Type:        "string",
					Description: "File or directory path",
				},
				"content": {
					Type:        "string",
					Description: "File content (required for write operation)",
				},
			},
			Required: []string{"operation", "path"},
		},
		Handler: fot.Execute,
		Timeout: 10 * time.Second,
	}
}

// HTTPRequestTool HTTP 请求工具
type HTTPRequestTool struct {
	// 允许的域名
	allowedDomains []string

	// 最大响应大小
	maxResponseSize int64

	// 超时时间
	timeout time.Duration
}

// NewHTTPRequestTool 创建 HTTP 请求工具
func NewHTTPRequestTool() *HTTPRequestTool {
	return &HTTPRequestTool{
		allowedDomains:  []string{"*"}, // 允许所有域名
		maxResponseSize: 1024 * 100,    // 100KB
		timeout:         10 * time.Second,
	}
}

// Execute 执行 HTTP 请求
func (hrt *HTTPRequestTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	method, ok := args["method"].(string)
	if !ok {
		method = "GET"
	}

	urlStr, ok := args["url"].(string)
	if !ok || urlStr == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	// 验证 URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// 检查域名白名单
	if !hrt.isDomainAllowed(parsedURL.Host) {
		return nil, fmt.Errorf("domain not allowed: %s", parsedURL.Host)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 添加请求体
	if body, ok := args["body"].(string); ok && body != "" {
		req.Body = io.NopCloser(strings.NewReader(body))
	}

	// 添加请求头
	if headers, ok := args["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if headerVal, ok := v.(string); ok {
				req.Header.Add(k, headerVal)
			}
		}
	}

	// 执行请求
	client := &http.Client{
		Timeout: hrt.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, hrt.maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(bodyBytes),
		"size":        len(bodyBytes),
	}, nil
}

// isDomainAllowed 检查域名是否允许
func (hrt *HTTPRequestTool) isDomainAllowed(domain string) bool {
	for _, allowed := range hrt.allowedDomains {
		if allowed == "*" {
			return true
		}

		if allowed == domain {
			return true
		}

		// 支持子域名匹配
		if strings.HasPrefix(allowed, "*.") && strings.HasSuffix(domain, allowed[1:]) {
			return true
		}
	}

	return false
}

// GetTool 获取工具定义
func (hrt *HTTPRequestTool) GetTool() *Tool {
	return &Tool{
		Name:        "http_request",
		Description: "Make HTTP requests to URLs",
		Parameters: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"url": {
					Type:        "string",
					Description: "URL to request",
				},
				"method": {
					Type:        "string",
					Description: "HTTP method (GET, POST, PUT, DELETE)",
					Enum: []interface{}{"GET", "POST", "PUT", "DELETE"},
				},
				"headers": {
					Type:        "object",
					Description: "HTTP headers",
				},
				"body": {
					Type:        "string",
					Description: "Request body",
				},
			},
			Required: []string{"url"},
		},
		Handler: hrt.Execute,
		Timeout: 10 * time.Second,
	}
}

// BuiltinToolsFactory 内置工具工厂
type BuiltinToolsFactory struct {
	engine *FunctionEngine
}

// NewBuiltinToolsFactory 创建内置工具工厂
func NewBuiltinToolsFactory(engine *FunctionEngine) *BuiltinToolsFactory {
	return &BuiltinToolsFactory{
		engine: engine,
	}
}

// RegisterAllTools 注册所有内置工具
func (btf *BuiltinToolsFactory) RegisterAllTools() error {
	tools := []interface {
		GetTool() *Tool
	}{
		NewWebSearchTool("google", ""),
		NewCodeExecutorTool(),
		NewFileOperationTool("/tmp"),
		NewHTTPRequestTool(),
	}

	for _, tool := range tools {
		if err := btf.engine.RegisterTool(tool.GetTool()); err != nil {
			return err
		}
	}

	return nil
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

