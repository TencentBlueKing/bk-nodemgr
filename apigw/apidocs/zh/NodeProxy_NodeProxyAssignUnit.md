### 描述

- 该接口提供版本：v3.0.1-alpha.20+.
- 该接口所需权限：networkunit_use_for_proxy（使用管控单元部署 Proxy）、proxy_operate（操作 Proxy）。
- 该接口功能描述：批量将未分配管控单元的运行中 Proxy 主机分配到指定管控单元，并启动分配工作流。

### URL

POST /api/v3/node/proxy/assign_unit

### 输入参数

| 参数名称          | 参数类型    | 必选 | 描述                                |
| ----------------- | ----------- | ---- | ----------------------------------- |
| bk_host_id        | int64 array | 是   | 待分配的 Proxy 主机ID列表，不能为空 |
| bk_networkunit_id | int64       | 是   | 目标管控单元ID，必须大于等于0       |

**参数说明**：

- 所有已找到的主机必须是运行中的 Proxy 主机，且尚未分配管控单元
- 所有已找到的主机必须属于同一网络区域，且该网络区域必须与目标管控单元的网络区域一致
- 重复的主机ID会自动去重
- 不存在的主机ID会记录为失败原因；符合条件的主机仍会提交分配工作流

### 调用示例

将3台 Proxy 主机分配到管控单元ID为5的管控单元。

```json
{
  "bk_host_id": [10001, 10002, 10003],
  "bk_networkunit_id": 5
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "success_count": 2,
    "failed_count": 1,
    "failed_reasons": ["host-id(10003) not found"],
    "workflow_id": "wf-20240101-001"
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
| data       | object   | 响应数据               |

#### data

| 参数名称       | 参数类型     | 描述                                                              |
| -------------- | ------------ | ----------------------------------------------------------------- |
| success_count  | int64        | 成功提交分配工作流的 Proxy 主机数量                               |
| failed_count   | int64        | 分配失败的 Proxy 主机数量，包括不存在的主机和工作流启动失败的主机 |
| failed_reasons | string array | 失败原因列表，每条说明具体的失败主机或工作流启动失败原因          |
| workflow_id    | string       | 分配工作流ID；工作流启动失败时为空                                |

**常见失败原因**：

- `host-id(xxx) not found`：指定的主机ID不存在
- `host-id(xxx) node role is xxx, not proxy`：主机不是 Proxy 节点
- `host-id(xxx) node status is xxx, not RUNNING`：Proxy 主机不是运行中状态
- `host-id(xxx) already assigned to networkunit-id(yyy)`：Proxy 主机已分配到其他管控单元
- `host-id(xxx) networkarea-id(xxx) does not match networkunit-id(xxx) networkarea-id(xxx)`：主机网络区域与目标管控单元网络区域不一致
- `failed to launch workflow: xxx`：分配工作流启动失败

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
