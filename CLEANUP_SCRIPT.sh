#!/bin/bash

# Oblivious 项目清理脚本
# 删除冗余和未使用的模块

echo "🧹 开始清理 Oblivious 项目..."
echo "⚠️  请确保已备份重要数据！"
echo ""

# 设置项目根目录
PROJECT_ROOT="/home/shirosora/windsurf-storage/oblivious/backend"

# 确认操作
read -p "是否继续？(y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    echo "❌ 操作已取消"
    exit 1
fi

echo "📂 进入项目目录: $PROJECT_ROOT"
cd "$PROJECT_ROOT" || exit 1

# ============================================
# 阶段 1: 删除完全未使用的模块
# ============================================
echo ""
echo "🗑️  阶段 1: 删除未使用的模块..."

# 删除 analytics 模块
if [ -d "internal/analytics" ]; then
    echo "  ❌ 删除 internal/analytics/"
    rm -rf internal/analytics/
fi

# 删除 performance 模块
if [ -d "internal/performance" ]; then
    echo "  ❌ 删除 internal/performance/"
    rm -rf internal/performance/
fi

# 删除 monitoring 模块
if [ -d "internal/monitoring" ]; then
    echo "  ❌ 删除 internal/monitoring/"
    rm -rf internal/monitoring/
fi

# 删除 queue 模块
if [ -d "internal/queue" ]; then
    echo "  ❌ 删除 internal/queue/"
    rm -rf internal/queue/
fi

# ============================================
# 阶段 2: 删除重复的服务文件
# ============================================
echo ""
echo "🗑️  阶段 2: 删除重复的服务文件..."

# 删除高级计费服务（保留基础版本）
if [ -f "internal/service/advanced_billing_service.go" ]; then
    echo "  ❌ 删除 internal/service/advanced_billing_service.go"
    rm internal/service/advanced_billing_service.go
fi

# 删除 billing 中的重复 token_counter
if [ -f "internal/billing/token_counter.go" ]; then
    echo "  ❌ 删除 internal/billing/token_counter.go (保留 tokenizer 模块)"
    rm internal/billing/token_counter.go
fi

# ============================================
# 阶段 3: 删除过度抽象的适配器文件
# ============================================
echo ""
echo "🗑️  阶段 3: 删除适配器过度抽象..."

# 删除适配器工厂（保留简单的 adapter.go）
if [ -f "internal/adapter/factory.go" ]; then
    echo "  ❌ 删除 internal/adapter/factory.go"
    rm internal/adapter/factory.go
fi

# 删除适配器注册中心
if [ -f "internal/adapter/registry.go" ]; then
    echo "  ❌ 删除 internal/adapter/registry.go"
    rm internal/adapter/registry.go
fi

# ============================================
# 阶段 4: 删除冲突的实现
# ============================================
echo ""
echo "🗑️  阶段 4: 删除冲突的实现..."

# 删除 relay 中的重复渠道选择器
if [ -f "internal/relay/channel_selector.go" ]; then
    echo "  ❌ 删除 internal/relay/channel_selector.go (保留 selector 模块)"
    rm internal/relay/channel_selector.go
fi

# 删除 billing 中的配额管理器（保留 quota 模块）
if [ -f "internal/billing/quota_manager.go" ]; then
    echo "  ⚠️  标记删除 internal/billing/quota_manager.go (需要迁移逻辑到 quota 模块)"
    # 暂时不删除，需要先迁移逻辑
    # rm internal/billing/quota_manager.go
fi

# ============================================
# 阶段 5: 删除未使用的 token 功能
# ============================================
echo ""
echo "🗑️  阶段 5: 删除未使用的 token 功能..."

# 删除 token 生命周期管理
if [ -f "internal/token/lifecycle.go" ]; then
    echo "  ❌ 删除 internal/token/lifecycle.go"
    rm internal/token/lifecycle.go
fi

# 删除 token 权限管理
if [ -f "internal/token/permissions.go" ]; then
    echo "  ❌ 删除 internal/token/permissions.go"
    rm internal/token/permissions.go
fi

# ============================================
# 阶段 6: 删除未使用的 chat 高级功能
# ============================================
echo ""
echo "🗑️  阶段 6: 删除未使用的 chat 功能..."

# 删除高级格式化器
if [ -f "internal/chat/advanced_formatter.go" ]; then
    echo "  ❌ 删除 internal/chat/advanced_formatter.go"
    rm internal/chat/advanced_formatter.go
fi

# 删除函数引擎
if [ -f "internal/chat/function_engine.go" ]; then
    echo "  ❌ 删除 internal/chat/function_engine.go"
    rm internal/chat/function_engine.go
fi

# ============================================
# 阶段 7: 清理测试文件中的错误
# ============================================
echo ""
echo "🧪 阶段 7: 清理测试文件..."

# 删除有问题的测试文件（可选）
# find internal/ -name "*_test.go" -type f -exec grep -l "ILLEGAL\|syntax error" {} \; | while read file; do
#     echo "  ⚠️  发现有问题的测试文件: $file"
# done

# ============================================
# 阶段 8: 生成清理报告
# ============================================
echo ""
echo "📊 生成清理报告..."

REPORT_FILE="../CLEANUP_REPORT.txt"
cat > "$REPORT_FILE" << EOF
Oblivious 项目清理报告
生成时间: $(date)

已删除的模块:
- internal/analytics/
- internal/performance/
- internal/monitoring/
- internal/queue/

已删除的文件:
- internal/service/advanced_billing_service.go
- internal/billing/token_counter.go
- internal/adapter/factory.go
- internal/adapter/registry.go
- internal/relay/channel_selector.go
- internal/token/lifecycle.go
- internal/token/permissions.go
- internal/chat/advanced_formatter.go
- internal/chat/function_engine.go

需要手动处理:
- internal/billing/quota_manager.go (需要迁移逻辑)
- internal/cache/ 中的 CacheStats 重复定义
- 各种测试文件的语法错误

下一步:
1. 运行 go mod tidy 清理依赖
2. 运行 gofmt -w . 格式化代码
3. 运行 go build 检查编译错误
4. 补全缺失的调用链
EOF

echo "✅ 清理报告已生成: $REPORT_FILE"

# ============================================
# 阶段 9: 清理依赖
# ============================================
echo ""
echo "📦 阶段 9: 清理 Go 模块依赖..."

go mod tidy

# ============================================
# 阶段 10: 格式化代码
# ============================================
echo ""
echo "✨ 阶段 10: 格式化代码..."

gofmt -w internal/
gofmt -w cmd/

# ============================================
# 完成
# ============================================
echo ""
echo "✅ 清理完成！"
echo ""
echo "📋 下一步操作:"
echo "  1. 查看清理报告: cat $REPORT_FILE"
echo "  2. 检查编译错误: go build ./..."
echo "  3. 运行测试: go test ./..."
echo "  4. 查看模块审计报告: cat ../MODULE_AUDIT_REPORT.md"
echo ""
echo "⚠️  注意: 某些文件需要手动迁移逻辑后才能删除"
