.PHONY: help build run test clean migrate-up migrate-down docker-build

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## 构建所有服务
	@echo "Building all services..."
	@for service in gateway user chat relay agent rag file plugin billing worker; do \
		echo "Building $$service..."; \
		cd cmd/$$service && go build -o ../../bin/$$service main.go && cd ../..; \
	done

build-gateway: ## 构建网关服务
	cd cmd/gateway && go build -o ../../bin/gateway main.go

run-gateway: ## 运行网关服务
	cd cmd/gateway && go run main.go

test: ## 运行测试
	go test -v -cover ./...

test-coverage: ## 生成测试覆盖率报告
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## 清理构建产物
	rm -rf bin/
	rm -f coverage.out coverage.html

# 数据库配置
DATABASE_URL ?= postgres://postgres:password@localhost:5433/oblivious?sslmode=disable

migrate-up: ## 执行数据库迁移（升级）
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## 回滚数据库迁移
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-force: ## 强制设置迁移版本 (使用: make migrate-force VERSION=1)
	migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

migrate-version: ## 查看当前迁移版本
	migrate -path migrations -database "$(DATABASE_URL)" version

migrate-create: ## 创建新的迁移文件 (使用: make migrate-create NAME=create_users_table)
	migrate create -ext sql -dir migrations -seq $(NAME)

docker-build: ## 构建所有 Docker 镜像
	@echo "Building Docker images..."
	@for service in gateway user chat relay agent rag file plugin billing worker; do \
		echo "Building $$service image..."; \
		docker build -f ../deploy/docker/$$service.Dockerfile -t oblivious-$$service:latest .; \
	done

docker-push: ## 推送 Docker 镜像到仓库
	@echo "Pushing Docker images..."
	@for service in gateway user chat relay agent rag file plugin billing worker; do \
		docker push oblivious-$$service:latest; \
	done

lint: ## 运行代码检查
	golangci-lint run ./...

fmt: ## 格式化代码
	go fmt ./...
	gofmt -s -w .

mod-tidy: ## 整理依赖
	go mod tidy

dev: ## 启动开发环境（所有服务）
	@echo "Starting development environment..."
	@docker-compose -f ../deploy/docker-compose.yml up -d

.DEFAULT_GOAL := help

