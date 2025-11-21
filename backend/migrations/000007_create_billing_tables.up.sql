-- 创建计费日志表
CREATE TABLE billing_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID,
    message_id UUID,
    model VARCHAR(100) NOT NULL,
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    cost BIGINT DEFAULT 0,
    cost_usd FLOAT8 DEFAULT 0,
    status INT DEFAULT 1,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_billing_logs_user ON billing_logs(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_logs_session ON billing_logs(session_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_logs_message ON billing_logs(message_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_logs_model ON billing_logs(model) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_logs_created ON billing_logs(created_at) WHERE deleted_at IS NULL;

-- 创建定价计划表
CREATE TABLE pricing_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_pricing_plans_name ON pricing_plans(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_pricing_plans_active ON pricing_plans(active) WHERE deleted_at IS NULL;

-- 创建额度变更日志表 (如果不存在)
CREATE TABLE IF NOT EXISTS quota_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation_type VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    reason TEXT,
    billing_log_id INT REFERENCES billing_logs(id),
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_quota_logs_user ON quota_logs(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_quota_logs_operation ON quota_logs(operation_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_quota_logs_created ON quota_logs(created_at) WHERE deleted_at IS NULL;

-- 创建发票表
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invoice_no VARCHAR(50) UNIQUE NOT NULL,
    total_cost BIGINT DEFAULT 0,
    total_usd FLOAT8 DEFAULT 0,
    item_count INT DEFAULT 0,
    status INT DEFAULT 1,
    issued_at TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_invoices_user ON invoices(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_no ON invoices(invoice_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_status ON invoices(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_created ON invoices(created_at) WHERE deleted_at IS NULL;

