---
name: changelog-doc
description: Use when writing versioned bk-nodemgr changelog, release notes, or release.md from tags, diffs, PR facts, issue summaries, or reference templates. Trigger when the user asks to write changelog, 发布说明, 版本变更记录, compare `from-tag...to-tag`, add a `Full Changelog` link, or generate a reusable Node Manager release template.
---

# Changelog 文档编写 Skill

为 bk-nodemgr 生成可直接发布的 changelog / release note Markdown，**全部写入 `release.md`**（前端可能加载整个文件）。

**定位：概要，不是详细变更说明**。详细内容用户应该点击 `Full Changelog` compare 链接自己查看。

目标是让发布人员、运维人员、对接方快速判断：

1. 这次发布影响哪些组件
2. 是否存在升级风险、前置条件或兼容性限制
3. 需要执行哪些升级、灰度、回滚相关动作
4. 哪些变化属于新增、优化、修复或已知问题

## 必须使用本 skill

在以下场景必须使用本 skill：

- 用户要“写 changelog / release note / 发布说明 / 版本变更记录”
- 用户给出版本号、tag 范围、PR facts 或 issue 摘要，希望整理成版本化 Markdown
- 用户要求生成 `release.md`、兼容性说明、已知问题（**所有 changelog 内容都写入 `release.md`，不分散到多个版本文件**）
- 用户要求参考 GitHub Releases 或外部 release note 模板，为 bk-nodemgr 生成统一格式
- 用户明确提到 `Full Changelog` compare 链接、`from-tag...to-tag`、版本间 diff

不要用于：

- 纯 API 参数/reference 文档，优先 `api-doc`
- 纯 PR 描述、提单摘要、合并说明，优先 `create-github-pr`
- 纯 README/流程文档整理且不涉及版本变更，优先 `readme-logic-first`

## 核心原则

- 中文主述，English technical terms 保持稳定：`Agent`、`Plugin`、`Package`、`Backend`、`Control Plane`、`Installer`、`compatibility`
- 面向发布执行者和使用方写，不写研发内部实现流水账
- **changelog 是概要，不是详细变更说明**。用户应该点击 `Full Changelog` compare 链接查看详细内容
- 先写“影响和动作”，再写“变化内容”
- 先基于证据归纳，再输出面向用户的结论
- 结构固定，章节可删减，但不要自由改写顺序
- 宁可标记 `待确认`，也不要编造版本号、日期、tag、兼容性

## Workflow Overview

按以下 6 步执行：

1. **确定范围**：确认目标版本、起止 tag、受影响组件、目标读者、是否要模板还是正文
2. **收集证据**：按优先级读取用户资料、tag diff、现有 release/changelog 文档、PR/issue 摘要、必要实现上下文
3. **归类变更**：收敛到固定分类：新增功能 / 功能优化 / 缺陷修复 / 兼容性 / 已知问题
4. **套用输出契约**：按固定 Markdown 版本块生成，不自由改章节
5. **处理缺口**：无法确认的信息显式标记 `待确认`
6. **审查收尾**：检查 compare link、范围、风险、动作、术语、一致性

## 证据优先级与 fallback

证据优先级：

1. **用户给定资料**：版本号、日期、起止 tag、PR、issue、发布单、外部参考模板
2. **tag diff / compare range**：`from-tag...to-tag` 的 commit 和文件变化，这是 changelog 的核心证据
3. **仓库内模板与既有 release 文档**：`support-files/`、`docs/`、既有 `release.md`
4. **PR / issue / 发布记录**：用于补充语义、确认影响对象
5. **必要实现上下文**：只用于确认影响范围、兼容性；不要把实现细节直接搬进 changelog

Fallback 规则：

- 用户只给目标版本、没给起始版本：先尝试识别上一个发布 tag；无法可靠确认时写 `待确认`
- diff 范围过大：先按组件拆分，再合并成用户可理解的发布结果
- 只有零散事实、没有可靠 diff：输出"草稿"或"模板化整理"，标注为概要，不要假装完整
- raw commit message 只能作为证据，不能直接拼成发布说明
- compare link 的文案与 URL 必须使用同一组 `from-tag...to-tag`

## 输出契约

### 模板模式

当用户只要模板、示例骨架或格式参考时：

- 输出空模板
- 保留固定章节与占位提示
- 不要虚构具体版本事实

### 正文模式（写入 `release.md`）

当用户给了版本事实、diff 范围或发布要求时，默认输出完整版本块。

**必备项**

- 标题：`## [Version: vX.Y.Z] - YYYY-MM-DD`
- 标题下的 `Full Changelog` compare 链接；若 tag 缺失，显式写 `待确认`
- `发布范围`
- 至少一个变化分类：`新增功能` / `功能优化` / `缺陷修复`
- 与发布风险相关的 `compatibility` / `组件依赖与升级前置条件`

**按需保留**

- `重要提示`：有 breaking change、强制升级、额外运维动作、重大兼容性风险时保留
- `已知问题`：确有遗留限制时保留

**删除规则**

- 章节明确“无影响”且用户不要求保留空项时，可删除整节或写“无”，同一文档内保持一致
- 用户明确要求“完整 release.md 风格”时，优先保留完整章节骨架
- 信息未知但该节对发布决策重要时，不删除，改为 `待确认`

## 标准模板

```markdown
## [Version: vX.Y.Z] - YYYY-MM-DD

Full Changelog: [from-tag...to-tag](https://github.com/TencentBlueKing/bk-nodemgr/compare/from-tag...to-tag)

> 重要提示
- [存在 breaking change / 强制升级 / 额外动作时保留；无则删除]

**发布范围**
- Agent:
- Plugin:
- Package:
- Backend/Control Plane:
- Installer/Deployment:

**新增功能**
- [新增的用户/运维能力]

**功能优化**
- [体验、性能、稳定性或可维护性优化]

**缺陷修复**
- [用户可感知或运维可感知的问题修复]

**Agent / Plugin / Package compatibility**
- [版本范围、OS / Arch 限制、包格式兼容性]

**组件依赖与升级前置条件**
- [依赖的 GSE / 制品库 / 平台能力 / 操作系统条件]

**已知问题**
- [问题现象 / 影响范围 / 临时规避方案]

```

## 章节填写规则

### 标题与 compare link

- 标题固定为 `## [Version: vX.Y.Z] - YYYY-MM-DD`
- compare link 默认保留，且放在标题正下方
- link text 和 URL 必须使用同一组 `from-tag...to-tag`
- 不能确认起止 tag 时，不要伪造 link；改写为 `⚠️ **待确认**`

### 发布范围

- 至少判断 `Agent`、`Plugin`、`Package`、`Backend/Control Plane`、`Installer/Deployment`
- 没有影响可写“无”或删除该行，但全文策略保持一致
- 这一节用于回答“这次到底动了什么层”

### 新增 / 优化 / 修复

- **新增功能**：全新能力、支持范围扩展、发布能力新增
- **功能优化**：已有能力的体验、性能、稳定性、兼容性增强
- **缺陷修复**：异常、失败路径、错误行为的修正
- 每条都写发布结果，不写代码实现
- 每条尽量说明影响对象，例如“仅新安装节点生效”“仅 Windows Agent 受影响”

### compatibility / 升级 / 回滚

- 优先回答最低兼容版本、OS / Arch 限制、格式兼容性、依赖关系、是否支持灰度 rollout
- 这些是 changelog 和普通更新摘要的核心差异，不能一笔带过
- 没有可靠信息时用 `待确认`，不要猜

### 已知问题

- 仅在确有遗留限制时保留
- 推荐顺序：问题现象 → 影响范围 → 临时规避方案

## 缺失信息处理

核心原则：宁可保留缺口，也不要补写未经确认的信息。

绝对不要臆造：

- 版本号
- 发布日期
- 起止 tag
- compatibility 结论
- compatibility 结论
- 发布范围细节

使用统一格式：

```markdown
⚠️ **待确认**: [问题描述]
- 来源: [用户输入 / 仓库文档 / PR / issue / diff]
- 需要确认: [版本号 / 起止 tag / compatibility / 发布范围细节]
```

## diff 使用注意事项

- diff 用来确认“发生了什么变化”，不是直接拷贝 commit message
- 先判断哪些变化真正影响发布对象，再写进 changelog
- 对纯重构、纯注释、纯内部整理，默认不要写进面向用户的版本说明，除非它们影响升级、compatibility 或运维动作
- 一个条目可以对应多个 commit，但最终应合并成一条用户可理解的发布结果
- 同一能力横跨多个模块时，优先按“一个发布结果”合并，而不是按文件逐条列出

## 输出风格要求

- **概要优先**：changelog 是概要，不是详细变更说明。详细内容用户应该点击 `Full Changelog` compare 链接查看
- 短句、可执行、少修辞
- 多用 bullet list，少写大段背景
- 先给结论，再给限制
- 每个版本块应能独立阅读
- 用户只要模板时输出空模板；用户给了事实时输出可发布正文
- 术语前后一致，不混用同义词

## Quick Review Checklist

- [ ] 版本号和日期是否明确
- [ ] 是否明确起止 tag，或显式标记 `待确认`
- [ ] 是否使用 tag diff 作为核心证据
- [ ] 是否生成正确的 `Full Changelog` compare 链接
- [ ] 是否写清发布范围
- [ ] 是否区分新增 / 优化 / 修复
- [ ] 是否覆盖 compatibility 与前置条件
- [ ] 是否对不确定信息使用 `待确认`
- [ ] 是否避免实现细节和空泛措辞
- [ ] 是否适合直接放入 `release.md` 或发布公告（概要定位，不是详细变更说明）

## Eval / Acceptance Criteria

这个 skill 至少应通过以下场景：
1. **模板生成**：给定外部参考风格时，能输出可复用的 bk-nodemgr changelog 模板
2. **事实归类**：给定变更事实时，能正确归类为新增 / 优化 / 修复，并补齐发布范围与 compatibility 视角
3. **缺口保留**：信息不全时，保留结构并把未知信息标记为 `待确认`
4. **diff 优先**：给定 `from-tag...to-tag` 时，先基于 diff 归纳，再输出发布结论，而不是复制 raw commits
5. **compare link 正确**：需要 compare link 时，link text 和 URL 使用完全一致的 tag 范围

## 参考资料

- `support-files/bkiamv3/README.md`
- `support-files/bkiamv3/templates/README.md`
- `docs/developer/api_doc_template.md`
- 外部参考：[bk-cmdb release.md](https://github.com/TencentBlueKing/bk-cmdb/blob/master/docs/support-file/changelog/release.md)
