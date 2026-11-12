# 依赖资源与引用

Node Manager 按实际功能选择依赖，不为复现 Demo 同时引入所有资源。MCP 服务由 APIGW 路径维护，其余需要同步的资源加入 Package 清单，再由 Agent 按 code 关联。命名与 Agent 基础约束见 [Package 约束](agent-package.md)。

## APIGW MCP

项目已有拓扑只读 MCP 声明，但声明存在不代表目标环境已发布或授权。项目 APIGW 入口、Chart 覆盖与执行顺序见[集成方案](integration.md#与-apigwmcp-的依赖关系)。

- 接入前创建、发布 APIGW MCP，并完成接入应用及目标 Agent 所需授权；公开 MCP 也不免授权。工具只引用，不创建、不审批、不授予权限。
- `spec.mcps[].type` 固定 `apigw`，`code` 填网关返回的完整准确编码，不填 URL、平台 ID 或仅声明短名。`id` 禁止填写，包括 null。
- 需要固定工具范围时，填 `is_selected_all: false` 和 `selected_tools: [{tool_name: ...}]`，工具名称以实际发布 MCP 为准；不从本地 RPC 名推断。
- 默认全选：省略开关且未选工具或列表为空时为 true，只有非空工具列表时为 false；显式开关优先。全选会包含后续新增工具，不能用于保证长期固定的只读范围。
- 网关 `target_app_codes` 可用于主动授权应用，但具体支持以目标网关版本为准，不把现有 `grant_permissions` 等同于此能力。MCP 查询成功也不证明运行调用授权已验证。

工具以 Agent 的 `metadata.code` 作为 `agent_code` 查询 MCP，不要求它与调用应用编码相同。目标租户和空间来自调用参数：

| 当前租户内的 Agent | 初始化规则                                                                   |
| ------------------ | ---------------------------------------------------------------------------- |
| 尚不存在           | 校验应用对目标空间的权限后查询 MCP；目标平台需支持此规则，不要求先创建 Agent |
| 已存在             | 应用需有 Agent 所在空间权限，该空间须与本次 `--space` 一致                   |

其它租户的同名 Agent 不影响上述存在性判断。查询不到或无权访问时停止，不创建空壳 Agent、不使用其它环境的 ID 绕过。

## 角色 Collection

内联 Prompt 已满足需求时无需独立角色。需要角色复用时，在 `collections/` 维护 Collection，并加入清单：

- `apiVersion: bkai.tencent.com/v1`、`kind: Collection`，metadata 提供 code 和 name，`spec.collection_type` 为 `role`。
- `spec.content` 为有序的非空消息数组，每项使用 `role`、`content`；首项必须为 system，且只能有一个 system，其余为 user / assistant。不合并为单段字符串，不使用角色运行时变量。
- `spec.generate_type` 默认 space；public 强制公开，space 使用 `is_public`。全租户约束见 [Agent 约束](agent-package.md#prompt模型与使用范围)。`tag_names` 为二维字符串数组，每项一条标签路径。
- Agent 用 `prompt: {type: collection, code: ...}` 引用；清单内角色先同步，不在清单内的角色仅引用线上资源。角色不定义默认模型，模型仍由 Agent 配置。
- 更新角色后必须重新同步引用它的 Agent 才刷新正文；切回 inline 时平台清空角色关联和快照。需要对外生效时另行发布。

仅排除角色资源，Agent 仍可引用线上角色并刷新快照；要保护平台手工维护的 Agent，应同时排除 Agent。

## Skill

只有需要脚本或特定执行环境时才交付 Skill。每个 Skill 目录包含 `SKILL.md` 和根目录 `skill.yaml`，后者使用 `kind: Skill`、metadata 的 code/name，以及 spec 的生成范围、可见性与标签；环境变量依赖由 Skill 声明。

### 镜像与版本

Skill 执行镜像不同于运行 `bkai-init` 的初始化镜像。自定义镜像须先进入目标环境仓库，通过模块 Chart 上报并由平台开白；本目录不定义尚未实现的上报 values。

仅支持 Skill 根目录已有 Dockerfile 引用基础镜像，不提供 `spec.base_image`，不自动生成 Dockerfile。如下仅为模板机制示例：

```dockerfile
FROM {{ BKAI_AGENT_IMAGE_REPO | required }}/skill:1.0
```

`BKAI_AGENT_IMAGE_REPO` 通过 `--var` 显式传入且不带末尾 `/`；完整镜像名必须在目标环境有效允许列表内。修改 Skill 文件或 Dockerfile 后重新构建模块资源镜像。

Skill 版本可省略，不以 Package 版本兜底。平台比较 ZIP hash：首次为 `0.0.1`，内容变化递增最新补丁号，未变化覆盖当前版本；取不到旧 hash 时按覆盖处理，新上传 ZIP 读不到则停止。显式版本须为 `x.x.x`，可同版本覆盖；版本快照不存在时平台创建。

### Agent 关联

Agent 的 `spec.skills` 按 code 关联当前 Skill 版本，不填写 version，不要求 SkillVersion 快照存在；目标平台需支持此关联行为。

- `envs` 每项 key 必填、最长 64 字符，同一 Skill 不重复，且必须已在 Skill 依赖中声明。
- value 可省略或为空字符串，保留首尾空白；省略、空串、默认值与 `envs: []` 语义不同，`[]` 不表示清空全部变量，最终需满足依赖声明。
- 仓库仅保存非敏感值；排除 Skill 资源不能阻止 Agent 内的关联和 envs 更新。
- 明确返回“已有版本正在发布”时工具跳过本次更新，其它错误仍停止；跳过不证明新内容已生效。

## 子智能体

只有存在独立职责或指令复用需求时才拆分子 Agent，不要求沿用 Demo 的主子结构。Agent 在 `spec.subagents` 按 code 引用，来源指令在 `conversation_settings.commands` 中用 `agent_code + id` 定位，不填平台 ID。

同包本轮 `--publish` 按先子后主提交，不等待子 Agent 部署，来源指令取本轮已同步配置。包外、未选中或本轮未发布的子 Agent 必须可访问并已有可用发布版本，指令须在该版本存在。首次主子初始化使用整包常规发布，不能假定仅同步草稿就能关联从未发布的来源。

## 知识库

只有需要检索项目文档时才交付知识库。每个目录包含根 `knowledgebase.yaml` 和完整 Markdown/图片树：

- 配置使用 `apiVersion: bkai.tencent.com/v1`、`kind: KnowledgeBase`；metadata 的 code 仅小写字母、数字、下划线，字母开头、最长 64 位，不能使用中划线。未填 code 时用目录名。
- spec 维护生成范围、公开属性、标签和平台支持的 `config`；Agent 的检索策略单独位于 `spec.knowledgebases`，不能混入知识库资源 config。
- Agent 用 `items[].code` 关联。`retriever_code`、`query_function`、`recall_channels`、`rrf_weights`、`document_fragment_count`、`polish`、`origin`、`is_response_when_no_knowledgebase_match`、`rejection_message` 直接放在 knowledgebases 下，统一作用于关联列表，不增加 retrieval 层。
- `query_function` 的 semantic / mixed / sql 须与 dense / sparse 召回通道及平台能力匹配，不代表所有知识库支持 SQL；RRF 权重每项 0–1，不要求总和为 1。是否允许未命中时通识回答需按业务明确，不默认照搬 Demo。

### 文档同步与验收

先同步知识库配置，再按目录相对路径打包 Markdown 和图片，申请临时上传 URL 直传 BKRepo，由平台检查状态、大小及摘要后提交 ZIP 导入。工具取得任务 ID 即继续，不等待导入或向量处理完成。

**含文档时按完整目录镜像同步，删除线上多余内容；无文档时只同步配置，不清空线上文档。** 不用不完整目录更新线上知识库，`show` / `plan` 不比较文档内容。

仅接受 Markdown、图片，排除根目录 `knowledgebase.yaml`、`.DS_Store`、`.git`、`__pycache__`，拒绝符号链接和其它文件类型。ZIP 最大 500 MiB，文件总大小最大 2 GiB，最多 10000 个文件，平台继续校验路径及解压安全。

导入需按任务 ID 确认，无查询入口时交由平台维护方核实；失败、部分成功或未知均不算成功。导入完成后还需验证向量处理、检索与实际问答效果，不能用 CLI 正常退出替代验收。
