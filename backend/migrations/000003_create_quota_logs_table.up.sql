CREATE TABLE quota_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type INT NOT NULL,
    amount BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    description TEXT,
    related_id VARCHAR(100),
    related_type VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_quota_logs_user ON quota_logs(user_id, created_at DESC);
CREATE INDEX idx_quota_logs_type ON quota_logs(type, created_at DESC);
CREATE INDEX idx_quota_logs_created ON quota_logs(created_at DESC);


