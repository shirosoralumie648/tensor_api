CREATE TABLE user_settings (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) DEFAULT 'zh-CN',
    theme VARCHAR(20) DEFAULT 'auto',
    font_size INT DEFAULT 14,
    tts_enabled BOOLEAN DEFAULT FALSE,
    tts_voice VARCHAR(50),
    tts_speed FLOAT DEFAULT 1.0,
    stt_enabled BOOLEAN DEFAULT FALSE,
    send_key VARCHAR(20) DEFAULT 'Enter',
    custom_config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


