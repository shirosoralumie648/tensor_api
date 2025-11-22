package database

import (
	"testing"
)

func TestSchemaManager(t *testing.T) {
	sm := NewSchemaManager()

	schema := &TableSchema{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: "BIGINT", NotNull: true},
			{Name: "username", Type: "VARCHAR(255)", NotNull: true},
			{Name: "email", Type: "VARCHAR(255)", NotNull: true},
			{Name: "created_at", Type: "TIMESTAMP", Default: "CURRENT_TIMESTAMP"},
		},
		Indexes: []IndexDefinition{
			{Name: "idx_username", Table: "users", Columns: []string{"username"}, Unique: true},
			{Name: "idx_email", Table: "users", Columns: []string{"email"}},
		},
		Constraints: []string{
			"PRIMARY KEY (id)",
		},
	}

	err := sm.RegisterSchema(schema)
	if err != nil {
		t.Fatalf("RegisterSchema failed: %v", err)
	}

	// 验证 Schema
	retrieved, err := sm.GetSchema("users")
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	if retrieved.Name != "users" {
		t.Errorf("Expected schema name 'users', got '%s'", retrieved.Name)
	}

	// 测试 Schema 验证
	err = sm.ValidateSchema("users")
	if err != nil {
		t.Fatalf("ValidateSchema failed: %v", err)
	}

	// 生成 CREATE TABLE SQL
	sql, err := sm.GenerateCreateTableSQL("users")
	if err != nil {
		t.Fatalf("GenerateCreateTableSQL failed: %v", err)
	}

	if sql == "" {
		t.Error("Expected SQL to be generated")
	}
}

func TestMigrationManager(t *testing.T) {
	mm := NewMigrationManager()

	// 注册迁移
	upSQL := []string{
		"CREATE TABLE users (id BIGINT PRIMARY KEY, username VARCHAR(255))",
	}
	downSQL := []string{
		"DROP TABLE users",
	}

	err := mm.RegisterMigration(1, "Create users table", upSQL, downSQL)
	if err != nil {
		t.Fatalf("RegisterMigration failed: %v", err)
	}

	// 验证迁移
	err = mm.ValidateMigrations()
	if err != nil {
		t.Fatalf("ValidateMigrations failed: %v", err)
	}

	// 检查待执行迁移
	pending := mm.GetPendingMigrations()
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending migration, got %d", len(pending))
	}

	// 执行迁移
	err = mm.ExecuteMigration(1)
	if err != nil {
		t.Fatalf("ExecuteMigration failed: %v", err)
	}

	// 检查执行历史
	executed := mm.GetExecutedMigrations()
	if len(executed) != 1 {
		t.Errorf("Expected 1 executed migration, got %d", len(executed))
	}

	// 检查当前版本
	version := mm.GetCurrentVersion()
	if version != 1 {
		t.Errorf("Expected current version 1, got %d", version)
	}
}

func TestMigrationRollback(t *testing.T) {
	mm := NewMigrationManager()

	upSQL := []string{"CREATE TABLE test (id BIGINT)"}
	downSQL := []string{"DROP TABLE test"}

	mm.RegisterMigration(1, "Create test table", upSQL, downSQL)
	mm.ExecuteMigration(1)

	// 回滚迁移
	err := mm.RollbackMigration(1)
	if err != nil {
		t.Fatalf("RollbackMigration failed: %v", err)
	}

	migration, _ := mm.GetMigration(1)
	if migration.Status != StatusRolledBack {
		t.Errorf("Expected status RolledBack, got %s", migration.Status)
	}
}

func TestGenerateMigrationScript(t *testing.T) {
	mm := NewMigrationManager()

	mm.RegisterMigration(1, "Migration 1", []string{"SQL1"}, []string{"DOWN1"})
	mm.RegisterMigration(2, "Migration 2", []string{"SQL2"}, []string{"DOWN2"})

	script, err := mm.GenerateMigrationScript(0, 2)
	if err != nil {
		t.Fatalf("GenerateMigrationScript failed: %v", err)
	}

	if script == "" {
		t.Error("Expected script to be generated")
	}

	if len(script) < 10 {
		t.Error("Generated script too short")
	}
}

func TestMigrationHealth(t *testing.T) {
	mm := NewMigrationManager()

	mm.RegisterMigration(1, "Migration 1", []string{"SQL1"}, []string{"DOWN1"})
	mm.ExecuteMigration(1)

	health := mm.CheckMigrationHealth()

	if health["total"] != 1 {
		t.Errorf("Expected total 1, got %v", health["total"])
	}

	if health["completed"] != 1 {
		t.Errorf("Expected completed 1, got %v", health["completed"])
	}
}

func TestMultipleMigrations(t *testing.T) {
	mm := NewMigrationManager()

	// 注册多个迁移
	for i := 1; i <= 5; i++ {
		upSQL := []string{
			"CREATE TABLE table" + string(rune(48+i)) + " (id BIGINT)",
		}
		downSQL := []string{
			"DROP TABLE table" + string(rune(48+i)),
		}

		err := mm.RegisterMigration(i, "Create table "+string(rune(48+i)), upSQL, downSQL)
		if err != nil {
			t.Fatalf("RegisterMigration %d failed: %v", i, err)
		}
	}

	// 验证
	err := mm.ValidateMigrations()
	if err != nil {
		t.Fatalf("ValidateMigrations failed: %v", err)
	}

	// 执行所有迁移
	for i := 1; i <= 5; i++ {
		err := mm.ExecuteMigration(i)
		if err != nil {
			t.Fatalf("ExecuteMigration %d failed: %v", i, err)
		}
	}

	// 检查最终版本
	version := mm.GetCurrentVersion()
	if version != 5 {
		t.Errorf("Expected current version 5, got %d", version)
	}
}

func BenchmarkSchemaValidation(b *testing.B) {
	sm := NewSchemaManager()

	schema := &TableSchema{
		Name: "test",
		Columns: []ColumnDef{
			{Name: "id", Type: "BIGINT", NotNull: true},
			{Name: "name", Type: "VARCHAR(255)"},
			{Name: "created_at", Type: "TIMESTAMP"},
		},
		Indexes: []IndexDefinition{
			{Name: "idx_name", Table: "test", Columns: []string{"name"}},
		},
	}

	sm.RegisterSchema(schema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.ValidateSchema("test")
	}
}

func BenchmarkMigrationExecution(b *testing.B) {
	mm := NewMigrationManager()

	mm.RegisterMigration(1, "Test Migration",
		[]string{"SELECT 1"},
		[]string{"SELECT 0"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.GetMigration(1)
	}
}
