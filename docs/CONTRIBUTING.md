# 贡献指南

## 欢迎贡献

感谢你对 Oblivious 项目的关注！我们欢迎所有形式的贡献。

## 贡献方式

### 报告 Bug

在 [GitHub Issues](https://github.com/your-org/oblivious/issues) 提交问题时请包含：

- 问题描述
- 复现步骤
- 预期行为
- 实际行为
- 环境信息（操作系统、Go/Node 版本等）
- 相关日志

### 提出新功能

提交 Feature Request 时请说明：

- 功能描述
- 使用场景
- 预期收益
- 可能的实现方案

### 提交代码

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'feat: add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

## 开发流程

### 环境搭建

参考 [快速开始指南](QUICK_START.md)。

### 代码规范

**Go 代码**：
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 格式化代码
- 通过 `golangci-lint` 检查

**TypeScript 代码**：
- 遵循 Airbnb 代码规范
- 使用 ESLint 和 Prettier
- 所有组件需要类型定义

### Commit 规范

使用 [Conventional Commits](https://www.conventionalcommits.org/)：

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
test: 测试相关
chore: 构建/工具链更新
```

示例：
```
feat(chat): 添加流式响应支持
fix(auth): 修复 token 过期问题
docs: 更新 API 文档
```

### 测试要求

- 新功能必须包含单元测试
- 测试覆盖率不低于 80%
- 提交前运行 `make test`

### Pull Request 流程

1. 确保代码通过所有测试
2. 更新相关文档
3. 填写 PR 模板
4. 等待代码审查
5. 根据反馈修改
6. 合并到主分支

## 代码审查

所有 PR 需要至少一名维护者审查通过。

审查重点：
- 代码质量
- 性能影响
- 安全性
- 可维护性
- 测试覆盖

## 许可证

提交代码即表示同意以 MIT 协议开源。

## 联系方式

- GitHub Issues
- Discord: https://discord.gg/oblivious
- Email: dev@oblivious.ai
