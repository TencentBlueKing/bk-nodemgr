### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：proxy_operate（操作Proxy）。
- 该接口功能描述：批量卸载节点Proxy。

### URL

POST /api/v3/node/proxy/uninstall

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|---------|------|------|
| host | array | 是 | 主机列表，详见下方 host 参数说明 |

**host[n]**

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|---------|------|------|
| bk_host_id | int64 | 是 | 主机ID |

### 调用示例

```json
{
    "host": [
        {
            "bk_host_id": 1001
        }
    ]
}
```

### 响应示例

```json
{
    "code": 0,
    "message": "ok",
    "request_id": "abc123",
    "data": {
        "workflow_id": "wf-20240101-005"
    }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|---------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求ID |
| error | object | 错误信息，成功时为 null |
| data | object | 响应数据 |

**data**

| 参数名称 | 参数类型 | 描述 |
|---------|---------|------|
| workflow_id | string | 工作流ID |

**error**

| 参数名称 | 参数类型 | 描述 |
|---------|---------|------|
| system | string | 错误来源系统 |
| message | string | 错误信息 |
| details | array | 详细错误列表 |

**error.details[n]**

| 参数名称 | 参数类型 | 描述 |
|---------|---------|------|
| code | string | 错误码 |
| message | string | 错误详情 |