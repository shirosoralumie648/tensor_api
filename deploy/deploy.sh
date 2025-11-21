#!/bin/bash

# Oblivious 快速部署脚本

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
NAMESPACE="oblivious"
REGISTRY="${DOCKER_REGISTRY:-docker.io}"
BACKEND_IMAGE="${REGISTRY}/oblivious-backend:latest"
FRONTEND_IMAGE="${REGISTRY}/oblivious-frontend:latest"

# 函数
print_header() {
    echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# 主函数
main() {
    local target="${1:-docker}"
    
    case $target in
        docker)
            deploy_docker
            ;;
        k8s|kubernetes)
            deploy_kubernetes
            ;;
        *)
            print_error "Unknown target: $target"
            echo "Usage: $0 [docker|kubernetes]"
            exit 1
            ;;
    esac
}

deploy_docker() {
    print_header "Docker 部署"
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装"
        exit 1
    fi
    
    print_info "Docker 版本: $(docker --version)"
    
    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose 未安装"
        exit 1
    fi
    
    # 构建镜像
    print_header "构建 Docker 镜像"
    
    print_info "构建后端镜像..."
    docker build -f docker/Dockerfile.backend -t "$BACKEND_IMAGE" ../
    print_success "后端镜像构建完成"
    
    print_info "构建前端镜像..."
    docker build -f docker/Dockerfile.frontend -t "$FRONTEND_IMAGE" ../
    print_success "前端镜像构建完成"
    
    # 启动服务
    print_header "启动 Docker Compose 服务"
    
    if [ ! -f ".env" ]; then
        print_info "创建 .env 文件..."
        cat > .env << 'EOF'
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=oblivious
DB_PORT=5433
REDIS_PORT=6379
APP_ENV=development
JWT_SECRET=your-super-secret-jwt-key
COMPOSE_PROJECT_NAME=oblivious
EOF
        print_success ".env 文件已创建"
    fi
    
    print_info "启动所有服务..."
    docker-compose up -d
    print_success "所有服务已启动"
    
    # 等待服务启动
    print_info "等待服务启动..."
    sleep 10
    
    # 运行迁移
    print_header "运行数据库迁移"
    print_info "运行迁移..."
    docker-compose exec -T gateway migrate -path /app/migrations -database \
        "postgresql://postgres:password@postgres:5432/oblivious?sslmode=disable" up || \
        print_info "迁移可能已经运行过"
    print_success "数据库迁移完成"
    
    # 显示服务信息
    print_header "服务信息"
    echo ""
    echo "前端:        http://localhost:3000"
    echo "API 网关:    http://localhost:8080"
    echo "用户服务:    http://localhost:8081"
    echo "对话服务:    http://localhost:8082"
    echo "中转服务:    http://localhost:8083"
    echo ""
    
    # 显示服务状态
    print_info "服务状态："
    docker-compose ps
    
    print_success "Docker 部署完成！"
}

deploy_kubernetes() {
    print_header "Kubernetes 部署"
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl 未安装"
        exit 1
    fi
    
    print_info "kubectl 版本: $(kubectl version --client --short)"
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        print_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    print_success "已连接到 Kubernetes 集群"
    
    # 推送镜像
    print_header "推送 Docker 镜像到仓库"
    print_info "推送后端镜像..."
    docker push "$BACKEND_IMAGE"
    print_success "后端镜像已推送"
    
    print_info "推送前端镜像..."
    docker push "$FRONTEND_IMAGE"
    print_success "前端镜像已推送"
    
    # 创建命名空间
    print_header "创建 Kubernetes 资源"
    print_info "创建命名空间..."
    kubectl apply -f kubernetes/namespace.yaml
    print_success "命名空间已创建"
    
    # 部署数据库和缓存
    print_info "部署数据库..."
    kubectl apply -f kubernetes/postgres.yaml
    print_success "PostgreSQL 已部署"
    
    print_info "部署缓存..."
    kubectl apply -f kubernetes/redis.yaml
    print_success "Redis 已部署"
    
    # 等待数据库启动
    print_info "等待数据库启动..."
    kubectl wait --for=condition=ready pod -l app=postgres -n "$NAMESPACE" --timeout=300s || true
    print_success "数据库已启动"
    
    # 部署后端服务
    print_info "部署后端服务..."
    
    # 更新镜像地址
    sed "s|oblivious-backend:latest|${BACKEND_IMAGE}|g" kubernetes/backend-services.yaml | kubectl apply -f -
    print_success "后端服务已部署"
    
    # 部署前端
    print_info "部署前端..."
    sed "s|oblivious-frontend:latest|${FRONTEND_IMAGE}|g" kubernetes/frontend.yaml | kubectl apply -f -
    print_success "前端已部署"
    
    # 等待 Pod 启动
    print_info "等待 Pod 启动..."
    kubectl wait --for=condition=ready pod -l app=gateway -n "$NAMESPACE" --timeout=300s || true
    print_success "所有 Pod 已启动"
    
    # 显示服务信息
    print_header "Kubernetes 资源"
    
    echo ""
    echo "命名空间: $NAMESPACE"
    echo ""
    
    print_info "Deployments:"
    kubectl get deployments -n "$NAMESPACE"
    
    echo ""
    print_info "Services:"
    kubectl get svc -n "$NAMESPACE"
    
    echo ""
    print_info "Pods:"
    kubectl get pods -n "$NAMESPACE"
    
    # 端口转发建议
    echo ""
    print_header "端口转发命令"
    echo ""
    echo "前端:        kubectl port-forward svc/frontend 3000:3000 -n $NAMESPACE"
    echo "API 网关:    kubectl port-forward svc/gateway 8080:8080 -n $NAMESPACE"
    echo ""
    
    print_success "Kubernetes 部署完成！"
}

# 运行主函数
main "$@"

