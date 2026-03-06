---
name: openspec-analyze-change
description: Use when the user wants to review, question, or validate an OpenSpec proposal before implementation. Use when artifacts exist but have not been confirmed, or when the user says "analyze", "review proposal", or "check before implementing".
---

Interactive multi-round Q&A to analyze and confirm an OpenSpec proposal before implementation.

**Core principle:** Analyze in phases, confirm each phase, never dump everything at once.

**Input**: Optionally specify a change name. If omitted, prompt for selection from available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes. Use the **AskUserQuestion tool** to let the user select.

   **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

2. **Load artifacts**
   ```bash
   openspec status --change "<name>" --json
   openspec instructions apply --change "<name>" --json
   ```
   Read all available artifacts (`proposal.md`, `design.md`, `tasks.md`, delta specs).

3. **Summarize in 3-5 bullet points**

   Present a concise summary of the proposal. Do NOT ask questions yet — just confirm you understood what the change is about.

4. **Run phases sequentially — each phase is a separate round**

   ```dot
   digraph analyze_flow {
     "Start" [shape=doublecircle];
     "Phase: Scope" [shape=box];
     "User confirmed scope?" [shape=diamond];
     "Phase: Requirements" [shape=box];
     "User confirmed requirements?" [shape=diamond];
     "Phase: Design" [shape=box];
     "User confirmed design?" [shape=diamond];
     "Phase: Tasks" [shape=box];
     "User confirmed tasks?" [shape=diamond];
     "All confirmed" [shape=doublecircle];

     "Start" -> "Phase: Scope";
     "Phase: Scope" -> "User confirmed scope?" ;
     "User confirmed scope?" -> "Phase: Requirements" [label="yes"];
     "User confirmed scope?" -> "Phase: Scope" [label="no, iterate"];
     "Phase: Requirements" -> "User confirmed requirements?";
     "User confirmed requirements?" -> "Phase: Design" [label="yes"];
     "User confirmed requirements?" -> "Phase: Requirements" [label="no, iterate"];
     "Phase: Design" -> "User confirmed design?";
     "User confirmed design?" -> "Phase: Tasks" [label="yes"];
     "User confirmed design?" -> "Phase: Design" [label="no, iterate"];
     "Phase: Tasks" -> "User confirmed tasks?";
     "User confirmed tasks?" -> "All confirmed" [label="yes"];
     "User confirmed tasks?" -> "Phase: Tasks" [label="no, iterate"];
   }
   ```

   **Phase details:**

   | Phase | Focus | Example Questions |
   |-------|-------|-------------------|
   | **Scope** | Problem boundary, impact, success criteria | "X 的影响范围包括哪些服务？" "成功标准是什么？" |
   | **Requirements** | Scenario coverage, edge cases, backward compat | "当 X 为空时如何处理？" "是否向后兼容？" |
   | **Design** | Technical approach, risks, trade-offs | "为什么选方案 A 而非 B？" "风险如何缓解？" |
   | **Tasks** | Order, dependencies, verification criteria | "任务 3 是否依赖任务 1？" "如何验证完成？" |

5. **Update artifacts based on confirmed decisions**

   After all phases, update proposal/design/tasks files with any changes that emerged from the analysis.

6. **Final validation**

   If `openspec validate` is available, run `openspec validate <name> --strict`. Otherwise, summarize manually. Report what was confirmed and any remaining open items.

**Graceful Degradation**

Not all changes have all artifacts. Adapt phases to what exists:

| Available Artifacts | Phases to Run |
|---------------------|---------------|
| Only proposal | Scope only, then recommend creating design/tasks |
| Proposal + design | Scope → Design, note tasks phase skipped |
| Proposal + design + tasks | All 4 phases |
| Proposal + design + tasks + specs | All 4 phases with spec cross-referencing |

**Question Types**

| Type | Purpose | Example |
|------|---------|---------|
| Clarify | Ambiguous terms or scope | "X 具体指什么？" |
| Boundary | Edge cases and limits | "当 X 为空/无效时如何处理？" |
| Integration | Interaction with existing code | "这与现有的 Y 功能如何交互？" |
| Verify | Confirmation criteria | "如何确认 X 正常工作？" |

**Tracking**

Maintain a running tracker across rounds. Display at each phase transition:

```
## 分析进度
- [x] 范围确认: 影响仅限 pkg/rest/server, 无 breaking change
- [x] 需求验证: 覆盖空值保护、wrapped error、omitempty
- [ ] 设计审查: (当前阶段)
- [ ] 任务检查: (待开始)

待解决: Q2 测试策略待确认
```

**Guardrails**

- **2-3 questions per round, MAX.** Do NOT ask 4+ questions. Do NOT dump all analysis in one response.
- **Wait for user response** before proceeding to next round or phase. NEVER auto-advance.
- **No implementation during analysis.** Do not write NEW code, suggest code changes, or create implementation checklists. Quoting existing artifact content (e.g., design decisions) is fine; writing new code is not.
- **Phase gates are mandatory.** Do NOT skip phases or combine multiple phases in one round.
- **Track everything.** Update the progress tracker at every phase transition.

**Red Flags — STOP and re-read this skill**

- You're about to output more than ~30 lines in a single analysis round
- You're asking 4+ questions at once
- You're including code snippets or implementation details
- You're skipping a phase because "it seems clear"
- You haven't waited for user confirmation before moving on

**Output**

After all phases confirmed:
- Summary of confirmed decisions and open items
- List of artifact updates made
- Readiness assessment: "Ready for `/openspec-apply`" or "N items still need resolution"
