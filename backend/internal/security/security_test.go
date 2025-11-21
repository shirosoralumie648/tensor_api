package security

import (
	"testing"
)

// TestInputValidator 测试输入验证器
func TestInputValidator(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name      string
		input     string
		fieldName string
		shouldErr bool
	}{
		{
			name:      "valid string",
			input:     "valid_input",
			fieldName: "test",
			shouldErr: false,
		},
		{
			name:      "empty string",
			input:     "",
			fieldName: "test",
			shouldErr: true,
		},
		{
			name:      "sql injection attempt",
			input:     "'; DROP TABLE users; --",
			fieldName: "test",
			shouldErr: true,
		},
		{
			name:      "xss attempt",
			input:     "<script>alert('xss')</script>",
			fieldName: "test",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateString(tt.input, tt.fieldName)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateString() error = %v, shouldErr = %v", err, tt.shouldErr)
			}
		})
	}
}

// TestValidateEmail 测试邮箱验证
func TestValidateEmail(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name      string
		email     string
		shouldErr bool
	}{
		{
			name:      "valid email",
			email:     "test@example.com",
			shouldErr: false,
		},
		{
			name:      "invalid email - no @",
			email:     "testexample.com",
			shouldErr: true,
		},
		{
			name:      "invalid email - no domain",
			email:     "test@",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateEmail() error = %v, shouldErr = %v", err, tt.shouldErr)
			}
		})
	}
}

// TestValidateURL 测试 URL 验证
func TestValidateURL(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{
			name:      "valid https url",
			url:       "https://example.com/path",
			shouldErr: false,
		},
		{
			name:      "valid http url",
			url:       "http://example.com/path",
			shouldErr: false,
		},
		{
			name:      "javascript protocol",
			url:       "javascript:alert('xss')",
			shouldErr: true,
		},
		{
			name:      "data protocol",
			url:       "data:text/html,<script>alert('xss')</script>",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateURL() error = %v, shouldErr = %v", err, tt.shouldErr)
			}
		})
	}
}

// TestValidateIntRange 测试整数范围验证
func TestValidateIntRange(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		shouldErr bool
	}{
		{
			name:      "value in range",
			value:     5,
			min:       1,
			max:       10,
			shouldErr: false,
		},
		{
			name:      "value below min",
			value:     0,
			min:       1,
			max:       10,
			shouldErr: true,
		},
		{
			name:      "value above max",
			value:     11,
			min:       1,
			max:       10,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateIntRange(tt.value, tt.min, tt.max, "test")
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateIntRange() error = %v, shouldErr = %v", err, tt.shouldErr)
			}
		})
	}
}

// TestEncryption 测试加密
func TestEncryption(t *testing.T) {
	manager, err := NewEncryptionManager("test_master_key_32_characters_long")
	if err != nil {
		t.Fatalf("NewEncryptionManager() error = %v", err)
	}

	plaintext := "sensitive data"

	// 测试加密
	encrypted, err := manager.EncryptAES256(plaintext)
	if err != nil {
		t.Fatalf("EncryptAES256() error = %v", err)
	}

	if encrypted == plaintext {
		t.Error("EncryptAES256() returned plaintext instead of ciphertext")
	}

	// 测试解密
	decrypted, err := manager.DecryptAES256(encrypted)
	if err != nil {
		t.Fatalf("DecryptAES256() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("DecryptAES256() got %q, want %q", decrypted, plaintext)
	}
}

// TestSanitizeString 测试字符串清理
func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no dangerous chars",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "html tags",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#x27;xss&#x27;)&lt;/script&gt;",
		},
		{
			name:     "quotes",
			input:    `"hello" 'world'`,
			expected: `&quot;hello&quot; &#x27;world&#x27;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGenerateRandomToken 测试随机 Token 生成
func TestGenerateRandomToken(t *testing.T) {
	token1, err := GenerateRandomToken(32)
	if err != nil {
		t.Fatalf("GenerateRandomToken() error = %v", err)
	}

	token2, err := GenerateRandomToken(32)
	if err != nil {
		t.Fatalf("GenerateRandomToken() error = %v", err)
	}

	if token1 == token2 {
		t.Error("GenerateRandomToken() generated duplicate tokens")
	}

	if len(token1) == 0 || len(token2) == 0 {
		t.Error("GenerateRandomToken() returned empty token")
	}
}

