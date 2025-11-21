#!/bin/bash

# 手动测试脚本 - 用于验证 Oblivious AI Platform 部署
# 需要先确保 Docker 服务正在运行

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Oblivious 手动测试工具  ${NC}"
echo -e "${BLUE}================================${NC}\n"

# 测试计数
PASS=0
FAIL=0
WARN=0

test_url() {
    local name=$1
    local url=$2
    local expected=$3
    
    echo -n "测试 $name ... "
    response=$(curl -s "$url" 2>&1)
    
    if echo "$response" | grep -q "$expected"; then
        echo -e "${GREEN}✅ 通过${NC}"
        PASS=$((PASS + 1))
        return 0
    elif curl -sf "$url" > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  服务运行但响应异常${NC}"
        WARN=$((WARN + 1))
        return 1
    else
        echo -e "${RED}❌ 失败${NC}"
        FAIL=$((FAIL + 1))
        return 1
    fi
}

# ==================== 1. 容器状态检查 ====================
echo -e "\n${YELLOW}[1/5] 容器状态检查${NC}\n"

if command -v docker &> /dev/null; then
    echo -e "${GREEN}✓${NC} Docker 已安装"
else
    echo -e "${RED}✗${NC} Docker 未安装"
    exit 1
fi

if docker ps > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Docker 服务正在运行\n"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep oblivious || echo "未找到 Oblivious 容器"
else
    echo -e "${RED}✗${NC} Docker 服务未运行，请先启动 Docker"
    exit 1
fi

# ==================== 2. 基础健康检查 ====================
echo -e "\n${YELLOW}[2/5] 基础健康检查${NC}\n"

test_url "API 网关" "http://localhost:8080/health" "ok\|healthy\|UP"
test_url "用户服务" "http://localhost:8081/health" "ok\|healthy\|UP"
test_url "对话服务" "http://localhost:8082/health" "ok\|healthy\|UP"
test_url "中转服务" "http://localhost:8083/health" "ok\|healthy\|UP"
test_url "助手服务" "http://localhost:8084/health" "ok\|healthy\|UP"
test_url "知识库服务" "http://localhost:8085/health" "ok\|healthy\|UP"
test_url "前端服务" "http://localhost:3000" ".*"

# ==================== 3. 功能测试 ====================
echo -e "\n${YELLOW}[3/5] 功能测试${NC}\n"

# 用户注册
echo -n "测试用户注册 ... "
TIMESTAMP=$(date +%s)
REGISTER_DATA="{\"username\":\"test$TIMESTAMP\",\"email\":\"test$TIMESTAMP@example.com\",\"password\":\"Test123456\"}"
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8081/api/users/register \
    -H "Content-Type: application/json" \
    -d "$REGISTER_DATA" 2>&1)

if echo "$REGISTER_RESPONSE" | grep -q "success\|token\|id"; then
    echo -e "${GREEN}✅ 通过${NC}"
    PASS=$((PASS + 1))
elif echo "$REGISTER_RESPONSE" | grep -q "exists\|duplicate"; then
    echo -e "${YELLOW}⚠️  用户已存在${NC}"
    WARN=$((WARN + 1))
else
    echo -e "${RED}❌ 失败 - $REGISTER_RESPONSE${NC}"
    FAIL=$((FAIL + 1))
fi

# 创建对话
echo -n "测试创建对话 ... "
CHAT_RESPONSE=$(curl -s -X POST http://localhost:8082/api/chats \
    -H "Content-Type: application/json" \
    -d '{"title":"测试对话"}' 2>&1)

if echo "$CHAT_RESPONSE" | grep -q "id\|chat_id\|success"; then
    echo -e "${GREEN}✅ 通过${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${YELLOW}⚠️  响应: $CHAT_RESPONSE${NC}"
    WARN=$((WARN + 1))
fi

# 获取模型列表
echo -n "测试模型列表 ... "
MODELS_RESPONSE=$(curl -s http://localhost:8083/v1/models 2>&1)

if echo "$MODELS_RESPONSE" | grep -q "data\|models\|id"; then
    echo -e "${GREEN}✅ 通过${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${YELLOW}⚠️  响应: $MODELS_RESPONSE${NC}"
    WARN=$((WARN + 1))
fi

# ==================== 4. 性能测试 ====================
echo -e "\n${YELLOW}[4/5] 性能测试${NC}\n"

echo -n "测试 API 响应时间 ... "
TOTAL_TIME=0
COUNT=5

for i in $(seq 1 $COUNT); do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/health 2>&1)
    if [[ $TIME =~ ^[0-9.]+$ ]]; then
        TOTAL_TIME=$(echo "$TOTAL_TIME + $TIME" | bc)
    fi
done

AVG_TIME=$(echo "scale=3; $TOTAL_TIME / $COUNT * 1000" | bc)
if (( $(echo "$AVG_TIME < 100" | bc -l) )); then
    echo -e "${GREEN}✅ 优秀 (${AVG_TIME}ms)${NC}"
    PASS=$((PASS + 1))
elif (( $(echo "$AVG_TIME < 500" | bc -l) )); then
    echo -e "${GREEN}✅ 良好 (${AVG_TIME}ms)${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${YELLOW}⚠️  较慢 (${AVG_TIME}ms)${NC}"
    WARN=$((WARN + 1))
fi

# ==================== 5. 资源使用 ====================
echo -e "\n${YELLOW}[5/5] 资源使用情况${NC}\n"

docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep oblivious

# ==================== 总结 ====================
TOTAL=$((PASS + FAIL + WARN))

echo -e "\n${BLUE}================================${NC}"
echo -e "${BLUE}  测试结果总结  ${NC}"
echo -e "${BLUE}================================${NC}"
echo -e "总测试数: $TOTAL"
echo -e "${GREEN}通过: $PASS${NC}"
echo -e "${RED}失败: $FAIL${NC}"
echo -e "${YELLOW}警告: $WARN${NC}"

if [ $FAIL -eq 0 ]; then
    echo -e "\n${GREEN}🎉 所有关键测试通过！${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  部分测试失败，请检查服务状态${NC}"
    exit 1
fi
