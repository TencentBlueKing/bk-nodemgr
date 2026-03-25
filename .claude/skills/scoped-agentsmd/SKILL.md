---
name: scoped-agentsmd
description: Use when the user wants to create, regenerate, or compress a scoped (subtree-specific) AGENTS.md inside bk-nodemgr, especially when the scope is given as a directory path or described in natural language (e.g. “backend api-v3 router”, “pkg/rest layer”, “proto conversion helpers”).
---

# Scoped AGENTS.md (bk-nodemgr)

Write a **scoped** `AGENTS.md` for a **specific subtree** of the current bk-nodemgr repo.

This skill is NOT for writing a single repo-wide AGENTS.md. It is for writing **one AGENTS.md per scope** (directory/module area).

## Output contract (non-negotiable)

You MUST produce a **complete** `AGENTS.md` file located at the chosen scope root directory.

The output MUST be in **pipe-index compressed format** (Vercel style):

- Prefer single-line `|...` index entries.
- Do not write long prose blocks.
- Do not use traditional Markdown document structure (no `##` sections as the main structure).
- Avoid fenced code blocks as the main content (if you must include code, keep it minimal and consider linking to files instead).

Compression guide reference (read when you are about to write/validate the file):
- `references/AGENTS-compression-guide.md`

This bundled reference is a full local copy of the compression guide and should travel with the skill. Prefer this local file over any host-specific global path.

Before drafting ANY scoped `AGENTS.md`, you MUST read `references/AGENTS-compression-guide.md` in full during the current invocation. Do not rely on memory or a previously read global copy. Re-read the bundled local file before writing and again before final validation if there is any doubt about format compliance.

## Scope input forms

The user may specify scope in either form:

1) **Path form** (preferred):
- Example: `internal/backend/router/api-v3/`

2) **Natural language form** (must support):
- Example: “backend api-v3 router”, “router/api-v3 callback flow”, “pkg rest framework”, “frontend api modules”.

If natural language matches multiple plausible scope roots, you MUST:
- list 2-6 candidate directories,
- explain the difference in 1 line each,
- ask the user to pick one.

Never guess silently.

## Mandatory tool usage (bk-nodemgr)

Use retrieval-led reasoning. Prefer tools over memory.

- Use `serena_find_file` to discover existing `AGENTS.md` files.
- Use `serena_list_dir` to understand local directory layout.
- Use `serena_search_for_pattern` to map natural language → directory candidates (search for key path fragments like `router/api-v3`, `pkg/rest`, `dao/mongo`, etc.).

Avoid reading large files unless necessary.

## Generation workflow

### Step 0: Confirm the scope root

1) If the user provided a path:
- verify it exists,
- treat it as scope root.

2) If the user provided natural language:
- derive keywords and likely path hints,
- locate candidates,
- ask for confirmation if ambiguous.

### Step 1: Discover existing knowledge in the scope

Within the chosen scope root:

- Find all nested `AGENTS.md` under it.
- Identify 0-3 closest “peer” scopes (siblings) that already have good `AGENTS.md`.

Goal: avoid duplicating what is already documented in nested scopes; instead, point to them.

### Step 2: Extract scope-specific content

You are writing **only what is needed for this scope**:

- What this subtree owns (responsibilities)
- Where to look (key files/dirs) for common tasks in this scope
- Conventions specific to this scope (boundaries, type flow, error handling conventions)
- Anti-patterns specific to this scope
- Commands (only if scope-specific; otherwise omit)

Do NOT copy the entire repo structure into a scoped file.

### Step 3: Write the file in pipe-index format

Use short categories. Keep it scannable.

#### Required header lines

Always include these near the top (exact content can be copied verbatim):

|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)

Then include a scope line:

|Scope:<relative/path/from/repo/root>

#### Recommended categories (pick what fits)

- `|Overview:...`
- `|Structure:<dir>:{child1,child2,...}`
- `|Where to look:<task>:<path(s)>:<note>` (keep notes short)
- `|Conventions:...`
- `|Anti-patterns:...`
- `|Dependencies:...` (only if it helps avoid wrong imports/boundary violations)
- `|Child AGENTS:<path1>|<path2>|...` (list nested scopes with their own AGENTS.md)
- `|Commands:...` (scope-specific only)

Guidelines for density:

- Prefer `{}` grouping for sibling files.
- Prefer `|` alternatives on the same line when it stays readable.
- Use `→` to express flows (e.g., request → types → manager → storage → response).

### Step 4: Validate before finishing

Checklist:

- [ ] File is named exactly `AGENTS.md` and saved under the chosen scope root.
- [ ] Starts with `|IMPORTANT:` line.
- [ ] Most lines start with `|`.
- [ ] No long explanatory paragraphs.
- [ ] No repo-wide duplication; content is scope-focused.
- [ ] If scope resolution was ambiguous, user confirmed selection.

## Handling “compress existing AGENTS.md” requests

If the user asks to “rewrite/compress” an existing `AGENTS.md` under a scope:

- Preserve all facts.
- Convert to pipe-index entries.
- Prefer linking to existing files instead of embedding long code examples.

## Examples (inputs)

1) Path-based:
- “为 `internal/backend/router/` 写一个 scoped AGENTS.md（压缩格式）”

2) Natural language:
- “我最近主要在做 backend 的 api-v3 路由开发，帮我生成对应 scope 的 AGENTS.md”

3) Compress:
- “把 `internal/backend/router/AGENTS.md` 改写成 pipe-index 压缩格式，并保留信息”
