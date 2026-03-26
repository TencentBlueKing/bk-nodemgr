### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：networkunit_use_for_agent（使用管控单元部署 Agent）、agent_operate（操作 Agent）。
- 该接口功能描述：批量将未分配网络单元的主机分配到指定网络单元（仅修改元数据，不执行远程操作）。

### URL

POST /api/v3/node/agent/assign_unit

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 array | 是 | 待分配的主机ID列表，不能为空 |
| bk_networkunit_id | int64 | 是 | 目标网络单元ID，必须大于等于0 |

**参数说明**：
- 所有主机必须属于同一网络区域，且该网络区域必须与目标网络单元的网络区域一致
- 已分配网络单元的主机将跳过并记录为失败
- 重复的主机ID会自动去重

### 调用示例

将3台主机分配到网络单元ID为5的网络单元。

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
    "failed_reasons": [
      "host-id(10003) already assigned to networkunit-id(3)"
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求ID |
| error | object | 错误信息，成功时为null |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| success_count | int64 | 成功分配的主机数量 |
| failed_count | int64 | 分配失败的主机数量 |
| failed_reasons | string array | 失败原因列表，每条说明具体的失败主机和原因 |

**常见失败原因**：
- `host-id(xxx) not found`：指定的主机ID不存在
- `host-id(xxx) already assigned to networkunit-id(yyy)`：主机已分配到其他网络单元

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误系统标识 |
| message | string | 错误消息 |
| details | array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误代码 |
| message | string | 错误消息 |
