### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：proxy_operate（操作Proxy）。
- 该接口功能描述：批量检查主机是否可以安装Proxy，并返回每个主机的安装前置状态。

### URL

POST /api/v3/node/proxy/install_check

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|---------|------|------|
| host | array | 是 | 主机列表，详见下方 host 参数说明 |

**host[n]**

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|---------|------|------|
| bk_networkunit_id | int64 | 是 | 网络单元ID |
| bk_host_id | int64 | 否 | 主机ID，-1 表示新注册主机 |
| bk_host_innerip_list | array[string] | 否 | 内网IPv4地址列表 |
| bk_host_innerip_v6_list | array[string] | 否 | 内网IPv6地址列表 |
| bk_biz_id | int64 | 否 | 业务ID |

### 调用示例

```json
{
    "host": [
        {
            "bk_networkunit_id": 1,
            "bk_host_id": 1001,
            "bk_host_innerip_list": ["10.0.0.1"],
            "bk_host_innerip_v6_list": [],
            "bk_biz_id": 2
        },
        {
            "bk_networkunit_id": 1,
            "bk_host_id": -1,
            "bk_host_innerip_list": ["10.0.0.2"],
            "bk_biz_id": 2
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
        "result": [
            {
                "bk_host_id": 1001,
                "status": "normal_install",
                "matched": {
                    "bk_host_id": 1001,
                    "bk_networkunit_id": 1,
                    "bk_host_innerip_list": ["10.0.0.1"],
                    "bk_host_innerip_v6_list": []
                }
            },
            {
                "bk_host_id": -1,
                "status": "register_to_cmdb_and_install",
                "matched": null
            }
        ]
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
| result | array | 检查结果列表，与请求 host 列表顺序对应 |

**data.result[n]**

| 参数名称 | 参数类型 | 描述 |
|---------|---------|------|
| bk_host_id | int64 | 主机ID，新注册主机为 -1 |
| status | string | 检查状态，详见下方状态说明 |
| matched | object | 匹配到的已有主机信息，无匹配时为 null |

**data.result[n].status 取值说明**

| 取值 | 描述 |
|------|------|
| register_to_cmdb_and_install | 需先注册至CMDB再安装 |
| normal_install | 可直接安装 |
| host_not_found | 主机不存在 |
| network_unit_not_found | 网络单元不存在 |
| duplicated_inner_ip | 内网IPv4地址重复 |
| duplicated_inner_ip_v6 | 内网IPv6地址重复 |
| mismatched_inner_ip | 内网IPv4地址不匹配 |
| mismatched_inner_ip_v6 | 内网IPv6地址不匹配 |

**data.result[n].matched**

| 参数名称 | 参数类型 | 描述 |
|---------|---------|------|
| bk_host_id | int64 | 匹配主机ID |
| bk_networkunit_id | int64 | 匹配主机所属网络单元ID |
| bk_host_innerip_list | array[string] | 匹配主机内网IPv4地址列表 |
| bk_host_innerip_v6_list | array[string] | 匹配主机内网IPv6地址列表 |

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
