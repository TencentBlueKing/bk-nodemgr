### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：无。
- 该接口功能描述：查询主机字段去重结果，支持按主机条件进行过滤。

### URL

POST /api/v3/topo/host/distinct

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 array | 否 | 主机 ID |
| bk_biz_id | int64 array | 否 | 业务 ID |
| bk_networkarea_id | int64 array | 否 | 管控区域 ID |
| os_type | string array | 否 | 操作系统类型 |
| node_role | string array | 否 | 节点角色 |
| node_status | string array | 否 | node_status 字段 |
| node_version | string array | 否 | 节点版本 |
| bk_agent_id | string array | 否 | bk_agent_id 字段 |
| bk_networkunit_id | int64 array | 否 | 管控单元 ID |
| node_generation | int64 array | 否 | node_generation 字段 |
| arch | string array | 否 | arch 字段 |
| proxy_tags | string array | 否 | proxy_tags 字段 |
| bk_set_id | int64 array | 否 | bk_set_id 字段 |
| bk_module_id | int64 array | 否 | bk_module_id 字段 |

#### fuzzy_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_name | string array | 否 | bk_host_name 字段 |
| dept_name | string array | 否 | dept_name 字段 |
| bk_host_innerip | string array | 否 | 主机内网 IPv4 地址 |
| bk_host_innerip_v6 | string array | 否 | 主机内网 IPv6 地址 |
| bk_host_outerip | string array | 否 | bk_host_outerip 字段 |
| bk_host_outerip_v6 | string array | 否 | bk_host_outerip_v6 字段 |

### 调用示例

```json
{
  "exact_include_conditions": {
    "bk_host_id": [
      1
    ],
    "bk_biz_id": [
      1
    ],
    "bk_networkarea_id": [
      1
    ],
    "os_type": [
      "linux"
    ],
    "node_role": [
      "agent"
    ],
    "node_status": [
      "running"
    ],
    "node_version": [
      "3.2.1"
    ],
    "bk_agent_id": [
      "id-001"
    ]
  },
  "fuzzy_include_conditions": {
    "bk_host_name": [
      "default"
    ],
    "dept_name": [
      "default"
    ],
    "bk_host_innerip": [
      "10.0.0.1"
    ],
    "bk_host_innerip_v6": [
      "2001:db8::1"
    ],
    "bk_host_outerip": [
      "10.0.0.1"
    ],
    "bk_host_outerip_v6": [
      "2001:db8::1"
    ]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "node_role": [
      "agent"
    ],
    "node_status": [
      "running"
    ],
    "node_version": [
      "3.2.1"
    ],
    "dept_name": [
      "default"
    ],
    "os_type": [
      "linux"
    ],
    "arch": [
      "string"
    ],
    "addressing": [
      "string"
    ],
    "bk_networkarea_id": [
      1
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| node_role | string array | 节点角色 |
| node_status | string array | node_status 字段 |
| node_version | string array | 节点版本 |
| dept_name | string array | dept_name 字段 |
| os_type | string array | 操作系统类型 |
| arch | string array | arch 字段 |
| addressing | string array | addressing 字段 |
| bk_networkarea_id | int64 array | 管控区域 ID |
| bk_networkunit_id | int64 array | 管控单元 ID |
| bk_biz_id | int64 array | 业务 ID |
