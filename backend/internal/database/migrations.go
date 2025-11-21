package database

import (
	"fmt"
	"sync"
	"time"
)

// MigrationStatus 迁移状态
type MigrationStatus string

const (
	StatusPending   MigrationStatus = "pending"
	StatusRunning   MigrationStatus = "running"
	StatusCompleted MigrationStatus = "completed"
	StatusFailed    MigrationStatus = "failed"
	StatusRolledBack MigrationStatus = "rolled_back"
)

// Migration 迁移记录
type Migration struct {
	ID        int64
	Version   int
	Name      string
	UpSQL     []string
	DownSQL   []string
	Status    MigrationStatus
	CreatedAt time.Time
	ExecutedAt time.Time
	Duration  int64 // 毫秒
	Error     string
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	mu         sync.RWMutex
	migrations map[int]*Migration
	history    []*Migration
	maxVersion int
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrations: make(map[int]*Migration),
		history:    make([]*Migration, 0),
	}
}

// RegisterMigration 注册迁移
func (mm *MigrationManager) RegisterMigration(version int, name string, upSQL, downSQL []string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.migrations[version]; exists {
		return fmt.Errorf("migration version %d already exists", version)
	}

	migration := &Migration{
		Version:   version,
		Name:      name,
		UpSQL:     upSQL,
		DownSQL:   downSQL,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	mm.migrations[version] = migration
	if version > mm.maxVersion {
		mm.maxVersion = version
	}

	return nil
}

// ExecuteMigration 执行迁移
func (mm *MigrationManager) ExecuteMigration(version int) error {
	mm.mu.Lock()
	migration, exists := mm.migrations[version]
	mm.mu.Unlock()

	if !exists {
		return fmt.Errorf("migration version %d not found", version)
	}

	if migration.Status != StatusPending {
		return fmt.Errorf("migration version %d is already %s", version, migration.Status)
	}

	startTime := time.Now()
	migration.Status = StatusRunning

	// 模拟执行 SQL
	// 实际应该连接数据库执行
	for _, sql := range migration.UpSQL {
		if sql == "" {
			continue
		}
		// 执行 SQL
		fmt.Printf("Executing: %s\n", sql)
	}

	duration := time.Since(startTime)
	migration.Status = StatusCompleted
	migration.ExecutedAt = time.Now()
	migration.Duration = duration.Milliseconds()

	mm.mu.Lock()
	mm.history = append(mm.history, migration)
	mm.mu.Unlock()

	return nil
}

// RollbackMigration 回滚迁移
func (mm *MigrationManager) RollbackMigration(version int) error {
	mm.mu.Lock()
	migration, exists := mm.migrations[version]
	mm.mu.Unlock()

	if !exists {
		return fmt.Errorf("migration version %d not found", version)
	}

	if migration.Status != StatusCompleted {
		return fmt.Errorf("migration version %d is not completed", version)
	}

	startTime := time.Now()

	// 执行回滚 SQL
	for _, sql := range migration.DownSQL {
		if sql == "" {
			continue
		}
		fmt.Printf("Rollback: %s\n", sql)
	}

	duration := time.Since(startTime)
	migration.Status = StatusRolledBack
	migration.Duration = duration.Milliseconds()

	return nil
}

// GetMigration 获取迁移
func (mm *MigrationManager) GetMigration(version int) (*Migration, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	migration, exists := mm.migrations[version]
	if !exists {
		return nil, fmt.Errorf("migration version %d not found", version)
	}

	return migration, nil
}

// ListMigrations 列出迁移
func (mm *MigrationManager) ListMigrations() []*Migration {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	migrations := make([]*Migration, 0, len(mm.migrations))
	for _, m := range mm.migrations {
		migrations = append(migrations, m)
	}

	return migrations
}

// GetExecutedMigrations 获取已执行的迁移
func (mm *MigrationManager) GetExecutedMigrations() []*Migration {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return append([]*Migration{}, mm.history...)
}

// GetPendingMigrations 获取待执行迁移
func (mm *MigrationManager) GetPendingMigrations() []*Migration {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var pending []*Migration
	for _, m := range mm.migrations {
		if m.Status == StatusPending {
			pending = append(pending, m)
		}
	}

	return pending
}

// ValidateMigrations 验证迁移
func (mm *MigrationManager) ValidateMigrations() error {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	// 检查迁移版本连续性
	for v := 1; v <= mm.maxVersion; v++ {
		if _, exists := mm.migrations[v]; !exists {
			return fmt.Errorf("missing migration version %d", v)
		}
	}

	// 检查 SQL 不为空
	for v, m := range mm.migrations {
		if len(m.UpSQL) == 0 {
			return fmt.Errorf("migration version %d has no up SQL", v)
		}
		if len(m.DownSQL) == 0 {
			return fmt.Errorf("migration version %d has no down SQL", v)
		}
	}

	return nil
}

// GetCurrentVersion 获取当前版本
func (mm *MigrationManager) GetCurrentVersion() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	maxCompleted := 0
	for _, m := range mm.history {
		if m.Status == StatusCompleted && m.Version > maxCompleted {
			maxCompleted = m.Version
		}
	}

	return maxCompleted
}

// GetTargetVersion 获取目标版本
func (mm *MigrationManager) GetTargetVersion() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return mm.maxVersion
}

// GenerateMigrationScript 生成迁移脚本
func (mm *MigrationManager) GenerateMigrationScript(fromVersion, toVersion int) (string, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if fromVersion >= toVersion {
		return "", fmt.Errorf("fromVersion must be less than toVersion")
	}

	script := fmt.Sprintf("-- Migration from v%d to v%d\n", fromVersion, toVersion)
	script += fmt.Sprintf("-- Generated at %s\n\n", time.Now().Format(time.RFC3339))

	for v := fromVersion + 1; v <= toVersion; v++ {
		if m, exists := mm.migrations[v]; exists {
			script += fmt.Sprintf("-- Version %d: %s\n", v, m.Name)
			for _, sql := range m.UpSQL {
				if sql != "" {
					script += sql + ";\n"
				}
			}
			script += "\n"
		}
	}

	return script, nil
}

// CheckMigrationHealth 检查迁移健康状态
func (mm *MigrationManager) CheckMigrationHealth() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	status := map[string]interface{}{
		"total":     len(mm.migrations),
		"completed": len(mm.history),
		"pending":   0,
		"failed":    0,
	}

	for _, m := range mm.migrations {
		switch m.Status {
		case StatusPending:
			status["pending"] = status["pending"].(int) + 1
		case StatusFailed:
			status["failed"] = status["failed"].(int) + 1
		}
	}

	return status
}


