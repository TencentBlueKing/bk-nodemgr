# APIGateway 同步 FAQ

## 1. 如何通过 Helm 开启 APIGateway 同步镜像？

这个问题对应 Helm 配置里的 `apigwSync` 和 `apiManagerImage`。

`apigwSync.enabled` 开启后，Chart 会创建一个一次性 Job。这个 Job 使用 `apiManagerImage` 指定的 `bk-nodemgr-apigw-sync`
镜像，挂载由 `apigwSync.config` 渲染出的 `/data/definition.yaml`，然后执行镜像内的 `/data/bin/sync-apigateway.sh`。

一个可直接参考的配置片段：

```yaml
apiManagerImage:
  registry: "hub.bktencent.com"
  repository: "blueking/bk-nodemgr-apigw-sync"
  tag: v3.0.1-alpha.40
  pullPolicy: IfNotPresent

apigwSync:
  enabled: true
  extraEnvVars:
    - name: BK_APIGW_NAME
      value: "bk-nodemgr"
    - name: BK_APP_CODE
      value: "bk-nodemgr"
    - name: BK_APP_SECRET
      value: "<bk_app_secret>"
    - name: BK_API_URL_TMPL
      value: "http://bkapi.example.com/api/{api_name}"
    - name: RELEASE_STAGES
      value: "stage"
    - name: NO_PUB
      value: "true"
  config:
    spec_version: 1
    release:
      version: "2.14.1"
      title: "bk-nodemgr"
      comment: "v3.0.1-alpha.40"
    apigateway:
      description: "蓝鲸节点管理"
      description_en: "bk-nodemgr"
      is_public: false
      maintainers:
        - "admin"
    stages:
      - name: "stage"
        description: "预发布环境"
        description_en: "stage"
        backends:
          - name: "default"
            config:
              timeout: 60
              loadbalance: "roundrobin"
              hosts:
                - host: "https://<bk-nodemgr-backend-host>"
                  weight: 100
    grant_permissions:
      - bk_app_code: "bk-nodemgr"
        grant_dimension: "resource"
        resource_names:
          - "<resource_name>"
```

`BK_APP_SECRET` 示例中只保留占位符，不要把真实 Secret 提交到仓库或长期保存在公开的 values 文件中。

## 2. 每个配置项分别影响什么？

| 配置项                                    | 作用                                                                                      |
| ----------------------------------------- | ----------------------------------------------------------------------------------------- |
| `apiManagerImage.registry/repository/tag` | 指定同步 Job 使用的 `bk-nodemgr-apigw-sync` 镜像。                                        |
| `apigwSync.enabled`                       | 是否创建 APIGateway 同步 Job。                                                            |
| `BK_APIGW_NAME`                           | 提供给 `apigw-manager` 的网关名配置；当前脚本命令显式使用 `bk-nodemgr`，建议保持一致。    |
| `BK_APP_CODE` / `BK_APP_SECRET`           | 调用 `bk-apigateway` 管理接口所需的应用身份。                                             |
| `BK_APP_TENANT_ID`                        | 调用 `bk-apigateway` 管理接口时透传的应用租户 ID；单租户传 `default`，多租户传 `system`。 |
| `BK_API_URL_TMPL`                         | 网关管理 API 地址模板，必须保留 `{api_name}` 占位符。                                     |
| `RELEASE_STAGES`                          | 创建版本后发布到哪个环境，应与 `apigwSync.config.stages[].name` 对齐。                    |
| `NO_PUB`                                  | 是否额外执行一次只生成版本不发布的命令。                                                  |
| `apigwSync.config.release`                | 渲染到 `definition.yaml` 的版本信息。                                                     |
| `apigwSync.config.apigateway`             | 渲染到 `definition.yaml` 的网关基础信息和维护人。                                         |
| `apigwSync.config.stages`                 | 渲染到 `definition.yaml` 的网关环境和后端回源地址。                                       |
| `apigwSync.config.grant_permissions`      | 渲染到 `definition.yaml` 的主动授权配置。                                                 |

`stages[].backends[].config.hosts[].host` 只写协议、域名或 IP，不包含 Path，例如 `https://bk-nodemgr.example.com`。

Helm 会根据 `backend.config.tenantMode` 自动注入 `BK_APP_TENANT_ID`：`single` 对应 `default`，`multiple` 对应
`system`，不要在 `apigwSync.extraEnvVars` 中重复配置。docker-compose 通过 `BK_NODEMGR_APP_TENANT_ID` 设置该值。
`apigw-manager` 会把它转换成请求头 `X-Bk-Tenant-Id`。如果同步日志出现 `1640302 Cross tenant forbidden`，先确认同步镜像
基于支持多租户的 `apigw-manager` 版本，再检查该变量是否按租户模式设置。

> notice: 一个环境的 `BK_API_URL_TMPL` 可以通过我们注册到 bk-apigw 后生成的 API 地址来获取。比如我们生成的 API 地址是
> `https://bk-nodemgr.apigw.example.com/api/v1/nodemgr/nodes`，那么 BK_API_URL_TMPL 就是 `https://bk-nodemgr.apigw.example.com/api/{api_name}`。

## 3. `NO_PUB=true` 是否表示一定不会发布？

不是。

当前同步脚本会先读取：

```bash
release_stages="${RELEASE_STAGES:-}"
no_pub="${NO_PUB:-true}"
```

当 `NO_PUB=true` 时，脚本会先执行一次 `create_version_and_release_apigw --no-pub`，只生成版本不发布。

随后，只要 `RELEASE_STAGES` 非空，脚本仍会继续执行一次带 `--stage` 的发布命令：

```bash
create_version_and_release_apigw definition.yaml --gateway-name=bk-nodemgr --stage "${RELEASE_STAGES}"
```

因此，默认的 `NO_PUB=true` 与空 `RELEASE_STAGES` 组合表示：只生成一个未发布版本，不发布任何环境。
如需在生成版本后发布到指定环境，需要显式设置 `RELEASE_STAGES`。

## 4. `resources.yaml` 和接口文档从哪里来？

Helm values 只会渲染并挂载 `/data/definition.yaml`。

同步脚本还会读取 `/data/resources.yaml` 和 `/data/apidocs/`。这些内容不是由 `apigwSync.config` 生成的，而是在构建
`bk-nodemgr-apigw-sync` 镜像时从仓库的 `apigw/` 目录复制进镜像：

```bash
make docker-build-apigw-sync
```

如果只修改了 Helm values，通常只会影响网关定义、环境、后端地址和授权配置；如果需要更新网关资源列表或接口文档，需要更新
`apigw/resources.yaml`、`apigw/apidocs/`，并重新构建同步镜像。

这个同步镜像基于 BlueKing `apigw-manager`
，相关命令能力可参考 [apigw-manager](https://github.com/TencentBlueKing/bkpaas-python-sdk/tree/master/sdks/apigw-manager)。

## 5. 如何确认同步 Job 已经执行？

升级或安装 Chart 后，先确认 Job 是否创建：

```bash
kubectl get job -n <namespace> | grep apigw-sync
```

再查看 Pod 日志：

```bash
kubectl logs -n <namespace> job/<fullname>-apigw-sync-<revision>
```

如果需要确认 Helm 渲染出的网关定义，可以查看 ConfigMap：

```bash
kubectl get configmap -n <namespace> <fullname>-apigw-sync-definition -o yaml
```

重点检查 `definition.yaml` 里的 `stages[].name`、后端 `host`、`grant_permissions` 是否和预期一致。

## 6. 已设置 `BK_APP_TENANT_ID=system`，为什么还反复提示设置租户 ID？

### 现象与原因

同步日志反复出现以下提示，但请求头中的租户 ID 已经是 `system`：

```text
[warn] if the syncing to apigateway failed, and your app(bk-nodemgr) is a global tenant app, please set the environment variable BK_APP_TENANT_ID to `system` (or set django settings.BK_APP_TENANT_ID to `system`) and try again. the current value [X-Bk-Tenant-Id=system].
```

**这条 warning 本身不表示租户配置错误，也不表示同步失败。** 在 `apigw-manager:4.2.4` 中，未设置
`BKPAAS_APP_TENANT_ID` 时，SDK 会读取 `BK_APP_TENANT_ID`，并无条件输出这条提示；即使值已经是 `system` 也会输出。
每次构造 API 请求都会经过该逻辑，因此一次同步可能重复出现多次。

Helm 和 docker-compose 通过 `BK_APP_TENANT_ID` 传入应用租户 ID，属于该版本 SDK 支持的独立部署方式，
不需要因为这条提示改用 PaaS 的配置入口。

### 如何处理

- 如果 `X-Bk-Tenant-Id` 与应用所属租户一致，且最终同步成功，可以忽略这条提示，无需修改配置或重跑 Job。
- 不要仅为消除 warning 而添加 `BKPAAS_APP_TENANT_ID`：它会优先于 `BK_APP_TENANT_ID` 生效，且空字符串表示全租户应用，
  SDK 会将其转换为 `system`，并非单纯的日志开关。也不要全局屏蔽 WARNING，以免隐藏其他告警。
- 如果同步确实失败，应结合接口错误码和错误信息排查；遇到 `1640302 Cross tenant forbidden` 时，按第 2 节核对租户配置。
  不要只凭这条 warning 判断失败原因。

## 7. MCP 同步报 `mcp_servers: 该字段是必填项` 怎么处理？

### 现象与原因

旧版 Chart 使用 `ENABLE_SYNC_MCP_SERVERS=true`，并在环境 values 中显式覆盖 `apigwSync.config.stages`，但没有为覆盖后的 stage
配置 `mcp_servers` 时，同步 Job 在 MCP 同步步骤报错：

```text
Response: {"error":{"code":"INVALID_ARGUMENT","message":"校验失败: mcp_servers: 该字段是必填项。","data":null}}
```

先检查环境 values 是否重新定义了 `apigwSync.config.stages`。**Helm 对列表采用整体替换，不会按 `name` 合并列表项内的字段。**
即使没有显式设置 `mcp_servers`，只要覆盖了 `stages`，Chart 默认 stage 中的 MCP 配置也不会自动保留。例如：

```yaml
apigwSync:
  config:
    stages:
      - name: "prod"
        description: "正式环境"
        description_en: "Production"
        backends:
          - name: "default"
            config:
              timeout: 60
              loadbalance: "roundrobin"
              hosts:
                - host: "https://<bk-nodemgr-backend-host>"
                  weight: 100
```

上面的覆盖结果只有后端配置，没有 `mcp_servers`。旧版 Chart 模板会省略空或缺失的 `mcp_servers`，而 MCP 同步命令会遍历
`stages` 发起请求，不会自动跳过缺少该字段的环境，因此接口返回必填项错误。

支持 `apigwSync.mcp.enabled` 的新版 Chart 将内置定义与环境列表分离，由同步镜像在运行时生成只包含非空 MCP 定义的同步文件。
开启新特性后，上述 stage 会自动获得产品内置 MCP；显式列表完整优先，显式 `[]` 则跳过该 stage 的 MCP 同步，不会删除或停用已有 MCP。

### 排查与修复

1. 查看失败 Job 使用的网关定义，而不只是仓库中的 `apigw/definition.yaml`：

   ```bash
   kubectl get configmap -n <namespace> <fullname>-apigw-sync-definition \
     -o jsonpath='{.data.definition\.yaml}'
   ```

   检查 `stages` 下每个环境的 `mcp_servers` 是否存在且非空，并与部署时所有 values 文件及覆盖参数核对。
   如果同步 Pod 仍在运行，也可以查看其实际挂载文件：

   ```bash
   kubectl exec -n <namespace> <apigw-sync-pod> -- cat /data/definition.yaml
   ```

2. 按实际部署版本选择修复方式，保留已有环境名、后端地址及其他所需配置：
   - 旧版 Chart：从该版本 `values.yaml` 复制完整的 `apigwSync.config.stages[].mcp_servers`，补到对应 stage 下，与 `backends` 同级。
     多个 stage 都需要分别检查；不要仅补 `mcp_servers: []`，旧模板同样会省略空列表。
   - 支持新特性的 Chart：设置 `apigwSync.mcp.enabled=true`，让未显式配置 `mcp_servers` 的 stage 自动使用并启用产品内置 MCP。
     显式非空列表仍完整优先，不合并字段；显式 `[]` 不参与同步。需要停用已有 MCP 时，提供完整定义并设置 `status: 0`。
     此开关优先于旧同步环境变量，但不改变发布策略、公开性或 OAuth 设置。
   - 升级时，若此前仅通过 `ENABLE_SYNC_MCP_SERVERS=true` 继承默认清单，需要迁移到新开关，或显式保留原完整定义及状态。
     新版默认 `stages` 不再包含 MCP 清单；Chart 与同步镜像需配套升级，旧镜像不支持运行时装配。
     当前产品内置清单使用 `OpenTopo_*` 资源和 `Official` 分类；显式覆写不会自动迁移到该清单。
3. 确认 `apigwSync.config.spec_version` 为 `2`，且 `mcp_servers[].resource_names` 引用的资源已发布到对应网关环境。
   Helm Job 用 ConfigMap 覆盖 `/data/definition.yaml`，其中保留部署者的显式配置。新版镜像另有不受挂载影响的
   `/data/mcp_servers.yaml`，来源为构建时的 `apigw/mcp_servers.yaml`，仅用于新特性开启且未显式覆写的 stage。
   修改产品清单后必须重新构建并升级镜像，不会改变已有显式覆写。
   新版需在同步 Pod 中检查运行时生成的 `/data/mcp-definition.yaml`，它才是本次 MCP 同步的实际输入，不在 ConfigMap 中：

   ```bash
   kubectl exec -n <namespace> <apigw-sync-pod> -- cat /data/mcp-definition.yaml
   ```

4. 使用实际部署的 Chart 和完整覆盖参数重新渲染，确认 `definition.yaml` 中的显式配置和 Job 的特性开关符合预期，再按原部署流程更新并重新执行同步 Job。
   新版在 Pod 运行时生成 MCP 清单，不能仅凭 `helm template` 判断最终同步范围。
   重试前检查 `RELEASE_STAGES`、`NO_PUB` 和资源变更范围：同步脚本不只处理 MCP，还会同步资源（带 `--delete`）并可能发布版本，
   不能把重跑整个 Job 当作只读验证。完成后检查 Job 日志和网关对应环境的 MCP Server 配置，确认同步成功。

## 8. 如何使用系统诊断 MCP？

产品清单 `apigw/mcp_servers.yaml` 中的 `system-diagnosis` 展示名称为“节点管理-系统诊断”，与 `topo`、`execution` 独立。
新增系统诊断能力统一扩展此 MCP，不按诊断能力另建 MCP；各工具仍须遵循对应功能域的权限和数据访问边界。
当前提供定时任务调度诊断的以下四个只读工具：

| 网关资源名                                                         | 用途                                                                                             |
| ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| `OpenScheduleWorkflow_OpenScheduleWorkflowList`                    | 查询调度任务及 `enabled`、`interval` 等配置                                                      |
| `OpenScheduleWorkflow_OpenScheduleWorkflowOperationList`           | 使用 `workflow_id` 查询当前触发器关联的操作，配合 `parent_oper_inst_id` 查询某轮派发的直接子操作 |
| `OpenScheduleWorkflow_OpenScheduleWorkflowOperationInstanceList`   | 使用 `operation_id` 或 `oper_inst_id` 查询执行实例，两者均提供时取交集                           |
| `OpenScheduleWorkflow_OpenScheduleWorkflowOperationInstanceLogGet` | 使用 `oper_inst_id` 查询步骤状态、双语日志和额外执行日志                                         |

接口参数、返回值及记录保留规则见[任务列表](../../apigw/apidocs/zh/OpenScheduleWorkflow_OpenScheduleWorkflowList.md)、
[操作列表](../../apigw/apidocs/zh/OpenScheduleWorkflow_OpenScheduleWorkflowOperationList.md)、
[实例列表](../../apigw/apidocs/zh/OpenScheduleWorkflow_OpenScheduleWorkflowOperationInstanceList.md)和
[实例日志](../../apigw/apidocs/zh/OpenScheduleWorkflow_OpenScheduleWorkflowOperationInstanceLogGet.md)。
`enabled`、`interval` 是长期调度配置，不表示执行完成；派发实例成功也不代表子任务成功。
查询仅覆盖当前触发器关联且尚未清理的任务链，不是完整历史归档。

### 授权与启用

- `is_public: false` 仅控制可见性，不等于调用授权。四个资源均禁止公开申请权限，同时要求应用认证、用户认证和资源调用权限。
  网关管理员只向获准的内部应用授予所需资源及 MCP 的调用权限；不向普通外部应用、`public` 或 `personal` 客户端主动授权。
  已有网关级授权也会覆盖新增资源，上线前须核对已有授权范围，不能仅凭禁止申请认定没有外部应用可访问。
- 调用应用须携带网关支持的有效用户凭证。后端身份认证应使用 `api-gateway`，由网关签名及现有认证链建立用户名和租户上下文；
  不通过请求体指定用户或租户，也不要通过关闭认证或信任任意用户名请求头解决调用失败。
- 用户必须同时拥有 `system_access`（系统访问）和 `platform_manage`（平台管理），后者资源类型为 `platform_resource`、
  资源 ID 为 `scheduler_workflow`（定时任务调度）。只有网关应用权限不能替代用户业务权限，计数和直接 ID 查询同样受限。
  后端须启用并正确同步 IAM V3 或 V4；两者都关闭时现有 authorizer 会跳过业务权限检查，不能用于提供上述隔离保证。
- `system-diagnosis` 的 public/personal OAuth 客户端均默认关闭，不沿用 `topo`、`execution` 的 personal token 开关。
  不能为接入个人客户端而直接复用普通执行诊断 MCP 的授权配置。
- 按现有流程导入并核对资源配置，使用新的 `release.version` 创建资源版本并发布四个只读资源到目标 stage 后，再同步 MCP。
  同步定义中必须保留 `isPublic: false`、
  `allowApplyPermission: false` 以及三项认证和调用权限检查；资源回源路径为 `/api/v3/open/schedule/workflow/...`，
  使用后端用户态认证入口，并复用系统态接口的查询、平台管理权限及租户归属校验。
  不绑定启用、停用、重试、终止或立即执行接口。
- 使用内置模式时，`apigwSync.mcp.enabled=true` 会为未显式配置清单的 stage 装配并启用产品 MCP，包括 `system-diagnosis`；
  产品清单中的 `status: 0` 不构成内置模式下的独立禁用开关。若只需 `topo`、`execution`，应显式完整覆写所需列表。
  显式非空列表不会自动加入 `system-diagnosis`；`[]` 只跳过同步，不会停用已有服务。其他启用及镜像升级要求见第 7 节。

### 环境验收

在目标 stage 核对已发布资源、MCP 实际工具列表及既有应用授权；使用隔离的验证应用和账号检查：

1. 缺少有效应用或用户凭证时，请求在认证阶段被拒绝。
2. 使用有效身份但未授予资源或 MCP 调用权限时，不能取得执行信息。
3. 应用已获授权、用户只有系统访问而没有定时任务调度资源权限时，列表、计数、实例和日志查询均被拒绝。
4. 内部应用和用户均获授权后，可以完成四层查询和按轮次子任务下钻；跨租户或不属于 scheduler 任务链的 ID 不返回执行信息。
5. 实际工具列表只有上述四个查询工具，已有 `topo`、`execution` 的资源及授权设置保持不变。

本地 YAML 或 Schema 检查不能证明上述认证授权已在目标环境生效。部署及同步仍按原流程执行，不要将整个同步 Job 当作只读探测。
