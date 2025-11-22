package adapter

// ProviderType 提供商类型常量（参考 New API）
type ProviderType int

const (
	ProviderOpenAI ProviderType = iota + 1
	ProviderAnthropic
	ProviderGoogle
	ProviderBaidu
	ProviderQwen
	ProviderAzure
	ProviderCohere
	ProviderMistral
	ProviderDeepSeek
	ProviderMoonshot
	ProviderZhipu
	ProviderXunfei
	ProviderTencent
	ProviderVolcEngine
)

// String 返回提供商类型的字符串表示
func (pt ProviderType) String() string {
	switch pt {
	case ProviderOpenAI:
		return "openai"
	case ProviderAnthropic:
		return "anthropic"
	case ProviderGoogle:
		return "google"
	case ProviderBaidu:
		return "baidu"
	case ProviderQwen:
		return "qwen"
	case ProviderAzure:
		return "azure"
	case ProviderCohere:
		return "cohere"
	case ProviderMistral:
		return "mistral"
	case ProviderDeepSeek:
		return "deepseek"
	case ProviderMoonshot:
		return "moonshot"
	case ProviderZhipu:
		return "zhipu"
	case ProviderXunfei:
		return "xunfei"
	case ProviderTencent:
		return "tencent"
	case ProviderVolcEngine:
		return "volcengine"
	default:
		return "unknown"
	}
}

// ParseProviderType 从字符串解析提供商类型
func ParseProviderType(s string) ProviderType {
	switch s {
	case "openai":
		return ProviderOpenAI
	case "anthropic", "claude":
		return ProviderAnthropic
	case "google", "gemini":
		return ProviderGoogle
	case "baidu", "wenxin":
		return ProviderBaidu
	case "qwen", "tongyi":
		return ProviderQwen
	case "azure":
		return ProviderAzure
	case "cohere":
		return ProviderCohere
	case "mistral":
		return ProviderMistral
	case "deepseek":
		return ProviderDeepSeek
	case "moonshot":
		return ProviderMoonshot
	case "zhipu":
		return ProviderZhipu
	case "xunfei":
		return ProviderXunfei
	case "tencent":
		return ProviderTencent
	case "volcengine":
		return ProviderVolcEngine
	default:
		return 0
	}
}

// RelayMode 中继模式（参考 New API）
type RelayMode int

const (
	RelayModeChatCompletions RelayMode = iota
	RelayModeEmbeddings
	RelayModeImages
	RelayModeAudio
	RelayModeRerank
	RelayModeModeration
)

func (rm RelayMode) String() string {
	switch rm {
	case RelayModeChatCompletions:
		return "chat_completions"
	case RelayModeEmbeddings:
		return "embeddings"
	case RelayModeImages:
		return "images"
	case RelayModeAudio:
		return "audio"
	case RelayModeRerank:
		return "rerank"
	case RelayModeModeration:
		return "moderation"
	default:
		return "unknown"
	}
}
