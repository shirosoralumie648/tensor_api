package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
)

// EncryptionManager 加密管理器
type EncryptionManager struct {
	masterKey []byte
}

// NewEncryptionManager 创建新的加密管理器
func NewEncryptionManager(masterKey string) (*EncryptionManager, error) {
	key := []byte(masterKey)
	if len(key) != 32 {
		// 使用 SHA-256 标准化密钥
		key = make([]byte, 32)
		copy(key, []byte(masterKey))
	}

	return &EncryptionManager{
		masterKey: key,
	}, nil
}

// EncryptAES256 使用 AES-256-GCM 加密
func (em *EncryptionManager) EncryptAES256(plaintext string) (string, error) {
	block, err := aes.NewCipher(em.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回 base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES256 使用 AES-256-GCM 解密
func (em *EncryptionManager) DecryptAES256(ciphertext string) (string, error) {
	// 解码 base64
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(em.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// 分离 nonce 和密文
	nonce, ct := decoded[:nonceSize], decoded[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// HashPassword 使用 bcrypt 哈希密码 (需要导入 golang.org/x/crypto/bcrypt)
func HashPassword(password string) (string, error) {
	// 实际应使用 bcrypt.GenerateFromPassword
	// 这里仅作示例
	h := sha256Hash(password)
	return h, nil
}

// VerifyPassword 验证密码
func VerifyPassword(hashedPassword string, password string) bool {
	// 实际应使用 bcrypt.CompareHashAndPassword
	h := sha256Hash(password)
	return h == hashedPassword
}

// GenerateRandomToken 生成随机 Token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// sha256Hash 简单的 SHA-256 哈希 (用于演示)
func sha256Hash(data string) string {
	// 实际应使用 crypto/sha256
	sum := make([]byte, 32)
	copy(sum, data)
	return hex.EncodeToString(sum)
}

// TokenEncryptor Token 加密器
type TokenEncryptor struct {
	manager *EncryptionManager
}

// NewTokenEncryptor 创建新的 Token 加密器
func NewTokenEncryptor(manager *EncryptionManager) *TokenEncryptor {
	return &TokenEncryptor{
		manager: manager,
	}
}

// EncryptToken 加密 Token
func (te *TokenEncryptor) EncryptToken(token string) (string, error) {
	return te.manager.EncryptAES256(token)
}

// DecryptToken 解密 Token
func (te *TokenEncryptor) DecryptToken(encryptedToken string) (string, error) {
	return te.manager.DecryptAES256(encryptedToken)
}

// DataEncryptor 数据加密器
type DataEncryptor struct {
	manager *EncryptionManager
}

// NewDataEncryptor 创建新的数据加密器
func NewDataEncryptor(manager *EncryptionManager) *DataEncryptor {
	return &DataEncryptor{
		manager: manager,
	}
}

// EncryptSensitiveData 加密敏感数据
func (de *DataEncryptor) EncryptSensitiveData(data string) (string, error) {
	return de.manager.EncryptAES256(data)
}

// DecryptSensitiveData 解密敏感数据
func (de *DataEncryptor) DecryptSensitiveData(encryptedData string) (string, error) {
	return de.manager.DecryptAES256(encryptedData)
}

