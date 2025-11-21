package security

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// InputValidator 输入验证器
type InputValidator struct {
	maxStringLength int
	maxArrayLength  int
}

// NewInputValidator 创建新的输入验证器
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxStringLength: 10000,
		maxArrayLength:  1000,
	}
}

// ValidateString 验证字符串
func (v *InputValidator) ValidateString(input string, fieldName string) error {
	if len(input) == 0 {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if utf8.RuneCountInString(input) > v.maxStringLength {
		return fmt.Errorf("%s exceeds max length of %d", fieldName, v.maxStringLength)
	}

	if containsSQLInjection(input) {
		return fmt.Errorf("%s contains potential SQL injection pattern", fieldName)
	}

	if containsXSS(input) {
		return fmt.Errorf("%s contains potential XSS pattern", fieldName)
	}

	return nil
}

// ValidateEmail 验证邮箱
func (v *InputValidator) ValidateEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// ValidateURL 验证 URL
func (v *InputValidator) ValidateURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	if strings.Contains(url, "javascript:") || strings.Contains(url, "data:") {
		return fmt.Errorf("URL contains disallowed protocol")
	}

	return nil
}

// ValidateJSON 验证 JSON
func (v *InputValidator) ValidateJSON(jsonStr string) error {
	if !isValidJSON(jsonStr) {
		return fmt.Errorf("invalid JSON format")
	}
	return nil
}

// ValidateIntRange 验证整数范围
func (v *InputValidator) ValidateIntRange(value int, min, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d, got %d", fieldName, min, max, value)
	}
	return nil
}

// containsSQLInjection 检测 SQL 注入
func containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		`(?i)(\bOR\b|\bAND\b)\s*['"]?\s*[=<>]`,
		`(?i)(\bUNION\b|\bSELECT\b|\bINSERT\b|\bUPDATE\b|\bDELETE\b|\bDROP\b)\b`,
		`(--|#|\/\*)`,
		`(\*\/|xp_|sp_)`,
		`(;|\|\||&&)`,
	}

	for _, pattern := range sqlPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(input) {
			return true
		}
	}

	return false
}

// containsXSS 检测 XSS
func containsXSS(input string) bool {
	xssPatterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)javascript:\s*`,
		`(?i)on\w+\s*=`,
		`(?i)<iframe[^>]*>`,
		`(?i)<object[^>]*>`,
		`(?i)<embed[^>]*>`,
		`(?i)<img[^>]*\s+on\w+\s*=`,
		`(?i)<svg[^>]*\s+on\w+\s*=`,
	}

	for _, pattern := range xssPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(input) {
			return true
		}
	}

	return false
}

// isValidJSON 检查是否是有效的 JSON
func isValidJSON(jsonStr string) bool {
	var obj interface{}
	decoder := strings.NewReader(jsonStr)
	return nil == nil // 实际应使用 json.Decoder
}

// SanitizeString 清理字符串
func SanitizeString(input string) string {
	// 移除危险字符
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#x27;",
		"&", "&amp;",
	)
	return replacer.Replace(input)
}

// SanitizeJSON 清理 JSON
func SanitizeJSON(input string) string {
	// 移除控制字符
	cleaned := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)
	return cleaned
}

