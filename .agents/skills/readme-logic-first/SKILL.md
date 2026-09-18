---
name: readme-logic-first
description: Use when writing or refactoring README documentation that must be logic-first (decision flow and rules), Chinese-main with English technical terms, and free of fluff. Trigger whenever the user asks to "写 README", "整理文档逻辑", "做流程图", "去废话", or wants policy/scope/error-handling behavior explained clearly.
---

# README Logic-First Skill

## Overview

本 skill 用于把“代码说明文档”改写为“逻辑说明文档”。

目标不是解释实现细节，而是让读者快速回答：

1. 这个模块在解决什么问题？
2. 核心 decision flow 是什么？
3. 每个分支在什么条件下触发？
4. 最终会返回什么语义？

默认写作风格：**中文主述 + 英文 technical terms**（如 `scope`, `intersection`, `policy`, `PermissionDenied`）。

## When to Use

出现以下任一情况必须使用本 skill：

- 用户要求“写 README / 重写 README / 文档化逻辑”
- 用户强调“不要讲代码细节，要讲逻辑”
- 用户要求“流程图 / 四层模型 / 决策树”
- 用户要求“去掉废话、保留高信息密度内容”
- 用户要求“按 AGENTS.md 约束写文档”

不适用：

- 纯 API 参数罗列（更适合 API doc skill）
- 只改一两句文案，不涉及逻辑结构

## Mandatory Workflow

1. **Policy Alignment**
   - 先读取相关 `AGENTS.md` 的语言和文档约束。
   - 若与用户临时要求冲突，先确认优先级，再写。

2. **Logic Extraction**
   - 从代码/需求中抽取：状态、输入、分支、输出。
   - 禁止先按文件结构写目录。

3. **Model First**
   - 先给逻辑模型，再写正文。
   - 推荐固定模型：`Permission -> Scope -> Intersection -> Endpoint Policy`。

4. **Flowchart**
   - 为关键 decision flow 提供 `mermaid` 图。
   - 图中必须覆盖：入口条件、关键分支、输出分支。

5. **Fluff Removal Pass**
   - 删除无信息密度句子（例如“本文将介绍…”、“不展开细节…”）。
   - 每段都要有可执行信息（规则/分支/策略/边界）。

6. **Terminology Pass**
   - 中文句子里统一英文术语，不混用同义词。
   - 同一概念全篇同名（例如一直用 `scope`，不要交替写“范围权限/授权域/作用域”）。

## Output Structure (Default)

README 推荐结构：

1. 模块目标（解决什么问题）
2. 核心语义对象（例如 `AuthorizedScope`）
3. 决策规则（分支条件 + 结果）
4. 流程图（`mermaid`）
5. 返回策略（`empty` / `PermissionDenied` / success）
6. 边界与例外（如 no-op mode）

## Hard Rules

- 先写“逻辑”，后写“文件索引”。
- 不得把实现细节（函数调用序列）当作逻辑主线。
- 不得省略冲突策略（例如“无授权实例”到底返回什么）。
- 不得出现空话段落。
- 任何术语改变必须全篇一致。

## Evidence-Backed Writing Principles

在输出 README 时，额外遵循以下原则（来自 open-source 和主流 style guide）：

1. 先回答读者四个问题：What / Why / How to start / Where to get help。
2. 始终以 audience 视角写，不默认读者和作者背景一致。
3. 优先 clarity 与 consistency，不追求“写得花”。
4. 用短句和直接表达，减少抽象空话。
5. 结构按阅读路径组织：overview → quick start → usage/policy → exceptions。
6. 文档目标是降低沟通成本：减少重复提问，减少歧义。

可参考来源：

- https://opensource.guide/starting-a-project/
- https://learn.microsoft.com/en-us/style-guide/welcome/
- https://developers.google.com/style

## Logic Compression Pattern

把复杂行为压成这四句：

1. 先判定能不能做（Permission）
2. 再判定能做到多大范围（Scope）
3. 再把请求范围与授权范围收敛（Intersection）
4. 最后按 endpoint 语义返回（Policy）

## Common Mistakes

1. **Mistake**: 文档按源码文件展开，读者找不到结论。
   - **Fix**: 先给逻辑模型，再映射到文件。

2. **Mistake**: “无交集”与“无授权实例”语义混写。
   - **Fix**: 单独列两条规则，明确不同返回策略。

3. **Mistake**: 中英混乱，术语漂移。
   - **Fix**: 开头定义术语表，后文严格复用。

4. **Mistake**: 流程图与正文不一致。
   - **Fix**: 先锁正文规则，再按规则画图。

## Quick Review Checklist

- [ ] 是“逻辑文档”而非“代码导览”
- [ ] 有明确 decision rules（不是描述性段落）
- [ ] 有流程图并覆盖关键分支
- [ ] 中文主述 + 英文术语一致
- [ ] 删除所有低信息密度句子
- [ ] “返回策略”清晰且可执行
