# IAM Migration 模板参考（中文）

本文档用于约束 `support-files/bkiamv3/templates` 下模板的维护依据。后续调整模板时，请以以下参考链接和规则为准。

## 源文档

- Migration 指南：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/HowTo/Solutions/Migration.md>

## 迁移工具参考

- `do_migrate.py`：<https://raw.githubusercontent.com/TencentBlueKing/iam-python-sdk/blob/master/iam/contrib/iam_migration/utils/do_migrate.py>
- `example.json`：<https://raw.githubusercontent.com/TencentBlueKing/iam-python-sdk/blob/master/iam/contrib/iam_migration/utils/example.json>

## 按 operation 对应的模型文档

### system

- `upsert_system` / `add_system` / `update_system`
- 结构参考：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/10-System.md>

### resource_type

- `upsert_resource_type` / `add_resource_type` / `update_resource_type` / `delete_resource_type`
- 结构参考：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/11-ResourceType.md>

### instance_selection

- `upsert_instance_selection` / `add_instance_selection` / `update_instance_selection` / `delete_instance_selection`
- 结构参考：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/12-InstanceSelection.md>

### action

- `upsert_action` / `add_action` / `update_action` / `delete_action`
- 结构参考：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/13-Action.md>

### action_groups

- `upsert_action_groups`
- 结构参考：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/14-ActionGroup.md>

### resource_creator_actions

- `upsert_resource_creator_actions`
- 结构参考：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/19-ResourceCreatorAction.md>

### common_actions

- `upsert_common_actions`
- 结构参考：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/17-CommonActions.md>

## 模板维护规则（来自 Migration 指南）

- 优先使用 `upsert_*`，保证 migration 可重复执行（幂等）。
- migration 文件命名遵循：`{seq}_{APP_CODE}_{YYYYmmdd-HHMM}_iam.json`。
- 注册/更新 system 时，`clients` 必须包含当前系统自身 `app_code`，否则后续可能失去模型变更权限。
