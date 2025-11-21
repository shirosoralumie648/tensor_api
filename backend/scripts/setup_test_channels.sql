-- 测试渠道配置脚本

-- 插入测试渠道 (OpenAI)
INSERT INTO channels (
    name, 
    type, 
    api_key, 
    base_url, 
    weight, 
    support_models, 
    enabled
) VALUES (
    'openai-test',
    'openai',
    'sk-test-key-12345',  -- 测试 API 密钥
    'https://api.openai.com',
    1,
    'gpt-3.5-turbo,gpt-4,gpt-4-turbo,gpt-4o',
    true
);

-- 获取插入的渠道 ID
-- SELECT id FROM channels WHERE name = 'openai-test';

-- 插入模型价格 (假设渠道 ID 为 1)
-- 输入价格: 0.0005 美元/1K tokens
-- 输出价格: 0.0015 美元/1K tokens
INSERT INTO model_prices (
    channel_id, 
    model, 
    input_price, 
    output_price
) VALUES
    (1, 'gpt-3.5-turbo', 0.0005, 0.0015),
    (1, 'gpt-4', 0.03, 0.06),
    (1, 'gpt-4-turbo', 0.01, 0.03),
    (1, 'gpt-4o', 0.005, 0.015);

-- 验证插入结果
SELECT 'Channels:' as info;
SELECT * FROM channels WHERE enabled = true;

SELECT 'Model Prices:' as info;
SELECT * FROM model_prices LIMIT 10;

