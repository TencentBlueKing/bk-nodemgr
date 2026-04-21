### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_view（查看 Agent）、proxy_view（查看 Proxy）。
- 该接口功能描述：按任务流 ID 维度返回节点任务流的操作实例状态统计结果。

### URL

POST /api/v3/node/workflow/statistics

### 输入参数

| 参数名称        | 参数类型         | 必选 | 描述        |
|-------------|--------------|----|-----------|
| workflow_id | string array | 否  | 任务流 ID 列表 |

**权限说明**：

- 接口请求不包含 `node_role` 过滤条件，后端按安全策略同时校验 `agent_view` 和 `proxy_view`。

### 调用示例

统计两个任务流的状态分布。

```json
{
  "workflow_id": [
    "wf-20240601-0001",
    "wf-20240601-0002"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": [
      {
        "workflow_id": "wf-20240601-0001",
        "total_count": 10,
        "init_count": 1,
        "launched_count": 1,
        "running_count": 2,
        "success_count": 4,
        "failed_count": 1,
        "timeout_count": 1,
        "terminated_count": 0
      },
      {
        "workflow_id": "wf-20240601-0002",
        "total_count": 0,
        "init_count": 0,
        "launched_count": 0,
        "running_count": 0,
        "success_count": 0,
        "failed_count": 0,
        "timeout_count": 0,
        "terminated_count": 0
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述           |
|------------|--------|--------------|
| code       | int32  | 状态码，`0` 表示成功 |
| message    | string | 请求信息         |
| request_id | string | 请求 ID        |
| error      | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息         |
| data       | object | 响应数据         |

#### error

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| system  | string | 错误系统标识 |
| message | string | 错误消息   |
| details | array  | 错误详情列表 |

#### error.details[n]

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | string | 错误代码 |
| message | string | 错误消息 |

#### permission

| 参数名称        | 参数类型   | 描述      |
|-------------|--------|---------|
| system      | string | 权限系统 ID |
| system_name | string | 权限系统名称  |
| apply_url   | string | 权限申请链接  |
| actions     | array  | 关联动作列表  |

#### permission.actions[n]

| 参数名称                   | 参数类型   | 描述     |
|------------------------|--------|--------|
| id                     | string | 动作 ID  |
| name                   | string | 动作名称   |
| related_resource_types | array  | 关联资源类型 |

#### data

| 参数名称  | 参数类型  | 描述      |
|-------|-------|---------|
| items | array | 任务流统计列表 |

#### data.items[n]

| 参数名称             | 参数类型   | 描述                         |
|------------------|--------|----------------------------|
| workflow_id      | string | 任务流 ID                     |
| total_count      | int64  | 总实例数，等于各状态实例数之和，并包含未初始化实例数 |
| init_count       | int64  | `init` 状态实例数               |
| launched_count   | int64  | `launched` 状态实例数           |
| running_count    | int64  | `running` 状态实例数            |
| success_count    | int64  | `success` 状态实例数            |
| failed_count     | int64  | `failed` 状态实例数             |
| timeout_count    | int64  | `timeout` 状态实例数            |
| terminated_count | int64  | `terminated` 状态实例数         |
