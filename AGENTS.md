<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# BlueKing Node Manager (bk-nodemgr)

Go 微服务应用，用于蓝鲸平台的节点管理。

**技术栈**: Go 1.23.10 | MongoDB | Redis | Gin | Vue.js + TypeScript

## 项目结构

```
cmd/         服务入口 (backend, application, file, relay)
pkg/         共享库 (dao, rest, workflow, thirdparty, logger, config)
internal/    业务逻辑
proto/       API 定义
docs/        📚 详细文档入口 → 查看 docs/README.md
```

## 快速开始

```bash
make pre     # 准备环境 (下载 Go 1.23.10)
make all     # 构建所有组件
make clean   # 清理构建产物
```

## 编码规范

### 核心原则
1. **遵循配置** - `.golangci.yml` 定义了所有 linter 规则
2. **错误处理** - 始终检查并处理错误
3. **结构化日志** - 使用 `pkg/logger` 包
4. **英文注释** - 公共函数和类型必须有注释

## 详细文档

📖 查看 **[docs/README.md](docs/README.md)** 获取完整文档导航：
- 开发手册 (编译、插件开发)
- API 开发流程
- 运维手册 (架构、部署)
- 概念文档

## 快速参考

- **Proto 构建**: `cd proto && make clean && make all`
- **代码检查**: `.golangci.yml` 包含所有规则配置
