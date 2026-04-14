### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：config_policy_manage（管理配置策略）。
- 该接口功能描述：更新指定配置策略。

### URL

POST /api/v3/policy/config/update

### 输入参数

| 参数名称              | 参数类型        | 必选 | 描述                                                                            |
|-------------------|-------------|----|-------------------------------------------------------------------------------|
| configpolicy_id   | int64       | 是  | 配置策略 ID                                                                       |
| configpolicy_name | string      | 是  | 配置策略名称                                                                        |
| configpolicy_type | string      | 是  | 配置策略类型，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| bk_biz_id         | int64       | 是  | 业务 ID                                                                         |
| remark            | string      | 否  | 备注                                                                            |
| scopes            | array       | 是  | 策略生效范围列表                                                                      |
| configs_string    | object      | 否  | 字符串类型配置键值对                                                                    |
| configs_int       | object      | 否  | 整型配置键值对                                                                       |
| configs_bool      | object      | 否  | 布尔配置键值对                                                                       |
| enabled           | bool        | 是  | 是否启用                                                                          |
| operator          | string      | 否  | 操作人                                                                           |
| priority          | int64       | 是  | 优先级，必须大于 `0`                                                                  |
| target_host_ids   | int64 array | 否  | 定向目标主机 ID 列表                                                                  |

#### scopes[n]

| 参数名称              | 参数类型   | 必选 | 描述                |
|-------------------|--------|----|-------------------|
| bk_networkarea_id | int64  | 否  | 管控区域 ID，`-1` 表示任意 |
| bk_networkunit_id | int64  | 否  | 管控单元 ID，`-1` 表示任意 |
| os_type           | string | 否  | 操作系统类型，空字符串表示任意   |
| cpu_arch          | string | 否  | CPU 架构，空字符串表示任意   |

### 调用示例

```json
{
  "configpolicy_id": 10001,
  "configpolicy_name": "prod-agent-config-v2",
  "configpolicy_type": "config_policy_agent",
  "bk_biz_id": 2,
  "remark": "生产环境 Agent 配置升级",
  "scopes": [
    {
      "bk_networkarea_id": -1,
      "bk_networkunit_id": -1,
      "os_type": "",
      "cpu_arch": ""
    }
  ],
  "configs_string": {
    "bk_cloud_id": "0"
  },
  "configs_int": {
    "heartbeat_interval": 30
  },
  "configs_bool": {
    "enable_metrics": true
  },
  "enabled": true,
  "operator": "admin",
  "priority": 1,
  "target_host_ids": [
    1001,
    1002
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123459",
  "data": {
    "configpolicy_id": 10001
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

#### data

| 参数名称            | 参数类型  | 描述           |
|-----------------|-------|--------------|
| configpolicy_id | int64 | 更新成功的配置策略 ID |
