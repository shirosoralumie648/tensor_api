CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    role INT DEFAULT 1,
    quota BIGINT DEFAULT 0,
    total_quota BIGINT DEFAULT 0,
    used_quota BIGINT DEFAULT 0,
    invite_code VARCHAR(20) UNIQUE,
    invited_by INT REFERENCES users(id),
    status INT DEFAULT 1,
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_invite_code ON users(invite_code);
CREATE INDEX idx_users_status ON users(status, created_at DESC) WHERE deleted_at IS NULL;


