-- 创建适配器配置表
-- Version: 000013
-- Description: 存储适配器配置，实现动态加载和配置驱动

BEGIN;

CREATE TABLE IF NOT EXISTS adapter_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    version VARCHAR(20) DEFAULT 'v1.0.0',
    handler_class VARCHAR(200),
    supported_models TEXT,
    default_config JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT true,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_adapter_configs_name 
ON adapter_configs(name) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_adapter_configs_type 
ON adapter_configs(type) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_adapter_configs_enabled 
ON adapter_configs(enabled) WHERE deleted_at IS NULL;

-- 插入默认适配器配置
INSERT INTO adapter_configs (name, type, handler_class, supported_models, default_config, description)
VALUES 
    ('openai', 'chat', 'openai.OpenAIAdapter', 
     'gpt-4,gpt-4-turbo,gpt-4-turbo-preview,gpt-4o,gpt-4o-mini,gpt-3.5-turbo,gpt-3.5-turbo-16k', 
     '{"timeout": 30, "max_retries": 3, "api_version": "v1"}', 
     'OpenAI 官方 API 适配器'),
    
    ('claude', 'chat', 'claude.ClaudeAdapter', 
     'claude-3-opus-20240229,claude-3-sonnet-20240229,claude-3-haiku-20240307,claude-3.5-sonnet-20241022,claude-2.1,claude-2.0', 
     '{"timeout": 60, "max_retries": 2, "api_version": "2023-06-01"}', 
     'Anthropic Claude API 适配器'),
    
    ('gemini', 'chat', 'gemini.GeminiAdapter', 
     'gemini-pro,gemini-1.5-pro,gemini-1.5-flash,gemini-1.5-flash-8b,gemini-2.0-flash-exp', 
     '{"timeout": 45, "max_retries": 2, "api_version": "v1"}', 
     'Google Gemini API 适配器'),
    
    ('azure', 'chat', 'azure.AzureOpenAIAdapter', 
     'gpt-4,gpt-35-turbo,gpt-4-32k,gpt-4-turbo,gpt-4o', 
     '{"timeout": 30, "max_retries": 3, "api_version": "2024-02-01"}', 
     'Azure OpenAI 适配器'),
    
    ('baidu', 'chat', 'baidu.BaiduAdapter', 
     'ERNIE-Bot,ERNIE-Bot-turbo,ERNIE-Bot-4,ERNIE-Speed,ERNIE-3.5-8K', 
     '{"timeout": 30, "max_retries": 2}', 
     '百度文心一言适配器'),
    
    ('qwen', 'chat', 'qwen.QwenAdapter', 
     'qwen-turbo,qwen-plus,qwen-max,qwen-max-longcontext,qwen2.5-72b-instruct', 
     '{"timeout": 30, "max_retries": 2}', 
     '阿里通义千问适配器'),
    
    ('deepseek', 'chat', 'deepseek.DeepSeekAdapter', 
     'deepseek-chat,deepseek-coder', 
     '{"timeout": 30, "max_retries": 2}', 
     'DeepSeek 适配器'),
    
    ('moonshot', 'chat', 'moonshot.MoonshotAdapter', 
     'moonshot-v1-8k,moonshot-v1-32k,moonshot-v1-128k', 
     '{"timeout": 30, "max_retries": 2}', 
     'Moonshot (Kimi) 适配器'),
    
    ('zhipu', 'chat', 'zhipu.ZhipuAdapter', 
     'glm-4,glm-4-plus,glm-4-air,glm-3-turbo', 
     '{"timeout": 30, "max_retries": 2}', 
     '智谱 GLM 适配器'),
    
    ('ollama', 'chat', 'ollama.OllamaAdapter', 
     'llama2,mistral,codellama,vicuna,mixtral', 
     '{"timeout": 60, "max_retries": 2}', 
     'Ollama 本地模型适配器')
ON CONFLICT (name) DO NOTHING;

-- 添加注释
COMMENT ON TABLE adapter_configs IS '适配器配置表 - 支持动态加载和配置驱动';
COMMENT ON COLUMN adapter_configs.handler_class IS '处理器类名（用于动态加载）';
COMMENT ON COLUMN adapter_configs.default_config IS '默认配置（JSON格式，包含超时、重试等参数）';

COMMIT;
