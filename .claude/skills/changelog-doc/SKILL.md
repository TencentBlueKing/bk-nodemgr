---
name: changelog-doc
description: Use when writing versioned bk-nodemgr changelog, release notes, or changelog markdown under `support-files/changelog/{zh,en}/` from tags, diffs, PR facts, issue summaries, or reference templates. Trigger when the user asks to write changelog, 发布说明, 版本变更记录, compare `from-tag...to-tag`, add a `Full Changelog` link, generate versioned changelog files, or prepare markdown for the changelog API/page.
---

# Changelog 文档编写 Skill

为 bk-nodemgr 生成可直接发布的 changelog / release note Markdown，默认写入 **`support-files/changelog/{zh,en}/` 下的版本化 Markdown 文件**，供 changelog API 和前端页面直接消费。

`release.md` 仅视为**历史单文件形态**或用户显式指定的兼容输出；默认不要再把它当成唯一目标文件。

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
- 用户要求生成版本化 changelog 文件、兼容性说明、已知问题，或明确提到 `support-files/changelog/zh`、`support-files/changelog/en`、changelog 页面 / changelog API
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
- **证据语言只用于内部判断，最终 changelog 必须是正式发布说明语气**：不要把“当前证据 / 无已知 / No known / current evidence / does not show / identified”这类审查口吻写进正文
- 结构固定，章节可删减，但不要自由改写顺序
- 宁可标记 `待确认`，也不要编造版本号、日期、tag、兼容性
- 输出默认面向 **版本化文件 + API + 前端 Markdown 渲染链路**，避免依赖复杂 HTML 或页面私有能力

## Workflow Overview

按以下 7 步执行：

1. **确定范围**：确认目标版本、起止 tag、受影响组件、目标读者、输出语言（`zh` / `en`）、是模板/正文/版本化文件
2. **收集证据**：按优先级读取用户资料、tag diff、现有 release/changelog 文档、PR/issue 摘要、必要实现上下文
3. **归类变更**：收敛到固定分类：新增功能 / 功能优化 / 缺陷修复 / 兼容性 / 已知问题
4. **确定承载形态**：明确是单个版本文件、双语文件、还是用户显式要求的 `release.md` 兼容输出
5. **套用输出契约**：按固定 Markdown 版本块生成，不自由改章节
6. **处理缺口**：无法确认的信息显式标记 `待确认`
7. **审查收尾**：检查 compare link、范围、风险、动作、术语、一致性、文件命名与 Markdown 兼容性

## 证据优先级与 fallback

证据优先级：

1. **用户给定资料**：版本号、日期、起止 tag、PR、issue、发布单、外部参考模板
2. **tag diff / compare range**：`from-tag...to-tag` 的 commit 和文件变化，这是 changelog 的核心证据
3. **仓库内模板与既有 changelog 文档**：`support-files/changelog/`、`support-files/`、`docs/`、历史 `release.md`
4. **PR / issue / 发布记录**：用于补充语义、确认影响对象
5. **必要实现上下文**：只用于确认影响范围、兼容性；不要把实现细节直接搬进 changelog

Fallback 规则：

- 用户只给目标版本、没给起始版本：先尝试识别上一个发布 tag；无法可靠确认时写 `待确认`
- diff 范围过大：先按组件拆分，再合并成用户可理解的发布结果
- 只有零散事实、没有可靠 diff：输出"草稿"或"模板化整理"，标注为概要，不要假装完整
- raw commit message 只能作为证据，不能直接拼成发布说明
- compare link 的文案与 URL 必须使用同一组 `from-tag...to-tag`
- 需要落盘为版本化文件时：优先使用 `support-files/changelog/{locale}/<version>_<date>.md`；无法确认文件名时显式标记 `待确认`

## 输出契约

### 模板模式

当用户只要模板、示例骨架或格式参考时：

- 输出空模板
- 保留固定章节与占位提示
- 不要虚构具体版本事实

### 正文模式（写入版本化 changelog 文件）

当用户给了版本事实、diff 范围或发布要求时，默认输出完整版本块，并优先落到版本化 changelog 文件。

**默认目标位置**

- 中文：`support-files/changelog/zh/<version>_<date>.md`
- 英文：`support-files/changelog/en/<version>_<date>.md`
- `<version>` 使用版本标识（例如 `v3.0.1-alpha.17`）
- `<date>` 使用 `YYYY-MM-DD`
- 仅当用户明确要求兼容旧流程时，才输出到单个 `release.md`

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

## 文件布局与命名规则

- 默认按语言拆分目录：`support-files/changelog/zh/` 与 `support-files/changelog/en/`
- 默认一个版本对应一个 Markdown 文件，不要把多个版本持续堆叠到同一个 `release.md`
- 文件名格式优先：`<version>_<YYYY-MM-DD>.md`
- 文件名中的 version/date 同时是 API 列表与详情的重要元数据来源；不确定时不要猜，改写成 `待确认`
- 如果用户只要正文不要求落盘，正文内容仍应保持可直接放入上述版本文件的形态

## API / 前端消费兼容性

- 输出应视为 **raw markdown 内容**，可被 changelog API 返回并由前端页面直接渲染
- 优先使用标准 Markdown 结构：标题、blockquote、bullet list、普通链接
- 避免依赖复杂 HTML、自定义脚本、内联样式或仅在特定渲染器下有效的写法
- compare link、章节标题、列表层级应保持稳定，避免给前端渲染和 XSS 清洗带来歧义

## 章节填写规则

### 标题与 compare link

- 标题固定为 `## [Version: vX.Y.Z] - YYYY-MM-DD`
- compare link 默认保留，且放在标题正下方
- link text 和 URL 必须使用同一组 `from-tag...to-tag`
- 不能确认起止 tag 时，不要伪造 link；改写为 `⚠️ **待确认**`
- 若输出需要落盘，确保标题内容与文件名中的 version/date 不冲突

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
- commit 可以作为候选证据来源，但最终必须先做去噪与归并，再写成面向发布对象的结果
- `feat/fix` 不等于一定要逐条入 changelog；仍要判断是否影响发布范围、compatibility、升级动作或用户可感知结果
- 对 merge / revert / cherry-pick / 纯整理提交，要优先回到实际发布结果，而不是照着提交粒度输出
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
- 禁止审查/分析口吻进入正文：
  - 中文避免：`当前证据`、`依据当前变更证据`、`无已知`、`未发现`、`未显示`
  - English avoid: `No known`、`Based on current evidence`、`Current evidence does not show`、`is identified`
  - 改写为发布说明语气：`本版本不涉及 ... 变更`、`... 保持兼容`、`本版本不引入新增组件依赖`、`This version does not change ...`、`... remain compatible`、`This version does not introduce ...`
- 默认让正文可直接被 Markdown 渲染器消费，不额外依赖人工二次整理

## Quick Review Checklist

- [ ] 版本号和日期是否明确
- [ ] 是否明确起止 tag，或显式标记 `待确认`
- [ ] 是否使用 tag diff 作为核心证据
- [ ] 是否生成正确的 `Full Changelog` compare 链接
- [ ] 是否确认输出目标是 `support-files/changelog/{zh,en}/` 版本化文件，或用户显式要求的旧式 `release.md`
- [ ] 文件名是否符合 `<version>_<YYYY-MM-DD>.md` 约定，或显式标记 `待确认`
- [ ] 是否写清发布范围
- [ ] 是否区分新增 / 优化 / 修复
- [ ] 是否覆盖 compatibility 与前置条件
- [ ] 是否对不确定信息使用 `待确认`
- [ ] 是否避免实现细节和空泛措辞
- [ ] 是否清理“当前证据 / 无已知 / No known / current evidence”等审查口吻，并改为正式发布说明语气
- [ ] 是否适合直接被 changelog API / 前端页面消费（概要定位，不是详细变更说明）

## Eval / Acceptance Criteria

这个 skill 至少应通过以下场景：
1. **模板生成**：给定外部参考风格时，能输出可复用的 bk-nodemgr changelog 模板
2. **事实归类**：给定变更事实时，能正确归类为新增 / 优化 / 修复，并补齐发布范围与 compatibility 视角
3. **缺口保留**：信息不全时，保留结构并把未知信息标记为 `待确认`
4. **diff 优先**：给定 `from-tag...to-tag` 时，先基于 diff 归纳，再输出发布结论，而不是复制 raw commits
5. **compare link 正确**：需要 compare link 时，link text 和 URL 使用完全一致的 tag 范围
6. **版本文件正确**：需要落盘时，优先输出到 `support-files/changelog/{zh,en}/<version>_<date>.md` 形态，而不是默认回退到 `release.md`
7. **Markdown 可消费**：输出能被 changelog API 返回并由前端页面直接渲染，不依赖复杂 HTML

## 参考资料

- `support-files/bkiamv3/README.md`
- `support-files/bkiamv3/templates/README.md`
- `docs/developer/api_doc_template.md`
- 外部参考：[bk-cmdb release.md](https://github.com/TencentBlueKing/bk-cmdb/blob/master/docs/support-file/changelog/release.md)
