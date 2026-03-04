---
name: openspec-enhance-change
description: Enhanced implementation with Serena tool constraints, milestone-driven execution, and auto-commit. Use when the user wants stricter tool discipline and milestone tracking during implementation.
license: MIT
compatibility: Requires openspec CLI and Serena MCP tools.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.0.2"
---

Enhanced implementation mode for an OpenSpec change — enforces Serena tool constraints, milestone-driven execution, and auto-commit per milestone.

**Input**: Specify a change name. If omitted, infer from context or prompt for available changes.

**Steps**

1. **Select the change**

   If a name is provided, use it. Otherwise:
   - Infer from conversation context if the user mentioned a change
   - Auto-select if only one active change exists
   - If ambiguous, run `openspec list --json` to get available changes and use the **AskUserQuestion tool** to let the user select

   Always announce: "Using change: <name>" and how to override (e.g., `/opsx:enhance <other>`).

2. **Load change context**

   ```bash
   openspec status --change "<name>" --json
   openspec instructions apply --change "<name>" --json
   ```

   Read all context files from the apply instructions output. Understand the full scope of work including:
   - Schema, artifact structure, task list
   - Affected files and code areas
   - Dependencies between tasks

3. **Generate milestone plan**

   Based on the change's tasks and scope, generate a milestone plan. Each milestone contains:

   ```
   ### Milestone: <name>
   **Completion criteria**: <what must be true when done>
   **Tool constraints**: <which Serena tools to use and how>
   **Commit rule**: auto-commit with message format "enhance(<change>): <milestone-name>"
   **Tasks**: <list of tasks covered by this milestone>
   ```

   **Milestone design principles:**
   - Group related tasks into logical milestones
   - Each milestone should be independently verifiable
   - Order milestones by dependency (analysis before implementation)
   - Typical milestone types: Analysis, Refactoring, Implementation, Verification

   Present the milestone plan to the user via **AskUserQuestion** for confirmation. Allow the user to adjust, reorder, or redefine milestones before proceeding.

4. **Execute milestones sequentially**

   For each milestone:

   **a. Announce milestone start**
   ```
   ## Milestone N/M: <name>
   Starting...
   ```

   **b. Save milestone context to memory**
   Use `write_memory` to save current milestone state for recovery across sessions.

   **c. Execute with tool constraints (MANDATORY)**

   The following tool usage rules are STRICTLY ENFORCED:

   | Phase | Required Tools | Forbidden Actions |
   |-------|---------------|-------------------|
   | **Analysis** | `get_symbols_overview`, `find_symbol(include_body=False)` | Reading entire files with `cat`/`Read` |
   | **Impact check** | `find_referencing_symbols` before ANY modification | Editing without checking references |
   | **Code reading** | `find_symbol(include_body=True)` for specific symbols | Reading full files when symbol access suffices |
   | **Editing** | `replace_symbol_body`, `insert_after_symbol`, `insert_before_symbol` for symbol-level changes | Using `Write` to overwrite entire files |
   | **Micro-edits** | `replace_content` for few-line changes within a symbol | Using `replace_symbol_body` for single-line fixes |
   | **Renaming** | `rename_symbol` for cross-codebase renames | Manual find-and-replace across files |
   | **Search** | `search_for_pattern` for cross-file pattern discovery | Using `grep`/`rg` via Bash |
   | **File discovery** | `list_dir`, `find_file` for locating files | Guessing file paths |

   **d. Verify milestone completion**
   - Check all completion criteria are met
   - Run relevant checks (build, lint, test) if applicable
   - If a criterion fails → **PAUSE**, report the issue, wait for user guidance

   **e. Auto-commit**
   On milestone completion:
   ```bash
   git add <changed-files>
   git commit -m "enhance(<change-name>): <milestone-name>"
   ```
   Record the commit hash for the milestone report.

   **f. Generate milestone report**
   ```
   ### Milestone N/M Complete: <name>
   **Commit**: <hash>
   **Files changed**: <list>
   **Tools used**: <list of Serena tools invoked>
   **Completion criteria**: all met ✓
   ```

5. **Handle failures**

   When a milestone fails:
   ```
   ## Milestone N/M Paused: <name>

   ### Issue Encountered
   <description of the problem>

   ### Diagnostic Info
   - Tool that detected the issue: <tool>
   - Affected symbol/file: <location>
   - Error details: <details>

   ### Options
   1. Fix the issue and retry this milestone
   2. Skip this milestone and continue
   3. Adjust the milestone plan
   4. Abort enhanced mode

   What would you like to do?
   ```

6. **Final summary**

   After all milestones complete:
   ```
   ## Enhanced Implementation Complete

   **Change:** <change-name>
   **Milestones:** M/M complete ✓

   ### Milestone Summary
   | # | Name | Commit | Files | Status |
   |---|------|--------|-------|--------|
   | 1 | <name> | <hash> | N files | ✓ |
   | 2 | <name> | <hash> | N files | ✓ |
   ...

   ### Tool Usage Statistics
   - Symbol analyses: N
   - Reference checks: N
   - Symbol edits: N
   - Pattern searches: N
   - Renames: N

   ### Memory
   Decision log saved to memory: enhance/<change-name>
   ```

**Guardrails**

- **NEVER** read entire source files — use `get_symbols_overview` first, then `find_symbol` with `include_body=True` for needed symbols only
- **ALWAYS** run `find_referencing_symbols` before modifying any symbol to check impact
- **ALWAYS** use `rename_symbol` for cross-codebase renames instead of manual replacement
- **ALWAYS** auto-commit after each milestone completes successfully
- **NEVER** commit if milestone verification fails — pause and report
- Use `write_memory` / `read_memory` to persist milestone state across sessions
- Milestone plan must be confirmed by user before execution begins
- Keep each milestone focused and independently verifiable
- If implementation reveals design issues, pause and suggest artifact updates

**Fluid Workflow Integration**

This skill supports the "actions on a change" model:

- **Can be invoked anytime**: After artifacts exist, during partial implementation, or to re-enhance remaining work
- **Complements /opsx:apply**: Use `enhance` when you want stricter tool discipline and milestone tracking; use `apply` for lighter-weight task execution
- **Resumes gracefully**: Reads memory to recover milestone state if a session was interrupted
