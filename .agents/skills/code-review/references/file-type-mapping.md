# 文件类型识别与规范映射

本文档定义了如何根据文件路径自动识别代码类型，并加载相应的审查规范。

## 简要说明

在代码审查过程中，不同类型的文件需要遵循不同的规范和最佳实践。本映射表提供了一个系统化的方法，帮助你：

1. **快速识别文件类型**：根据文件路径特征自动判断文件所属的代码类别
2. **加载对应规范**：找到该类型文件需要遵循的规范文档
3. **明确审查重点**：了解每种类型文件的核心关注点

## 文件类型映射表

| 文件类型 | 路径特征 | 规范文档 | 关注点 |
|---------|---------|---------|--------|
| **API 接口** | `proto/`, API 相关 | `docs/api/API接口开发流程.md` | 接口设计、错误码、HTTP 状态码 |
| **Proto 文件** | `*.proto` | `proto/README.md` | 命名规范、字段编号、注释 |
| **Executor 模式** | `internal/backend/dpmgr/executor.go` | [patterns/dpmgr-executor.md](../patterns/dpmgr-executor.md) | 函数系列一致性、日志格式 |
| **DAO 层** | `pkg/dao/` | `docs/developer/README.md` | 数据访问模式、错误处理 |
| **REST 框架** | `pkg/rest/` | `docs/api/API接口开发流程.md` | 错误响应、中间件 |

## 如何使用这个映射表

### 1. 识别文件类型

当你审查一个文件时，按照以下步骤识别其类型：

```bash
# 查看文件路径
# 例如：pkg/dao/deployment.go
```

根据路径特征匹配映射表：
- 路径包含 `pkg/dao/` → **DAO 层**
- 路径包含 `proto/` 或文件名是 `*.proto` → **API 接口** 或 **Proto 文件**
- 路径是 `internal/backend/dpmgr/executor.go` → **Executor 模式**
- 路径包含 `pkg/rest/` → **REST 框架**

### 2. 加载规范文档

识别出文件类型后，读取对应的规范文档：

```bash
# 示例：审查 DAO 层文件
Read file_path="/data/home/xyuzou/items/bk-nodemgr/docs/developer/README.md"
```

### 3. 应用审查重点

根据"关注点"列，在审查时重点检查相应的方面：

**示例：审查 DAO 层文件时**
- ✅ 数据访问模式是否符合规范
- ✅ 错误处理是否完整和一致
- ✅ 是否正确使用 ORM 或数据库操作

**示例：审查 API 接口文件时**
- ✅ 接口设计是否合理（RESTful 原则）
- ✅ 错误码定义是否规范
- ✅ HTTP 状态码使用是否正确

## 如何添加新映射

当项目中出现新的代码模块或文件类型时，需要添加新的映射规则。

### 步骤 1：确定路径特征

分析新模块的文件路径模式，找出唯一的识别特征：

```
例如：pkg/plugin/manager/*.go
特征：包含 pkg/plugin/manager/
```

### 步骤 2：创建规范文档

根据模块的特点，在合适的位置创建规范文档：

- **通用规范**：放在 `docs/developer/` 或 `docs/api/`
- **模块特定规范**：放在 `patterns/<module-name>.md`

可以使用 `patterns/_TEMPLATE.md` 作为模板。

### 步骤 3：明确关注点

列出该类型文件审查时的核心关注点（3-5 个最重要的方面）：

```markdown
例如：
- 插件注册机制
- 配置管理
- 生命周期管理
```

### 步骤 4：更新映射表

在本文档的映射表中添加新行：

```markdown
| **插件管理** | `pkg/plugin/manager/` | `patterns/plugin-manager.md` | 插件注册、配置管理、生命周期 |
```

### 步骤 5：更新 SKILL.md

同时更新 `/data/home/xyuzou/items/bk-nodemgr/.agents/skills/code-review/SKILL.md` 文件中的映射表（第 52-64 行），保持两处同步。

## 最佳实践

1. **保持映射简单明确**：路径特征应该足够具体，避免歧义
2. **规范文档要完整**：每个规范文档都应包含具体的检查清单和示例
3. **定期审查和更新**：随着项目演进，及时更新映射关系
4. **优先使用已有规范**：新类型如果可以使用现有规范，不要重复创建
5. **关注点要具体**：避免泛泛而谈，给出可操作的审查方向

## 相关文档

- [SKILL.md](../SKILL.md) - Code Review Skill 主文档
- [workflow-guide.md](./workflow-guide.md) - 详细的审查工作流程
- [go-standards.md](./go-standards.md) - Go 代码规范速查表
- [patterns/README.md](../patterns/README.md) - 模块规范说明
- [patterns/_TEMPLATE.md](../patterns/_TEMPLATE.md) - 模块规范模板
