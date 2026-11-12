# Package 与 Agent 约束

本文保留 Node Manager 维护 Agent Package 时需要的输入约束。下面是资源结构示例，不是已交付的项目 Package；示例 code 不代表已注册的线上资源。构建与部署方式见[集成方案](integration.md)，MCP 等依赖见[依赖资源与引用](resources.md)。

## 最小资源组织

仅使用内联 Prompt 时，一个清单和一个 Agent 文件即可表达 Package 结构；实际 MCP 关联需按目标环境补齐。角色、Skill、知识库和子 Agent 按业务需要增加，不要求复制完整 Demo。

```text
package/
├── bkai.yaml
└── agents/
    └── nodemgr.yaml
```

`bkai.yaml` 中路径相对清单所在目录：

```yaml
apiVersion: bkai.tencent.com/v1
kind: AgentPackage
metadata:
  name: bk-nodemgr
  version: "1.0.0"
resources:
  - agents/nodemgr.yaml
```

`agents/nodemgr.yaml` 展示 Agent 的基本结构，不包含环境凭据或平台资源 ID：

```yaml
apiVersion: bkai.tencent.com/aidev-agent/v1
kind: Agent
metadata:
  code: ai-nodemgr
  name: 节点管理助手
  description: 协助理解节点管理信息
spec:
  prompt:
    type: inline
    content: |
      你是节点管理助手，帮助用户理解拓扑、网络与主机查询结果。
      仅解释已获取的信息，不编造查询结果，不承诺执行未提供的操作。
  model:
    llm_code: aidev-chat-auto
    context_window: 16
  tools: []
```

模型需在目标平台可用。该示例未关联 MCP，不代表已经具备查询能力；关联时使用网关返回的完整 code 和实际工具列表，不凭短名猜测。

## code 与资源定位

code 是创建、更新和关联的稳定标识，不是展示名称，全部使用小写；`name` 可以使用中文。

| 资源               | 完整匹配正则                | 约束                                   |
| ------------------ | --------------------------- | -------------------------------------- |
| Agent / 子 Agent   | `^ai-[a-z][a-z0-9-]{1,12}$` | `ai-` 开头，总长 5–16 位，不支持下划线 |
| 角色 / MCP / Skill | `^[a-z][a-z0-9_-]{0,63}$`   | 字母开头，最长 64 位                   |
| 知识库             | `^[a-z][a-z0-9_]{0,63}$`    | 字母开头，最长 64 位，不支持中划线     |

- 定义、关联、`commands[].agent_code` 及 CLI 选择/排除均按相应规则校验，变量替换后再校验；不自动转小写或迁移旧编码。
- YAML 的 kind 为 `Agent` / `Collection` / `Skill` / `KnowledgeBase`，CLI 类型为小写。文件名不代替 code，外部模型编码、指令 ID、组件 key 和环境变量名不适用本表。
- 资源文件不指定空间，统一由调用方传入租户和目标空间；资源不存在或无法唯一定位时停止，不迁移、恢复或覆盖冲突资源。
- Agent 禁止配置 `version`（包括 null），发布版本由平台分配；Package 的 `metadata.version` 不作为 Agent 发布版本，也不作为 Skill 版本兜底。

## Prompt、模型与使用范围

| 字段                     | 维护约束                                                                                                                                  |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `metadata.code` / `name` | 必填；名称由平台校验                                                                                                                      |
| `spec.prompt`            | `inline` 只填 `type`、`content`；`collection` 只填 `type`、`code`。两种模式互斥，不填平台 ID                                              |
| `spec.model.llm_code`    | 协议默认 `aidev-chat-auto`，省略由工具补齐；目标模型仍需实际可用                                                                          |
| `model.context_window`   | 会话轮数，只接受 1–30 的整数，不是 token 数或 K 单位                                                                                      |
| 其它模型限制             | `temperature` 为 0–2；`llm_token_limit` 最小 1024；`max_tokens` 为 1–20480；`tool_output_compress_thrd` 最小 1024，并满足模型自身能力限制 |
| `spec.user_scope`        | 可选 `public` / `custom`，不接受 null；省略时新建遵循平台默认，已有保持原范围。主子 Agent 分别配置                                        |
| `spec.admins`            | 禁止，包括 null 或空数组；管理员只能通过 `--admins` 追加                                                                                  |
| `spec.tools`             | 暂不支持，保持 `[]`；APIGW MCP 通过 `spec.mcps` 关联                                                                                      |

`public` 不免除入口鉴权、租户隔离和依赖授权；`custom` 的人员名单通过平台管理。使用范围修改会影响已有发布版本，不依赖本次 `--publish`，同步前应明确变更影响；设置失败或回读不一致时停止。

如需全租户资源，`spec.generate_type: global` 仅允许调用方空间为 `system-bkaidev`；平台还要求多租户模式及 `system` 租户，创建后不能切换全租户范围。Agent 兼容 `metadata.generate_type` 且优先于 spec，不能用资源内空间字段绕过调用方校验；不要从“模块统一交付”推断应创建全租户资源。

## 对话指令与来源引用

需要快捷指令时，在 `spec.conversation_settings.commands` 中维护，不因示例而默认加入。

- 本地指令使用 `id`、`name`、`icon`、`content`、`components`；无图标显式填 `icon: null`。正文使用 `{{ key }}` 引用组件，不用 `prompt` 字段；同一 Agent 的指令 ID 唯一。
- 组件支持 `text`、`textarea`、`number`、`select`，key 在同一指令内唯一；默认值须匹配类型、数值边界和选项，不配置 `multiple`。
- 回填由指令的 `enable_fill_back`、`fill_back_component_key`、`fill_regx` 控制；启用时目标 key 必须存在，不在组件内重复维护平台回填字段。
- 来源指令只填 `id`、`agent_code`、可选 `alias`，来源必须在 `subagents` 中；正文、组件及上传设置由工具展开。需要改正文或组件时定义本地指令，不修改引用简写。
- 同包本轮发布时读取已同步来源配置，其它情况读取来源最新发布版本。来源不存在、无权访问、缺少发布版本或指令时停止，不以空配置替代。
- 平台默认指令无需手工复制；`support_upload` 的 `vision` / `file` 声明不证明执行链路支持上传。

部署变量与运行时组件占位符共享相似语法，不要重名；替换白名单及未传变量行为见[变量替换](execution.md#变量替换)。
