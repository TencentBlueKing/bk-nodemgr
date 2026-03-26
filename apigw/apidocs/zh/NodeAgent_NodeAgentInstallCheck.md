### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：agent_operate（操作Agent）。
- 该接口功能描述：批量检查主机是否满足Agent安装条件，包括网络单元可用性、IP冲突检测、主机属性一致性校验等。

### URL

POST /api/v3/node/agent/install_check

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| host | object array | 是 | 待检查的主机信息列表 |

#### host[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_biz_id | int64 | 是 | 业务ID，必须大于等于0 |
| bk_host_id | int64 | 否 | 主机ID，-1或不填表示新主机（不在CMDB中） |
| bk_host_innerip_list | string array | 否 | 主机内网IPv4地址列表，与bk_host_innerip_v6_list至少填写一个 |
| bk_host_innerip_v6_list | string array | 否 | 主机内网IPv6地址列表，与bk_host_innerip_list至少填写一个 |
| bk_networkunit_id | int64 | 否 | 网络单元ID，不填时默认为-1 |

**参数说明**：
- `bk_host_innerip_list` 和 `bk_host_innerip_v6_list` 不能同时为空，列表中不允许包含空字符串
- 当 `bk_host_id` 为 -1 或未提供时，表示该主机尚未注册到 CMDB，系统将检查 IP 是否与已有主机冲突
- 当 `bk_host_id` >= 0 时，表示该主机已在 CMDB 中，系统将校验主机属性（业务ID、网络区域ID、IP 地址等）是否一致

### 调用示例

检查两台主机的安装条件：一台已在 CMDB 中的主机和一台新主机。

```json
{
  "host": [
    {
      "bk_biz_id": 100,
      "bk_host_id": 12345,
      "bk_host_innerip_list": ["10.0.0.1"],
      "bk_networkunit_id": 1
    },
    {
      "bk_biz_id": 100,
      "bk_host_id": -1,
      "bk_host_innerip_list": ["10.0.0.2"],
      "bk_networkunit_id": 1
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "results": [
      {
        "status": "normal_install",
        "message_en": "Install Agent",
        "message_zh": "安装Agent",
        "category": "normal_install",
        "matched": {
          "bk_host_id": 12345,
          "bk_biz_id": 100,
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "os_type": "linux",
          "node_role": "agent",
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": []
        }
      },
      {
        "status": "register_to_cmdb_and_install",
        "message_en": "Import node to CMDB and install Agent",
        "message_zh": "将节点导入CMDB并安装Agent",
        "category": "register_to_cmdb_and_install"
      }
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
| results | object array | 检查结果列表，与请求中host数组一一对应 |

#### data.results[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| status | string | 检查状态（枚举值见下方说明） |
| message_en | string | 检查结果英文描述 |
| message_zh | string | 检查结果中文描述 |
| category | string | 结果分类（枚举值见下方说明） |
| matched | object | 匹配到的已有主机信息，部分状态下为null |

**status 枚举说明**：

| 枚举值 | 描述 |
|-------|------|
| normal_install | 正常安装，主机已在CMDB中且所有校验通过 |
| register_to_cmdb_and_install | 需要先注册到CMDB再安装，新主机且无IP冲突 |
| duplicated_inner_ip | 内网IPv4地址冲突，同一网络区域下已存在相同IP的主机 |
| duplicated_inner_ipv6 | 内网IPv6地址冲突，同一网络区域下已存在相同IPv6的主机 |
| host_not_found | 指定的主机ID在系统中不存在 |
| networkunit_not_found | 指定的网络单元ID不存在 |
| networkunit_not_support_install | 网络单元不支持安装（非直连网络单元且无可用的安装代理Proxy） |
| mismatched_inner_ip | 主机内网IPv4地址与CMDB记录不匹配 |
| mismatched_inner_ipv6 | 主机内网IPv6地址与CMDB记录不匹配 |
| mismatched_biz_id | 请求中的业务ID与主机实际所属业务ID不匹配 |
| mismatched_networkarea_id | 网络单元所属的网络区域ID与主机实际所属网络区域ID不匹配 |
| invalid_node_role | 主机的节点角色为Proxy，不能安装Agent |

**category 枚举说明**：

| 枚举值 | 描述 |
|-------|------|
| normal_install | 正常安装，无需用户确认 |
| register_to_cmdb_and_install | 新主机需先注册到CMDB再安装 |
| need_confirm | 需要用户确认（如IP冲突将触发重装） |
| error | 检查失败，无法进行安装 |

#### data.results[n].matched

当检查涉及已有主机时返回该主机的信息，新主机无冲突时为null。

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| bk_host_id | int64 | 主机ID |
| bk_biz_id | int64 | 业务ID |
| bk_networkarea_id | int64 | 网络区域ID |
| bk_networkunit_id | int64 | 网络单元ID |
| os_type | string | 操作系统类型（枚举值：linux、windows、darwin） |
| node_role | string | 节点角色（枚举值：blank、agent、proxy） |
| bk_host_innerip_list | string array | 主机内网IPv4地址列表 |
| bk_host_innerip_v6_list | string array | 主机内网IPv6地址列表 |

**node_role 枚举说明**：
- `blank`：空白节点，尚未安装任何程序
- `agent`：Agent节点
- `proxy`：Proxy节点

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
