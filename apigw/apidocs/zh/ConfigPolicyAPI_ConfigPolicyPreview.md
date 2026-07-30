### 描述

- 该接口提供版本：v3.0.1-alpha.63+。
- 该接口所需权限：config_policy_view（查看配置策略）。
- 该接口功能描述：按业务、策略类型和主机列表预览配置策略匹配结果与配置合并结果。

### URL

POST /api/v3/policy/config/preview

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                                                                       |
| ----------- | -------- | ---- | ------------------------------------------------------------------------------------------ |
| bk_biz_id   | int64    | 是   | 业务 ID                                                                                    |
| policy_type | string   | 是   | 配置策略类型，可选值：`config_policy_agent`、`config_policy_proxy`、`config_policy_plugin` |
| plugin_name | string   | 否   | 插件名称；当 `policy_type` 为 `config_policy_plugin` 时必填                                |
| hosts       | array    | 是   | 预览目标主机列表，至少一个元素                                                             |

#### hosts[n]

| 参数名称          | 参数类型 | 必选 | 描述                                   |
| ----------------- | -------- | ---- | -------------------------------------- |
| bk_host_id        | int64    | 是   | 主机 ID，必须大于 `0`                  |
| bk_networkunit_id | int64    | 否   | 管控单元 ID，未传时协议层自动补为 `-1` |
| bk_networkarea_id | int64    | 否   | 管控区域 ID，未传时协议层自动补为 `-1` |
| os_type           | string   | 否   | 操作系统类型                           |
| cpu_arch          | string   | 否   | CPU 架构                               |

### 调用示例

```json
{
  "bk_biz_id": 2,
  "policy_type": "config_policy_plugin",
  "plugin_name": "bkmonitorbeat",
  "hosts": [
    {
      "bk_host_id": 1001,
      "bk_networkunit_id": 2001,
      "bk_networkarea_id": 1,
      "os_type": "linux",
      "cpu_arch": "x86_64"
    },
    {
      "bk_host_id": 1002,
      "os_type": "linux",
      "cpu_arch": "x86_64"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123464",
  "data": {
    "reliable_items": [
      {
        "bk_host_id": 1001,
        "matched_policies": [
          {
            "configpolicy_id": 10001,
            "configpolicy_name": "prod-plugin-config",
            "priority": 1
          },
          {
            "configpolicy_id": 10003,
            "configpolicy_name": "fallback-plugin-config",
            "priority": 2
          }
        ],
        "merged_config": "{\"plugin\":{\"base\":{\"cpu_percent_limit\":10,\"mem_percent_limit\":10}}}"
      }
    ],
    "unreliable_items": [
      {
        "bk_host_id": 1002,
        "matched_policies": [],
        "merged_config": "{}"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| code       | int32    | 状态码，`0` 表示成功     |
| message    | string   | 请求信息                 |
| request_id | string   | 请求 ID                  |
| error      | object   | 错误信息，成功时通常为空 |
| permission | object   | 权限信息                 |
| data       | object   | 预览结果                 |

#### data

| 参数名称         | 参数类型 | 描述                       |
| ---------------- | -------- | -------------------------- |
| reliable_items   | array    | 匹配结果可靠的主机预览项   |
| unreliable_items | array    | 匹配结果不可靠的主机预览项 |

#### data.reliable_items[n] / data.unreliable_items[n]

| 参数名称         | 参数类型 | 描述                                                 |
| ---------------- | -------- | ---------------------------------------------------- |
| bk_host_id       | int64    | 主机 ID                                              |
| matched_policies | array    | 命中的策略列表，按优先级顺序                         |
| merged_config    | string   | 合并后的嵌套 JSON 配置，键按字典序排列，值为原始类型 |

#### data.reliable_items[n].matched_policies[n]

| 参数名称          | 参数类型 | 描述         |
| ----------------- | -------- | ------------ |
| configpolicy_id   | int64    | 配置策略 ID  |
| configpolicy_name | string   | 配置策略名称 |
| priority          | int64    | 策略优先级   |
