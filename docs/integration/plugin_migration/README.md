# plugin_migration 接入指南

本文面向需要通过 bk-nodemgr backend API 把主机上的 V2 插件进程迁移为 V3 插件进程的第三方平台开发者或运维人员。

完整流程由本文统一描述；`list_v2_processes`、`list_v3_processes`、`sync_v2_alive_process_info`、`sync_v2_stopped_process_info`、`install_v3_no_start`、`stop_v2`、`start_plugin` 七篇文档只描述各自的接口契约。

## 接入结果

迁移完成后，目标主机应达到：

- V3 插件包和配置已安装。
- V2 插件进程已停止。
- V3 插件进程已启动并处于 `running`，`auto_start=true`。

如果当前是蓝鲸 7.2 升级到 7.3、V2 插件进程记录仍在 NodeMan，必须先完成 V2 进程数据迁移；否则 `process/list_v2` 可能返回空，`stop_v2` 也没有可停止的 V2 记录。

## 完整流程

| 步骤 | 操作                       | 接口或方式                                                                      | 即时输出                           |
| ---- | -------------------------- | ------------------------------------------------------------------------------- | ---------------------------------- |
| 0    | 迁移 V2 进程数据           | 离线导出/导入，见下文                                                           | 无 API 响应、无 workflow_id        |
| 1    | 查询 V2 进程               | `POST /api/v3/process/list_v2`                                                  | `data.total` / `data.items[]`      |
| 2    | 同步 V2 进程状态并重新查询 | `sync/gse/plugin/process_v2/info/alive` + `.../stopped`；再查 `process/list_v2` | `data.trigger_id` / `data.items[]` |
| 3    | 安装 V3 插件但不启动       | `POST /api/v3/plugin/install_no_start`                                          | `data.workflow_id`                 |
| 4    | 停止 V2 插件进程           | `POST /api/v3/plugin/stop_v2`                                                   | `data.workflow_id`                 |
| 5    | 查询 V2 状态               | `POST /api/v3/process/list_v2`                                                  | `data.items[]`                     |
| 6    | 启动 V3 插件进程           | `POST /api/v3/plugin/start`                                                     | `data.workflow_id`                 |
| 7    | 查询 V3 状态               | `POST /api/v3/process/list`                                                     | `data.items[]`                     |
| 8    | 读取 workflow 执行结果     | `plugin/workflow/*`                                                             | workflow / operation / 日志        |

> 步骤 4 到步骤 6 之间，该批主机不上报数据，属于预期停机窗口；确认 V2 停止后应尽快启动 V3。

## 0. 迁移 V2 进程数据（运维前置）

完整筛选条件和执行说明见 [NodeMan 插件进程迁移手册](../../operation/migrate/nodeman_process_to_nodemgr/README.md)。

适用场景：

- 蓝鲸 7.2 升级至 7.3。
- 从 NodeMan 迁移到 NodeMgr。
- 当前只回填 NodeMan 安装的主插件进程和主配置。

### 0.1 在 NodeMan pod 内导出

把 `export_nodeman_process.py` 放到 bk-nodeman backend-api pod 的 `/app/apps/node_man/management/commands/`，执行：

```bash
python manage.py export_nodeman_process
```

导出文件保存在 NodeMan pod 的 `/tmp/nodeman_export/`，按 `tenant/biz/<plugin_name>` 组织，并生成 `import_selected.js` 批量导入清单。

### 0.2 拷贝到 NodeMgr pod 并导入

把整个 `/tmp/nodeman_export/` 目录拷贝到 bk-nodemgr backend pod，然后执行：

```bash
cd /bk-nodemgr/script/ops
./mongo.sh login
```

导入前确认数据库：

```javascript
db.getName();
```

批量导入：

```javascript
load("/tmp/nodeman_export/import_selected.js");
```

单独导入：

```javascript
load("/tmp/nodeman_export/cw/2/bkmonitorbeat.js");
```

### 0.3 数据迁移结果

| 产物                         | 说明                                                                  |
| ---------------------------- | --------------------------------------------------------------------- |
| `process_v2_<tenant>` 记录   | 主机、业务、平台、进程状态/版本、identity、controller、monitor policy |
| `process_v2_config_<tenant>` | 主配置内容、md5、`file_path`、`custom_config_context`                 |
| 状态归一化                   | 源 `RUNNING` 写为 `running`，其他状态统一写为 `stopped`               |

导入使用 `bulkWrite + upsert`，只覆盖文件中出现的记录，不删除 NodeMgr 已有数据；同一份文件重复执行结果一致。导入后 nodemgr 会按每 10 分钟一次的同步任务继续同步 `running` 和 `stopped` 的进程信息。

数据迁移完成后，继续执行步骤 1。

## 1. 查询 V2 进程

调用 `POST /api/v3/process/list_v2` 查询 V2 兼容进程记录。接口契约和完整字段见 [list_v2_processes](list_v2_processes.md)。

```json
{
  "page": { "offset": 0, "limit": 100 },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "plugin_name": ["bkunifylogbeat"],
    "status": ["running"]
  }
}
```

把 `status=running` 的主机纳入迁移候选集合。接口分页，单页 `limit` 取值 `1`–`500`；数据量大时逐页拉取。

## 2. 同步 V2 进程状态

数据迁移后，`process_v2` 记录可能还未与 GSE 实际状态对齐。对步骤 1 得到的主机 ID，分别触发存活和已停止进程同步：

- `POST /api/v3/sync/gse/plugin/process_v2/info/alive`：同步 `running` 以及 `unknown + auto_start=true` 的记录，见 [sync_v2_alive_process_info](sync_v2_alive_process_info.md)。
- `POST /api/v3/sync/gse/plugin/process_v2/info/stopped`：同步 `stopped` 记录，见 [sync_v2_stopped_process_info](sync_v2_stopped_process_info.md)。

```json
{
  "host_ids": [1001, 1002]
}
```

两个接口都返回 `data.trigger_id`。`trigger_id` 只表示同步工作流已创建并激活，不表示同步已完成。同步完成后重新调用 `process/list_v2`，以其结果为准，再进入安装 V3 步骤。

## 3. 安装 V3 插件但不启动

调用 `POST /api/v3/plugin/install_no_start`。该接口只下发并安装 V3 插件包和配置，不启动进程；若 V3 进程正在运行，会先停止该进程，任务结束后进程记录为 `stopped`。

接口契约和完整字段见 [install_v3_no_start](install_v3_no_start.md)。

```json
{
  "plugin": [{ "bk_host_id": 1001, "plugin_name": "bkunifylogbeat" }]
}
```

`version` 建议省略，由服务端使用该插件当前默认发布版本；如果固定版本，必须确认所有目标平台都有该版本。V2 的自定义参数不会自动继承，需要通过 `custom_config_context` 显式传入。

保存响应中的 `data.workflow_id`，等待 workflow 成功。

## 4. 停止 V2 插件

确认 V3 已安装且进程为 `stopped` 后，调用 `POST /api/v3/plugin/stop_v2`。接口契约和完整字段见 [stop_v2](stop_v2.md)。

```json
{
  "plugin": [{ "bk_host_id": 1001, "plugin_name": "bkunifylogbeat" }]
}
```

保存响应中的 `data.workflow_id`，等待 workflow 成功。

## 5. 查询 V2 状态

停止 workflow 成功后，再次调用 `POST /api/v3/process/list_v2`：

```json
{
  "page": { "offset": 0, "limit": 20 },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "bk_host_id": [1001],
    "plugin_name": ["bkunifylogbeat"]
  }
}
```

确认 `process_info.status` 不再是 `running`。如果仍为 `running` 或 `unknown`，不要启动 V3；先查看 `stop_v2` workflow 日志或重试。

## 6. 启动 V3 插件

确认 V2 已停止后，立即调用 `POST /api/v3/plugin/start`。接口契约和完整字段见 [start_plugin](start_plugin.md)。

```json
{
  "plugin": [{ "bk_host_id": 1001, "plugin_name": "bkunifylogbeat" }]
}
```

保存响应中的 `data.workflow_id`，等待 workflow 成功。

## 7. 查询 V3 状态

启动 workflow 成功后，调用 `POST /api/v3/process/list`。接口契约和完整字段见 [list_v3_processes](list_v3_processes.md)。

```json
{
  "page": { "offset": 0, "limit": 20 },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "bk_host_id": [1001],
    "plugin_name": ["bkunifylogbeat"]
  }
}
```

确认：

- `process_info.status = running`。
- `process_info.auto_start = true`。
- `process_info.version` 符合本次安装预期。

## 8. 读取 workflow 执行结果

`install_no_start`、`stop_v2`、`start` 各自返回独立的 `workflow_id`。读取路径：

1. `POST /api/v3/plugin/workflow/list`：按 `exact_include_conditions.workflow_id` 查询整体状态。
2. `POST /api/v3/plugin/workflow/operation/list`：按 `workflow_id` 查询每主机 operation。
3. `POST /api/v3/plugin/workflow/operation/instance/list`：按 `operation_id` 查询执行实例。
4. `POST /api/v3/plugin/workflow/operation/instance/log/get`：按 `oper_inst_id` 查询实例日志。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export MIGRATION_WORKFLOW_ID="${MIGRATION_WORKFLOW_ID}"  # 上一步返回的 workflow_id

WORKFLOW_LIST="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/list" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "page": {"offset": 0, "limit": 20},
    "exact_include_conditions": {
      "workflow_id": ["${MIGRATION_WORKFLOW_ID}"]
    }
  }
EOF
)"

WORKFLOW_STATUS="$(printf '%s' "${WORKFLOW_LIST}" | jq -r '.data.items[0].status')"
```

整体 `status` 的终态取值：

| 终态值           | 含义                       |
| ---------------- | -------------------------- |
| `success`        | 全部执行单位成功结束       |
| `failed`         | 全部执行单位失败结束       |
| `partial_failed` | 部分主机成功、部分主机失败 |

多主机时，每个 operation 对应一台主机；各 operation 独立读日志、独立判定。workflow 成功仍需通过步骤 5 和步骤 7 的进程状态查询确认。

同步 V2 进程状态的两个接口返回的是 `trigger_id`，不是 `workflow_id`；它们的同步结果以 `process/list_v2` 的查询结果为准。

## 失败与边界

- 顺序不能反：先启动 V3 再停止 V2 会出现 V2 与 V3 同时上报；停止 V2 后不启动 V3 会出现数据断档。
- 未完成 V2 进程数据迁移时，`process/list_v2` 可能没有 V2 记录，`stop_v2` 也没有可停止的 V2 数据。
- `process/list_v2` 是 V2 进程列表接口，不是迁移专用接口；它只返回查询时可见的 V2 兼容进程记录。
- 同步 V2 进程状态的接口返回 `trigger_id`，只表示同步工作流已创建；同步结果需重新查询 `process/list_v2`。
- 同步 V2 进程状态的 `host_ids` 至少包含一个主机 ID。
- `stop_v2` 和 `start` 都是异步 workflow；成功响应只表示任务已提交，不证明进程状态已经改变。
- `install_no_start` 只负责 V3 下发和安装，不负责启动。
- `stop_v2` 只停止 V2 进程，不卸载 V2 二进制、配置或运行目录。
- 公开 contract 不定义迁移请求的幂等键、重复调用合并行为、单批上限、失败重试窗口或推荐轮询间隔。
- `plugin_name` 必须在 V2 查询、V3 安装、V2 停止、V3 启动四个步骤中保持一致。

## Contract 参考

- [list_v2_processes](list_v2_processes.md)
- [list_v3_processes](list_v3_processes.md)
- [sync_v2_alive_process_info](sync_v2_alive_process_info.md)
- [sync_v2_stopped_process_info](sync_v2_stopped_process_info.md)
- [install_v3_no_start](install_v3_no_start.md)
- [stop_v2](stop_v2.md)
- [start_plugin](start_plugin.md)
- [NodeMan 插件进程迁移手册](../../operation/migrate/nodeman_process_to_nodemgr/README.md)
- [Plugin Swagger](../../api/swagger/backend/api/v3/plugin.swagger.json)
- [Process Swagger](../../api/swagger/backend/api/v3/process.swagger.json)
- [Sync Swagger](../../api/swagger/backend/api/v3/sync.swagger.json)
- [Plugin workflow Swagger](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json)
- [Plugin Proto](../../../proto/backend/api/v3/plugin.proto)
- [Process Proto](../../../proto/backend/api/v3/process.proto)
- [Sync Proto](../../../proto/backend/api/v3/sync.proto)
