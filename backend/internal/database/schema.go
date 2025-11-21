package database

import (
	"fmt"
	"sync"
	"time"
)

// SchemaVersion 数据库版本
type SchemaVersion struct {
	Version   int
	CreatedAt time.Time
	Description string
}

// IndexDefinition 索引定义
type IndexDefinition struct {
	Name       string
	Table      string
	Columns    []string
	Unique     bool
	Type       string // btree, hash, gin, etc
	Partial    string // 部分索引条件
}

// TableSchema 表结构
type TableSchema struct {
	Name        string
	Columns     []ColumnDef
	Indexes     []IndexDefinition
	Constraints []string
	PartitionBy string // 分区字段
}

// ColumnDef 列定义
type ColumnDef struct {
	Name        string
	Type        string
	NotNull     bool
	Default     string
	Comment     string
}

// SchemaManager 数据库 Schema 管理器
type SchemaManager struct {
	mu              sync.RWMutex
	schemas         map[string]*TableSchema
	versions        []SchemaVersion
	currentVersion  int
	migrations      map[int][]string // 迁移 SQL
}

// NewSchemaManager 创建 Schema 管理器
func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		schemas:    make(map[string]*TableSchema),
		versions:   make([]SchemaVersion, 0),
		migrations: make(map[int][]string),
	}
}

// RegisterSchema 注册表结构
func (sm *SchemaManager) RegisterSchema(schema *TableSchema) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.schemas[schema.Name]; exists {
		return fmt.Errorf("schema %s already exists", schema.Name)
	}

	sm.schemas[schema.Name] = schema
	return nil
}

// AddMigration 添加迁移
func (sm *SchemaManager) AddMigration(version int, sqls []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.migrations[version]; exists {
		return fmt.Errorf("migration version %d already exists", version)
	}

	sm.migrations[version] = sqls
	sm.currentVersion = version
	sm.versions = append(sm.versions, SchemaVersion{
		Version:     version,
		CreatedAt:   time.Now(),
		Description: fmt.Sprintf("Migration v%d", version),
	})

	return nil
}

// GetSchema 获取表结构
func (sm *SchemaManager) GetSchema(tableName string) (*TableSchema, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	schema, exists := sm.schemas[tableName]
	if !exists {
		return nil, fmt.Errorf("schema %s not found", tableName)
	}

	return schema, nil
}

// GetMigrationSQL 获取迁移 SQL
func (sm *SchemaManager) GetMigrationSQL(fromVersion, toVersion int) ([]string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sqls []string
	for v := fromVersion + 1; v <= toVersion; v++ {
		if migration, exists := sm.migrations[v]; exists {
			sqls = append(sqls, migration...)
		}
	}

	if len(sqls) == 0 {
		return nil, fmt.Errorf("no migrations found from %d to %d", fromVersion, toVersion)
	}

	return sqls, nil
}

// GenerateCreateTableSQL 生成创建表 SQL
func (sm *SchemaManager) GenerateCreateTableSQL(tableName string) (string, error) {
	schema, err := sm.GetSchema(tableName)
	if err != nil {
		return "", err
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName)

	// 添加列
	for i, col := range schema.Columns {
		sql += fmt.Sprintf("  %s %s", col.Name, col.Type)
		if col.NotNull {
			sql += " NOT NULL"
		}
		if col.Default != "" {
			sql += fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		if i < len(schema.Columns)-1 {
			sql += ",\n"
		} else {
			sql += "\n"
		}
	}

	// 添加约束
	for _, constraint := range schema.Constraints {
		sql += fmt.Sprintf("  %s,\n", constraint)
	}

	sql += ");\n"

	// 添加索引
	for _, idx := range schema.Indexes {
		sql += sm.generateCreateIndexSQL(idx) + ";\n"
	}

	// 添加分区
	if schema.PartitionBy != "" {
		sql += fmt.Sprintf("-- PARTITION BY %s\n", schema.PartitionBy)
	}

	return sql, nil
}

// generateCreateIndexSQL 生成创建索引 SQL
func (sm *SchemaManager) generateCreateIndexSQL(idx IndexDefinition) string {
	indexType := "INDEX"
	if idx.Unique {
		indexType = "UNIQUE INDEX"
	}

	columnsStr := ""
	for i, col := range idx.Columns {
		columnsStr += col
		if i < len(idx.Columns)-1 {
			columnsStr += ", "
		}
	}

	sql := fmt.Sprintf("CREATE %s %s ON %s (%s)", indexType, idx.Name, idx.Table, columnsStr)

	if idx.Type != "" && idx.Type != "btree" {
		sql += fmt.Sprintf(" USING %s", idx.Type)
	}

	if idx.Partial != "" {
		sql += fmt.Sprintf(" WHERE %s", idx.Partial)
	}

	return sql
}

// ValidateSchema 验证 Schema
func (sm *SchemaManager) ValidateSchema(tableName string) error {
	schema, err := sm.GetSchema(tableName)
	if err != nil {
		return err
	}

	// 检查列名重复
	columnNames := make(map[string]bool)
	for _, col := range schema.Columns {
		if columnNames[col.Name] {
			return fmt.Errorf("duplicate column name: %s", col.Name)
		}
		columnNames[col.Name] = true
	}

	// 检查索引列存在
	for _, idx := range schema.Indexes {
		for _, col := range idx.Columns {
			if !columnNames[col] {
				return fmt.Errorf("index column %s not found in table %s", col, tableName)
			}
		}
	}

	return nil
}

// ListSchemas 列出所有表
func (sm *SchemaManager) ListSchemas() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	tables := make([]string, 0, len(sm.schemas))
	for name := range sm.schemas {
		tables = append(tables, name)
	}

	return tables
}

// GetCurrentVersion 获取当前版本
func (sm *SchemaManager) GetCurrentVersion() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.currentVersion
}

// GetVersionHistory 获取版本历史
func (sm *SchemaManager) GetVersionHistory() []SchemaVersion {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return append([]SchemaVersion{}, sm.versions...)
}

// InitializeDatabase 初始化数据库
func (sm *SchemaManager) InitializeDatabase() ([]string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sqls []string

	// 为每个表生成创建 SQL
	for tableName := range sm.schemas {
		sql, err := sm.GenerateCreateTableSQL(tableName)
		if err != nil {
			return nil, err
		}
		sqls = append(sqls, sql)
	}

	return sqls, nil
}

// OptimizeSchema 优化 Schema
func (sm *SchemaManager) OptimizeSchema(tableName string) ([]string, error) {
	schema, err := sm.GetSchema(tableName)
	if err != nil {
		return nil, err
	}

	var optimizations []string

	// 建议添加缺失的索引
	// 根据常见查询模式建议索引

	// 建议分区策略
	if schema.PartitionBy == "" {
		optimizations = append(optimizations, 
			fmt.Sprintf("-- Consider partitioning table %s", tableName))
	}

	// 建议列类型优化
	for _, col := range schema.Columns {
		if col.Type == "TEXT" {
			optimizations = append(optimizations,
				fmt.Sprintf("-- Consider using VARCHAR for column %s", col.Name))
		}
	}

	return optimizations, nil
}


