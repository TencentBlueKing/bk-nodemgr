# 运行参数与执行

本文用于维护 Node Manager 同步任务的工具调用。目标版本为 `bkai-init==0.1.6`，由平台基础镜像提供；不要求开发者安装 Python 包或运行上游 Demo。项目配置与凭据的接入边界见[集成方案](integration.md)。

## 连接与运行参数

以下是工具参数，不是已实现的 Node Manager Helm values。

| 配置                                      | 接入要求                                                                                                                              |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `BK_API_URL_TMPL`                         | 例如 `https://gateway.example.com/api/{api_name}/{stage}`，CLI 将 `api_name` 分别解析为 `bk-aidev`、`bk-user`；没有 stage 时补 `prod` |
| `BKAI_BASE_URL` / `--base-url`            | 可替代模板，填 AIDEV 网关完整环境入口（含 stage）；命令行优先，未提供才使用模板。专用域名不按通用模板猜测                             |
| `BKAI_APP_CODE` / `BKAI_APP_SECRET`       | 必填应用身份；未配置时读取 `BKPAAS_APP_ID` / `BKPAAS_APP_SECRET`，命令参数优先。不自动读取 APIGW 的 `BK_APP_CODE` / `BK_APP_SECRET`   |
| `BKPAAS_MULTI_TENANT_MODE`                | 默认 `false`，多租户显式设 `true`                                                                                                     |
| `BKAI_TENANT_ID` / `--tenant-id`          | 多租户下传应用实际租户，写入 `X-Bk-Tenant-Id`；未设默认 `system`，不自动推断，单租户忽略                                              |
| `BKAI_SPACE_ID` / `--space`               | 显式传完整目标 ID；默认 `bkaidev`，不拼接租户前缀。命令行优先且不接受空值；未传命令行参数时读取环境变量，空值回退默认空间             |
| `BKAI_USERNAME` / `--username`            | 显式值必须是实际 `bk_username`，原样透传。未传时默认 `bk_admin`，仅多租户下查询转换                                                   |
| `BK_USER_BASE_URL` / `--bk-user-base-url` | 仅多租户且未提供执行用户名时需要，可由模板推导；默认用户查询失败或结果不唯一时停止，应用需有相应 lookup 权限                          |
| `BKAI_ACCESS_TOKEN` / `--access-token`    | 可选，通常无需配置；兼容 `ACCESS_TOKEN`，不能替代应用、空间或 MCP 授权                                                                |
| `-f`                                      | Package 清单路径，镜像机制示例为 `/app/package/bkai.yaml`                                                                             |
| `--resource` / `--exclude-resource`       | 可重复，格式为 `agent/code`、`collection/code`、`skill/code` 或 `knowledgebase/code`；排除优先于选择。`show` 不支持排除参数           |
| `--admins`                                | 空格分隔的实际 `bk_username`，作用于所有选中的主子 Agent；只追加，不删除已有管理员，YAML 禁止 `spec.admins`                           |
| `--confirm`                               | `sync` 写入确认；不传只预览，不上传、不写入、不发布                                                                                   |
| `--publish` / `--publish_config_only`     | 显式开启 Agent 发布；后者为 `0` 或 `1`，默认 `0`，须同时传 `--publish`。`0` 请求常规发布到开发者中心，`1` 仅发布配置                  |
| `--var KEY=VALUE`                         | 可重复传部署变量，不自动读取环境变量；限制见[变量替换](#变量替换)                                                                     |

## 校验、预览与写入

以下命令在具备目标工具版本及完整 Package 的环境中执行，用于说明部署任务应调用的 CLI。先通过部署凭据机制注入应用身份、地址、租户及执行用户；示例不设置密钥。

```bash
: "${BKAI_SPACE_ID:?请设置完整目标空间 ID}"
: "${BKAI_APP_CODE:?请注入应用编码}"
: "${BKAI_APP_SECRET:?请注入应用密钥}"

bkai-init validate -f /app/package/bkai.yaml --space "$BKAI_SPACE_ID"
bkai-init show -f /app/package/bkai.yaml --space "$BKAI_SPACE_ID"
bkai-init plan -f /app/package/bkai.yaml --space "$BKAI_SPACE_ID" --publish
```

`validate` 校验本地配置；`show` 查看线上配置，不查询权限；`plan` 只检查和预览，不保证后续写权限或服务端执行成功。`plan --format json` / `--format yaml` 可输出结构化计划。

确认预览与覆盖范围后，常规发布调用如下；仅同步草稿则同时从预览和写入命令移除发布参数：

```bash
bkai-init sync -f /app/package/bkai.yaml --space "$BKAI_SPACE_ID" \
  --confirm --publish --publish_config_only=0
```

资源选择仅接受清单内对象，未选中的依赖不会自动同步，所有本地配置仍须通过校验。MCP 不是独立同步资源。排除平台手工维护的 Agent 时，对 `plan` 和 `sync` 使用相同排除列表。

## 发布与管理员

- 不传 `--publish` 只同步草稿；常规发布请求显式传递 `is_publish_to_paas_v3=true`，平台可能按代码版本一致策略复用部署。
- 工具只调用发布接口，不轮询、不等待部署完成。`pending` / `running` 表示已提交；`successful` 按接口结果记录，实际部署和调用效果仍需验收。
- 同包本轮发布按先子后主提交，来源指令使用本轮已同步配置。包外或未选中的子 Agent 需有可用发布版本；首次主子初始化不应仅同步主 Agent 草稿。
- 仅明确返回 HTTP 400 且原因是“智能体状态异常，无法发布，当前状态为：publishing”时，记录并继续，不计为失败。Skill 明确返回已有版本正在发布时也跳过本次更新；其它错误仍停止。
- `--admins` 仅支持插件智能体，在配置写入后、发布前追加；调用方负责将 `login_name` 转为实际 `bk_username`，工具不转换管理员。
- 管理员追加失败或结果未确认时，工具保留原因并继续发布和后续同步，不能仅凭退出状态认定追加成功。`plan` 仅展示追加意图，不校验用户映射，不把无法回读的名单判为一致。

## 更新覆盖保护

- **配置完整**：Agent 同步可能重建关联，省略字段不等于保留线上原值。
- **按资源排除**：排除整个 Agent 才能保护其手工配置。仅排除 Skill 不阻止 Agent 内 Skill 关联和 envs 更新；仅排除角色不阻止 Agent 刷新线上角色快照。
- **环境变量语义**：Skill 的 `envs` 省略、`[]`、缺少 `value`、空字符串不同；工具不将缺少 value 转为空串，`[]` 不表示清空全部变量，最终需满足 Skill 依赖声明。
- **引用更新**：角色更新后重新同步引用它的 Agent；Agent 关联 Skill 使用当前版本。资源版本号不等于 Agent 发布版本。
- **知识库目录完整**：有文档时镜像同步会删除线上多余内容；无文档时只同步配置，不清空线上文档。`show` / `plan` 不比较文档内容。
- **失败不回滚**：先查线上状态再重跑，不能假设 Job 失败时所有事项均未写入。

## 变量替换

仅支持 `{{ KEY }}`、`{{ KEY | default("默认值") }}`、`{{ KEY | required }}`。变量由任务显式传入 `--var`，替换结果不再次渲染。

| 文件范围                                                            | 处理方式                                                    |
| ------------------------------------------------------------------- | ----------------------------------------------------------- |
| `bkai.yaml`、清单引用的 Agent / Collection YAML                     | 解析后仅替换字符串值，不替换键、结构或类型                  |
| 已声明 Skill 根目录 `skill.yaml`、知识库根目录 `knowledgebase.yaml` | 同上，不扫描任意 YAML                                       |
| 已声明 Skill 根目录已有 Dockerfile                                  | 替换文本并写入上传包，不修改本地文件，不自动创建 Dockerfile |

- 变量名为 1–64 位英文、数字或下划线，不以数字开头；重复 key 报错，按第一个等号分隔 key/value。
- `default` 仅接受单个字符串参数，只在未传变量时生效，显式空字符串保持为空。`required` 不接受参数，未传、空字符串或纯空白均报错。
- 不支持表达式、属性访问、函数调用、其它过滤器、循环或条件。
- YAML 独立占位符必须加引号。未传且无默认值的 YAML 占位符保留，供运行时组件引用；部署变量不要与组件 key 重名。
- Dockerfile 不允许遗留占位符，替换值不能含换行或 NUL；变量来源须可信，不通过变量传递密钥。
- `SKILL.md`（含 frontmatter）、脚本、知识库文档、普通 YAML、嵌套 Dockerfile 不处理；镜像构建和允许列表仍由平台校验。

## 故障定位

保留失败资源、配置路径、步骤、执行时间、CLI 版本、Trace ID 与服务端请求标识，不粘贴原始凭据。查询请求遇到连接失败、超时、HTTP 429 或 5xx 时按 1、2、4 秒重试，最多请求 4 次；写入及其它业务错误不会盲目重试。

| 现象                                   | 优先检查                                                                                    |
| -------------------------------------- | ------------------------------------------------------------------------------------------- |
| `CROSS_TENANT_FORBIDDEN` / BK User 403 | 应用实际租户、多租户开关；不要直接套用 APIGW Job 的租户变量                                 |
| `bk_username is not verified`          | 失败步骤和实际下传的 `bk_username`；追加管理员不能替代身份校验                              |
| `API_NOT_FOUND` / `1640401`            | 网关地址、`api_name=bk-aidev`、stage 与路由发布；不等于 MCP 不存在                          |
| Agent 不在当前空间、资源不存在或冲突   | 租户、完整空间 ID、稳定 code、资源归属或软删除占用；不换用其它环境 ID 绕过                  |
| 首次 MCP 查询失败                      | MCP 发布、应用/Agent 授权、空间权限及平台初始化查询能力；不创建空壳 Agent 绕过              |
| `INITIALIZE_APP_MEMBERS_ERROR`         | 将 IAM `create_group` 路径、HTTP 状态和 request_id 提供给平台维护方，不按本地 YAML 错误处理 |
| Skill 镜像不在允许列表                 | 渲染后的完整镜像名、目标环境有效允许列表；改 Dockerfile 后重新构建模块镜像                  |
| 关联 Skill 版本不存在                  | Skill 是否存在可访问，以及平台是否支持不依赖 SkillVersion 快照的关联；不要补写关联 version  |
| 知识导入或发布已提交                   | 按任务 ID 或发布记录核对后台结果；CLI 正常退出不等于完成                                    |

不支持的配置应先核对 CLI 和平台能力，不通过删除重要配置后覆盖既有 Agent 来绕过。同步与运行的分层验收见[集成方案](integration.md#升级保护与验收)。
