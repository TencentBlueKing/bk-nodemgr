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

- `templates/0003_bk_nodemgr_actions.json.tpl` 注册无资源绑定的 `system_access`（系统访问），`related_resource_types` 为 `[]`。该权限仅用于系统准入，不替代具体功能和资源权限。原有 32 个 Action 均增加 `system_access` 依赖，保留已有依赖和资源绑定。
- `templates/0004_bk_nodemgr_action_groups.json.tpl` 增加“系统”操作分组，保留原有分组及成员。
- `templates/0005_bk_nodemgr_common_actions.json.tpl` 为全部四个常用操作组合增加 `system_access`，保留原有成员。

已完成历史同步的环境也需重新执行上述模板，使远端模型包含 `system_access` 及其依赖、分组和常用操作配置。

`related_actions` 用于申请权限时同时创建依赖权限，不是运行时自动鉴权，也不会为已有用户自动补授权。常用操作组合是申请权限的快捷配置，不等同于 V4 Role。接入系统访问中间件前，应先确认已有用户获得 `system_access`，避免原有功能权限仍有效但系统准入被拒绝。本迁移不包含中间件或用户授权变更。

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
