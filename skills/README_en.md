# BlueKing Node Manager Skills

English | [简体中文](README.md)

Reusable skills for AI agents that help users deploy, configure, and operate BlueKing Node Manager. Choose a skill for your task, install it into your agent, and describe what you want to accomplish.

## Available Skills

| Skill | When to use it |
| --- | --- |
| [bk-nodemgr-deploy](bk-nodemgr-deploy/SKILL.md) | First deployment, runtime configuration changes, upgrades, or deployment and startup troubleshooting with Helm, Helmfile, Docker Compose, or binaries. |

Each skill's `SKILL.md` defines its scope, workflow, prerequisites, and access or execution boundaries. Check the selected skill rather than assuming all skills have the same permissions.

## Install

Use an agent supported by the [skills CLI](https://github.com/vercel-labs/skills#supported-agents) and Node.js 22.20 or later. Run the command from the project where you want to use the skill:

```bash
npx skills add TencentBlueKing/bk-nodemgr --skill <skill-name>
```

Replace `<skill-name>` with a name from the table above. For example:

```bash
npx skills add TencentBlueKing/bk-nodemgr --skill bk-nodemgr-deploy
```

Follow the prompts to choose the target agent. Installation is project-local by default; add `-g` to make the skill available across projects.

To list the repository's discoverable skills without installing them:

```bash
npx skills add TencentBlueKing/bk-nodemgr --list
```

The repository also contains contributor skills under `.agents/skills/`, which the CLI can discover. Use the table above and `--skill` to select a user-facing skill.

## Use

In your agent, name the installed skill and describe your goal. Include the relevant version, environment, or deployment method when known. The skill can ask for missing information that affects the recommendation.

For example, with `bk-nodemgr-deploy`:

- "Use bk-nodemgr-deploy to help me plan a Helm deployment of BlueKing Node Manager."
- "Use bk-nodemgr-deploy to review a runtime configuration change in my Docker Compose deployment."
- "Use bk-nodemgr-deploy to investigate a startup failure after an upgrade."

The deployment skill prepares and reviews configuration and instructions; you execute changes to the target environment. A local source checkout is optional. Share only task-relevant, redacted information, and never paste passwords or tokens into the conversation.

## Related Documentation

- [BlueKing Node Manager installation guide](../docs/operation/installation.md)
- [skills CLI documentation](https://skills.sh/docs/cli)
