#!/bin/bash

# Oblivious 项目整合前备份脚本
# 将所有代码备份到 _backup 目录，不删除任何功能

echo "📦 Oblivious 项目整合前备份"
echo "================================"
echo ""

# 设置项目根目录
PROJECT_ROOT="/home/shirosora/windsurf-storage/oblivious/backend"
BACKUP_DATE=$(date +%Y-%m-%d_%H%M%S)
BACKUP_DIR="$PROJECT_ROOT/_backup/$BACKUP_DATE"

# 创建备份目录
echo "📁 创建备份目录: $BACKUP_DIR"
mkdir -p "$BACKUP_DIR/modules"
mkdir -p "$BACKUP_DIR/alternative_implementations"
mkdir -p "$BACKUP_DIR/original_services"

# 确认操作
echo ""
echo "⚠️  即将备份以下内容:"
echo "  - analytics 模块"
echo "  - performance 模块"
echo "  - monitoring 模块"
echo "  - queue 模块"
echo "  - 替代实现文件"
echo ""
read -p "是否继续？(y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    echo "❌ 操作已取消"
    exit 1
fi

cd "$PROJECT_ROOT" || exit 1

# ============================================
# 阶段 1: 备份完整模块
# ============================================
echo ""
echo "📦 阶段 1: 备份完整模块..."

# 备份 analytics 模块
if [ -d "internal/analytics" ]; then
    echo "  ✅ 备份 analytics 模块"
    cp -r internal/analytics "$BACKUP_DIR/modules/"
fi

# 备份 performance 模块
if [ -d "internal/performance" ]; then
    echo "  ✅ 备份 performance 模块"
    cp -r internal/performance "$BACKUP_DIR/modules/"
fi

# 备份 monitoring 模块
if [ -d "internal/monitoring" ]; then
    echo "  ✅ 备份 monitoring 模块"
    cp -r internal/monitoring "$BACKUP_DIR/modules/"
fi

# 备份 queue 模块
if [ -d "internal/queue" ]; then
    echo "  ✅ 备份 queue 模块"
    cp -r internal/queue "$BACKUP_DIR/modules/"
fi

# ============================================
# 阶段 2: 备份替代实现
# ============================================
echo ""
echo "📦 阶段 2: 备份替代实现..."

# 备份高级计费服务
if [ -f "internal/service/advanced_billing_service.go" ]; then
    echo "  ✅ 备份 advanced_billing_service.go"
    cp internal/service/advanced_billing_service.go \
       "$BACKUP_DIR/alternative_implementations/billing_advanced_v1.go"
fi

# 备份 billing 中的 token_counter
if [ -f "internal/billing/token_counter.go" ]; then
    echo "  ✅ 备份 billing/token_counter.go"
    cp internal/billing/token_counter.go \
       "$BACKUP_DIR/alternative_implementations/token_counter_v2.go"
fi

# 备份 billing 中的 quota_manager
if [ -f "internal/billing/quota_manager.go" ]; then
    echo "  ✅ 备份 billing/quota_manager.go"
    cp internal/billing/quota_manager.go \
       "$BACKUP_DIR/alternative_implementations/quota_manager_v2.go"
fi

# 备份适配器工厂
if [ -f "internal/adapter/factory.go" ]; then
    echo "  ✅ 备份 adapter/factory.go"
    cp internal/adapter/factory.go \
       "$BACKUP_DIR/alternative_implementations/adapter_factory_v1.go"
fi

# 备份适配器注册中心
if [ -f "internal/adapter/registry.go" ]; then
    echo "  ✅ 备份 adapter/registry.go"
    cp internal/adapter/registry.go \
       "$BACKUP_DIR/alternative_implementations/adapter_registry_v1.go"
fi

# 备份 relay 中的渠道选择器
if [ -f "internal/relay/channel_selector.go" ]; then
    echo "  ✅ 备份 relay/channel_selector.go"
    cp internal/relay/channel_selector.go \
       "$BACKUP_DIR/alternative_implementations/channel_selector_v2.go"
fi

# ============================================
# 阶段 3: 备份原始服务文件
# ============================================
echo ""
echo "📦 阶段 3: 备份原始服务文件..."

# 备份所有 service 文件
if [ -d "internal/service" ]; then
    echo "  ✅ 备份 service 目录"
    cp -r internal/service "$BACKUP_DIR/original_services/"
fi

# ============================================
# 阶段 4: 备份 token 和 chat 高级功能
# ============================================
echo ""
echo "📦 阶段 4: 备份高级功能..."

# 备份 token 生命周期
if [ -f "internal/token/lifecycle.go" ]; then
    echo "  ✅ 备份 token/lifecycle.go"
    cp internal/token/lifecycle.go \
       "$BACKUP_DIR/alternative_implementations/token_lifecycle.go"
fi

# 备份 token 权限管理
if [ -f "internal/token/permissions.go" ]; then
    echo "  ✅ 备份 token/permissions.go"
    cp internal/token/permissions.go \
       "$BACKUP_DIR/alternative_implementations/token_permissions.go"
fi

# 备份 chat 高级格式化
if [ -f "internal/chat/advanced_formatter.go" ]; then
    echo "  ✅ 备份 chat/advanced_formatter.go"
    cp internal/chat/advanced_formatter.go \
       "$BACKUP_DIR/alternative_implementations/chat_advanced_formatter.go"
fi

# 备份 chat 函数引擎
if [ -f "internal/chat/function_engine.go" ]; then
    echo "  ✅ 备份 chat/function_engine.go"
    cp internal/chat/function_engine.go \
       "$BACKUP_DIR/alternative_implementations/chat_function_engine.go"
fi

# ============================================
# 阶段 5: 创建备份说明文档
# ============================================
echo ""
echo "📝 阶段 5: 创建备份说明..."

cat > "$BACKUP_DIR/README.md" << 'EOF'
# Oblivious 项目备份说明

## 备份信息

- **备份时间**: 自动生成
- **备份原因**: 功能整合前的完整备份
- **备份策略**: 保留所有代码，不删除任何功能

## 备份内容

### 1. 完整模块 (modules/)

#### analytics/
- **用途**: 数据分析和统计
- **文件**:
  - `analytics_api.go` - API统计接口
  - `realtime_stats.go` - 实时统计
  - `usage_logger.go` - 使用日志记录
- **整合计划**: 集成到 relay 流程，添加仪表板和导出功能

#### performance/
- **用途**: 性能优化
- **文件**:
  - `optimization.go` - 性能优化工具
- **整合计划**: 扩展为完整的性能系统（缓存、连接池、批处理）

#### monitoring/
- **用途**: 系统监控
- **文件**: 监控相关实现
- **整合计划**: 集成健康检查、指标收集、链路追踪

#### queue/
- **用途**: 异步任务队列
- **文件**:
  - `async_queue.go` - 异步队列
  - `rabbitmq.go` - RabbitMQ实现
- **整合计划**: 支持多级队列（内存、Redis、RabbitMQ）

### 2. 替代实现 (alternative_implementations/)

#### 计费系统
- `billing_advanced_v1.go` - 高级计费服务（订阅、发票）
- `quota_manager_v2.go` - 配额管理器v2
- **整合计划**: 创建三层架构（引擎、基础服务、高级服务）

#### Token计数
- `token_counter_v2.go` - billing 模块的Token计数实现
- **整合计划**: 统一使用 tokenizer 模块

#### 适配器系统
- `adapter_factory_v1.go` - 适配器工厂
- `adapter_registry_v1.go` - 适配器注册中心
- **整合计划**: 保留并增强（热插拔、版本管理、降级策略）

#### 渠道选择
- `channel_selector_v2.go` - relay 模块的渠道选择实现
- **整合计划**: 统一使用 selector 模块

#### Token高级功能
- `token_lifecycle.go` - Token生命周期管理
- `token_permissions.go` - Token权限管理
- **整合计划**: 集成到 TokenService

#### Chat高级功能
- `chat_advanced_formatter.go` - 高级消息格式化
- `chat_function_engine.go` - 函数调用引擎
- **整合计划**: 集成到 ChatService

### 3. 原始服务 (original_services/)

完整备份 `internal/service/` 目录，包含所有服务的原始实现。

## 恢复方法

### 恢复完整模块
```bash
# 恢复 analytics 模块
cp -r _backup/2025-11-22_*/modules/analytics backend/internal/

# 恢复 performance 模块
cp -r _backup/2025-11-22_*/modules/performance backend/internal/
```

### 恢复单个文件
```bash
# 恢复高级计费服务
cp _backup/2025-11-22_*/alternative_implementations/billing_advanced_v1.go \
   backend/internal/service/advanced_billing_service.go
```

### 恢复所有服务
```bash
# 恢复整个 service 目录
cp -r _backup/2025-11-22_*/original_services/service backend/internal/
```

## 整合计划参考

详细的整合计划请参考项目根目录的以下文档：
- `INTEGRATION_PLAN.md` - 完整的整合计划
- `MODULE_AUDIT_REPORT.md` - 模块审计报告
- `REFACTOR_ACTION_PLAN.md` - 重构行动计划

## 注意事项

1. **不要删除备份**: 这些备份包含所有原始实现，可能在未来需要参考
2. **版本控制**: 建议同时使用 Git 进行版本控制
3. **测试验证**: 整合后务必进行完整的功能测试
4. **文档更新**: 整合完成后更新相关文档

## 联系方式

如有问题，请参考项目文档或联系开发团队。
EOF

# 创建整合笔记模板
cat > "$BACKUP_DIR/../integration_notes.md" << 'EOF'
# 整合笔记

## 整合进度

- [ ] 第1天: 备份和准备
- [ ] 第2天: 计费系统整合
- [ ] 第3天: Analytics 和 Performance 整合
- [ ] 第4天: Monitoring 和 Queue 整合
- [ ] 第5天: 适配器和 Token 功能整合

## 遇到的问题

### 问题1: 
**描述**: 
**解决方案**: 

### 问题2:
**描述**: 
**解决方案**: 

## 重要决策

### 决策1:
**日期**: 
**内容**: 
**原因**: 

## 测试记录

### 测试1:
**功能**: 
**结果**: 
**问题**: 

## 待办事项

- [ ] 任务1
- [ ] 任务2
EOF

# ============================================
# 阶段 6: 生成备份报告
# ============================================
echo ""
echo "📊 生成备份报告..."

REPORT_FILE="$BACKUP_DIR/backup_report.txt"
cat > "$REPORT_FILE" << EOF
Oblivious 项目备份报告
=====================

备份时间: $(date)
备份目录: $BACKUP_DIR

备份内容统计:
-------------
模块数量: $(find "$BACKUP_DIR/modules" -maxdepth 1 -type d | wc -l)
替代实现: $(find "$BACKUP_DIR/alternative_implementations" -type f | wc -l)
服务文件: $(find "$BACKUP_DIR/original_services" -type f -name "*.go" | wc -l)

详细列表:
---------

完整模块:
$(ls -la "$BACKUP_DIR/modules")

替代实现:
$(ls -la "$BACKUP_DIR/alternative_implementations")

备份大小:
---------
$(du -sh "$BACKUP_DIR")

下一步:
-------
1. 查看备份说明: cat $BACKUP_DIR/README.md
2. 开始整合: 参考 INTEGRATION_PLAN.md
3. 记录笔记: 编辑 $BACKUP_DIR/../integration_notes.md
EOF

echo "✅ 备份报告已生成: $REPORT_FILE"

# ============================================
# 完成
# ============================================
echo ""
echo "✅ 备份完成！"
echo ""
echo "📁 备份位置: $BACKUP_DIR"
echo "📄 备份说明: $BACKUP_DIR/README.md"
echo "📊 备份报告: $REPORT_FILE"
echo ""
echo "📋 下一步操作:"
echo "  1. 查看备份内容: ls -la $BACKUP_DIR"
echo "  2. 阅读备份说明: cat $BACKUP_DIR/README.md"
echo "  3. 开始整合: 参考 INTEGRATION_PLAN.md"
echo "  4. 提交备份到Git: git add backend/_backup && git commit -m '备份: 整合前的完整代码'"
echo ""
echo "⚠️  重要提示:"
echo "  - 所有原始代码已备份，不会丢失任何功能"
echo "  - 可以随时从备份恢复"
echo "  - 建议同时使用 Git 进行版本控制"
