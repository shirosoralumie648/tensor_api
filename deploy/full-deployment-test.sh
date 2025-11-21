#!/bin/bash

# 完整部署和测试脚本
# 部署所有前后端服务并执行全面测试

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Oblivious AI 平台 - 完整部署测试  ${NC}"
echo -e "${BLUE}========================================${NC}\n"

# 记录开始时间
START_TIME=$(date +%s)

# 1. 构建前端镜像
echo -e "\n${YELLOW}[1/6] 构建前端 Docker 镜像...${NC}"
if docker build -f docker/Dockerfile.frontend -t oblivious-frontend:latest .. 2>&1 | grep -q "ERROR"; then
    echo -e "${RED}❌ 前端镜像构建失败${NC}"
    echo -e "${YELLOW}继续后端部署...${NC}"
else
    echo -e "${GREEN}✅ 前端镜像构建成功${NC}"
fi

# 2. 构建后端镜像（已完成）
echo -e "\n${YELLOW}[2/6] 验证后端 Docker 镜像...${NC}"
if docker images oblivious-backend:latest | grep -q "oblivious-backend"; then
    echo -e "${GREEN}✅ 后端镜像已存在${NC}"
else
    echo -e "${YELLOW}构建后端镜像...${NC}"
    docker build -f docker/Dockerfile.backend -t oblivious-backend:latest ..
fi

# 3. 停止旧容器
echo -e "\n${YELLOW}[3/6] 停止旧容器...${NC}"
docker compose down
echo -e "${GREEN}✅ 旧容器已停止${NC}"

# 4. 启动所有服务
echo -e "\n${YELLOW}[4/6] 启动所有服务...${NC}"
echo "  - PostgreSQL (数据库)"
echo "  - Redis (缓存)"
echo "  - Gateway (API 网关)"
echo "  - Frontend (前端界面)"

docker compose up -d postgres redis gateway

# 尝试启动前端
if docker images oblivious-frontend:latest | grep -q "oblivious-frontend"; then
    docker compose up -d frontend
    echo -e "${GREEN}✅ 前端服务已启动${NC}"
else
    echo -e "${YELLOW}⚠️  前端镜像不存在，跳过前端启动${NC}"
fi

# 等待服务启动
echo -e "\n${YELLOW}等待服务启动 (15秒)...${NC}"
sleep 15

# 5. 服务健康检查
echo -e "\n${YELLOW}[5/6] 服务健康检查...${NC}"

# 后端服务检查
echo -e "\n${BLUE}后端服务:${NC}"
SERVICES=("Gateway:8080" "Relay:8083")
for SERVICE in "${SERVICES[@]}"; do
    NAME="${SERVICE%:*}"
    PORT="${SERVICE#*:}"
    echo -n "  $NAME ($PORT): "
    if curl -sf "http://localhost:$PORT/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 运行中${NC}"
    else
        echo -e "${RED}❌ 不可用${NC}"
    fi
done

# 前端服务检查
echo -e "\n${BLUE}前端服务:${NC}"
echo -n "  Frontend (3000): "
if curl -sf "http://localhost:3000" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 运行中${NC}"
elif docker compose ps frontend | grep -q "Up"; then
    echo -e "${YELLOW}⏳ 启动中...${NC}"
else
    echo -e "${RED}❌ 未启动${NC}"
fi

# 数据层检查
echo -e "\n${BLUE}数据层:${NC}"
echo -n "  PostgreSQL (5433): "
if docker compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 正常${NC}"
else
    echo -e "${RED}❌ 异常${NC}"
fi

echo -n "  Redis (6379): "
if docker compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 正常${NC}"
else
    echo -e "${RED}❌ 异常${NC}"
fi

# 6. 功能测试
echo -e "\n${YELLOW}[6/6] 执行功能测试...${NC}"

# 后端API测试
echo -e "\n${BLUE}后端 API 测试:${NC}"

# 健康检查
echo -n "  GET /health: "
RESPONSE=$(curl -s http://localhost:8080/health)
if echo "$RESPONSE" | grep -q "ok"; then
    echo -e "${GREEN}✅ 通过${NC}"
else
    echo -e "${RED}❌ 失败${NC}"
fi

# 性能测试
echo -n "  响应时间: "
TOTAL=0
for i in {1..5}; do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/health)
    TOTAL=$(echo "$TOTAL + $TIME" | bc)
done
AVG=$(echo "scale=1; $TOTAL / 5 * 1000" | bc)
echo -e "${GREEN}${AVG}ms${NC}"

# 前端测试
echo -e "\n${BLUE}前端测试:${NC}"
echo -n "  访问首页: "
if curl -sf "http://localhost:3000" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 可访问${NC}"
else
    echo -e "${RED}❌ 不可访问${NC}"
fi

# 容器状态
echo -e "\n${BLUE}容器状态:${NC}"
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" | grep -v "NAME"

# 计算总用时
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo -e "\n${BLUE}========================================${NC}"
echo -e "${GREEN}✅ 部署测试完成！${NC}"
echo -e "总用时: ${DURATION} 秒"
echo -e "${BLUE}========================================${NC}"

# 访问信息
echo -e "\n${YELLOW}📱 访问地址:${NC}"
echo "  前端: http://localhost:3000"
echo "  API:  http://localhost:8080"
echo "  健康检查: http://localhost:8080/health"

echo -e "\n${YELLOW}📋 查看日志:${NC}"
echo "  docker compose logs -f gateway"
echo "  docker compose logs -f frontend"

echo -e "\n${YELLOW}🔧 管理命令:${NC}"
echo "  停止服务: docker compose down"
echo "  重启服务: docker compose restart"
echo "  查看状态: docker compose ps"
