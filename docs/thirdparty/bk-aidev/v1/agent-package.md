# Package 与 Agent 协议

Agent Package 通过 `bkai.yaml` 声明同步清单，由独立 Agent、Collection 配置和 Skill、知识库目录组成。本文对应 `bkai-init==0.1.6`；依赖资源定义见[依赖资源与引用](resources.md)。

## 资源 code 命名

code 是创建、更新和关联的稳定标识，不是展示名称。所有资源 code 使用小写，`name` 可以使用中文。

| 资源               | 规则                                                                                           | 完整匹配正则                | 示例                                           |
| ------------------ | ---------------------------------------------------------------------------------------------- | --------------------------- | ---------------------------------------------- |
| Agent / 子智能体   | `ai-` 开头，前缀后首位为小写英文字母，其余仅小写英文、数字、中划线；总长 5–16 位，不支持下划线 | `^ai-[a-z][a-z0-9-]{1,12}$` | `ai-job`、`ai-job-exec`                        |
| 角色 / MCP / Skill | 小写英文字母开头，其余仅小写英文、数字、下划线、中划线；1–64 位                                | `^[a-z][a-z0-9_-]{0,63}$`   | `bk_job_role`、`bk-job-apigw-mcp`、`job_skill` |
| 知识库             | 小写英文字母开头，其余仅小写英文、数字、下划线；1–64 位，不支持中划线                          | `^[a-z][a-z0-9_]{0,63}$`    | `bk_job_docs`                                  |

错误示例：Agent 的 `job-agent`、`ai-Job`、`ai-job_exec`、`ai-a`；角色等资源的 `Job_role`、`1job`；知识库的 `bk-job-docs`、`Job_docs`。知识库 code 直接用于 Milvus 集合名，Milvus 不支持中划线。

资源定义、关联引用、`commands[].agent_code`、CLI 筛选与黑名单均按对应规则校验，变量替换后再校验。工具不兼容不符合规则的旧 code，不自动转小写、重命名或迁移；平台保留旧数据不代表 CLI 接受旧编码。

模型编码、外部应用编码、指令 ID、组件 key、环境变量名不适用本表，例如 `BK_JOB_MODULE` 可以大写。YAML 的 `kind` 使用 `Agent` / `Collection` / `Skill` / `KnowledgeBase`；CLI 类型使用小写，如 `agent/ai-job`。文件名不代替资源 code。

## 目录结构与同步清单

Demo 目录可以完整复制；未使用的依赖无需创建目录，凭据不放入接入目录或 Git。Demo 镜像内 `/app/.env` 通过运行时挂载：

```text
demo/
├── .env
├── .env.example
├── Makefile
├── Dockerfile
└── package/
    ├── bkai.yaml
    ├── agents/
    ├── skills/
    └── knowledgebases/

/app/
├── .env
└── package/
    └── bkai.yaml
```

使用角色时另建 `collections/`。每个知识库目录通过根目录 `knowledgebase.yaml` 定义 code 和配置，未指定 code 时使用目录名；Markdown 和图片按目录层级打包。

清单路径相对 `bkai.yaml` 所在目录，仅声明本次需要同步的资源：

```yaml
apiVersion: bkai.tencent.com/v1
kind: AgentPackage
metadata:
  name: bk-job
  version: "1.0.0"
resources:
  - collections/job_role.yaml
  - skills/job_skill
  - knowledgebases/bk_job_docs
  - agents/bk_job_exec.yaml
  - agents/bk_job_ai.yaml
```

仅使用内联 Prompt 时可移除角色资源项。Package 的 `metadata.version` 不用于 Agent 发布版本，也不作为 Skill 版本兜底。

## Agent 完整示例

`agents/bk_job_ai.yaml`：

```yaml
# AIDEV Agent CLI Manifest v1
apiVersion: bkai.tencent.com/aidev-agent/v1
kind: Agent

metadata:
  code: ai-job
  name: 蓝鲸作业智能助手
  description: 面向蓝鲸作业平台的智能助手

spec:
  user_scope: public
  prompt:
    type: inline
    content: |
      ## Role

      蓝鲸作业智能助手

      ## Constraints

      - 仅回答与蓝鲸作业平台相关的问题。

  model:
    llm_code: aidev-chat-auto
    fallback_model: null
    temperature: 0.7
    context_window: 16
    llm_token_limit: 28000
    max_tokens: 4096
    tool_output_compress_thrd: 4096

  conversation_settings:
    opening_remark: 你好，我可以帮助你查询作业执行情况。
    predefined_questions:
      - 如何查询作业执行状态？
    enable_chat_session: true
    enable_word_selection_popup: true
    commands:
      - id: query_job_status
        name: 查询作业状态
        icon: null
        agent_code: ai-job
        alias: null
        content: |
          查询作业 {{ job_id }} 在 {{ environment }} 环境中的执行状态。
          最多返回 {{ limit }} 条记录。
          补充信息：{{ context }}
        components:
          - type: text
            key: job_id
            name: 作业 ID
            placeholder: 请输入作业 ID
            default: ""
            required: true
            hide: false
          - type: select
            key: environment
            name: 环境
            placeholder: 请选择环境
            default: test
            required: true
            hide: false
            options:
              - label: 测试
                value: test
              - label: 生产
                value: production
          - type: number
            key: limit
            name: 返回数量
            default: 10
            min: 1
            max: 100
            required: false
            hide: false
          - type: textarea
            key: context
            name: 补充信息
            placeholder: 输入或引用需要分析的文本
            default: ""
            rows: 4
            required: false
            hide: false
        enable_fill_back: true
        fill_back_component_key: context
        fill_regx: null
        support_upload:
          vision: false
          file: false

      # Reference a command by its source agent_code and id.
      - id: execute_job
        agent_code: ai-job-exec
        alias: 作业执行助手

  subagents:
    - code: ai-job-exec

  mcps:
    - type: apigw
      code: bk-job-apigw-mcp
      is_selected_all: false
      selected_tools:
        - tool_name: get_job_status

  skills:
    - code: job_skill
      envs:
        - key: BK_JOB_MODULE
          value: bk-job

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

  tools: []
```

先替换编码、名称、Prompt，再按实际依赖调整关联。专题示例是配置片段，应合并到同一 `spec` 下，不重复声明 `spec` 或覆盖其它配置；角色引用可替换示例中的内联 Prompt。

## 基础字段与模型

| 字段                              | 类型或约束                         | 说明                                                                                                |
| --------------------------------- | ---------------------------------- | --------------------------------------------------------------------------------------------------- |
| `apiVersion` / `kind`             | 固定字符串                         | `bkai.tencent.com/aidev-agent/v1` 与 `Agent`                                                        |
| `metadata.code`                   | 必填 string                        | 按 [code 规则](#资源-code-命名)定位智能体                                                           |
| `metadata.name` / `description`   | string                             | 名称必填，描述可选；名称由平台校验                                                                  |
| `spec.user_scope`                 | 可选 `public` / `custom`           | 所有人或指定人使用；不接受 null 或其它值，省略时新建沿用平台默认，已有 Agent 保持原范围             |
| `--admins`                        | 命令参数                           | 所选主子 Agent 的实际 `bk_username`；不传则不处理管理员。YAML 禁止 `spec.admins`，包括 null、空列表 |
| `prompt.type`                     | `inline` / `collection`            | 内联正文与角色引用互斥                                                                              |
| `prompt.content`                  | string                             | `inline` 必填正文，`collection` 不填写                                                              |
| `prompt.code`                     | string                             | `collection` 必填角色 `metadata.code`，不填平台 ID，`inline` 不填写                                 |
| `model.llm_code`                  | string，协议默认 `aidev-chat-auto` | 省略时由工具补齐协议默认值后提交平台                                                                |
| `model.fallback_model`            | string / null                      | null 表示不指定备用模型                                                                             |
| `model.temperature`               | number / null                      | 平台范围 0–2                                                                                        |
| `model.context_window`            | integer / null                     | 会话轮数，仅接受 1–30 的整数；不是 token 数或 K 单位，256、字符串 `"16"`、小数、布尔值均报错        |
| `model.llm_token_limit`           | integer / null                     | 上下文 token 上限，平台最小值 1024                                                                  |
| `model.max_tokens`                | integer / null                     | 回复 token 上限，平台范围 1–20480                                                                   |
| `model.tool_output_compress_thrd` | integer / null                     | 工具输出压缩阈值，平台最小值 1024                                                                   |

模型还需满足目标模型自身能力限制。资源文件不指定 `space`，由调用方统一传入。Agent 禁止配置 `version`（包括 null），发布版本由平台分配。

### 人员使用范围

```yaml
spec:
  user_scope: public
```

主、子 Agent 分别配置，不继承。`custom` 人员名单通过平台管理；`public` 不免除入口鉴权、租户隔离或依赖授权，也不等于资源公开属性。

使用范围修改会影响已有发布版本，不依赖本次 `--publish`。先 `plan` 再同步；CLI 和平台需支持该字段，设置失败或回读不一致时停止，不静默忽略。管理员追加的身份、失败处理与发布顺序见[管理员追加](execution.md#管理员追加)。

### 资源生成范围

Agent、Collection、Skill、KnowledgeBase 支持 `spec.generate_type: global`，表示全租户资源。`validate`、`plan`、`sync` 发送资源请求前校验调用方空间，仅 `system-bkaidev` 可配置；`metadata.space` 不能覆盖调用方空间。Agent 兼容 `metadata.generate_type`，优先于 `spec`。

平台仍要求多租户模式和 `system` 租户，创建后不能切换全租户范围。资源文件不应通过指定空间绕过调用方空间校验。

## 对话设置

在 `spec.conversation_settings` 配置开场白、预设问题和快捷指令。需要用户输入参数时，在 `commands` 定义 `components`，正文通过 `{{ key }}` 引用组件值。

| 字段                          | 类型或约束           | 说明                                         |
| ----------------------------- | -------------------- | -------------------------------------------- |
| `opening_remark`              | string，默认空字符串 | 最长 1024 字符                               |
| `predefined_questions`        | string[]，默认 `[]`  | 预设问题                                     |
| `commands`                    | Command[]，默认 `[]` | 本地指令或来源指令引用，展开后须满足指令结构 |
| `enable_chat_session`         | boolean，默认 true   | 启用多会话                                   |
| `enable_word_selection_popup` | boolean，默认 false  | 启用划词弹出菜单                             |

## 指令与组件

### commands[]

| 字段                      | 类型或约束                  | 说明                                                                       |
| ------------------------- | --------------------------- | -------------------------------------------------------------------------- |
| `id`                      | 必填 string                 | 指令标识，引用时填来源指令 ID；同一 Agent 内唯一，避免与默认指令 ID 冲突   |
| `name`                    | 必填 string                 | 本地指令必填，来源简写由 CLI 补齐                                          |
| `icon`                    | string / null，平台无默认值 | 完整提交须提供，无图标显式填 null                                          |
| `agent_code`              | string / null               | 引用时填来源 Agent code，本地指令省略表示当前 Agent                        |
| `alias`                   | string / null，默认 null    | 目标 Agent 的展示别名                                                      |
| `content`                 | string / null，默认 null    | 正文以 `{{ key }}` 引用组件，不是 `prompt` 字段                            |
| `components`              | Component[]，默认 `[]`      | 四类组件见下表                                                             |
| `enable_fill_back`        | boolean，默认 false         | 启用引用文本回填                                                           |
| `fill_back_component_key` | string / null               | 启用回填时必填，须匹配组件 key                                             |
| `fill_regx`               | string / null，默认 null    | 引用文本匹配正则，保留平台字段拼写 `regx`                                  |
| `support_upload`          | 对象，默认 `{}`             | 仅 `vision`、`file` 两键，值为 boolean；声明开关不代表执行链路必然支持上传 |

### components[]

| 字段                   | 类型或约束                               | 说明                                                                                     |
| ---------------------- | ---------------------------------------- | ---------------------------------------------------------------------------------------- |
| `type`                 | 必填 string                              | `text`、`textarea`、`number`、`select`                                                   |
| `key`                  | 必填 string                              | 同一指令内唯一的组件引用键                                                               |
| `name` / `placeholder` | string / null，默认 null                 | 展示名称与输入提示                                                                       |
| `default`              | string / integer / boolean / list / null | 按组件类型填写，select 默认值须匹配 options                                              |
| `required`             | boolean / null，默认 false               | 建议显式布尔值                                                                           |
| `hide`                 | boolean，默认 false                      | 是否隐藏                                                                                 |
| `rows`                 | integer / null，默认 null                | textarea 行数                                                                            |
| `min` / `max`          | integer / null，默认 null                | 数值边界，平台校验整数默认值处于范围内                                                   |
| `options`              | 对象数组 / null，默认 null               | select 选项 `{label: string, value: string或number}`；label 为展示文本，value 为实际取值 |

只填写支持的组件字段，不配置 `multiple`。文本回填统一在指令层设置，无需在组件内重复配置平台维护字段。

### 引用与展开规则

- **组件引用**：`content` 使用 `{{ key }}`；CLI 校验 key 唯一、引用存在、默认值与类型及选项匹配、数值范围有效。
- **本地指令**：填写完整定义，`agent_code` 可省略，不手工填平台 ID。
- **来源指令**：以 `agent_code + id` 定位，来源 Agent 须在 `subagents` 中。同包本轮发布时读取已同步配置，其它情况读取来源最新发布版本；复制名称、图标、正文、组件、回填与上传设置，允许本地覆盖 `alias`。
- **来源失败**：不存在、无权访问、缺少本轮同步配置或可用发布版本、指令删除、code 无法唯一定位时终止，不以空配置替代；简写展开由 CLI 负责。
- **资源定位**：引用资源须在指定租户与目标空间内可访问，不存在或无法唯一定位时修正资源或空间，不换用其它环境 ID。
- **回填**：`enable_fill_back=true` 时目标 key 须存在；平台将目标组件标为 `fill_back` 并同步 `fill_regx`，关闭时重置组件回填字段。
- **默认指令**：平台可能自动生成当前 Agent 默认指令，无需复制到接入文件手工维护。

本地定义与来源引用共用 `commands` 数组，不新增 `reference` 层。来源引用只接受 `id`、`agent_code`、`alias`；需修改正文或组件时定义新的本地指令。资源引用只填 code，不填平台资源 ID；指令 `id` 是指令标识，仍需保留。
