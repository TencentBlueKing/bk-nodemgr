# 部署、验收与排障

本文说明接入模块如何使用平台基础镜像、接入部署任务，并确认同步与后台处理结果。Demo 命令见 [Quick Start](quick-start.md)，运行参数见[运行参数与执行](execution.md)。

## 基础镜像与模块镜像

交付顺序为：先构建并发布 pip 包，再构建安装该包的基础镜像，最后由各模块加入自己的接入目录。基础镜像通过 pip 安装已发布版本，不复制 `bkai-init` 源码。

| 交付项    | 地址或版本                  | 使用要求                                                                  |
| --------- | --------------------------- | ------------------------------------------------------------------------- |
| Python 包 | `bkai-init==0.1.6`          | 固定版本，由平台提供可访问包源，不假设已发布公共 PyPI                     |
| 包源      | `BKAI_PIP_INDEX_URL`        | 提供 `bkai-init` 的 `/simple` 地址                                        |
| 依赖源    | `BKAI_DEPENDENCY_INDEX_URL` | 提供 Jinja2、PyYAML、requests 等依赖                                      |
| 基础镜像  | `BKAI_INIT_IMAGE`           | 平台提供的可拉取完整地址，使用 0.1.6 配套镜像，通过 `make build` 参数传入 |

Demo 的 `BKAI_INIT_VERSION` 设为 `0.1.6`，本地模块镜像默认名为 `bkai-init-demo:local`。确认所用构建包含常规发布及 admins 透传能力。

模块 Dockerfile 继承基础镜像内的入口：

```dockerfile
ARG BKAI_INIT_IMAGE=bkai-init:local
FROM ${BKAI_INIT_IMAGE}

ARG BKAI_INIT_VERSION=0.1.6
USER root
RUN python -m pip install --no-cache-dir "bkai-init==${BKAI_INIT_VERSION}"
USER 10001:10001

WORKDIR /app
COPY --chown=10001:10001 package /app/package

CMD ["sync", "--confirm", "-f", "/app/package/bkai.yaml"]
```

构建前确认包源中的目标版本可用。连接信息在运行时注入，不写入镜像或资源文件；`user_scope` 在 Agent YAML 的 `spec` 中设置。

### 进入已有 Demo 容器

宿主机在 Demo 目录执行：

```bash
make exec
```

容器内工作目录为 `/app`，`.env` 由运行时挂载，每项需使用 `export`：

```bash
source .env
bkai-init validate -f ./package/bkai.yaml --space "$BKAI_SPACE_ID"
bkai-init sync -f ./package/bkai.yaml --space "$BKAI_SPACE_ID" --confirm --publish
```

仅开启 `BKPAAS_MULTI_TENANT_MODE` 时使用 `BKAI_TENANT_ID`。不要打印凭据。

## Helm 接入

以下为业务 Chart 的接入示例，不是 bk-nodemgr 现有 values 契约：

```yaml
image:
  repository: registry.example.com/modules/bk-job-init
  tag: "0.1.6"
  pullPolicy: IfNotPresent

bkai:
  enabled: true
  tenantId: "<应用所在租户>"
  space: "<目标空间ID>"
  packagePath: /app/package/bkai.yaml
  excludeResources: []
```

- 模块 Chart 使用 `post-install` / `post-upgrade` Job，加载包含接入目录的模块镜像；`enabled` 控制运行，`excludeResources` 转换为重复的 `--exclude-resource`。
- Job 参数必须带 `sync --confirm`，并传入 `packagePath`、`space`。目标空间必须替换为真实完整 ID。
- 多租户还需注入 `BKPAAS_MULTI_TENANT_MODE=true` 和应用实际 `tenantId`；单租户不要求 `tenantId`。
- 示例 Job 默认只同步草稿。首次初始化主子智能体时，显式增加 `--publish --publish_config_only=0`，按先子后主顺序提交，不等待部署完成。
- 连接信息复用业务 Chart 现有环境变量注入机制，不新增 `credentialsSecretName`；真实凭据不进入 values 示例、镜像或 Git。
- 部署变量通过多个 `--var` 传入，限制见[变量替换](execution.md#变量替换)。
- Job 失败先看日志和线上状态再决定重跑，不能假设资源均已回滚。

## 验收同步结果

1. 检查命令结果，确认目标资源处理成功，黑名单中的资源未覆盖。
2. 再次执行 `show` / `plan`，核对名称、Prompt、模型、指令和资源关联。`show` 只查看线上配置，不查询权限；`plan` 预览同步动作和差异。Skill 文件、`envs` 值、角色标签、知识库文档等无法回读的内容，不能据此认定一致。
3. 知识库按任务 ID 确认 ZIP 导入完成、目录和文档符合预期；无查询入口时交由平台维护方核实。失败、部分成功或结果未知都不视为导入成功，不能用 CLI 正常退出替代任务验收。
4. 导入完成后，单独检查向量处理与检索效果。
5. CLI 未指定 `--publish` 时确认 Agent 草稿；`make sync` 默认发布，需核对平台发布版本，再验证对话、MCP 工具、Skill 和子智能体调用。

配置同步成功不等于运行验证完成，提交发布也不等于应用部署完成。

## 排障

先定位失败事项与步骤，携带脱敏日志反馈，不要只截取最后一行错误。反馈保留执行时间、CLI 版本、`trace_id` 与服务端请求标识，不粘贴原始凭据。

| 现象                                        | 排查方向                                                                                                                                              |
| ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| BK User HTTP 403 / `CROSS_TENANT_FORBIDDEN` | 请求租户与应用租户不一致；在 `.env` 设置应用实际 `BKAI_TENANT_ID`，不要用默认 `system`，重新 `source` 后执行                                          |
| 资源不存在或编码冲突                        | 检查创建状态、code 与唯一定位；同包本轮发布使用已同步配置，其它子智能体引用需有可用发布版本                                                           |
| 智能体不存在或不在当前空间                  | 核对租户、空间、已有智能体归属；首次未创建时确认平台支持 MCP 初始化查询规则，不创建空壳智能体绕过检查                                                 |
| 不支持的配置                                | 确认 CLI 与平台版本，先确认支持范围，不直接删除重要配置后覆盖已有 Agent                                                                               |
| `API_NOT_FOUND` / `1640401`                 | 网关路由未找到，不等同于 MCP 不存在；检查 `api_name=bk-aidev`、网关环境、资源路径与接口发布状态。仅业务明确返回 MCP 不存在时检查 MCP 创建、发布、授权 |
| `INITIALIZE_APP_MEMBERS_ERROR`              | 应用成员初始化调用 BK IAM 失败；提供 `create_group` 路径、HTTP 状态、`request_id` 给权限中心或平台维护方，不是 Skill 或本地 YAML 错误                 |
| 关联 Skill 版本不存在                       | 检查 Skill 存在且可访问，并确认平台部署了关联不依赖版本快照的修复；无需在 Agent 关联中补版本                                                          |
| `bk_username is not verified`               | 核对失败步骤、Trace ID 与下游实际传递的 `bk_username`；追加管理员不能替代上游身份校验                                                                 |
| 基础镜像不在允许列表                        | 核对渲染后完整镜像名及目标租户有效白名单；修改 Dockerfile 后重新构建再同步                                                                            |
| 同步中途失败                                | 先查看线上状态和差异，再修正重试；此前成功的资源可能已写入，不会自动全部回滚                                                                          |

仅明确的 Agent `publishing` HTTP 400 或 Skill 版本构建中响应可按约定记录并继续，其它写入或发布错误不自动重试。完整行为见[发布规则](execution.md#发布规则)。
