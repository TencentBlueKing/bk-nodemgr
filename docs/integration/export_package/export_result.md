# export_result

## 目的与适用场景

当第三方平台希望读取一次插件包导出 workflow 的整体状态，或在成功完成后获取插件包下载 URL 时，使用 `POST /api/v3/package/workflow/export_result`。

本接口只返回 workflow 整体结果和下载信息，**不能直接返回具体 action / instance 日志**。需要具体流程日志时，转到 [read_export_logs](read_export_logs.md)，按 `workflow/list` → `operation/list` → `operation/instance/list` → `log/get` 四步组合查询。

## 输入

请求字段以 [pkg_workflow.proto `PackageExportResultReq`](../../../proto/backend/api/v3/pkg_workflow.proto) 和 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json) 为准：

| 字段          | Required | 含义                                       |
| ------------- | -------- | ------------------------------------------ |
| `workflow_id` | yes      | `export_plugin` 成功响应返回的 workflow ID |

## 最小 payload 和 curl template

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export WORKFLOW_ID="${WORKFLOW_ID}"  # 来自 export_plugin 响应

RESULT_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/package/workflow/export_result" \
  -H "Content-Type: application/json" \
  -d "{\"workflow_id\": \"${WORKFLOW_ID}\"}")"

# workflow 整体状态
WORKFLOW_STATUS="$(printf '%s' "${RESULT_RESPONSE}" | jq -r '.data.status')"
IS_FINISH="$(printf '%s' "${RESULT_RESPONSE}" | jq -r '.data.is_finish')"

# 仅当响应提供下载信息时读取；字段缺失时 jq 返回 null
DOWNLOAD_URL="$(printf '%s' "${RESULT_RESPONSE}" | jq -r '.data.download_url // empty')"
DOWNLOAD_URL_EXPIRED_AT="$(printf '%s' "${RESULT_RESPONSE}" | jq -r '.data.download_url_expired_at // empty')"
```

## 即时输出

`data` 字段以 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json) 为准：

| 字段                           | 类型   | 含义                           |
| ------------------------------ | ------ | ------------------------------ |
| `data.status`                  | string | workflow 整体状态              |
| `data.is_finish`               | bool   | workflow 是否进入终态          |
| `data.download_url`            | string | 成功完成时可选的插件包下载 URL |
| `data.download_url_expired_at` | int64  | 可选的下载 URL 过期信息        |

字段的存在性和具体取值以实际响应及公开 contract 为准。成功响应或某次查询结果本身不应被解释为已经可以下载，调用方应结合 `data.is_finish`、`data.status` 和可用的下载字段判断下一步。

## 最终或机器侧可见产物

当 workflow 成功完成时，`data.download_url` 可作为插件包的机器侧下载入口；调用方应同时关注 `data.download_url_expired_at`，并在 URL 可用期间使用它。本文档不声明该字段的确切时间单位或 URL 的具体有效期。

`export_result` 的 `data.status` / `data.is_finish` 反映的是 workflow 整体状态，不提供具体 action / instance 日志。具体日志必须按 [read_export_logs](read_export_logs.md) 的四步组合接口读取。

## 重复行为

该接口用于读取指定 `workflow_id` 的导出结果；重复读取不会启动新的导出 workflow。公开 contract 不定义推荐轮询间隔、并发读取时的取值合并行为，调用方应自行决定读取策略。

## 失败情况与限制

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)。
- `workflow_id` 不存在或不属于可查询的导出 workflow 时，按 API 错误处理。
- `data.is_finish = false` 时，不应把缺少 `download_url` 解释为最终失败；workflow 尚未进入终态。
- 进入终态不等同于成功完成；调用方应结合 `data.status` 判定是否可以使用下载信息。
- 未提供 `data.download_url` 时，不要自行拼接下载地址，也不要调用 file service 内部 endpoint。
- 公开 contract 不定义轮询间隔、最大等待时长、重试窗口、回滚行为，以及 `download_url_expired_at` 的时间单位和具体有效期。

## Contract 参考

- [接入总览](README.md)
- [启动导出插件包](export_plugin.md)
- [读取具体导出日志](read_export_logs.md)
- [Package workflow Swagger](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)
- [Package workflow Proto](../../../proto/backend/api/v3/pkg_workflow.proto)
