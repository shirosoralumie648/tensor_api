#!/bin/bash

# 快速功能测试脚本

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Oblivious AI 平台快速测试 ===${NC}\n"

# 1. 健康检查
echo "1. 健康检查测试"
echo -n "  Gateway (8080): "
if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 正常${NC}"
else
    echo -e "${RED}❌ 失败${NC}"
fi

echo -n "  Relay (8083): "
if curl -s -f http://localhost:8083/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 正常${NC}"
else
    echo -e "${YELLOW}⚠️  不可用 (这是预期的，relay 可能未编译)${NC}"
fi

echo -n "  PostgreSQL (5433): "
if docker compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 正常${NC}"
else
    echo -e "${RED}❌ 失败${NC}"
fi

echo -n "  Redis (6379): "
if docker compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 正常${NC}"
else
    echo -e "${RED}❌ 失败${NC}"
fi

# 2. API 响应测试
echo -e "\n2. API 响应测试"
echo -n "  GET /health: "
RESPONSE=$(curl -s http://localhost:8080/health)
if echo "$RESPONSE" | grep -q "ok"; then
    echo -e "${GREEN}✅ $RESPONSE${NC}"
else
    echo -e "${RED}❌ 响应异常${NC}"
fi

# 3. 响应时间测试
echo -e "\n3. 性能测试"
echo -n "  平均响应时间: "
TOTAL=0
for i in {1..10}; do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/health)
    TOTAL=$(echo "$TOTAL + $TIME" | bc)
done
AVG=$(echo "scale=3; $TOTAL / 10 * 1000" | bc)
echo -e "${GREEN}${AVG}ms${NC}"

# 4. 容器状态
echo -e "\n4. 容器状态"
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

# 5. 日志检查 (最近10行)
echo -e "\n5. Gateway 最近日志"
docker compose logs gateway --tail=5 2>&1 | grep -E "INFO|ERROR|WARN" | tail -5 || echo "无日志"

echo -e "\n${GREEN}=== 测试完成 ===${NC}"
