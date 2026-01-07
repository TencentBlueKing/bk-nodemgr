# BlueKing Node Manager (bk-nodemgr) - Cursor Rules

## 项目概述

这是一个基于 Go 的微服务应用，用于蓝鲸平台的节点管理系统。

## 技术栈

- **语言**: Go 1.23.10（固定版本，通过 `make pre` 自动下载）
- **数据库**: MongoDB (go.mongodb.org/mongo-driver)
- **缓存**: Redis (github.com/redis/go-redis/v9)
- **Web框架**: Gin (github.com/gin-gonic/gin)
- **CLI框架**: Cobra (github.com/spf13/cobra)
- **任务队列**: Machinery v2 (github.com/RichardKnop/machinery/v2)
- **前端**: Vue.js + TypeScript (使用 pnpm 管理依赖)
- **代码检查**: golangci-lint (配置在 `.golangci.yml`)

## 项目结构

### 服务组件
- `cmd/backend/` - 后端服务，核心节点管理逻辑
- `cmd/application/` - 应用服务，应用管理和调度
- `cmd/file/` - 文件服务，文件操作和存储
- `cmd/relay/` - 中继服务，代理和中继功能

### 核心包
- `cmd/` - 主入口和命令行
- `pkg/` - 共享工具和库
  - `pkg/dao/` - 数据访问层
  - `pkg/rest/` - REST API 框架
  - `pkg/workflow/` - 工作流管理
  - `pkg/thirdparty/` - 第三方服务集成（CMDB, GSE, BKRepo 等）
  - `pkg/logger/` - 结构化日志
  - `pkg/config/` - 配置管理
- `internal/` - 具体业务代码（backend, application, file 服务）
- `tools/` - 节点管理工具代码
- `proto/` - Protocol Buffers 定义文件
- `plugin/` - 插件定义和模板
- `script_tools/` - 脚本工具（gsectl, plugin_scripts）


## 构建命令

### 核心构建目标
- `make all` - 构建所有组件（backend, application, file, relay, front, tools, script_tools）
- `make backend` - 构建后端服务
- `make application` - 构建应用服务
- `make file` - 构建文件服务
- `make relay` - 构建中继服务
- `make front` - 构建前端（需要 pnpm）
- `make tools` - 构建额外工具
- `make script_tools` - 构建脚本工具
- `make clean` - 清理构建产物


### 开发命令
- `make pre` - 准备开发环境（下载 Go 1.23.10，运行 mod tidy）

### Docker 构建
- `make docker-build-server` - 构建服务器 Docker 镜像
- `make docker-build-apigw-sync` - 构建 API 网关同步 Docker 镜像

### proto 构建
进入 `proto/` 目录执行：
- `make clean` - 清理 proto 构建产物
- `make all` - 构建所有 proto 文件（生成 Go 代码和 Swagger 文档）
- proto 文件位于 `proto/{backend,application,file}/api/v3/` 目录
- 生成的 Go 代码输出到 `pkg/proto/` 目录
- 生成的 Swagger 文档输出到 `docs/swagger/` 目录

## 编码规范

### 项目特定规范
1. 遵循 'docs/style' 目录下的风格文档

### Go 代码规范

1. **导入别名**: 遵循 `.golangci.yml` 中定义的导入别名规则
   - 使用 `importas` linter 确保导入别名一致性

2. **错误处理**:
   - 始终检查并处理错误，不要忽略错误

3. **日志记录**:
   - 使用 `pkg/logger` 包进行结构化日志记录
   - 日志应包含足够的上下文信息

4. **代码质量**:
   - 遵循 `.golangci.yml` 中配置的所有 linter 规则
   - 运行 `golangci-lint run` 确保代码质量
   - 避免使用全局变量（`gochecknoglobals`）
   - 避免使用 init 函数（`gochecknoinits`）
   - 控制函数复杂度（`gocyclo`, `gocognit`）

5. **命名规范**:
   - 错误类型以 `Error` 结尾
   - 错误变量以 `Err` 开头
   - 接口命名简洁明了，避免接口污染（`interfacebloat`）

6. **注释规范**:
   - 公共函数和类型必须有注释
   - 注释以句号结尾（`godot`）
   - 使用中文注释（项目主要使用中文）

### 配置管理

- 服务通过 `-f` 标志接收配置文件路径
- 支持 debug/release 模式
- 支持单租户/多租户模式
- 配置文件位于 `etc/` 目录

### API 设计

- REST API 遵循 `pkg/rest` 框架的约定
- 使用统一的错误响应格式（通过 `pkg/rest/errf`）
- HTTP 状态码映射在 `pkg/rest/errf/http_status.go` 中定义

### 数据库操作

- 使用 `pkg/dao/mongo/` 包进行 MongoDB 操作
- 遵循自定义 DAO 模式
- 正确处理错误和上下文

## 开发工作流

### 开始开发前

1. 运行 `make pre` 准备开发环境
2. 确保 Go 1.23.10 已安装并可用
3. 运行 `go mod tidy` 确保依赖正确

### 编写代码时

1. 遵循项目的代码风格和规范
2. 使用统一的错误处理机制
3. 添加必要的注释和文档
4. 编写单元测试（如果适用）

### 提交前检查

1. 运行 `golangci-lint run` 检查代码质量
2. 运行 `go test ./...` 确保测试通过
3. 确保所有服务可以正常构建（`make all`）

## 特殊注意事项

### 版本管理

- 版本信息通过 ldflags 在构建时嵌入
- 版本格式：`${GITTAG}-${DATE}`（例如：v1.0.0-24.12.15）

### 前端开发

- 前端使用 pnpm（不是 npm）管理依赖
- 前端代码位于 `front/` 目录
- 构建前端：`make front`

### 服务启动

- 每个服务都有自己的主入口点和配置
- 服务使用 Cobra 进行 CLI 命令结构
- 服务需要配置文件路径（通过 `-f` 标志）
- 配置文件示例位于 `etc/` 目录

### 插件开发

- 插件定义位于 `plugin/` 目录
- 使用 `make plugin-pkg-relay` 构建插件包
- 插件支持多平台构建（linux/amd64, linux/arm64）

## 测试

- 使用标准 Go 测试命令运行测试
- 测试特定包：`go test ./pkg/dao/mongo/...`
- 运行所有测试：`go test ./...`

## 依赖管理

- 使用 Go modules（`go.mod`）
- 固定版本依赖确保构建一致性
- 某些依赖有版本替换规则（见 `go.mod` 的 `replace` 部分）

## 常见任务

### 添加新的错误码

1. 在 `pkg/rest/errf/code.go` 中添加错误码常量
2. 在 `pkg/rest/errf/error.go` 的 `errorMaps` 中添加错误映射
3. 在 `pkg/rest/errf/http_status.go` 中添加 HTTP 状态码映射
4. 遵循错误码规则：38号段 + 5位数字

### 添加新的服务端点

1. 在相应的服务目录下添加路由处理器
2. 使用 `pkg/rest` 框架的约定
3. 实现统一的错误处理
4. 添加必要的认证和授权检查

### 数据库操作

1. 在 `pkg/dao/mongo/` 中添加或修改 DAO 方法
2. 使用上下文传递
3. 正确处理错误
4. 考虑添加适当的索引

### 修改 proto 文件

1. 在 `proto/{service}/api/v3/` 目录下修改或添加 `.proto` 文件
2. 进入 `proto/` 目录运行 `make clean && make all` 生成代码
3. 检查生成的 Go 代码和 Swagger 文档
4. 更新相关的 API 处理逻辑

## 代码审查检查清单

- [ ] 代码遵循 `.golangci.yml` 中的 linter 规则
- [ ] 日志记录使用结构化日志
- [ ] 添加了必要的注释和文档
- [ ] 测试通过
- [ ] 构建成功
- [ ] 没有引入安全漏洞（`gosec`）
- [ ] 代码复杂度在合理范围内

## 参考文档

- `.golangci.yml` - 代码检查配置
- `README.md` - 项目 README
- `docs/` - 项目文档目录

