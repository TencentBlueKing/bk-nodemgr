# 代码目录结构

bk-nodemgr 是 Go 多服务后端 + Vue3/TypeScript 前端的 monorepo，服务启动入口位于 `cmd/`，各服务私有实现在 `internal/`，跨服务共享库在 `pkg/`，proto 源文件在 `proto/`。

## 顶层目录

```
bk-nodemgr/
├── cmd/            # 服务启动入口与依赖装配
├── internal/       # 各服务私有实现（router/service/storage 等）
├── pkg/            # 跨服务共享库（types/proto/dao/logger 等）
├── proto/          # protobuf 源文件（*.proto）
├── front/          # 前端工程（Vue3 + TypeScript，pnpm）
├── tools/          # 独立 Go module 的工具集
├── testsuite/      # Go package-level integration test support
├── install/        # 部署物料（docker-compose / helm / 镜像）
├── apigw/          # 蓝鲸 API 网关资源定义
├── script_tools/   # 运维与节点侧脚本（gsectl、iam、插件脚本等）
├── support-files/  # 部署支撑文件（配置策略、初始包、changelog 等）
├── docs/           # 项目文档
├── Makefile        # 根构建入口：二进制 / 前端 / 工具 / 镜像
└── ...
```

## cmd/ — 服务启动入口

| 目录                  | 说明                                                                                                                |
| --------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `cmd/backend/main.go` | backend 服务直启入口                                                                                                |
| `cmd/file/main.go`    | file 服务直启入口                                                                                                   |
| `cmd/relay/main.go`   | relay 服务直启入口                                                                                                  |
| `cmd/application/`    | 多命令 CLI（`root.go` 为根命令），子命令包括数据迁移（`migrate`）、调度器（`scheduler`）、Web 服务（`webserver`）等 |
| `cmd/adminclient/`    | 管理客户端入口                                                                                                      |

`cmd/backend`、`cmd/file`、`cmd/relay` 是直接启动的服务进程；`cmd/application` 是多命令 CLI。依赖装配（wire）集中在 `cmd/*` 完成。

## internal/ — 服务私有实现

按服务划分，各服务内部结构基本遵循同一分层模式：

```
internal/<service>/
├── options/    # 服务配置项定义
├── router/     # API 路由与 handler
├── service/    # 业务编排逻辑
├── storage/    # 该服务的存储访问层
└── ...         # 服务私有域组件
```

| 目录                    | 说明                                                                                                                                             |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `internal/backend/`     | 核心管理服务，含 `auth`（鉴权）、`dpmgr`、`manager`、`nodeconfig`、`periodictask` 等域组件；路由在 `router/api-v3`，另有 `admin`、`healthz` 路由 |
| `internal/application/` | 多命令 CLI 的实现，含 `distinctcache`、`frontsetting` 等                                                                                         |
| `internal/file/`        | 文件服务                                                                                                                                         |
| `internal/relay/`       | relay 服务，含 `handler`、`messagetracker`、`relayconstant` 等                                                                                   |
| `internal/adminclient/` | 管理客户端实现                                                                                                                                   |

约定：路由只做参数校验与协议转换，业务逻辑在 service，跨服务不互相引用内部包。

## pkg/ — 共享库

仅存放已被多个服务复用（或已证明共享价值）的代码，主要子包包括：

| 目录                                                   | 说明                                                                       |
| ------------------------------------------------------ | -------------------------------------------------------------------------- |
| `pkg/types`                                            | 跨层业务模型（层间载荷使用 types 而非 proto 结构体）                       |
| `pkg/proto`                                            | proto 生成代码目标目录，同时是 proto 结构体与业务模型之间的转换层          |
| `pkg/dao/mongo`                                        | MongoDB 持久化 DAO                                                         |
| `pkg/logger`                                           | 结构化日志                                                                 |
| `pkg/config` / `pkg/envx` / `pkg/contextx`             | 配置、环境变量、上下文                                                     |
| `pkg/installer`                                        | Agent 安装相关共享逻辑                                                     |
| `pkg/scheduler` / `pkg/workflow` / `pkg/batchexecutor` | 任务调度、工作流、批量执行                                                 |
| 其余                                                   | `metrics`、`tracing`、`rest`、`sshx`、`downloader`、`version` 等基础能力包 |

约定：proto 结构体仅作为传输边界类型，进入业务逻辑前必须经 `pkg/proto/*` 转换为 `pkg/types` 模型。

## proto/ — 接口定义

proto 源文件按服务组织：`proto/backend`（含 `api`、`callback`）、`proto/application`、`proto/file`、`proto/thirdparty`。

修改 proto 后重新生成：

```bash
cd proto && make clean && make all
```

生成产物落在 `pkg/proto/**`，禁止手工编辑生成文件。

## front/ — 前端工程

Vue3 + TypeScript + Vite，包管理器为 pnpm@9.8.0。

| 目录                   | 说明                                |
| ---------------------- | ----------------------------------- |
| `front/src/api`        | API 封装，命名与后端 proto 命名对齐 |
| `front/src/pages`      | 页面模块                            |
| `front/src/components` | 公共组件                            |
| `front/src/stores`     | 状态管理                            |
| `front/src/modules`    | 业务模块                            |

本地开发：

```bash
cd front && pnpm install && pnpm dev
```

## tools/ — 工具集

独立 Go module（有自己的 `go.mod`），包含构建辅助工具，由根 Makefile 统一编排。

## install/ — 部署

包含 docker-compose、Helm chart、镜像构建配置及 metrics 相关物料。

## 常用命令

| 命令         | 说明                  |
| ------------ | --------------------- |
| `make pre`   | 提交前检查（lint 等） |
| `make all`   | 构建全部产物          |
| `make lint`  | 代码静态检查          |
| `make clean` | 清理构建产物          |
