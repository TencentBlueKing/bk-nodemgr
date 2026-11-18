# Nodeman 插件进程迁移

> 本脚本仅适用于 `蓝鲸7.2` 升级至 `蓝鲸7.3` 版本时从 `NodeMan` 迁移进程数据到 `NodeMgr` 时使用

本目录存放 nodeman 到 nodemgr 的离线迁移脚本。当前提供 nodeman 安装的主插件进程回填：

```text
nodeman pod: export_nodeman_process.py (Django 已连数据库) -> /tmp/nodeman_export/<tenant>/<biz_id>/<plugin>.js
      ↓ kubectl cp 或运维中转
nodemgr pod: ./mongo.sh login -> load(<单个文件或 import_selected.js>)
```

脚本位于本目录的 [export_nodeman_process.py](export_nodeman_process.py)。

## 第一步：在 nodeman pod 内导出

把本目录的脚本放置到到 bk-nodeman backend-api pod的指定目录 `/app/apps/node_man/management/commands/` 下并执行：

```bash
python manage.py export_nodeman_process
```

导出的数据按 mongodb 可解析的格式保存在bk-nodeman-backend-api pod的 `/tmp/nodeman_export/` 目录下。


### 输出目录结构

```text
/tmp/nodeman_export/
  system/               # tenant_id 如果非多租户环境默认为default
    2/                  # bk_biz_id
      bkmonitorbeat.js  # plugin_name
  import_selected.js    # 批量导入清单，注释掉不需要的行即可跳过
```

目录按 `tenant/biz/<plugin_name>` 组织：每个文件只包含该分组的记录，可单独导入；`import_selected.js` 按行列出全部分组，默认全部启用，用于批量导入。

## 第二步：在 nodemgr pod 内导入

把导出的整个目录拷贝到 bk-nodemgr-backend pod（如 `/tmp/nodeman_export`），用镜像自带的 mongosh 执行。

```bash
cd /bk-nodemgr/script/ops
./mongo.sh login
```

确认当前使用的数据库:

```javascript
db.getName()
```

单独导入：

```javascript
load("/tmp/nodeman_export/cw/2/bkmonitorbeat.js")
```

批量导入：

```javascript
load("/tmp/nodeman_export/import_selected.js")
```

如需批量导入部分数据，编辑 `import_selected.js` 中的 `files` 列表，注释掉不需要导入的行，然后执行上面的批量导入操作：

```javascript
var files = [
  "cw/2/bkmonitorbeat.js",
  // "cw/2/processbeat.js",
  "default/3/processbeat.js",
];
```

> 导入只 upsert 文件中的记录，不删除数据库中已有数据。
> 同一份文件重复执行结果一致。

## 附加说明

### 导出筛选条件

只导出同时满足以下条件的记录：

| 数据源 | 数据库筛选条件 | 作用 |
| ------ | -------------- | ---- |
| `node_man_subscription` | `is_main = true`、`is_deleted = false`、`plugin_name` 非空 | 只处理主插件安装订阅 |
| `node_man_processstatus` | `source_type = default`、`is_latest = true`、`source_id ∈ 主订阅ID`、`name = 订阅.plugin_name` | 每台主机每个插件只取最新一条进程记录，且必须归属于对应主订阅 |
| `node_man_host` | `bk_host_id ∈ node_man_processstatus.bk_host_id`，且 `bk_host_id > 0` | 补全业务、OS、CPU 架构；缺失主机的进程记录跳过 |
| 订阅作用域 | `node_type = INSTANCE` 且 `object_type = HOST` 时，`bk_host_id ∈ node_man_subscription.nodes` | 静态主机范围只保留订阅范围内的主机；动态拓扑范围不按此过滤 |
| `node_man_pluginconfigtemplate` | `is_main = true`、`os` 与主机匹配、`cpu_arch` 与主机匹配、`plugin_version = 进程版本` 或 `*` | 选择主配置模板；同名条件优先精确版本、正式版本、模板 ID 较大者 |
| `node_man_processstatus.configs` | `config.name = 主模板.name` | 只导出主配置文件 |
| `node_man_packages` + `node_man_proccontrol` | 按 `tenant_id + project + version + os + cpu_arch` 匹配，同键取 `node_man_packages.id` 最大的包 | 回填 start/stop 等进程控制命令 |

租户取值顺序为 `node_man_subscription.tenant_id` → `node_man_processstatus.tenant_id` → `node_man_host.tenant_id`，都为空时归入 `default`。

### 导入目标

| 目标集合 | 内容 |
| -------- | ---- |
| `process_v2_<tenant>` | 主机、业务、平台、进程状态/版本、identity、controller、monitor policy |
| `process_v2_config_<tenant>` | 主配置文件内容、md5、`file_path`、`custom_config_context` |

进程状态会归一化：源状态 `RUNNING` 写为 `running`，其他状态统一写为 `stopped`，确保所有导入记录都进入 nodemgr 的同步范围。导入使用 `bulkWrite + upsert`，只覆盖文件中出现的记录，不删除 bk-nodemgr 已有数据。

### 进程状态同步

每10分钟会执行一次同步任务，同步进程状态为 running 和 stopped 的插件信息