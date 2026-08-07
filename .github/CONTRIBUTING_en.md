# Contributing

BlueKing Node Manager welcomes focused issues and pull requests. Before contributing, read the relevant project rules first so the change follows the repository's existing service boundaries and verification flow.

## Before You Start

- Read `README.md` or `README_en.md` to understand the project entry points.
- Read `AGENTS.md` for repository-wide development rules.
- Before changing a directory, read every applicable scoped `AGENTS.md` from the repository root down to that directory.
- For API changes, read `docs/api/API接口开发流程.md` first.
- For documentation changes, read `docs/AGENTS.md` first.

## Submit an Issue

Use the existing issue templates under `.github/ISSUE_TEMPLATE`.

### Bug Report

Use `.github/ISSUE_TEMPLATE/bug_report.yml`.

Before submitting, confirm that:

- You searched existing issues and did not find a duplicate.
- The problem can be reproduced reliably.
- The issue includes the affected version, environment, logs, reproduction steps, and expected behavior.

### Feature Request

Use `.github/ISSUE_TEMPLATE/feature_request.yml`.

Describe:

- The current problem or limitation.
- The proposed solution.
- The concrete use case.
- Alternatives you have considered, if any.

### Discussion or Proposal

Use `.github/ISSUE_TEMPLATE/discussion.yml` for topics that need design discussion before implementation, such as architecture changes, API design, engineering practice improvements, performance ideas, or feature proposals without a settled solution.

### Security Issues

Do not disclose security vulnerability details in public issues. Until a reporting channel is confirmed, avoid publishing sensitive details in public project spaces.

## Submit a Pull Request

Before opening a pull request:

1. Start from the latest target branch.
2. Keep the change focused on one issue or one behavior change.
3. Fill in `.github/PULL_REQUEST_TEMPLATE.md`.
4. Link the related issue, or write `N/A` when there is no issue.
5. Run the checks that match the changed area.
6. Wait for GitHub Actions, labels, related-file checks, review suggestions, and maintainers' review.

## Code Change Rules

### Go Backend

- Follow `AGENTS.md` and the scoped `AGENTS.md` files for the changed path.
- Preserve the handler -> service -> storage / proto converter layering.
- Do not put database logic in handlers.
- Do not leak proto structs into business logic.
- Do not hand-edit generated `*.pb.go` files.
- For proto changes, regenerate generated files with:

```bash
cd proto && make clean && make all
```

### Frontend

- Follow `front/AGENTS.md`.
- Use the existing pnpm, ESLint, Vue 3, and TypeScript conventions.
- Keep frontend API naming aligned with backend proto and route semantics.

### Documentation

- Follow `docs/AGENTS.md`.
- Place new documentation in the existing docs structure when possible:
  - `docs/api`
  - `docs/developer`
  - `docs/operation`
  - `docs/concepts`

## Verification

Choose verification commands according to the changed area. Common repository checks include:

```bash
make pre
make all
make lint
```

Frontend local verification:

```bash
cd front && pnpm install && pnpm dev
```

Integration tests:

```bash
cd test && make build && make test
```

In the pull request, record the commands you actually ran and their results in the "测试结果 / Test Results" section.

## Review Expectations

Contributors should:

- Keep pull requests small and focused.
- Respond to review comments.
- Avoid unrelated refactors or formatting-only changes in behavior changes.
- Never delete or bypass failing checks to make a pull request pass.
- Explain risks and rollback plans for public API, proto, frontend, deployment, permission, or storage changes.

Reviewers will check whether the change follows repository boundaries, scoped `AGENTS.md` rules, required verification, and the affected API, proto, frontend, deployment, permission, or storage contracts.
