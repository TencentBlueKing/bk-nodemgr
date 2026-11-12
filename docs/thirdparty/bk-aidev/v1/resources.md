# 依赖资源与引用

本文说明角色、APIGW MCP、Skill、子智能体和知识库的定义、准备与关联规则。先维护依赖资源，再通过 Agent 的 `spec` 引用；片段合并到同一 `spec` 下，不重复声明或覆盖其它配置。

命名和清单见 [Package 与 Agent 协议](agent-package.md)，同步与发布行为见[运行参数与执行](execution.md)。Tools 暂不支持，Manifest 中保留 `tools: []`。

## 角色 Collection

使用角色引用时，在 `collections/` 定义 Collection 并加入 `bkai.yaml.resources`，先创建或更新角色，再同步 Agent。

`collections/job_role.yaml`：

```yaml
apiVersion: bkai.tencent.com/v1
kind: Collection
metadata:
  code: bk_job_role
  name: 蓝鲸作业专家
spec:
  collection_type: role
  generate_type: space
  is_public: false
  tag_names: []
  content:
    - role: system
      content: |
        你是蓝鲸作业专家，仅回答与作业平台相关的问题。
    - role: user
      content: 如何查询作业执行状态？
    - role: assistant
      content: 我可以帮助你查询作业执行状态，请提供作业 ID。
```

| 字段                               | 类型或约束                             | 说明                                                                                                            |
| ---------------------------------- | -------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `apiVersion` / `kind`              | `bkai.tencent.com/v1` / `Collection`   | 加入清单                                                                                                        |
| `metadata.code` / `name`           | 必填 string                            | 重复同步使用 code 定位已有角色                                                                                  |
| `spec.icon`                        | 可选 string / null，最长 1024 字符     | 省略遵循平台默认或更新规则                                                                                      |
| `spec.collection_type`             | 仅 `role`，默认 `role`                 | 只管理角色资源                                                                                                  |
| `spec.content`                     | 非空消息数组                           | 每项 `role`、`content`，保持顺序，不合并为单段字符串                                                            |
| `spec.content[].role`              | `system` / `user` / `assistant`        | 首条必须为 system，且只能有一条 system                                                                          |
| `spec.content[].content`           | 非空 string                            | 固定消息，不支持角色运行时变量；部署文件变量见[变量替换](execution.md#变量替换)                                 |
| `spec.generate_type` / `is_public` | `space` / `public` / `global`；boolean | 默认选 space；public 强制公开，space 使用 is_public；global 约束见[资源生成范围](agent-package.md#资源生成范围) |
| `spec.tag_names`                   | 二维 string 数组，默认 `[]`            | 每项为一条标签路径                                                                                              |

角色不定义变量和默认模型，模型统一在 Agent 中配置。code 小写、字母开头、最长 64 位，允许下划线和中划线。

### Agent 角色引用

用以下片段替换内联 Prompt，其它配置不变：

```yaml
spec:
  prompt:
    type: collection
    code: bk_job_role
```

- 清单内 Collection 先同步，再解析 `prompt.code`；不在清单中的角色仅引用已有资源。
- 只填角色 code，不填平台 ID，不支持引用含模板变量的角色。
- collection 模式只填 `type`、`code`；inline 模式只填 `type`、`content`。切换 inline 时平台清空角色关联与快照。
- 更新角色后需重新同步引用它的 Agent，已有 Agent 不会自动使用新正文；对外生效还需另行发布。

### 定位与更新保护

在目标租户、空间按 code 定位，不存在创建，存在更新。其它空间同 code、软删除占用等冲突会报错，不擅自迁移、恢复或覆盖。

通过 `validate` / `sync` / `show` / `plan` 管理，重点检查消息正文与顺序、角色类型、标签、公开配置；无法回读的字段不能视为一致。

`--exclude-resource collection/bk_job_role` 只跳过角色写入，Agent 仍可引用线上角色；不希望刷新 Agent 快照时需同时排除 Agent。角色同步失败先修正配置或权限，再重试依赖它的 Agent，不假设已有成功事项会回滚。

## APIGW MCP

接入方先创建、发布 APIGW MCP Server，并完成接入应用及目标智能体所需授权，再在 `spec.mcps` 中引用。公开和非公开 MCP 都要提前授权；公开可见、已发布不等于允许初始化工具或目标 Agent 使用。

```yaml
spec:
  mcps:
    - type: apigw
      code: bk-job-apigw-mcp
      is_selected_all: false
      selected_tools:
        - tool_name: get_job_status
```

| 字段                     | 类型或约束        | 说明                                                                                                               |
| ------------------------ | ----------------- | ------------------------------------------------------------------------------------------------------------------ |
| `mcps[].type` / `code`   | `apigw` / string  | 只支持 APIGW，填网关返回的完整准确 MCP Server code，不填地址或平台 ID                                              |
| `mcps[].is_selected_all` | boolean，默认全选 | 显式 true / false 优先；省略且未选工具或为空列表时为 true，只提供非空 selected_tools 时为 false；plan 与 sync 一致 |
| `mcps[].selected_tools`  | 对象数组          | 每项 `tool_name: string`；全选时建议省略或 `[]`，固定范围时填写 false 与工具清单                                   |
| `mcps[].id`              | 禁止              | 包括 null，传入即报错，不接受环境相关 ID                                                                           |

默认全选只填 `type`、`code` 即可，也可显式 `is_selected_all: true`。全选包含 MCP 后续新增工具，需要限制范围时显式 false 并列出工具。

code 须小写、字母开头、最长 64 位，允许下划线和中划线。先完成网关授权及平台应用权限配置再 `plan` / `sync`；查询不到或权限不足时停止，不绕过授权。工具不创建 MCP、不自动申请或授予权限。

### 初始化与空间校验

MCP 创建、更新和主动授权由接入方完成，不属于 Package 同步。网关同步配置的 `target_app_codes` 可主动授权应用，`spec.mcps[].code` 使用网关返回的完整编码；支持项以目标网关版本为准。

初始化工具使用 Agent 的 `metadata.code` 作为 `agent_code` 查询 MCP，目标租户与空间来自 `--tenant-id`、`--space`，无需新增配置开关，不要求 Agent code 与调用应用 code 相同。

| 当前租户的 Agent 状态 | 接入规则                                                                  |
| --------------------- | ------------------------------------------------------------------------- |
| 尚未创建              | 校验调用应用对目标空间的权限后，按 APIGW 数据查询 MCP，不要求先创建 Agent |
| 已经创建              | 应用需有 Agent 所在空间权限，且该空间必须与 `--space` 相同，不匹配则停止  |

是否创建只按当前租户判断，其它租户同名 Agent 不影响初始化。目标平台需支持此规则；查询成功不代表运行调用已验证，授权要求不变。

## Skill

先通过清单同步 Skill，再在 Agent 的 `spec.skills` 中关联。

`skills/job_skill/skill.yaml`：

```yaml
apiVersion: bkai.tencent.com/v1
kind: Skill
metadata:
  code: job_skill
  name: 作业查询 Skill
  description: 查询作业执行信息
spec:
  generate_type: space
  is_public: false
  tag_names: []
```

目录必须包含 `SKILL.md`，内容与环境变量依赖声明由接入方维护。code 小写、字母开头、最长 64 位，允许下划线和中划线，Agent 引用相同 code。

### 版本与镜像

版本可省略，不以 Package 版本兜底。平台按 ZIP hash 决定版本：

- 首次创建为 `0.0.1`；内容变化时在最新版本上加补丁号，未变化则覆盖当前版本。
- 无法取得旧 ZIP hash 时按版本覆盖处理；无法读取新上传 ZIP 时停止。
- 显式版本须为 `x.x.x`，支持同版本覆盖，版本快照不存在时由平台创建。

仅在 Skill 根目录已有 Dockerfile 引用镜像，不提供 `spec.base_image`，不自动生成 Dockerfile：

```dockerfile
FROM {{ BKAI_AGENT_IMAGE_REPO | required }}/skill:1.0
```

自定义镜像先同步到环境仓库，由模块 Helm Chart 上报、平台开白。核对渲染后的完整镜像名与目标环境有效允许列表；仓库变量不带末尾 `/`，替换规则见[变量替换](execution.md#变量替换)。

### Agent 关联与环境变量

```yaml
spec:
  skills:
    - code: job_skill
      envs:
        - key: BK_JOB_MODULE
          value: bk-job
```

| 字段                    | 类型或约束                | 说明                                                                                          |
| ----------------------- | ------------------------- | --------------------------------------------------------------------------------------------- |
| `skills[].code`         | 必填 string               | Skill 已同步且在目标租户、空间可访问                                                          |
| `skills[].envs`         | 可选对象数组，可为 `[]`   | 每项 key 与可选 value                                                                         |
| `skills[].envs[].key`   | 必填 string，最长 64 字符 | 同一 Skill 下不可重复，须在 Skill 环境变量依赖中声明                                          |
| `skills[].envs[].value` | 可选 string，允许空字符串 | 保留首尾空白；省略不被 CLI 转为空字符串                                                       |
| `skills[].version`      | 不纳入关联输入            | 不在 Agent YAML 中指定；只要求 Skill 存在可访问，不要求 SkillVersion 快照；平台需部署对应修复 |

示例 `BK_JOB_MODULE` 须先在 Skill 依赖中声明，仓库只保存非敏感值。`envs` 省略、`[]`、缺少 value、空字符串语义不同；`[]` 不代表清空全部变量，最终取值须满足依赖声明。Agent 关联使用当前 Skill 版本；仅排除 Skill 资源不能阻止 Agent 内关联或 envs 更新。

## 子智能体

`spec.subagents` 填子智能体 code，规则与主智能体相同：`ai-` 开头，总长 5–16 位。同包本轮使用 `--publish` 时先同步子智能体、提交发布，再继续主智能体，不等待部署；来源指令使用本轮已同步配置。包外或本轮未发布的来源须可访问且有可用发布版本，指令须存在于该版本。

`agents/bk_job_exec.yaml`：

```yaml
apiVersion: bkai.tencent.com/aidev-agent/v1
kind: Agent
metadata:
  code: ai-job-exec
  name: 作业执行助手
spec:
  prompt:
    type: inline
    content: 你是作业执行助手，帮助用户理解任务信息。
  model:
    llm_code: aidev-chat-auto
    context_window: 16
  conversation_settings:
    commands:
      - id: execute_job
        name: 分析作业执行
        icon: null
        content: 请分析任务 {{ task_id }} 的执行信息。
        components:
          - type: text
            key: task_id
            name: 任务 ID
            required: true
  tools: []
```

主 Agent 的关联与可选来源指令引用：

```yaml
spec:
  subagents:
    - code: ai-job-exec
  conversation_settings:
    commands:
      - id: execute_job
        agent_code: ai-job-exec
        alias: 作业执行助手
```

来源需有 ID 为 `execute_job` 的指令；仅关联子 Agent 时不必添加 commands 项。

| 字段               | 说明                                                           |
| ------------------ | -------------------------------------------------------------- |
| `subagents[].code` | 必填 string，名称、平台 ID 和关联信息由工具获取                |
| 来源 `commands[]`  | 工具读取 `id`、`name`、`agent_id`，`status` 可选，无需重复填写 |

首次推荐整包 `sync --confirm --publish --publish_config_only=0`；单独同步主 Agent 时，包外或未选中的子 Agent 仍需已发布版本。指令展开与失败规则见[引用与展开规则](agent-package.md#引用与展开规则)。

## 知识库

每个知识库独立目录，维护根配置、Markdown 和图片。code 小写英文字母开头，仅小写英文、数字、下划线，最长 64 位，不支持中划线。

`knowledgebases/bk_job_docs/knowledgebase.yaml`：

```yaml
apiVersion: bkai.tencent.com/v1
kind: KnowledgeBase
metadata:
  code: bk_job_docs
  name: 作业平台知识库
  description: 作业平台使用文档
spec:
  generate_type: space
  is_public: false
  config: {}
  tag_names: []
```

`metadata` 定义标识、名称和描述；`spec` 定义可见性、知识库 config 和标签，具体 config 按目标平台能力填写，不与 Agent 检索配置混用。未指定 code 时使用目录名。

### Agent 检索配置

```yaml
spec:
  knowledgebases:
    items:
      - code: bk_job_docs
    retriever_code: ""
    query_function: mixed
    recall_channels: [dense, sparse]
    rrf_weights:
      dense: 0.7
      sparse: 0.3
    document_fragment_count: 20
    polish: true
    origin: false
    is_response_when_no_knowledgebase_match: true
    rejection_message: 未找到相关作业文档，请补充问题信息。
```

字段直接位于 `knowledgebases` 下，`items` 是关联列表，其余配置作用于全部关联知识库，不增加 `retrieval` 层。

| 字段                                      | 类型或约束                   | 说明                                                               |
| ----------------------------------------- | ---------------------------- | ------------------------------------------------------------------ |
| `items[].code`                            | string                       | 已同步知识库编码，遵循知识库命名规则                               |
| `retriever_code`                          | 可选 string                  | 检索器编码                                                         |
| `query_function`                          | `semantic` / `mixed` / `sql` | 兼容检索方式，不代表所有知识库支持 SQL，须与召回通道一致           |
| `recall_channels`                         | `dense` / `sparse` 数组      | 与 query_function 一致                                             |
| `rrf_weights`                             | string → number 对象         | 每项 0–1，示例为 dense、sparse，不要求总和为 1                     |
| `document_fragment_count`                 | integer                      | 返回片段数，示例 20                                                |
| `polish` / `origin`                       | boolean / boolean            | 结果处理开关，首次按示例填写，再结合目标平台版本的具体返回效果调整 |
| `is_response_when_no_knowledgebase_match` | boolean                      | 未命中时是否允许根据通识回答                                       |
| `rejection_message`                       | string，可为空               | 拒答文案                                                           |

建议显式填写需要的检索配置，更新前看差异，不通过省略字段假设保留线上原值。

### 文档 ZIP 同步

1. 同步知识库配置，打包目录内 Markdown 与图片，保留相对路径。
2. 向平台申请临时上传 URL，直接上传 BKRepo，再由平台核对上传状态、大小和摘要。
3. 提交 ZIP 导入，取得任务 ID 后继续后续资源，不等待后台完成。“已提交，后台处理中”不代表导入成功。
4. 按任务 ID 在平台确认导入状态，核对目录和文档；无查询入口时交由平台维护方核实。失败、部分成功、结果未知均不算成功，CLI 正常退出不能替代验收。
5. 单独检查向量处理及问答、检索效果，导入完成不代表这些步骤已通过。

包含文档时按整个知识库根目录镜像同步，**删除线上多余内容**；本地目录必须完整。没有文档时只同步配置，不清空线上文档。`show` / `plan` 不比较文档内容。

仅接受 Markdown、图片，排除根目录 `knowledgebase.yaml`、`.DS_Store`、`.git`、`__pycache__`；拒绝符号链接与其它文件类型。ZIP 最大 **500 MiB**，文件总大小最大 **2 GiB**，最多 **10000** 个文件；平台继续校验路径与解压安全。
