package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

// ChannelStatus 渠道状态
const (
	ChannelStatusEnabled      = 1 // 启用
	ChannelStatusDisabled     = 2 // 手动禁用
	ChannelStatusAutoDisabled = 3 // 自动禁用
)

// MultiKeyMode 多密钥模式
const (
	MultiKeyModeRandom  = 1 // 随机
	MultiKeyModePolling = 2 // 轮询
)

// Channel 代表一个 AI 服务渠道（例如 OpenAI、Claude、Gemini 等）
type Channel struct {
	ID      int    `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"size:100;uniqueIndex" json:"name"` // 渠道名称
	Type    string `gorm:"size:50" json:"type"`              // 渠道类型
	APIKey  string `gorm:"size:500" json:"api_key"`          // API 密钥
	BaseURL string `gorm:"size:500" json:"base_url"`         // 基础 URL

	// 负载均衡
	Weight   int   `gorm:"default:1" json:"weight"`         // 权重
	Priority int64 `gorm:"default:0;index" json:"priority"` // 优先级

	// 分组和标签
	Group string  `gorm:"type:varchar(64);default:'default';index" json:"group"` // 用户分组
	Tag   *string `gorm:"size:100;index" json:"tag"`                             // 标签

	// 限流和配额
	MaxRateLimit       int     `json:"max_rate_limit"`                        // 最大请求速率
	UsedQuota          int64   `gorm:"default:0" json:"used_quota"`           // 已使用配额
	Balance            float64 `gorm:"default:0" json:"balance"`              // 余额
	BalanceUpdatedTime int64   `gorm:"default:0" json:"balance_updated_time"` // 余额更新时间

	// 模型配置
	ModelMapping  *string `gorm:"type:jsonb" json:"model_mapping"` // 模型映射
	SupportModels string  `gorm:"type:text" json:"support_models"` // 支持的模型列表

	// 多密钥配置
	ChannelInfo ChannelInfo `gorm:"type:jsonb;default:'{}';column:channel_info" json:"channel_info"`

	// 状态和监控
	Status       int   `gorm:"default:1;index" json:"status"`  // 状态
	Enabled      bool  `gorm:"default:true" json:"enabled"`    // 兼容旧版
	ResponseTime int   `gorm:"default:0" json:"response_time"` // 平均响应时间(ms)
	TestTime     int64 `gorm:"default:0" json:"test_time"`     // 最后测试时间
	AutoBan      int   `gorm:"default:1" json:"auto_ban"`      // 自动禁用开关

	// 高级配置
	StatusCodeMapping *string `gorm:"type:varchar(1024)" json:"status_code_mapping"`
	ParamOverride     *string `gorm:"type:text" json:"param_override"`
	HeaderOverride    *string `gorm:"type:text" json:"header_override"`
	OtherInfo         string  `gorm:"type:text;default:'{}}'" json:"other_info"`
	OtherSettings     string  `gorm:"type:text;default:'{}}'" json:"other_settings"`
	Remark            *string `gorm:"type:varchar(255)" json:"remark"`

	// 时间戳
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`

	// 缓存字段（不持久化）
	Keys []string `gorm:"-" json:"-"`
}

// ChannelInfo 多密钥配置
type ChannelInfo struct {
	IsMultiKey             bool           `json:"is_multi_key"`
	MultiKeySize           int            `json:"multi_key_size"`
	MultiKeyMode           int            `json:"multi_key_mode"` // 1: random, 2: polling
	MultiKeyPollingIndex   int32          `json:"multi_key_polling_index"`
	MultiKeyStatusList     map[int]int    `json:"multi_key_status_list,omitempty"`
	MultiKeyDisabledReason map[int]string `json:"multi_key_disabled_reason,omitempty"`
	MultiKeyDisabledTime   map[int]int64  `json:"multi_key_disabled_time,omitempty"`
}

// Value 实现 driver.Valuer 接口
func (c ChannelInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ChannelInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

func (Channel) TableName() string {
	return "channels"
}

// GetKeys 获取所有密钥
func (c *Channel) GetKeys() []string {
	if c.Keys != nil {
		return c.Keys
	}

	if !c.ChannelInfo.IsMultiKey {
		c.Keys = []string{c.APIKey}
		return c.Keys
	}

	// 从 APIKey 解析（格式：key1,key2,key3 或 key1\nkey2\nkey3）
	var keys []string
	if strings.Contains(c.APIKey, "\n") {
		keys = strings.Split(c.APIKey, "\n")
	} else {
		keys = strings.Split(c.APIKey, ",")
	}

	c.Keys = make([]string, 0, len(keys))
	for _, key := range keys {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			c.Keys = append(c.Keys, trimmed)
		}
	}

	return c.Keys
}

// GetNextEnabledKey 获取下一个可用密钥（并发安全）
func (c *Channel) GetNextEnabledKey() (string, int) {
	if !c.ChannelInfo.IsMultiKey || c.ChannelInfo.MultiKeySize == 0 {
		return c.APIKey, 0
	}

	keys := c.GetKeys()
	if len(keys) == 0 {
		return c.APIKey, 0
	}

	// 随机模式
	if c.ChannelInfo.MultiKeyMode == MultiKeyModeRandom {
		enabledKeys := make([]string, 0)
		enabledIndices := make([]int, 0)
		for i, key := range keys {
			if status, ok := c.ChannelInfo.MultiKeyStatusList[i]; !ok || status == 1 {
				enabledKeys = append(enabledKeys, key)
				enabledIndices = append(enabledIndices, i)
			}
		}
		if len(enabledKeys) == 0 {
			return keys[0], 0
		}
		idx := rand.Intn(len(enabledKeys))
		return enabledKeys[idx], enabledIndices[idx]
	}

	// 轮询模式：使用原子操作递增索引（并发安全）
	for i := 0; i < len(keys); i++ {
		index := atomic.AddInt32(&c.ChannelInfo.MultiKeyPollingIndex, 1)
		idx := int(index) % len(keys)
		if status, ok := c.ChannelInfo.MultiKeyStatusList[idx]; !ok || status == 1 {
			return keys[idx], idx
		}
	}

	return keys[0], 0
}

// IsEnabled 检查渠道是否启用
func (c *Channel) IsEnabled() bool {
	return c.Status == ChannelStatusEnabled || c.Enabled
}

// GetSupportedModels 获取支持的模型列表
func (c *Channel) GetSupportedModels() []string {
	if c.SupportModels == "" {
		return []string{}
	}

	// 解析逗号分隔的模型列表
	models := make([]string, 0)
	for _, model := range strings.Split(c.SupportModels, ",") {
		model = strings.TrimSpace(model)
		if model != "" {
			models = append(models, model)
		}
	}

	return models
}
