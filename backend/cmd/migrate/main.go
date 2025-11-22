package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/shirosoralumie648/Oblivious/backend/internal/config"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载环境变量（从项目根目录）
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v\n", err)
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// 解析命令行参数
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		if err := migrateUp(db); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("✅ Migration up completed successfully!")

	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("✅ Migration down completed successfully!")

	case "status":
		if err := migrationStatus(db); err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		}

	case "sync":
		// 同步 channel_abilities 数据
		if err := syncChannelAbilities(db); err != nil {
			log.Fatalf("Failed to sync channel abilities: %v", err)
		}
		log.Println("✅ Channel abilities synced successfully!")

	default:
		log.Printf("Unknown command: %s\n", command)
		log.Println("Usage: migrate [up|down|status|sync]")
		os.Exit(1)
	}
}

// migrateUp 执行向上迁移
func migrateUp(db *gorm.DB) error {
	ctx := context.Background()
	migrations, err := loadMigrations("../../migrations", ".up.sql")
	if err != nil {
		return err
	}

	log.Printf("Found %d up migrations\n", len(migrations))

	for _, mig := range migrations {
		log.Printf("Executing migration: %s\n", mig.Name)

		// 读取 SQL 文件
		content, err := os.ReadFile(mig.Path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", mig.Path, err)
		}

		// 执行 SQL
		if err := db.WithContext(ctx).Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute %s: %w", mig.Name, err)
		}

		log.Printf("✅ Completed: %s\n", mig.Name)
	}

	return nil
}

// migrateDown 执行向下迁移
func migrateDown(db *gorm.DB) error {
	ctx := context.Background()
	migrations, err := loadMigrations("../../migrations", ".down.sql")
	if err != nil {
		return err
	}

	log.Printf("Found %d down migrations\n", len(migrations))

	// 反向执行
	for i := len(migrations) - 1; i >= 0; i-- {
		mig := migrations[i]
		log.Printf("Executing rollback: %s\n", mig.Name)

		content, err := os.ReadFile(mig.Path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", mig.Path, err)
		}

		if err := db.WithContext(ctx).Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute %s: %w", mig.Name, err)
		}

		log.Printf("✅ Rolled back: %s\n", mig.Name)
	}

	return nil
}

// migrationStatus 检查迁移状态
func migrationStatus(db *gorm.DB) error {
	log.Println("Checking migration status...")

	// 检查表是否存在
	tables := []string{
		"channels",
		"adapter_configs",
		"channel_abilities",
		"unified_logs",
		"model_pricing",
	}

	for _, table := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = ?
		)`
		if err := db.Raw(query, table).Scan(&exists).Error; err != nil {
			return err
		}

		status := "❌ Not exists"
		if exists {
			status = "✅ Exists"
		}
		log.Printf("%s: %s\n", table, status)
	}

	// 检查 channels 表的新字段
	log.Println("\nChecking channels table columns...")
	newColumns := []string{
		"priority", "group", "tag", "channel_info",
		"status", "response_time", "used_quota", "balance",
	}

	for _, col := range newColumns {
		var exists bool
		query := `SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'channels' AND column_name = ?
		)`
		if err := db.Raw(query, col).Scan(&exists).Error; err != nil {
			return err
		}

		status := "❌ Missing"
		if exists {
			status = "✅ Exists"
		}
		log.Printf("  %s: %s\n", col, status)
	}

	return nil
}

// syncChannelAbilities 同步渠道能力数据
func syncChannelAbilities(db *gorm.DB) error {
	ctx := context.Background()

	// 查询所有渠道
	var channels []model.Channel
	if err := db.WithContext(ctx).Where("deleted_at IS NULL").Find(&channels).Error; err != nil {
		return fmt.Errorf("failed to query channels: %w", err)
	}

	log.Printf("Found %d channels to process\n", len(channels))

	for _, ch := range channels {
		if ch.SupportModels == "" {
			continue
		}

		// 解析支持的模型
		models := strings.Split(ch.SupportModels, ",")

		for _, modelName := range models {
			modelName = strings.TrimSpace(modelName)
			if modelName == "" {
				continue
			}

			// 创建或更新 ability
			ability := &model.ChannelAbility{
				ChannelID: ch.ID,
				Model:     modelName,
				Group:     ch.Group,
				Enabled:   ch.IsEnabled(),
				Priority:  ch.Priority,
				Weight:    ch.Weight,
			}

			// 使用 FirstOrCreate 避免重复
			result := db.WithContext(ctx).Where(model.ChannelAbility{
				ChannelID: ch.ID,
				Model:     modelName,
				Group:     ch.Group,
			}).Assign(model.ChannelAbility{
				Enabled:  ch.IsEnabled(),
				Priority: ch.Priority,
				Weight:   ch.Weight,
			}).FirstOrCreate(ability)

			if result.Error != nil {
				log.Printf("Warning: failed to sync ability for channel %d, model %s: %v\n",
					ch.ID, modelName, result.Error)
			}
		}

		log.Printf("✅ Synced channel: %s (%d models)\n", ch.Name, len(models))
	}

	return nil
}

// Migration 迁移信息
type Migration struct {
	Name string
	Path string
}

// loadMigrations 加载迁移文件
func loadMigrations(dir string, suffix string) ([]Migration, error) {
	var migrations []Migration

	// 获取绝对路径
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// 读取目录
	files, err := os.ReadDir(absDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", absDir, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, suffix) {
			continue
		}

		migrations = append(migrations, Migration{
			Name: name,
			Path: filepath.Join(absDir, name),
		})
	}

	// 按文件名排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}
