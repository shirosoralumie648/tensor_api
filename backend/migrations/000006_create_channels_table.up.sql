-- 创建渠道表
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    api_key VARCHAR(500) NOT NULL,
    base_url VARCHAR(500),
    weight INT DEFAULT 1,
    max_rate_limit INT DEFAULT 0,
    model_mapping JSONB,
    support_models TEXT,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_channels_name ON channels(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_channels_enabled ON channels(enabled) WHERE deleted_at IS NULL;
CREATE INDEX idx_channels_type ON channels(type) WHERE deleted_at IS NULL;

-- 创建模型价格表
CREATE TABLE model_prices (
    id SERIAL PRIMARY KEY,
    channel_id INT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    model VARCHAR(100) NOT NULL,
    input_price FLOAT8 NOT NULL,
    output_price FLOAT8 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_model_prices_channel ON model_prices(channel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_model_prices_model ON model_prices(model) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_model_prices_channel_model ON model_prices(channel_id, model) WHERE deleted_at IS NULL;

