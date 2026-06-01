### 描述

- 该接口提供版本：v3.0.1-alpha.32+。
- 该接口所需权限：plugin_operate（操作插件）。
- 该接口功能描述：批量将指定主机上的插件进程从 v2 迁移到当前插件工作流。

### URL

POST /api/v3/plugin/migrate_from_v2

### 输入参数

| 参数名称 | 参数类型     | 必选 | 描述                             |
| -------- | ------------ | ---- | -------------------------------- |
| plugin   | object array | 是   | 待迁移插件列表，至少包含一个元素 |

#### plugin[n]

| 参数名称    | 参数类型 | 必选 | 描述       |
| ----------- | -------- | ---- | ---------- |
| bk_host_id  | int64    | 是   | 目标主机ID |
| plugin_name | string   | 是   | 插件名称   |

**参数说明**：

- `plugin`：至少需要传入一个待处理对象，否则请求会校验失败。
- `bk_host_id`：主机ID 必须为正整数。
- `plugin_name`：不能为空。

### 调用示例

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent"
    },
    {
      "bk_host_id": 1002,
      "plugin_name": "bk-log-collector"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260601-000001",
  "error": null,
  "data": {
    "workflow_id": "plugin-migrate-v2-workflow-123456"
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                   |
| ---------- | -------- | ---------------------- |
| code       | int32    | 状态码，0表示成功      |
| message    | string   | 请求信息               |
| request_id | string   | 请求ID                 |
| error      | object   | 错误信息，成功时为null |
| permission | object   | 权限信息               |
| data       | object   | 响应数据               |

#### data

| 参数名称    | 参数类型 | 描述                                 |
| ----------- | -------- | ------------------------------------ |
| workflow_id | string   | 工作流ID，可用于查询插件迁移任务状态 |

#### error

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| system   | string   | 错误系统标识 |
| message  | string   | 错误消息     |
| details  | array    | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| code     | string   | 错误代码 |
| message  | string   | 错误消息 |

#### permission

| 参数名称    | 参数类型 | 描述         |
| ----------- | -------- | ------------ |
| system      | string   | 权限系统 ID  |
| system_name | string   | 权限系统名称 |
| apply_url   | string   | 权限申请链接 |
| actions     | array    | 关联动作列表 |

#### permission.actions[n]

| 参数名称               | 参数类型 | 描述         |
| ---------------------- | -------- | ------------ |
| id                     | string   | 动作 ID      |
| name                   | string   | 动作名称     |
| related_resource_types | array    | 关联资源类型 |
