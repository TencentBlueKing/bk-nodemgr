# export_plugin

## 目的与适用场景

当第三方平台希望按插件包名称和版本启动一次插件包导出时，使用 `POST /api/v3/package/workflow/export/plugin`。

该接口的成功响应只确认导出 workflow 已创建。插件包是否已经准备好下载，需要后续使用 [export_result](export_result.md) 读取；具体 action / instance 日志需要使用 [read_export_logs](read_export_logs.md) 的四步组合链路读取。

## 输入

请求字段以 [pkg_workflow.proto](../../../proto/backend/api/v3/pkg_workflow.proto) 和 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json) 为准：

| 字段                 | Required | 含义               |
| -------------------- | -------- | ------------------ |
| `plugin_pkg_name`    | yes      | 要导出的插件包名称 |
| `plugin_pkg_version` | yes      | 要导出的插件包版本 |

## 最小 payload 和 curl template

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export PLUGIN_PKG_NAME="${PLUGIN_PKG_NAME}"
export PLUGIN_PKG_VERSION="${PLUGIN_PKG_VERSION}"

EXPORT_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/package/workflow/export/plugin" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "plugin_pkg_name": "${PLUGIN_PKG_NAME}",
  "plugin_pkg_version": "${PLUGIN_PKG_VERSION}"
}
EOF
)"

# 保存本次导出的 workflow 标识，供 export_result 和日志组合查询使用
WORKFLOW_ID="$(printf '%s' "${EXPORT_RESPONSE}" | jq -r '.data.workflow_id')"
```

## 即时输出

成功响应中的 `data.workflow_id` 是本次导出 workflow 的接入句柄。调用方应保存该值，并将其作为 `POST /api/v3/package/workflow/export_result` 的 `workflow_id`。

`workflow_id` 只表示 workflow 已创建，不能证明插件包已经完成导出，也不能证明已经产生可用的 `download_url`。

## 最终或机器侧可见产物

导出完成后的机器侧产物通过 [export_result](export_result.md) 读取：

- `data.is_finish` 表示 workflow 是否进入终态。
- `data.status` 表示 workflow 整体状态。
- 成功完成时，响应可能提供 `data.download_url`，调用方可使用该 URL 获取插件包。
- 若响应提供 `data.download_url_expired_at`，调用方应记录并关注该值。

启动接口本身不返回最终下载结果。具体流程日志也不在启动接口响应中；需要按 [read_export_logs](read_export_logs.md) 的四步组合接口读取。

## 重复行为

每次成功调用都会返回一个导出 workflow 的创建结果。公开 contract 不定义对相同 `plugin_pkg_name` + `plugin_pkg_version` 重复提交时的合并、替换、去重或幂等行为；调用方不应自行假定重复提交会复用已有 workflow。

## 失败情况与限制

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)。
- 缺少 `plugin_pkg_name` 或 `plugin_pkg_version`，或字段值不符合 contract 时，调用方应按 API 错误处理。
- 成功响应仅代表 workflow 已创建，不代表导出已完成。
- 公开 contract 不定义导出最大等待时长、推荐轮询间隔、重试窗口、回滚行为或重复提交的幂等保证。
- 不要根据启动响应调用 file service 内部 endpoint；下载 URL 仅从导出结果接口读取。

## Contract 参考

- [接入总览](README.md)
- [读取导出结果](export_result.md)
- [读取具体导出日志](read_export_logs.md)
- [Package workflow Swagger](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)
- [Package workflow Proto](../../../proto/backend/api/v3/pkg_workflow.proto)
