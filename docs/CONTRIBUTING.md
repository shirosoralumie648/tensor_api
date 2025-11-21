# 贡献指南

感谢您对 Oblivious 项目的关注！我们欢迎任何形式的贡献，包括但不限于：

- 🐛 报告 Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码修复或新功能
- 🌍 翻译文档

---

## 行为准则

参与本项目即表示您同意遵守我们的行为准则：

- 尊重所有贡献者
- 使用友好和包容的语言
- 接受建设性批评
- 关注对社区最有利的事项
- 对他人表现出同理心

---

## 开始之前

### 环境准备

确保您的开发环境满足以下要求：

**后端开发**:
- Go 1.24.10 或更高版本
- PostgreSQL 15+
- Redis 7+
- Make 工具

**前端开发**:
- Node.js 20+
- npm 或 yarn
- 现代浏览器

**工具**:
- Git
- Docker & Docker Compose（推荐）
- VSCode 或其他代码编辑器

### 项目设置

1. **Fork 仓库**

访问 [Oblivious GitHub 仓库](https://github.com/your-org/oblivious)，点击右上角的 "Fork" 按钮。

2. **克隆到本地**

```bash
git clone https://github.com/your-username/oblivious.git
cd oblivious
```

3. **添加上游仓库**

```bash
git remote add upstream https://github.com/your-org/oblivious.git
```

4. **安装依赖**

```bash
# 后端
cd backend
go mod download

# 前端
cd ../frontend
npm install
```

5. **启动开发环境**

参考 [快速开始指南](../docs/QUICK_START.md)

---

## 开发流程

### 1. 创建分支

始终从最新的 `main` 分支创建特性分支：

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

分支命名规范：
- `feature/xxx` - 新功能
- `fix/xxx` - Bug 修复
- `docs/xxx` - 文档更新
- `refactor/xxx` - 代码重构
- `test/xxx` - 测试相关
- `chore/xxx` - 构建/工具相关

### 2. 进行开发

遵循项目的代码规范（见下文）进行开发。

### 3. 提交代码

**提交信息规范** (遵循 [Conventional Commits](https://www.conventionalcommits.org/))：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型 (type)**:
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 重构（不是新功能也不是 Bug 修复）
- `perf`: 性能优化
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动

**示例**:

```bash
git commit -m "feat(chat): 添加流式消息发送功能"
git commit -m "fix(auth): 修复 token 刷新失败的问题"
git commit -m "docs(api): 更新 API 文档"
```

### 4. 推送到 Fork

```bash
git push origin feature/your-feature-name
```

### 5. 创建 Pull Request

1. 访问您的 Fork 仓库
2. 点击 "New Pull Request"
3. 选择 base: `main` <- compare: `feature/your-feature-name`
4. 填写 PR 标题和描述
5. 提交 Pull Request

**PR 描述模板**:

```markdown
## 变更说明

简要描述这个 PR 做了什么。

## 变更类型

- [ ] Bug 修复
- [ ] 新功能
- [ ] 代码重构
- [ ] 文档更新
- [ ] 性能优化
- [ ] 其他（请说明）

## 相关 Issue

Closes #123

## 测试

说明如何测试这个变更。

## 截图（如适用）

如果是 UI 变更，请提供截图。

## 检查清单

- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 测试全部通过
- [ ] 更新了相关文档
- [ ] 没有引入新的警告
- [ ] 代码已自我审查
```

---

## 代码规范

### Go 代码规范

遵循 [Effective Go](https://go.dev/doc/effective_go) 和以下规范：

1. **格式化**

```bash
# 格式化代码
gofmt -w .

# 或使用 goimports
goimports -w .
```

2. **命名**

- 包名：小写，简短，无下划线
- 函数/方法：驼峰式命名，公开方法首字母大写
- 变量：驼峰式命名，缩写词全大写（如 `userID`, `httpClient`）
- 常量：驼峰式或全大写下划线分隔

3. **注释**

```go
// GetUser 根据用户 ID 获取用户信息
// 如果用户不存在，返回 ErrUserNotFound 错误
func GetUser(id int) (*User, error) {
    // 实现...
}
```

4. **错误处理**

```go
// 好的做法
if err != nil {
    return nil, fmt.Errorf("failed to get user: %w", err)
}

// 避免忽略错误
result, _ := SomeFunction() // ❌ 不好
```

5. **单元测试**

```go
func TestGetUser(t *testing.T) {
    tests := []struct {
        name    string
        userID  int
        want    *User
        wantErr bool
    }{
        {"valid user", 1, &User{ID: 1}, false},
        {"user not found", 999, nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GetUser(tt.userID)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("GetUser() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### TypeScript/React 代码规范

遵循 [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript) 和以下规范：

1. **格式化**

项目使用 ESLint 和 Prettier：

```bash
npm run lint
npm run lint:fix
npm run format
```

2. **命名**

- 组件：PascalCase (`UserProfile.tsx`)
- 函数/变量：camelCase (`getUserProfile`)
- 常量：UPPER_SNAKE_CASE (`API_BASE_URL`)
- CSS 类：kebab-case (`user-profile`)

3. **组件结构**

```tsx
import React, { useState, useEffect } from 'react';

interface UserProfileProps {
  userId: number;
  onUpdate?: (user: User) => void;
}

export const UserProfile: React.FC<UserProfileProps> = ({ userId, onUpdate }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchUser(userId).then(setUser).finally(() => setLoading(false));
  }, [userId]);

  if (loading) return <Spinner />;
  if (!user) return <NotFound />;

  return (
    <div className="user-profile">
      <h1>{user.name}</h1>
      {/* ... */}
    </div>
  );
};
```

4. **Hooks 使用**

- 优先使用函数组件和 Hooks
- 自定义 Hook 以 `use` 开头
- 遵循 [Hooks 规则](https://react.dev/reference/rules/rules-of-hooks)

5. **类型定义**

```typescript
// 使用 interface 而非 type（除非需要联合类型）
interface User {
  id: number;
  name: string;
  email: string;
}

// API 响应类型
interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}
```

### SQL 规范

1. **表名和字段名**

- 小写下划线分隔
- 表名使用复数 (`users`, `sessions`)
- 字段名清晰明确 (`created_at`, `user_id`)

2. **索引命名**

```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_messages_session_date ON messages(session_id, created_at);
```

3. **迁移文件**

- 使用序号前缀 (`000001_`, `000002_`)
- 包含 `.up.sql` 和 `.down.sql`
- 每个迁移只做一件事

---

## 测试

### 运行测试

**后端**:
```bash
cd backend
go test ./... -v
go test ./... -cover
```

**前端**:
```bash
cd frontend
npm test
npm run test:coverage
```

### 测试覆盖率要求

- 新代码测试覆盖率应 ≥ 80%
- 核心业务逻辑应 ≥ 90%
- PR 不应降低整体覆盖率

### 测试类型

1. **单元测试**: 测试单个函数/方法
2. **集成测试**: 测试服务间交互
3. **端到端测试**: 测试完整用户流程

---

## 文档

### 更新文档

当您的代码变更影响到以下内容时，请更新相应文档：

- API 接口：更新 `docs/API_REFERENCE.md`
- 配置项：更新相关配置文档
- 部署流程：更新 `docs/DEPLOYMENT_QUICK_REFERENCE.md`
- 新功能：更新 `README.md` 和相关文档

### 文档风格

- 使用 Markdown 格式
- 保持简洁清晰
- 提供代码示例
- 保持中英文文档同步（如适用）

---

## 报告 Bug

### 在提交 Bug 前

1. 搜索现有 [Issues](https://github.com/your-org/oblivious/issues)
2. 确保您使用的是最新版本
3. 尝试复现问题

### Bug 报告模板

```markdown
## Bug 描述

简要描述遇到的问题。

## 复现步骤

1. 进入 '...'
2. 点击 '....'
3. 滚动到 '....'
4. 看到错误

## 期望行为

描述您期望发生什么。

## 实际行为

描述实际发生了什么。

## 环境信息

- OS: [e.g. macOS 14.0]
- 浏览器: [e.g. Chrome 120]
- 版本: [e.g. v1.0.0]

## 截图

如果适用，添加截图帮助说明问题。

## 额外信息

添加任何其他有关问题的信息。
```

---

## 功能请求

### 功能请求模板

```markdown
## 功能描述

简要描述您希望添加的功能。

## 问题/动机

这个功能解决了什么问题？为什么需要它？

## 建议的解决方案

描述您希望如何实现这个功能。

## 替代方案

描述您考虑过的其他替代方案。

## 额外信息

添加任何其他有关功能请求的信息、截图等。
```

---

## 代码审查

所有 PR 都需要通过代码审查才能合并。

### 审查者指南

- 检查代码逻辑是否正确
- 检查是否遵循代码规范
- 检查测试是否充分
- 检查文档是否更新
- 提供建设性反馈

### 被审查者指南

- 及时响应审查意见
- 虚心接受建议
- 如有不同意见，礼貌讨论
- 完成修改后请求重新审查

---

## 发布流程

项目维护者负责发布新版本：

1. 更新版本号
2. 更新 CHANGELOG
3. 创建 Git Tag
4. 发布 GitHub Release
5. 部署到生产环境

---

## 社区

### 获取帮助

- 📖 查看 [文档](../README.md)
- 💬 加入 [Discord](https://discord.gg/oblivious)
- 📧 发送邮件到 dev@oblivious.ai
- 🐛 提交 [Issue](https://github.com/your-org/oblivious/issues)

### 保持联系

- GitHub: [@oblivious](https://github.com/your-org/oblivious)
- Twitter: [@ObliviousAI](https://twitter.com/obliviousai)
- 官网: [https://oblivious.ai](https://oblivious.ai)

---

## 致谢

感谢所有贡献者！您的贡献让 Oblivious 变得更好。

---

**文档版本**: v1.0.0  
**最后更新**: 2024 年 11 月 21 日  
**维护团队**: Oblivious 开发团队
