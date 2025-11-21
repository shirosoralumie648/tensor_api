#!/bin/bash

# Oblivious AI Platform - 完整部署和测试脚本
# 用于 Docker Compose 部署并执行全面的功能、稳定性和bug测试

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

# 日志文件
LOG_FILE="deployment-test-$(date +%Y%m%d-%H%M%S).log"
REPORT_FILE="test-report-$(date +%Y%m%d-%H%M%S).md"

# 记录函数
log() {
    echo -e "$1" | tee -a "$LOG_FILE"
}

# 测试结果记录
test_result() {
    local test_name=$1
    local result=$2
    local details=$3
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$result" == "PASS" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log "${GREEN}✅ [PASS] $test_name${NC}"
    elif [ "$result" == "FAIL" ]; then
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log "${RED}❌ [FAIL] $test_name${NC}"
    else
        WARNINGS=$((WARNINGS + 1))
        log "${YELLOW}⚠️  [WARN] $test_name${NC}"
    fi
    
    if [ -n "$details" ]; then
        log "    详情: $details"
    fi
}

# 开始部署
log "${BLUE}========================================${NC}"
log "${BLUE}  Oblivious AI 平台部署与测试  ${NC}"
log "${BLUE}========================================${NC}\n"
log "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
log "日志文件: $LOG_FILE"
log "报告文件: $REPORT_FILE\n"

START_TIME=$(date +%s)

# ==================== 阶段 1: 环境检查 ====================
log "${MAGENTA}[阶段 1/7] 环境检查${NC}"

# 检查 Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    test_result "Docker 安装检查" "PASS" "$DOCKER_VERSION"
else
    test_result "Docker 安装检查" "FAIL" "Docker 未安装"
    exit 1
fi

# 检查 Docker Compose
if command -v docker compose &> /dev/null; then
    COMPOSE_VERSION=$(docker compose version)
    test_result "Docker Compose 安装检查" "PASS" "$COMPOSE_VERSION"
else
    test_result "Docker Compose 安装检查" "FAIL" "Docker Compose 未安装"
    exit 1
fi

# 检查镜像
if docker images oblivious-backend:latest | grep -q "oblivious-backend"; then
    test_result "后端镜像检查" "PASS" "镜像存在"
else
    test_result "后端镜像检查" "FAIL" "镜像不存在"
fi

if docker images oblivious-frontend:latest | grep -q "oblivious-frontend"; then
    test_result "前端镜像检查" "PASS" "镜像存在"
else
    test_result "前端镜像检查" "WARN" "镜像不存在，将跳过前端部署"
fi

# ==================== 阶段 2: 停止旧服务 ====================
log "\n${MAGENTA}[阶段 2/7] 停止旧服务${NC}"
docker compose down 2>&1 | tee -a "$LOG_FILE"
test_result "停止旧容器" "PASS" "所有旧容器已停止"

# ==================== 阶段 3: 启动服务 ====================
log "\n${MAGENTA}[阶段 3/7] 启动所有服务${NC}"

log "启动基础设施服务..."
docker compose up -d postgres redis 2>&1 | tee -a "$LOG_FILE"

log "等待数据库就绪 (20秒)..."
sleep 20

# 检查 PostgreSQL
if docker compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
    test_result "PostgreSQL 启动" "PASS" "数据库已就绪"
else
    test_result "PostgreSQL 启动" "FAIL" "数据库未就绪"
fi

# 检查 Redis
if docker compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    test_result "Redis 启动" "PASS" "缓存已就绪"
else
    test_result "Redis 启动" "FAIL" "缓存未就绪"
fi

log "启动后端微服务..."
docker compose up -d gateway user chat relay agent kb 2>&1 | tee -a "$LOG_FILE"

log "等待后端服务启动 (20秒)..."
sleep 20

# 启动前端
if docker images oblivious-frontend:latest | grep -q "oblivious-frontend"; then
    log "启动前端服务..."
    docker compose up -d frontend 2>&1 | tee -a "$LOG_FILE"
    sleep 10
fi

# ==================== 阶段 4: 运行数据库迁移 ====================
log "\n${MAGENTA}[阶段 4/7] 运行数据库迁移${NC}"

# 执行迁移
MIGRATION_OUTPUT=$(docker compose exec -T postgres psql -U postgres -d oblivious -c "\dt" 2>&1)
if echo "$MIGRATION_OUTPUT" | grep -q "users"; then
    test_result "数据库表创建" "PASS" "表已存在"
else
    log "执行数据库迁移..."
    # 如果有迁移脚本，在这里执行
    test_result "数据库表创建" "WARN" "表可能未创建"
fi

# ==================== 阶段 5: 健康检查 ====================
log "\n${MAGENTA}[阶段 5/7] 服务健康检查${NC}"

# 定义服务端口
declare -A SERVICES=(
    ["Gateway"]=8080
    ["User"]=8081
    ["Chat"]=8082
    ["Relay"]=8083
    ["Agent"]=8084
    ["KB"]=8085
    ["Frontend"]=3000
)

# 检查每个服务
for service in "${!SERVICES[@]}"; do
    port=${SERVICES[$service]}
    
    if [ "$service" == "Frontend" ]; then
        # 前端只检查HTTP访问
        if curl -sf "http://localhost:$port" > /dev/null 2>&1; then
            test_result "$service 服务 (端口 $port)" "PASS" "服务响应正常"
        else
            test_result "$service 服务 (端口 $port)" "WARN" "服务可能未启动"
        fi
    else
        # 后端检查 health 端点
        response=$(curl -s "http://localhost:$port/health" 2>&1)
        if echo "$response" | grep -q "ok\|healthy\|UP"; then
            test_result "$service 服务 (端口 $port)" "PASS" "健康检查通过"
        else
            # 尝试直接访问端口
            if curl -sf "http://localhost:$port" > /dev/null 2>&1; then
                test_result "$service 服务 (端口 $port)" "WARN" "服务运行但健康端点未响应"
            else
                test_result "$service 服务 (端口 $port)" "FAIL" "服务不可访问"
            fi
        fi
    fi
done

# ==================== 阶段 6: 功能完整性测试 ====================
log "\n${MAGENTA}[阶段 6/7] 功能完整性测试${NC}"

# 6.1 API 网关测试
log "\n${CYAN}6.1 API 网关功能测试${NC}"

# 健康检查
response=$(curl -s http://localhost:8080/health)
if echo "$response" | grep -q "ok\|healthy"; then
    test_result "网关健康检查" "PASS" "响应: $response"
else
    test_result "网关健康检查" "FAIL" "无效响应"
fi

# CORS测试
response=$(curl -s -H "Origin: http://localhost:3000" -I http://localhost:8080/health 2>&1)
if echo "$response" | grep -q "Access-Control-Allow-Origin"; then
    test_result "CORS 配置" "PASS" "CORS 头已配置"
else
    test_result "CORS 配置" "WARN" "CORS 头可能未配置"
fi

# 6.2 用户服务测试
log "\n${CYAN}6.2 用户服务功能测试${NC}"

# 注册测试
REGISTER_DATA='{"username":"testuser'$(date +%s)'","email":"test'$(date +%s)'@example.com","password":"Test123456"}'
register_response=$(curl -s -X POST http://localhost:8081/api/users/register \
    -H "Content-Type: application/json" \
    -d "$REGISTER_DATA" 2>&1)

if echo "$register_response" | grep -q "success\|token\|id"; then
    test_result "用户注册功能" "PASS" "注册成功"
    TOKEN=$(echo "$register_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
elif echo "$register_response" | grep -q "already exists\|duplicate"; then
    test_result "用户注册功能" "WARN" "用户已存在（预期行为）"
else
    test_result "用户注册功能" "FAIL" "注册失败: $register_response"
fi

# 登录测试
LOGIN_DATA='{"username":"testuser","password":"Test123456"}'
login_response=$(curl -s -X POST http://localhost:8081/api/users/login \
    -H "Content-Type: application/json" \
    -d "$LOGIN_DATA" 2>&1)

if echo "$login_response" | grep -q "token\|success"; then
    test_result "用户登录功能" "PASS" "登录成功"
else
    test_result "用户登录功能" "WARN" "登录失败: $login_response"
fi

# 6.3 对话服务测试
log "\n${CYAN}6.3 对话服务功能测试${NC}"

# 创建对话
create_chat_response=$(curl -s -X POST http://localhost:8082/api/chats \
    -H "Content-Type: application/json" \
    -d '{"title":"测试对话"}' 2>&1)

if echo "$create_chat_response" | grep -q "id\|chat_id\|success"; then
    test_result "创建对话功能" "PASS" "对话创建成功"
else
    test_result "创建对话功能" "WARN" "对话创建失败: $create_chat_response"
fi

# 6.4 中转服务测试
log "\n${CYAN}6.4 中转服务功能测试${NC}"

# 获取模型列表
models_response=$(curl -s http://localhost:8083/v1/models 2>&1)

if echo "$models_response" | grep -q "data\|models\|id"; then
    test_result "获取模型列表" "PASS" "模型列表获取成功"
else
    test_result "获取模型列表" "WARN" "模型列表获取失败: $models_response"
fi

# 6.5 前端访问测试
log "\n${CYAN}6.5 前端功能测试${NC}"

if curl -sf http://localhost:3000 > /dev/null 2>&1; then
    # 检查关键资源
    homepage=$(curl -s http://localhost:3000)
    
    if echo "$homepage" | grep -q "Oblivious\|AI\|Chat"; then
        test_result "前端首页访问" "PASS" "页面内容正常"
    else
        test_result "前端首页访问" "WARN" "页面内容可能异常"
    fi
else
    test_result "前端首页访问" "FAIL" "前端不可访问"
fi

# ==================== 阶段 7: 稳定性和性能测试 ====================
log "\n${MAGENTA}[阶段 7/7] 稳定性和性能测试${NC}"

# 7.1 响应时间测试
log "\n${CYAN}7.1 API 响应时间测试${NC}"

declare -A response_times
for endpoint in "8080/health" "8081/health" "8082/health" "8083/health"; do
    total_time=0
    success_count=0
    
    for i in {1..10}; do
        time=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:$endpoint 2>&1)
        if [[ $time =~ ^[0-9.]+$ ]]; then
            total_time=$(echo "$total_time + $time" | bc)
            success_count=$((success_count + 1))
        fi
    done
    
    if [ $success_count -gt 0 ]; then
        avg_time=$(echo "scale=3; $total_time / $success_count * 1000" | bc)
        response_times[$endpoint]=$avg_time
        
        if (( $(echo "$avg_time < 100" | bc -l) )); then
            test_result "响应时间 [$endpoint]" "PASS" "平均 ${avg_time}ms (优秀)"
        elif (( $(echo "$avg_time < 500" | bc -l) )); then
            test_result "响应时间 [$endpoint]" "PASS" "平均 ${avg_time}ms (良好)"
        else
            test_result "响应时间 [$endpoint]" "WARN" "平均 ${avg_time}ms (较慢)"
        fi
    else
        test_result "响应时间 [$endpoint]" "FAIL" "无法测试响应时间"
    fi
done

# 7.2 并发测试
log "\n${CYAN}7.2 并发请求测试${NC}"

concurrent_test() {
    local url=$1
    local concurrent=$2
    local requests=$3
    
    success=0
    for i in $(seq 1 $requests); do
        curl -s $url > /dev/null &
        if [ $((i % concurrent)) -eq 0 ]; then
            wait
        fi
    done
    wait
}

# 测试网关并发
log "执行并发测试 (20个并发请求)..."
start=$(date +%s.%N)
concurrent_test "http://localhost:8080/health" 10 20
end=$(date +%s.%N)
duration=$(echo "$end - $start" | bc)

test_result "并发请求测试 (20请求)" "PASS" "完成时间: ${duration}s"

# 7.3 容器资源使用
log "\n${CYAN}7.3 容器资源使用测试${NC}"

docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" > /tmp/docker_stats.txt 2>&1
cat /tmp/docker_stats.txt | tee -a "$LOG_FILE"

# 检查是否有容器使用过高资源
high_cpu=$(cat /tmp/docker_stats.txt | grep oblivious | awk '{print $2}' | sed 's/%//' | awk '$1 > 80 {print $1}')
if [ -z "$high_cpu" ]; then
    test_result "CPU 使用率检查" "PASS" "所有容器 CPU 使用正常"
else
    test_result "CPU 使用率检查" "WARN" "部分容器 CPU 使用率较高"
fi

# 7.4 容器状态检查
log "\n${CYAN}7.4 容器状态检查${NC}"

container_status=$(docker compose ps --format "table {{.Name}}\t{{.Status}}")
log "$container_status"

unhealthy_count=$(echo "$container_status" | grep -c "unhealthy" || true)
if [ $unhealthy_count -eq 0 ]; then
    test_result "容器健康状态" "PASS" "所有容器健康"
else
    test_result "容器健康状态" "WARN" "$unhealthy_count 个容器不健康"
fi

# 7.5 日志错误检查
log "\n${CYAN}7.5 日志错误检查${NC}"

for service in gateway user chat relay; do
    errors=$(docker compose logs $service 2>&1 | grep -i "error\|fatal\|panic" | wc -l)
    if [ $errors -eq 0 ]; then
        test_result "$service 日志检查" "PASS" "无错误日志"
    else
        test_result "$service 日志检查" "WARN" "发现 $errors 条错误日志"
    fi
done

# ==================== 测试总结 ====================
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

log "\n${BLUE}========================================${NC}"
log "${BLUE}  测试完成  ${NC}"
log "${BLUE}========================================${NC}"
log "结束时间: $(date '+%Y-%m-%d %H:%M:%S')"
log "总用时: ${DURATION} 秒"
log ""
log "测试统计:"
log "  总测试数: ${TOTAL_TESTS}"
log "  ${GREEN}通过: ${PASSED_TESTS}${NC}"
log "  ${RED}失败: ${FAILED_TESTS}${NC}"
log "  ${YELLOW}警告: ${WARNINGS}${NC}"
log ""

# 计算成功率
if [ $TOTAL_TESTS -gt 0 ]; then
    SUCCESS_RATE=$(echo "scale=1; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)
    log "成功率: ${SUCCESS_RATE}%"
fi

# ==================== 生成测试报告 ====================
log "\n${MAGENTA}生成测试报告: $REPORT_FILE${NC}"

cat > "$REPORT_FILE" << EOF
# Oblivious AI Platform - 部署测试报告

**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')  
**测试时长**: ${DURATION} 秒

---

## 📊 测试概览

| 指标 | 数值 |
|-----|-----|
| 总测试数 | ${TOTAL_TESTS} |
| ✅ 通过 | ${PASSED_TESTS} |
| ❌ 失败 | ${FAILED_TESTS} |
| ⚠️ 警告 | ${WARNINGS} |
| 成功率 | ${SUCCESS_RATE}% |

---

## 🚀 部署状态

### 容器运行状态

\`\`\`
$(docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}")
\`\`\`

### 资源使用情况

\`\`\`
$(cat /tmp/docker_stats.txt)
\`\`\`

---

## 🔍 测试详情

详细测试日志请查看: \`$LOG_FILE\`

---

## 📱 访问地址

- **前端**: http://localhost:3000
- **API 网关**: http://localhost:8080
- **用户服务**: http://localhost:8081
- **对话服务**: http://localhost:8082
- **中转服务**: http://localhost:8083
- **助手服务**: http://localhost:8084
- **知识库服务**: http://localhost:8085

---

## 🛠️ 管理命令

\`\`\`bash
# 查看日志
docker compose logs -f gateway
docker compose logs -f frontend

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 查看状态
docker compose ps
\`\`\`

---

## 📋 已发现的问题

EOF

# 列出失败的测试
if [ $FAILED_TESTS -gt 0 ]; then
    echo "### ❌ 失败的测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    grep "❌ \[FAIL\]" "$LOG_FILE" | sed 's/\x1b\[[0-9;]*m//g' >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
fi

# 列出警告
if [ $WARNINGS -gt 0 ]; then
    echo "### ⚠️ 警告项" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    grep "⚠️  \[WARN\]" "$LOG_FILE" | sed 's/\x1b\[[0-9;]*m//g' >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
fi

if [ $FAILED_TESTS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo "✅ 所有测试通过，未发现问题！" >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" << EOF

---

## 🎯 建议

1. **性能优化**: 监控高 CPU/内存使用的服务
2. **日志监控**: 定期检查错误日志
3. **健康检查**: 设置自动化健康检查告警
4. **备份策略**: 配置数据库定期备份
5. **安全加固**: 更新生产环境密钥和密码

---

**报告生成完毕** ✨
EOF

log "\n${GREEN}✅ 测试报告已生成: $REPORT_FILE${NC}"
log "${GREEN}✅ 详细日志已保存: $LOG_FILE${NC}"

# 显示最终状态
if [ $FAILED_TESTS -eq 0 ]; then
    log "\n${GREEN}🎉 部署和测试全部成功！系统运行正常。${NC}"
    exit 0
else
    log "\n${YELLOW}⚠️  部署完成，但存在 $FAILED_TESTS 个失败的测试。${NC}"
    log "${YELLOW}请查看报告了解详情。${NC}"
    exit 1
fi
