-- 删除触发器
DROP TRIGGER IF EXISTS trigger_update_kb_timestamp ON documents;
DROP FUNCTION IF EXISTS update_kb_timestamp();

-- 删除文本块表
DROP TABLE IF EXISTS chunks;

-- 删除文档表
DROP TABLE IF EXISTS documents;

-- 删除知识库表
DROP TABLE IF EXISTS knowledge_bases;

-- 禁用 pgvector 扩展（可选）
-- DROP EXTENSION IF EXISTS vector;

