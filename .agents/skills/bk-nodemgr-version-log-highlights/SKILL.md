---
name: bk-nodemgr-version-log-highlights
description: Use when writing or reviewing bk-nodemgr version highlights (版本亮点) from full changelogs under support-files/changelog/{zh,en}, including user-visible changes in the UI and installation workflows, or when providing Chinese and English highlights templates.
---

# bk-nodemgr Version Highlights

基于完整版本日志编写或审阅节点管理的版本亮点，也可单独提供中英文模板。亮点正文供版本日志弹窗默认展示，聚焦操作者在页面或安装器中能实际遇到的变化，说明适用场景，以及所需操作、步骤或可见结果。

## Source Anchors

- 生成或审查具体版本的亮点时，按用户给出的版本号或文件名定位 `support-files/changelog/<locale>/<version>_<YYYY-MM-DD>.md`。通读新增功能、功能优化、缺陷修复及其他章节中的适用条件。缺少完整日志时，使用 `bk-nodemgr-changelog-doc` 补齐。
- 以完整日志作为发布事实的依据。按目标版本核对实际的页面入口和调用、安装器步骤、操作文档及用户可见结果。变更是否随目标版本发布，应依据该版本的发布记录或变更范围。
- 使用 `bk-nodemgr-audience-boundary` 核对读者范围。读者是实际使用节点管理的人，包括安装、升级、配置或管理节点和插件的操作者。

## Workflow

1. 逐条确认读者能通过页面或安装器使用变更，或能在这些流程中看到变更结果。仅通过命令行操作、仅通过 API 调用、仅供内部使用的变化，以及无法核实使用影响的条目留在完整日志。
2. 写清读者会遇到的变化，包括所需操作、步骤和顺序、默认行为、任务状态或结果。配置项、默认配置来源和接口细节仅在读者需要据此操作或判断适用条件时写入。读者只会看到提示、链接或任务结果时，写清该可见变化。
3. 按使用场景或操作归并新增、优化、修复中的相关变化，并按用户影响排序。性能和稳定性变化以已证实的使用结果为准。
4. 编写或审查亮点正文时，沿用 `bk-nodemgr-changelog-doc` 的「核心原则」和「输出风格要求」中关于语言、文风和术语的规则。用简洁、正式的发布说明语气描述已核实的变化，先写用户影响和所需操作，再给出适用条件。按 `docs/concepts/glossary.md` 核对中英文领域术语，涉及界面名称时核对 `front/locales/{zh-CN,en-US}.yml` 中的实际文案。

## 输出契约

### 模板模式

- 用户只要模板、骨架或格式参考时，直接返回所需语言的标准模板。未指定语言时提供中英文两份。
- 保留标题和条目占位提示。模板无需指定版本，只在对话或指定的草稿位置交付。

### 正文模式（写入版本化摘要文件）

**默认目标位置**

- 中文 `support-files/changelog/highlights/zh/<version>_<YYYY-MM-DD>.md`
- 英文 `support-files/changelog/highlights/en/<version>_<YYYY-MM-DD>.md`
- 语言目录名及文件名与完整日志一致。中英文原文都有时生成双语摘要，用户指定单语时只生成该语言。用户只要草稿时在对话或指定的草稿位置交付。

**必备项**

- 中文标题用 `## 版本亮点`，英文标题用 `## Highlights`。有亮点条目时，在标题下使用单层 Markdown 列表，并用已核实的版本事实替换所有占位提示。
- 每条写清使用场景和变化后的操作、步骤、顺序或结果。有适用条件时一并说明。中英文的变更范围和适用条件应一致。
- 摘要文件供默认展示，完整版本日志及其 `Full Changelog` 链接供展开后查看。页面渲染和展开交互由前端负责。

**空摘要与冲突处理**

- 核实完整日志后，若没有符合范围的亮点条目，仍按正文模式交付相应语言的版本化摘要文件。保留标题，空一行后写入对应的一段正文。
  - 中文 `本版本的变更详见完整版本日志。`
  - English `See the full release notes for changes in this version.`
- 发布事实尚未核实时，向编写者说明待核实的事项，待核实后再判断是否使用空摘要。同版本已有摘要且与原文冲突时，向编写者指出差异。

## 标准模板

### 中文

```markdown
## 版本亮点

- [在什么使用场景下，操作、步骤、顺序或结果有什么变化。如有适用条件，一并说明。]
```

### English

```markdown
## Highlights

- [Name the usage scenario and describe what changed in the action, steps, order, or outcome. Include any applicable conditions.]
```

## Quick Review Checklist

- [ ] 每条亮点都能追溯到目标版本，并有已核实的读者可达入口或可见结果。正文只保留读者理解变化、执行操作或判断适用条件所需的信息。
- [ ] 文风符合完整日志的发布说明规范，中英文术语与项目 glossary 一致，界面名称与对应语言的实际文案一致。
- [ ] 摘要文件名与完整日志一致，完整日志内容保持原样。

## Evals

`evals/evals.json` 覆盖页面与安装器入口、可见结果与配置细节的取舍、默认选项、安装顺序、升级配置、仅限命令行的离线迁移、双语模板及空摘要处理。
