### 描述

- 该接口提供版本：v3.0.1-alpha.120+。
- 该接口所需权限：system_access（系统访问）、plugin_history_view（查看插件历史）、plugin_view（查看插件）；同时校验历史任务流涉及的所有业务的两项插件权限。
- 该接口功能描述：读取选中历史插件安装操作可恢复的参数，用于执行诊断，不执行安装、重试或修改历史任务。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。

**使用边界**：

- 仅支持 `install_plugin` 工作流，不限制操作执行状态。可通过[操作列表接口](OpenPluginWorkflow_OpenPluginWorkflowOperationList.md)取得操作 ID。
- 参数来自历史部署记录，不是原始请求的完整快照。记录过期或丢失时，该操作 ID 放入 `unavailable_operation_ids`，不出现在 `items` 中。
- `missing_fields` 标识无法恢复的字段。自定义配置原样返回，不包含系统生成的配置上下文、渲染结果或部署 Token。请妥善处理自定义配置中的业务敏感信息。
- 历史参数不代表主机、插件版本及配置当前仍然有效。历史记录不能证明安装是否成功，需要结合执行实例和步骤日志判断。

### URL

POST /api/v3/open/plugin/workflow/operation/install_params/get

### 输入参数

| 参数名称      | 参数类型     | 必选 | 描述                                                                                      |
| ------------- | ------------ | ---- | ----------------------------------------------------------------------------------------- |
| workflow_id   | string       | 是   | 当前租户的历史安装工作流 ID，不能为空或全为空白字符                                       |
| operation_ids | string array | 是   | 同一工作流下的操作 ID，列表非空、元素非空且不重复；不存在或不属于工作流的操作会使请求失败 |

### 调用示例

```json
{
  "workflow_id": "wf-plugin-0001",
  "operation_ids": ["op-plugin-0001", "op-plugin-0002"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "type": "install_plugin",
    "items": [
      {
        "operation_id": "op-plugin-0001",
        "bk_host_id": 1001,
        "plugin_name": "bkmonitorbeat",
        "version": "3.1.0",
        "custom_config_context": { "interval": 60 },
        "custom_spec": null,
        "missing_fields": []
      }
    ],
    "unavailable_operation_ids": ["op-plugin-0002"]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述               |
| ---------- | -------- | ------------------ |
| code       | int32    | 状态码，0 表示成功 |
| message    | string   | 请求结果信息       |
| request_id | string   | 请求 ID            |
| data       | object   | 恢复结果           |

#### data

| 参数名称                  | 参数类型     | 描述                                |
| ------------------------- | ------------ | ----------------------------------- |
| type                      | string       | 工作流类型，当前为 `install_plugin` |
| items                     | object array | 按请求顺序返回的可恢复参数          |
| unavailable_operation_ids | string array | 无可用部署记录的操作 ID             |

#### items[]

| 参数名称              | 参数类型     | 描述                                                 |
| --------------------- | ------------ | ---------------------------------------------------- |
| operation_id          | string       | 历史操作 ID                                          |
| bk_host_id            | int64        | 历史安装目标主机 ID                                  |
| plugin_name           | string       | 插件名称                                             |
| version               | string       | 历史部署记录中保存的版本                             |
| custom_config_context | object       | 历史自定义配置参数，可为空                           |
| custom_spec           | object       | 历史自定义部署规格，未指定时为空                     |
| missing_fields        | string array | 无法恢复的字段名称；不应将这些字段的空值视为历史选择 |

#### custom_spec

| 参数名称       | 参数类型 | 描述               |
| -------------- | -------- | ------------------ |
| resource       | object   | 自定义资源限制     |
| monitor_policy | object   | 自定义进程监控策略 |

#### resource

| 参数名称          | 参数类型 | 描述                 |
| ----------------- | -------- | -------------------- |
| cpu_limit_percent | double   | CPU 使用率限制百分比 |
| mem_limit_percent | double   | 内存使用率限制百分比 |

#### monitor_policy

| 参数名称                | 参数类型 | 描述                         |
| ----------------------- | -------- | ---------------------------- |
| restart_type            | string   | 重启方式：`auto` 或 `manual` |
| start_check_seconds     | int64    | 启动检查时长（秒）           |
| stop_check_seconds      | int64    | 停止检查时长（秒）           |
| operate_timeout_seconds | int64    | 操作超时时长（秒）           |
