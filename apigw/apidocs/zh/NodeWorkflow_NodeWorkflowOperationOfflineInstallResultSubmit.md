### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：无。
- 该接口功能描述：提交离线安装结果数据，唤醒并继续执行对应的离线安装任务步骤。

### URL

POST /api/v3/node/workflow/operation/offline/result

### 输入参数

| 参数名称         | 参数类型   | 必选 | 描述                                          |
|--------------|--------|----|---------------------------------------------|
| operation_id | string | 是  | 操作 ID，对应离线安装 operation                      |
| result_data  | string | 是  | 离线安装结果 JSON 字符串，通常为 installer.data.json 的内容 |

### 调用示例

```json
{
  "operation_id": "oper-20260421-0001",
  "result_data": "{\"status\":\"success\",\"message\":\"offline install completed\",\"instance_id\":\"inst-123\"}"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": null
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
| data       | object | 响应数据，当前为空    |

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
