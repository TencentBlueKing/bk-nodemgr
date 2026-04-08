# IAM 权限模型迁移文件

本目录存放 bk-nodemgr 的 IAM 权限模型迁移文件。

## 文件说明

| 文件                  | 说明 |
|---------------------|------|
| `iam-render`        | 渲染工具（可执行文件） |
| `new_template.sh`   | 创建新模板的脚本 |
| `templates/`        | 模板文件目录 |
| `vars.yaml.example` | 变量配置示例 |
| `render/`           | 渲染工具源码 |

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

## 变量配置

复制 `vars.yaml.example` 为 `vars.yaml` 并填入实际值：

| 变量路径 | 说明 | 默认值 |
|----------|------|--------|
| `provider.host` | IAM 回调 bk-nodemgr 的地址 | `""` (空) |
| `provider.auth` | 认证方式 | `"basic"` |

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
