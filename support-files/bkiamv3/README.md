# IAM 权限模型迁移文件

本目录存放 bk-nodemgr 的 IAM 权限模型迁移文件。

## 文件说明

| 文件                | 说明                   |
| ------------------- | ---------------------- |
| `iam-render`        | 渲染工具（可执行文件） |
| `new_template.sh`   | 创建新模板的脚本       |
| `templates/`        | 模板文件目录           |
| `vars.yaml.example` | 变量配置示例           |
| `render/`           | 渲染工具源码           |

## 使用方法

### 1. 渲染模板

```bash
cd support-files/bkiamv3
./iam-render -t templates -v vars.yaml -o output
```

渲染 `templates/` 目录下所有 `.tpl` 文件，输出到 `output/` 目录（去掉 `.tpl` 后缀）。

### 2. 执行迁移

使用 `support-files/bkiamv3/do_migrate.py` 脚本执行迁移：

```bash
cd support-files/bkiamv3
python do_migrate.py \
  -t "https://bkapi.example.com/api/bk-iam/prod/" \
  -f "output/0001_bk_nodemgr_init.json" \
  -a "bk-nodemgr" \
  -s "your-app-secret"
```

## 系统访问权限

`system_access` 直接包含在现有模型模板中，按 `0003`、`0004`、`0005` 的顺序同步：

- `templates/0003_bk_nodemgr_actions.json.tpl` 注册无资源绑定的 `system_access`（系统访问），`related_resource_types` 为 `[]`。该权限仅用于系统准入，不替代具体功能和资源权限。原有 32 个 Action 均增加 `system_access` 依赖，保留已有依赖；除下述 `package_history_view` 改绑外，其余资源绑定不变。
- `templates/0004_bk_nodemgr_action_groups.json.tpl` 增加“系统”操作分组，保留原有分组及成员。
- `templates/0005_bk_nodemgr_common_actions.json.tpl` 为全部四个常用操作组合增加 `system_access`，保留原有成员。

已完成历史同步的环境也需重新执行上述模板，使远端模型包含 `system_access` 及其依赖、分组和常用操作配置。

`related_actions` 用于申请权限时同时创建依赖权限，不是运行时自动鉴权，也不会为已有用户自动补授权。常用操作组合是申请权限的快捷配置，不等同于 V4 Role。接入系统访问中间件前，应先确认已有用户获得 `system_access`，避免原有功能权限仍有效但系统准入被拒绝。本迁移不包含中间件或用户授权变更。

## 显式替换 Action 资源绑定

模板继续使用 `upsert_action` 和 IAM Action API 字段，不增加替换操作或迁移控制字段。工具在更新前重新查询当前绑定，按顺序比较 `related_resource_types` 的 `(system_id, id)`；只修改实例选择器等元数据不会触发策略清理。缺省该字段表示不修改绑定，显式 `[]` 表示清空绑定。

- Action 不存在：直接创建目标定义。
- 资源绑定不变：普通更新，不删除策略。
- 资源绑定变化：默认停止；只有显式传入 `--allow-action-replace`，才校验目标资源类型、实例选择器及名称，提交策略清理、删除旧 Action，确认旧 Action 已不存在后按目标定义重新创建。查询或校验失败时停止，不清理策略。重建时保留请求中未提供的现有可写字段。

`--allow-action-replace` 适用于本次文件或目录执行中所有发生绑定变化的 `upsert_action`，不限定为 `package_history_view`，也不改变显式 `update_action` 的行为。执行前检查选定模板的全部绑定变更；需要缩小范围时使用 `-f`：

```bash
python do_migrate.py \
  -t "https://bkapi.example.com/api/bk-iam/prod/" \
  -f "output/0003_bk_nodemgr_actions.json" \
  -a "bk-nodemgr" \
  -s "your-app-secret" \
  --allow-action-replace
```

策略清理与 Action 删除都可能异步执行，接口成功不代表已完成。工具只提交一次策略清理和 Action 删除请求，随后最多查询 60 次，每次仍存在时等待 2 秒；确认旧 Action 已消失才创建，不通过 PUT 原位改绑。删除失败、查询失败或等待结束后仍存在时停止，不尝试创建，也不自动重发删除请求。操作分组、常用操作和新建关联等配置可能影响删除，应检查远端错误；工具不自动清理这些配置，重建后需同步并核对 `0003`～`0005` 中的相关配置。

该流程没有事务或自动回滚，删除与重建之间 Action 暂时不可用。任一步骤失败或结果未知时，应先检查远端状态：Action 已不存在时，可重跑同一目标模板直接创建；已是目标绑定时重跑只执行普通更新；仍是旧绑定时，先确认删除状态再恢复，避免重复提交清理请求。

**授权影响：** 策略删除不可逆，旧的资源包级授权不会转换为资源包类型级授权。执行前确认系统、租户和受影响的用户／用户组；执行后重新配置并验证 `package_type` 维度的权限。策略删除接口要求 IAM SaaS ≥ 1.4.15、后台 ≥ 1.7.7；不支持时迁移停止，不回退到其他删除方式。

`0003` 中的 `package_history_view` 直接声明最终 `package_type` 绑定，并保留 `system_access` 依赖；旧环境首次改绑需显式传入 `--allow-action-replace`。全目录迁移不会先写回 `package` 再替换，因此完成迁移后重跑不会再次清理该 Action 的策略。重新渲染时使用空输出目录，避免已删除模板对应的旧 JSON 文件残留。

`--dry-run` 仍然只列出迁移文件，不查询模型、验证依赖或输出操作级计划，不能据此确认策略清理的影响范围。

参考：[Action API](https://github.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/13-Action.md)、[Action Policy API](https://github.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/20-ActionPolicy.md)。

## 变量配置

复制 `vars.yaml.example` 为 `vars.yaml` 并填入实际值：

| 变量路径        | 说明                       | 默认值    |
| --------------- | -------------------------- | --------- |
| `provider.host` | IAM 回调 bk-nodemgr 的地址 | `""` (空) |
| `provider.auth` | 认证方式                   | `"basic"` |

## 添加新模板

```bash
cd support-files/bkiamv3
./new_template.sh add_resource_type
```

自动生成下一个序号的模板文件，如 `0002_bk_nodemgr_add_resource_type.json.tpl`。

## 模板语法

模板文件使用 Go template 语法：

- 变量引用：`{{ .parent.child }}`
- 默认值：`{{ .VarName | default "value" }}`

详见 [Go text/template 文档](https://pkg.go.dev/text/template)。

## 参考文档

- Migration 指南：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/HowTo/Solutions/Migration.md>
- `do_migrate.py`：<https://raw.githubusercontent.com/TencentBlueKing/iam-python-sdk/blob/master/iam/contrib/iam_migration/utils/do_migrate.py>
- `example.json`：<https://raw.githubusercontent.com/TencentBlueKing/iam-python-sdk/blob/master/iam/contrib/iam_migration/utils/example.json>
- System：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/10-System.md>
- ResourceType：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/11-ResourceType.md>
- InstanceSelection：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/12-InstanceSelection.md>
- Action：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/13-Action.md>
- ActionGroup：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/14-ActionGroup.md>
- CommonActions：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/17-CommonActions.md>
- ResourceCreatorAction：<https://raw.githubusercontent.com/TencentBlueKing/BKDocs/blob/main/ZH/IAM/IntegrateGuide/Reference/API/02-Model/19-ResourceCreatorAction.md>

为避免文档漂移，模板维护时请同步参照：

- `support-files/bkiamv3/templates/README.md`
