# IAM V4 权限模型迁移文件

本目录存放 bk-nodemgr 的 IAM V4 权限模型模板、渲染工具和迁移工具。

使用流程：配置变量 → 渲染模板 → 预检查 → 执行迁移。支持 `upsert_system`、`upsert_resource_type`、`upsert_action` 和 `upsert_role`：不存在时注册，已存在时有限更新；不自动展开操作依赖，不处理用户组授权、资源范围分配或审批。

## 文件说明

| 文件                | 说明                                                            |
| ------------------- | --------------------------------------------------------------- |
| `templates/`        | V4 System、五类本地 ResourceType、32 个 Action 与四个 Role 模板 |
| `vars.yaml.example` | 变量配置示例                                                    |
| `render/`           | 渲染工具源码，构建后生成 `render/iam-render`                    |
| `migrate/`          | 迁移工具源码，构建后生成 `migrate/iam-migrate`                  |

V4 工具独立维护，不修改 V3 工具。运行时不依赖 `bk-cli`。

## 使用方法

### 1. 构建工具并准备变量

以下步骤在 `support-files/bkiamv4` 目录执行，构建需要项目要求的 Go 环境。

```bash
cd support-files/bkiamv4
make -C render build
make -C migrate build
cp vars.yaml.example vars.yaml
```

按下方「变量配置」填写 `vars.yaml`。示例中的回调地址需替换为实际部署地址，不要将应用密钥写入变量文件或模板。

### 2. 渲染模板

```bash
./render/iam-render -t templates -v vars.yaml -o output
```

渲染 `templates/` 目录下所有 `.tpl` 文件，输出到 `output/`，文件名去掉 `.tpl` 后缀。当前生成 `0001_bk_nodemgr_system.json`、`0002_bk_nodemgr_resource_type.json`、`0003_bk_nodemgr_actions.json` 和 `0004_bk_nodemgr_roles.json`，执行迁移前检查四个文件。

### 3. 预检查

通过本地环境或部署平台注入 `BK_APP_SECRET`，再执行：

```bash
./migrate/iam-migrate \
  --gateway-url "https://bkapi.example.com/api/bkiam/prod/" \
  --app-code "bk-nodemgr" \
  --tenant-id "example-tenant" \
  --dir output \
  --dry-run
```

将网关地址、应用编码和租户 ID 替换为实际值。`--dry-run` 会校验文件并查询远端模型，输出创建、更新或跳过计划，**不会写入远端**。因此预检查也需要网络和有效的应用认证。

### 4. 执行迁移

确认目标环境、租户和预检查结果后，去掉 `--dry-run`：

```bash
./migrate/iam-migrate \
  --gateway-url "https://bkapi.example.com/api/bkiam/prod/" \
  --app-code "bk-nodemgr" \
  --tenant-id "example-tenant" \
  --dir output
```

该命令直接执行，不会再次询问确认。如只执行一个文件，将 `--dir output` 替换为 `--file output/0001_bk_nodemgr_system.json`，两者不能同时使用。

## 变量配置

变量由 `vars.yaml` 提供，示例值不是工具自动补充的默认值。

| 变量路径              | 说明                                                        | 示例值                                                  |
| --------------------- | ----------------------------------------------------------- | ------------------------------------------------------- |
| `system.id`           | IAM 系统标识，不是应用编码                                  | `bk_nodemgr`                                            |
| `system.name`         | 系统显示名称                                                | `BlueKing Node Manager`                                 |
| `system.description`  | 系统说明                                                    | 节点管理系统的功能说明                                  |
| `system.clients`      | 可调用该系统的应用编码数组，必须包含本次调用的 `--app-code` | `["bk-nodemgr"]`                                        |
| `system.callback_url` | IAM 可访问的 V4 资源回调完整地址                            | `https://bk-nodemgr.example.com/api/v3/iam/v4/resource` |

注意区分 `bk_nodemgr`（system ID）和 `bk-nodemgr`（app code）。默认模板不包含可选字段 `managers`；如需设置，应在模板的 `data` 中显式添加字符串数组。

## 迁移参数

| 参数            | 说明                                                             |
| --------------- | ---------------------------------------------------------------- |
| `--gateway-url` | 完整网关入口，包含网关名 `bkiam` 和 stage，如 `/api/bkiam/prod/` |
| `--app-code`    | 调用应用编码，必填                                               |
| `--app-secret`  | 调用应用密钥；未传时读取 `BK_APP_SECRET`，两者不能都为空         |
| `--tenant-id`   | 目标租户 ID，必填；不是环境名称，单次只处理一个租户              |
| `--file`        | 单个已渲染 JSON 文件，与 `--dir` 二选一                          |
| `--dir`         | 已渲染 JSON 文件目录，与 `--file` 二选一                         |
| `--dry-run`     | 只查询和输出计划，不写入                                         |

建议通过 `BK_APP_SECRET` 提供密钥，避免在命令历史和进程参数中留下明文。显式传入的 `--app-secret` 优先，即使传入空字符串也不会回退到环境变量。

## 模板与更新规则

模板使用 Go template 语法，通过 `toJson` 输出 JSON 字符串或数组，避免引号和换行破坏 JSON：

```gotemplate
"name": {{ .system.name | toJson }},
"clients": {{ .system.clients | toJson }}
```

迁移文件通过 `system_id` 指定系统，`operations` 中每项包含 `operation` 和 `data`。`upsert_system` 的 `data.id` 必须与 `system_id` 一致；`upsert_resource_type`、`upsert_action` 的 `data.id` 分别是该系统内的资源类型 ID、操作 ID，每项只处理一个对象。

| 场景                                         | 处理方式                                        |
| -------------------------------------------- | ----------------------------------------------- |
| 查询明确返回 HTTP 404 且错误码为 `NOT_FOUND` | 创建 system，要求提供 `id`、`name` 和 `clients` |
| system 已存在                                | 更新显式提供的可变字段，不发送不可变的 `id`     |
| 更新时未提供某字段                           | 保留远端值，不自动清空                          |
| 显式提供数组                                 | 整体替换，不与远端数组合并                      |
| 显式提供空字符串或空数组                     | 按原值提交，但仍须满足本地校验和 IAM 接口约束   |
| 提供 `clients` 但不包含调用应用              | 本地报错，防止更新后失去管理权限；不会自动追加  |

目录模式仅选择 `<数字>_*.json`，按数字序号排序；序号相同则按文件名排序。每个文件内按 `operations` 顺序执行。dry-run 会将同一 system 前面计划的创建纳入后续判断，不重复计划创建。

### ResourceType 更新规则

`0002` 模板按父资源在先的顺序注册 `biz`、`networkarea`、`networkunit`、`package_type`、`package`，保留 `networkarea → networkunit` 和 `package_type → package` 两条关系。`biz` 是本系统提供的顶层资源，`ancestors` 为 `[]`；业务 ID、含义和使用方式沿用原有约定，IAM 资源提供方由 `bk_cmdb` 改为 `bk_nodemgr`。

`data` 仅接受 `id`、`name`、`ancestors`。ID 最长 32 字符，以小写字母开头，只含小写字母、数字、`_`、`-`；`ancestors` 为从根到直接父级的 ID 数组，不得重复或包含自身。

| 场景                                     | 处理方式                                                                   |
| ---------------------------------------- | -------------------------------------------------------------------------- |
| 资源类型不存在                           | 要求 `name`，通过批量创建接口提交单元素数组；省略 `ancestors` 表示顶层资源 |
| 资源类型已存在，未提供 `ancestors`       | 保留远端祖先链                                                             |
| 显式祖先链与远端不一致，包括用 `[]` 清空 | 报错停止，不自动改拓扑或删除重建                                           |
| 名称改变                                 | 只更新 `name`，不发送 `id` 或 `ancestors`                                  |
| 名称未提供或相同，拓扑一致               | 跳过写入                                                                   |
| 祖先尚未存在、祖先链不一致或名称冲突     | 报错停止；先创建祖先，再创建子资源                                         |

IAM API 允许受限修改祖先链，但会受后代资源类型、角色和已有授权约束；本工具有意不自动执行这类变更。

```mermaid
flowchart TD
    A[读取全部分页资源类型] --> B{资源类型存在?}
    B -- 否 --> C[校验名称和祖先链后计划创建]
    B -- 是 --> D{显式祖先链是否冲突?}
    D -- 是 --> E[报错停止]
    D -- 否 --> F{名称是否变化?}
    F -- 否 --> G[跳过]
    F -- 是 --> H[计划仅更新名称]
    C --> I{dry-run?}
    H --> I
    I -- 是 --> J[只输出计划]
    I -- 否 --> K[执行写入]
```

每个系统的资源类型列表按 `page_size=100` 读取全部分页；失败或不完整响应不会被当作空列表。dry-run 中，前面计划新建的 System 使用虚拟空列表，计划新建的 ResourceType 对后续子资源可见，不请求尚不存在的系统。单独执行 `0002` 时，System 必须已存在。dry-run 不验证服务端全部约束，也不保证后续执行期间远端状态不变。

### Action 更新规则

`0003` 模板保留 V3 的全部 32 个操作 ID 和名称：16 个业务操作绑定本系统的 `biz`，15 个操作保持其他本地资源绑定，`networkarea_create` 的 `resource_type_id` 为 `""`，表示无资源绑定。`networkunit_create` 仍绑定父资源 `networkarea`，`package_type_upload` 仍绑定 `package_type`。

`data` 仅接受 `id`、`name`、`resource_type_id`；ID 格式与 ResourceType 一致，显式提供的 `name` 不得为空或仅含空白，所有字段都不接受 `null`。

| 场景                                 | 处理方式                                                                      |
| ------------------------------------ | ----------------------------------------------------------------------------- |
| 操作不存在                           | 要求 `name`，通过批量创建接口提交单元素数组；省略绑定或提供 `""` 均表示无资源 |
| 新操作绑定非空资源类型               | 资源类型必须在远端或前序计划中存在，不自动创建                                |
| 操作已存在，未提供名称或绑定         | 保留对应远端值                                                                |
| 显式绑定与远端不同，包括用 `""` 清空 | 报错停止，不更新绑定或删除重建                                                |
| 名称改变                             | 只提交 `name`，不发送 `id` 或 `resource_type_id`                              |
| 名称相同且绑定无冲突                 | 跳过写入                                                                      |
| 名称被其他操作占用                   | 报错停止                                                                      |

每个系统的 Action 按 `page_size=100` 读取全部分页，完整性检查通过后再决定写入。创建响应必须是 HTTP 201 且只返回对应 ID，更新必须是 HTTP 204。dry-run 复用前序 System、ResourceType 和 Action 的虚拟状态；实际成功写入也对后续操作可见。单独执行 `0003` 时，System 和所需 ResourceType 必须已存在。

绑定本系统 `biz` 的 16 个操作为：`biz_access`、`agent_view`、`agent_operate`、`agent_history_view`、`proxy_view`、`proxy_operate`、`proxy_history_view`、`plugin_view`、`plugin_operate`、`plugin_history_view`、`config_policy_view`、`config_policy_manage`、`config_policy_history_view`、`deploy_policy_view`、`deploy_policy_manage`、`deploy_policy_history_view`。先执行 `0002` 注册 `biz`，再执行 `0003`；已有 Action 的绑定冲突仍按上表报错，不自动改绑。

本工具仅补齐模型初始化，Provider、资源查询与运行时鉴权链路需另行建设和验证。**模型注册成功不代表业务权限链路已就绪。** V3 模型及已有授权数据不变，不自动迁移 V3 的操作依赖、分组和创建者授权配置。

### Role 更新规则

`0004` 在 Action 之后注册以下四个角色。成员以 V3 CommonActions 为起点；业务节点管理员显式补入 `networkarea_view`，不自动展开 `related_actions`。

| Role ID               | 名称           | Action 数 |
| --------------------- | -------------- | --------- |
| `agent_manager`       | 业务节点管理员 | 9         |
| `networkarea_manager` | 管控区域管理员 | 15        |
| `policy_manager`      | 策略管理员     | 7         |
| `package_manager`     | 资源包管理员   | 4         |

`data` 仅接受 `id`、`name`、`description`、`actions`。每次 upsert 都必须提供完整、非空的 `actions` 数组；每个成员必须显式提供 `id` 和 `resource_type_id`，不允许重复 Action ID、未知字段或 `null`。名称非空，描述允许空字符串。

模板成员保持 Action 当前直接资源绑定，包括 `networkunit_create → networkarea` 和无资源的 `networkarea_create → ""`。工具也支持 IAM 允许的祖先维度；非空维度必须是 Action 绑定或其祖先，所引用的 Action 和 ResourceType 必须在远端或前序计划中存在。

| 场景                                  | 处理方式                                                                             |
| ------------------------------------- | ------------------------------------------------------------------------------------ |
| Role 不存在                           | 要求名称，提交包含成员的单元素数组创建                                               |
| Role 已存在                           | 按 `(Action ID, resource_type_id)` 集合比较，忽略成员顺序                            |
| 保留全部已有成员及其维度，新增 Action | 仅通过追加接口提交缺少的成员，不重复提交已有成员                                     |
| 删除已有成员或改变其维度              | 报错停止，不自动删除、改维度或删除重建                                               |
| 显式名称或描述改变                    | 仅发送发生变化的元数据字段；空描述用于清除描述。若同时新增成员，先追加，再更新元数据 |
| 名称或描述省略                        | 保留远端对应值                                                                       |
| 成员一致且元数据无变化                | 跳过写入                                                                             |
| 名称冲突、Action 缺失或维度不合法     | 报错停止                                                                             |

Role 列表读取全部分页后再决策，异常或不完整响应不会被当作空列表。dry-run 复用前序 System、ResourceType、Action 和 Role 的虚拟状态，不写入远端；成功写入也对后续操作可见。单独执行 `0004` 时，其前置模型必须已存在。Role 注册不包含用户组授权或资源范围分配。

新增 Action 关联的资源类型已在 Role 中时，IAM 会使已有用户在原授权资源范围内获得该 Action 权限；关联新资源类型时，同样追加模型成员，但用户需要另行申请该资源类型的权限。工具不替用户授权或扩大资源范围。

追加成员与更新元数据是两次独立请求。如果追加成功、元数据更新失败，已追加成员不会回滚；检查远端后重跑，只补充仍缺少的成员并更新尚未生效的元数据。dry-run 输出 `add_role_actions` 或 `add_role_actions+update_role`，后续操作可见计划中的新成员。

## 失败处理

- 所有选中文件先完成本地校验，再访问远端。未知字段、未知 operation、`null` 字段值或不一致的 system ID 都会报错。
- 认证失败、超时、异常响应等错误不会被当成「系统不存在」，也不会触发创建。
- 单次 HTTP 请求超时为 30 秒，不跟随重定向，不自动重试写入。
- 任一步骤失败即停止，已成功的操作不会回滚。写入中断或失败后，先检查远端状态，再决定是否重跑。
- 执行计划输出到 stdout，诊断信息输出到 stderr；失败时返回非零退出码。
