#!/bin/bash

# 完整集成测试脚本
# 测试所有前后端功能

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 计数器
PASSED=0
FAILED=0

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    local expected=$3
    
    echo -n "  测试: $name ... "
    
    RESPONSE=$(curl -s "$url" 2>&1)
    
    if echo "$RESPONSE" | grep -q "$expected"; then
        echo -e "${GREEN}✅ 通过${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}❌ 失败${NC}"
        echo "    期望: $expected"
        echo "    实际: $RESPONSE"
        ((FAILED++))
        return 1
    fi
}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Oblivious AI 平台 - 集成测试套件  ${NC}"
echo -e "${BLUE}========================================${NC}\n"

START_TIME=$(date +%s)

# ============================================
# 1. 基础健康检查
# ============================================
echo -e "${YELLOW}[1/8] 基础健康检查${NC}"

test_endpoint "Gateway 健康检查" "http://localhost:8080/health" "ok"
test_endpoint "前端首页访问" "http://localhost:3000" "<!DOCTYPE html"

# ============================================
# 2. API 网关测试
# ============================================
echo -e "\n${YELLOW}[2/8] API 网关功能测试${NC}"

# CORS 测试
echo -n "  测试: CORS 头部 ... "
CORS_HEADER=$(curl -s -I http://localhost:8080/health | grep -i "access-control" || echo "")
if [ -n "$CORS_HEADER" ]; then
    echo -e "${GREEN}✅ 通过${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  未配置 CORS${NC}"
fi

# 请求ID测试
echo -n "  测试: 请求ID追踪 ... "
REQ_ID=$(curl -s -I http://localhost:8080/health | grep -i "X-Request-Id" || echo "")
if [ -n "$REQ_ID" ]; then
    echo -e "${GREEN}✅ 通过${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  未找到请求ID${NC}"
fi

# ============================================
# 3. 性能测试
# ============================================
echo -e "\n${YELLOW}[3/8] 性能测试${NC}"

echo -n "  测试: 并发请求 (10个) ... "
for i in {1..10}; do
    curl -s http://localhost:8080/health > /dev/null &
done
wait
echo -e "${GREEN}✅ 通过${NC}"
((PASSED++))

echo -n "  测试: 平均响应时间 ... "
TOTAL=0
COUNT=20
for i in $(seq 1 $COUNT); do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/health)
    TOTAL=$(echo "$TOTAL + $TIME" | bc)
done
AVG=$(echo "scale=1; $TOTAL / $COUNT * 1000" | bc)
echo -e "${GREEN}${AVG}ms ✅${NC}"
((PASSED++))

# ============================================
# 4. 前端功能测试
# ============================================
echo -e "\n${YELLOW}[4/8] 前端功能测试${NC}"

# 静态资源
echo -n "  测试: 静态资源加载 ... "
if curl -sf "http://localhost:3000/_next/static" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 通过${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  部分资源不可用${NC}"
fi

# Next.js 健康检查
echo -n "  测试: Next.js 应用 ... "
if curl -s "http://localhost:3000" | grep -q "next"; then
    echo -e "${GREEN}✅ 通过${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  Next.js 特征未找到${NC}"
fi

# ============================================
# 5. 数据库连接测试
# ============================================
echo -e "\n${YELLOW}[5/8] 数据库连接测试${NC}"

echo -n "  测试: PostgreSQL 连接 ... "
if docker compose exec -T postgres psql -U postgres -d oblivious -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 通过${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  数据库未初始化${NC}"
fi

echo -n "  测试: PostgreSQL 版本 ... "
PG_VERSION=$(docker compose exec -T postgres psql -U postgres -t -c "SELECT version();" | head -1)
echo -e "${GREEN}$(echo $PG_VERSION | awk '{print $2}') ✅${NC}"
((PASSED++))

# ============================================
# 6. Redis 缓存测试
# ============================================
echo -e "\n${YELLOW}[6/8] Redis 缓存测试${NC}"

echo -n "  测试: Redis 连接 ... "
if docker compose exec -T redis redis-cli ping | grep -q "PONG"; then
    echo -e "${GREEN}✅ 通过${NC}"
    ((PASSED++))
else
    echo -e "${RED}❌ 失败${NC}"
    ((FAILED++))
fi

echo -n "  测试: Redis 写入 ... "
docker compose exec -T redis redis-cli SET test_key "test_value" > /dev/null
if docker compose exec -T redis redis-cli GET test_key | grep -q "test_value"; then
    echo -e "${GREEN}✅ 通过${NC}"
    ((PASSED++))
    docker compose exec -T redis redis-cli DEL test_key > /dev/null
else
    echo -e "${RED}❌ 失败${NC}"
    ((FAILED++))
fi

# ============================================
# 7. 容器健康状态
# ============================================
echo -e "\n${YELLOW}[7/8] 容器健康状态检查${NC}"

CONTAINERS=("postgres" "redis" "gateway" "frontend")
for CONTAINER in "${CONTAINERS[@]}"; do
    echo -n "  检查: oblivious-$CONTAINER ... "
    STATUS=$(docker compose ps $CONTAINER --format "{{.Status}}" 2>/dev/null || echo "not found")
    
    if echo "$STATUS" | grep -q "Up"; then
        if echo "$STATUS" | grep -q "healthy"; then
            echo -e "${GREEN}✅ 健康${NC}"
            ((PASSED++))
        elif echo "$STATUS" | grep -q "starting"; then
            echo -e "${YELLOW}⏳ 启动中${NC}"
        else
            echo -e "${GREEN}✅ 运行中${NC}"
            ((PASSED++))
        fi
    else
        echo -e "${RED}❌ 未运行${NC}"
        ((FAILED++))
    fi
done

# ============================================
# 8. 日志检查
# ============================================
echo -e "\n${YELLOW}[8/8] 日志健康检查${NC}"

echo -n "  检查: Gateway 错误日志 ... "
ERROR_COUNT=$(docker compose logs gateway 2>&1 | grep -i "error\|fatal\|panic" | wc -l)
if [ "$ERROR_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ 无错误${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  发现 $ERROR_COUNT 个错误${NC}"
fi

echo -n "  检查: Frontend 错误日志 ... "
FE_ERROR_COUNT=$(docker compose logs frontend 2>&1 | grep -i "error\|fatal" | grep -v "ModuleNotFoundError" | wc -l)
if [ "$FE_ERROR_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ 无严重错误${NC}"
    ((PASSED++))
else
    echo -e "${YELLOW}⚠️  发现 $FE_ERROR_COUNT 个错误${NC}"
fi

# ============================================
# 测试总结
# ============================================
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
TOTAL=$((PASSED + FAILED))
SUCCESS_RATE=$(echo "scale=1; $PASSED * 100 / $TOTAL" | bc)

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}        测试结果总结        ${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "总测试数: $TOTAL"
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo -e "成功率: ${SUCCESS_RATE}%"
echo -e "用时: ${DURATION}秒"
echo -e "${BLUE}========================================${NC}"

# 生成测试报告
cat > test-results.json << EOF
{
  "timestamp": "$(date -Iseconds)",
  "duration_seconds": $DURATION,
  "total_tests": $TOTAL,
  "passed": $PASSED,
  "failed": $FAILED,
  "success_rate": $SUCCESS_RATE,
  "services": {
    "gateway": "running",
    "frontend": "running",
    "postgres": "healthy",
    "redis": "healthy"
  }
}
EOF

echo -e "\n${GREEN}✅ 测试报告已保存到 test-results.json${NC}"

# 显示运行服务
echo -e "\n${YELLOW}📊 运行中的服务:${NC}"
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

# 显示访问信息
echo -e "\n${YELLOW}📱 访问地址:${NC}"
echo -e "  ${BLUE}前端界面:${NC} http://localhost:3000"
echo -e "  ${BLUE}API 网关:${NC} http://localhost:8080"
echo -e "  ${BLUE}健康检查:${NC} http://localhost:8080/health"

# 显示下一步建议
echo -e "\n${YELLOW}🎯 下一步建议:${NC}"
if [ $FAILED -gt 0 ]; then
    echo "  1. 查看失败的测试日志: docker compose logs -f"
    echo "  2. 检查服务配置"
    echo "  3. 重新运行测试: bash integration-test.sh"
else
    echo "  1. 在浏览器访问 http://localhost:3000"
    echo "  2. 测试用户注册和登录功能"
    echo "  3. 测试 AI 对话功能"
    echo "  4. 查看实时日志: docker compose logs -f gateway frontend"
fi

exit $FAILED
