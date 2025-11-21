CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_id INT,
    group_id UUID,
    title VARCHAR(200),
    description TEXT,
    pinned BOOLEAN DEFAULT FALSE,
    archived BOOLEAN DEFAULT FALSE,
    model VARCHAR(100),
    temperature FLOAT DEFAULT 0.7,
    top_p FLOAT DEFAULT 1.0,
    max_tokens INT,
    system_role TEXT,
    context_length INT DEFAULT 4,
    plugin_ids INT[],
    knowledge_base_ids INT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_sessions_user ON sessions(user_id, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_pinned ON sessions(user_id, pinned DESC, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_archived ON sessions(user_id, archived, updated_at DESC) WHERE deleted_at IS NULL;

