# 贡献指南

BlueKing Node Manager 欢迎聚焦的问题反馈和 Pull Request。贡献前请先阅读相关项目规则，确保改动符合仓库现有的 service boundary 和 verification flow。

## 开始前先读

- 阅读 `README.md` 或 `README_en.md`，了解项目入口。
- 阅读 `AGENTS.md`，了解仓库级开发规则。
- 修改具体目录前，阅读从仓库根目录到目标目录路径上所有适用的 scoped `AGENTS.md`。
- 涉及 API 改动时，先阅读 `docs/api/API接口开发流程.md`。
- 涉及文档改动时，先阅读 `docs/AGENTS.md`。

## 提交 Issue

请使用 `.github/ISSUE_TEMPLATE` 下已有的 issue template。

### Bug 报告

使用 `.github/ISSUE_TEMPLATE/bug_report.yml`。

提交前请确认：

- 已搜索现有 issue，确认不是重复问题。
- 问题可以稳定复现。
- issue 中包含受影响版本、环境信息、相关日志、复现步骤和期望行为。

### 功能请求

使用 `.github/ISSUE_TEMPLATE/feature_request.yml`。

请说明：

- 当前问题或限制是什么。
- 期望方案是什么。
- 具体使用场景是什么。
- 是否考虑过其他替代方案。

### 讨论或提案

使用 `.github/ISSUE_TEMPLATE/discussion.yml` 提交需要先讨论再实现的议题，例如架构调整、API 设计、工程实践改进、性能优化方向，或尚未确定方案的功能想法。

### 安全问题

不要在公开 issue 中披露安全漏洞细节。在安全报告渠道确认前，请避免在公开项目空间发布敏感细节。

## 提交 Pull Request

提交 Pull Request 前请确认：

1. 基于最新目标分支创建改动分支。
2. 改动聚焦在一个 issue 或一个行为变化上。
3. 填写 `.github/PULL_REQUEST_TEMPLATE.md`。
4. 关联对应 issue；没有 issue 时写 `N/A`。
5. 根据改动范围运行对应检查。
6. 等待 GitHub Actions、labels、related-file checks、review suggestions 和维护者 review。

## 代码改动规则

### Go Backend

- 遵循 `AGENTS.md` 和改动路径下适用的 scoped `AGENTS.md`。
- 保持 handler -> service -> storage / proto converter 分层。
- 不在 handler 中写 database logic。
- 不把 proto struct 泄漏到 business logic。
- 不手动修改 generated `*.pb.go` 文件。
- 涉及 proto 改动时，使用以下命令重新生成代码：

```bash
cd proto && make clean && make all
```

### Frontend

- 遵循 `front/AGENTS.md`。
- 使用项目已有的 pnpm、ESLint、Vue 3 和 TypeScript 约定。
- Frontend API 命名需与 backend proto 和 route 语义保持一致。

### 文档

- 遵循 `docs/AGENTS.md`。
- 新文档优先放入已有 docs 结构：
  - `docs/api`
  - `docs/developer`
  - `docs/operation`
  - `docs/concepts`

## 验证

根据改动范围选择验证命令。常用仓库检查包括：

```bash
make pre
make all
make lint
```

Frontend 本地验证：

```bash
cd front && pnpm install && pnpm dev
```

请在 Pull Request 的 “测试结果 / Test Results” 中记录实际执行的命令和结果。

## Review 期望

贡献者需要：

- 保持 Pull Request 小而聚焦。
- 回应 review comments。
- 避免在行为改动中混入无关 refactor 或纯格式化改动。
- 不删除或绕过失败检查来让 Pull Request 通过。
- 对 public API、proto、frontend、deployment、permission 或 storage 改动说明风险和回滚方案。

Reviewer 会检查改动是否符合仓库边界、scoped `AGENTS.md` 规则、必要验证，以及受影响的 API、proto、frontend、deployment、permission 或 storage contract。
