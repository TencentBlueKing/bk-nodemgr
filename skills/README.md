# 蓝鲸节点管理 Skills

简体中文 | [English](README_en.md)

面向蓝鲸节点管理用户的 AI 助手技能，帮助你完成部署、配置和运维任务。选择适合当前任务的 Skill，安装到 AI 助手后，用自然语言描述你的目标即可开始。

## Skills 目录

| Skill | 适用场景 |
| --- | --- |
| [bk-nodemgr-deploy](bk-nodemgr-deploy/SKILL.md) | 通过 Helm、Helmfile、Docker Compose 或二进制方式首次部署、调整运行配置、升级，以及排查部署和启动问题。 |

每个 Skill 的 `SKILL.md` 分别定义适用范围、工作流程、前置条件和访问或执行权限边界。使用前请查看所选 Skill 的说明，不要假定所有 Skills 都具有相同权限。

## 安装

准备 [skills CLI 支持的 AI 助手](https://github.com/vercel-labs/skills#supported-agents)，以及 Node.js 22.20 或更高版本。在需要使用 Skill 的项目目录中运行：

```bash
npx skills add TencentBlueKing/bk-nodemgr --skill <skill-name>
```

将 `<skill-name>` 替换为上表中的技能名称。例如，安装部署助手：

```bash
npx skills add TencentBlueKing/bk-nodemgr --skill bk-nodemgr-deploy
```

按提示选择目标 AI 助手。默认只安装到当前项目；如需跨项目使用，可添加 `-g` 进行全局安装。

如需查看仓库中可发现的 Skills，而不安装它们：

```bash
npx skills add TencentBlueKing/bk-nodemgr --list
```

仓库的 `.agents/skills/` 中还包含面向贡献者的开发技能，CLI 也可能发现它们。请结合上表，通过 `--skill` 按名称选择面向用户的技能。

## 使用

在 AI 助手中说明要使用的 Skill 和希望达成的目标。如果已知相关版本、环境或部署方式，请一并提供；Skill 会在必要时追问影响建议的信息。

例如，使用 `bk-nodemgr-deploy`：

- “使用 bk-nodemgr-deploy，帮我规划通过 Helm 部署蓝鲸节点管理。”
- “使用 bk-nodemgr-deploy，帮我检查 Docker Compose 部署中的运行配置调整。”
- “使用 bk-nodemgr-deploy，帮我排查升级后的启动失败问题。”

部署助手负责准备和审查配置及操作步骤，由你执行目标环境中的变更，不要求本地已有源码。请只提供与任务相关且经过脱敏的信息，不要在对话中粘贴密码或 Token。

## 相关文档

- [蓝鲸节点管理安装部署指南](../docs/operation/installation.md)
- [skills CLI 使用说明](https://skills.sh/docs/cli)
