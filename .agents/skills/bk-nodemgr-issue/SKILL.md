---
name: bk-nodemgr-issue
metadata:
  internal: true
description: Use when drafting, creating, supplementing, or updating a GitHub Issue for bk-nodemgr, including bug reports, feature requests, unresolved proposals, and Issue lookup or reuse during PR preparation. Trigger for “写 Issue”, “整理需求”, “补充 Issue”, “提交 Issue”, or “找关联 Issue”. Not for implementing the change or deciding a PR's closing relationship.
---

# bk-nodemgr Issue Workflow

Own the Issue lifecycle for this project, both standalone and when called by `bk-nodemgr-create-github-pr`. Use the current Issue forms and contribution guide rather than a parallel generic Issue workflow. The intended readers are maintainers triaging or taking over the problem; apply `bk-nodemgr-audience-boundary` when drafting the body.

## Operating Modes

| Intent                                       | Mode     | Action                                                                              |
| -------------------------------------------- | -------- | ----------------------------------------------------------------------------------- |
| “写”“整理”“草稿”“补充一下”                   | `draft`  | Read facts and return a draft; no GitHub writes.                                    |
| Explicit request to create an Issue          | `create` | Prepare the exact draft, then obtain submission confirmation.                       |
| Explicit request to update an existing Issue | `update` | Read the current Issue, propose the specific change, then obtain edit confirmation. |
| Find an Issue for a PR or routine work       | `lookup` | Read and confirm reuse; do not mutate the Issue.                                    |

Default to `draft` when ambiguous. Text approval is not submission approval. PR preparation or submission does not authorize an Issue write. Never create an Issue as a substitute for updating an existing one; ask for its number or URL if missing.

## 1. Read the Project Contract

Read `.github/CONTRIBUTING.md`, the Issue-related sections of `docs/developer/git-contribution.md`, `.github/ISSUE_TEMPLATE/config.yml`, and the applicable form. Check relevant `AGENTS.md`, affected modules, and analogous behavior before asking questions. Resolve the canonical repository from remote URLs and GitHub metadata, not the name `origin`; confirm ambiguous targets. Use `gh` for CLI invocation guidance and explicit `--repo <owner/repo>` for Issue operations.

For suspected vulnerabilities, use the security contact in the chooser configuration; do not publish sensitive details in a public Issue. Redact credentials and sensitive log or conversation content in all drafts.

## 2. Find or Read the Issue

- If the user supplies an Issue, read its title, body, state, labels, and relevant discussion before proposing changes. Preserve unrelated facts and history; do not silently reopen, close, relabel, or rewrite it into a different type.
- Before creating, search open Issues by concrete behavior and affected area; inspect relevant closed matches when they may explain a duplicate or regression. Report the search scope and candidate numbers. Search limits are not proof that no duplicate exists.
- For routine `.github/**` CI maintenance, look for the current matching open `[DAILY]` or umbrella Issue. Do not create a Feature merely to satisfy PR title lint. Verify current state and scope rather than hardcoding a historical number.
- Confirm reuse when the match is not explicit in the user's request. A keyword match alone is insufficient. Reuse does not require editing, commenting on, or closing the Issue.

## 3. Select and Render the Form

| Intent                                                        | Form                  |
| ------------------------------------------------------------- | --------------------- |
| Existing behavior is wrong                                    | `bug_report.yml`      |
| Concrete new capability or enhancement                        | `feature_request.yml` |
| Unresolved design, architecture, API, or engineering proposal | `discussion.yml`      |

Use the current chooser policy. If maintenance or another request does not fit and blank Issues are disabled, ask which form to use; do not bypass the chooser with a generic body. If local forms are absent, inspect canonical forms and recent same-type Issues before proposing a fallback.

For a new Issue, render the selected YAML form into Markdown, not YAML answers:

1. Read the exact `title` prefix, default labels, input labels/order, required fields, checkbox claims, and dropdown options from the form. Keep Issue titles separate from PR title conventions.
2. Use each input label as a heading in order. Preserve exact selected option text and all selections for `multiple: true`. Retain specified log code fences; do not copy instructional placeholders as facts.
3. Include required fields. Fit scope, non-goals, and verifiable acceptance expectations inside the appropriate form fields rather than adding a competing generic template. Optional implementation fields can remain empty when no supported plan exists.
4. Check a checklist item only from actual evidence or explicit user confirmation. In particular, do not invent documentation reading, stable reproduction, environment versions, logs, or duplicate-search completion.
5. Verify proposed labels exist remotely before submission. Report missing labels and ask before omitting, replacing, or creating them; a YAML label entry does not create a remote label.

For updates, keep the existing template-compliant structure and make only the requested changes. Propose any necessary template migration separately rather than silently discarding existing context.

## 4. Ground the Draft and Resolve Gaps

Describe supported problem facts, user impact, affected component, boundaries, and the decision or outcome needed. Use actual paths and terminology. A problem handoff or undecided proposal is not permission to prescribe interface changes, shared abstractions, or an implementation plan.

Ask only material questions tied to the current form and project facts: missing version/reproduction evidence, affected service, scope/non-scope, expected behavior, failure semantics, category, or acceptance. Distinguish suspected causes from established facts. Do not turn a form placeholder into a requirement or invent alternatives and benefit measurements.

Return the reader-facing draft separately from author-facing questions and submission blockers. Unsupported required claims stay unchecked; do not claim the draft is submission-ready or publish it while material required evidence is missing. An intermittent bug does not become reliably reproducible merely to satisfy the form. Do not change the form or reclassify the problem solely to evade that gap.

## 5. Confirm, Submit, and Read Back

Before a write, show the target repository, mode, selected form, exact title and complete body, labels, and the specific creation or edit being requested. For an existing Issue, show what will change. Confirm submission separately from text review, push, PR creation, and unrelated metadata changes.

After approval, verify `gh auth status` and the exact target repository with `gh repo view <owner/repo> --json nameWithOwner`. Write the approved multiline body to a concrete temporary file and use the applicable command:

```bash
gh issue create --repo <owner/repo> --title "<approved-title>" --body-file <approved-file>
gh issue edit <number> --repo <owner/repo> --body-file <approved-file>
gh issue view <number> --repo <owner/repo> --json title,body,url,labels,state
```

Run only the approved operation, with only approved title/label changes. Compare readback with the approved draft. Show any corrective edit and obtain confirmation before applying it, then verify again. If a create request times out or otherwise leaves success uncertain, inspect remote state before retrying; never blindly repeat creation.

## Output and PR Handoff

For standalone drafting, return repository, selected form, exact title/body, proposed versus verified labels, and any material questions outside the body. For updates, identify the existing Issue and requested changes. Report writes performed or explicitly state `No GitHub writes performed`; return a real URL only from a verified existing or created Issue.

When called by `bk-nodemgr-create-github-pr`, return:

- repository and confirmed Issue number/URL, or no confirmed match;
- current state, covered behavior/scope and acceptance expectations;
- whether it is routine, `[DAILY]`, umbrella, or other long-lived tracking;
- unresolved evidence or reuse decisions;
- writes performed and specific actions still awaiting confirmation.

Do not decide that a PR completes the Issue. The PR workflow compares its actual delivery with this handoff and owns the `refs` versus closing decision under `docs/developer/git-contribution.md`.

## Final Check

- Current form, prefix, ordered fields, options, required claims, and labels match the project sources.
- Issue reuse is grounded; drafting or lookup did not cause remote writes.
- Public body contains facts and decision points, not author process or an unsupported plan.
- Submission and corrections were confirmed and read back; ambiguous creates were checked before retry.
- PR association uses the handoff rather than a second Issue-creation workflow.

Reusable scenario definitions live in `evals/evals.json`; they are not evidence of executed tests.
