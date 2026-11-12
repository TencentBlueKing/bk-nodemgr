# Quick Start

使用 `bkai-init==0.1.6` 的 Demo 完成首次资源同步。Demo 包含 Skill、知识库、子智能体 `ai-bkai-child` 和主智能体 `ai-bkai-demo`，自动按依赖顺序同步。

准备可用的 Podman（或 Docker）、平台提供的基础镜像和可安装 0.1.6 的包源。目标环境需开放应用接口、空间权限和 Skill 基础镜像白名单；MCP 使用前需完成授权。

## 1. 准备 Demo 与 .env

在 AIDEV 上游仓库根目录执行，也可完整复制 Demo 目录到自己的项目：

```bash
cd src/bkai-init/demo
# Preserve an existing .env.
test -f .env || cp .env.example .env
```

编辑 `package/bkai.yaml` 及其引用文件；修改资源 code 时同步修改关联，规则见[资源命名](agent-package.md#资源-code-命名)。`.env` 每项使用 `export`，替换全部占位符，真实文件不入库：

```bash
export BKAI_INIT_VERSION=0.1.6
export BKAI_APP_CODE='<应用编码>'
export BKAI_APP_SECRET='<应用密钥>'
export BK_API_URL_TMPL='https://gateway.example.com/api/{api_name}/{stage}'
export BKAI_SPACE_ID='<目标空间完整 ID>'

# Use true and the application's tenant in a multi-tenant environment.
export BKPAAS_MULTI_TENANT_MODE=false
export BKAI_TENANT_ID=

# Supply the actual bk_username, not a login_name to be converted.
export BKAI_USERNAME='<实际 bk_username>'

# Required by the Demo Skill Dockerfile; do not add a trailing slash.
export BKAI_AGENT_IMAGE_REPO=registry.example.com/team
```

专用网关不能套用模板时，另设 `BKAI_BASE_URL` 为平台提供的完整环境入口（含 stage，如 `/prod`）；它优先于模板。仅多租户且 `BKAI_USERNAME` 留空时查询默认 `bk_admin`，此时需可用的 BK User 地址与虚拟用户查询权限。显式填写 `bk_admin` 也会原样透传。

## 2. 构建、校验与预览

在 Demo 目录执行：

```bash
source .env
make build
make validate
make plan
```

- `make build` 从 `.env` 读取 `BKAI_INIT_VERSION`，基础镜像默认 `bkai-init:local`。使用平台提供的其它地址时，在 `.env` 配置 `export BKAI_INIT_IMAGE=...`，`source` 后生效；也可通过 `make build` 的 `BKAI_INIT_IMAGE` 参数传入。
- 构建产物为本地 `bkai-init-demo:local`，不会发布包或镜像。
- 使用 Docker 时，每条 make 命令增加 `CONTAINER_TOOL=docker`。
- `make validate` 不自行加载 `.env`，先 `source` 后自动读取已导出的 `BKAI_SPACE_ID`。
- `plan` 只预览，不保证后续写权限或服务端发布成功。

Demo Skill 根目录 Dockerfile 使用：

```dockerfile
FROM {{ BKAI_AGENT_IMAGE_REPO | required }}/bkdbm-aidev-skills-env:0.0.1-alpha.16
```

仓库变量必填且不带末尾 `/`，通过 make 的 `--var` 传入；渲染后的完整镜像名必须在目标平台允许列表内。`bkai-init` 包版本与 Skill 运行镜像版本相互独立。

修改包版本或 `package` 内的 YAML、文档、Skill Dockerfile 后，都要重新 `make build`。只修改 `.env` 的运行配置，下次同步即读取。

## 3. 同步并追加管理员

确认预览后执行；此步骤会写入平台并提交发布：

```bash
source .env
: "${BKAI_USERNAME:?请先填写实际 bk_username}"
make sync ADMINS="$BKAI_USERNAME"
```

不追加管理员时执行 `make sync`；多个管理员使用 `ADMINS="user_id_a user_id_b"`。Make 参数名为 `ADMINS`，值必须是实际 `bk_username`，由调用方自行转换。它应用于所有选中的主子 Agent，只追加成员，不删除已有管理员，不改写 YAML；YAML 不支持 `spec.admins`。

`make sync` 默认带 `--confirm` 和 `--publish`，先同步子智能体并提交发布，再关联主智能体，不等待后台完成。仅同步草稿使用 `PUBLISH=false`；首次主子初始化可能因子智能体没有可用发布版本而停止，此时使用默认整包发布流程。

## 4. 判断结果

| 日志结果                      | 含义与下一步                                                                              |
| ----------------------------- | ----------------------------------------------------------------------------------------- |
| 管理员追加完成                | 该应用的管理员追加接口成功，不等于后续发布身份校验一定通过                                |
| 已提交，后台处理中            | 知识导入或 Agent 发布已提交；到平台确认后台结果，CLI 不轮询                               |
| 当前状态为 publishing         | 发布接口明确以 HTTP 400 返回该状态时，记录原因并继续，不作为失败退出项                    |
| Skill 已有版本正在发布        | 明确的版本构建中响应会跳过本次更新并继续；其它上传或发布错误仍停止                        |
| `bk_username is not verified` | 检查失败步骤及 Trace ID，核对下游实际传递的 `bk_username`；追加管理员不能替代上游身份校验 |
| 基础镜像不在允许列表          | 核对渲染后的完整镜像名及目标租户有效白名单；修改 Dockerfile 后重新构建再同步              |

已完成操作不会自动回滚。限定重跑范围可用 `RESOURCES="agent/ai-bkai-child"` 等清单内资源；未选中的依赖不会自动同步。知识库目录同步会移除线上多余文档，应以完整本地目录为准。

下一步见[运行参数与执行](execution.md)、[部署、验收与排障](deployment.md)和[Package 与 Agent 协议](agent-package.md)。
