### 描述

- 该接口提供版本：v3.0.1-alpha.105+
- 该接口所需权限：system_access（系统访问）、agent_operate（操作 Agent）。
- 该接口功能描述：以用户身份批量检查 Agent 安装条件，返回主机匹配结果及可否继续安装的分类；不会执行安装。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。`agent_operate` 按请求中的业务校验；本检查接口不检查管控单元使用权限，实际安装仍需该权限。

### URL

POST /api/v3/open/node/agent/install_check

### 输入参数

| 参数名称 | 参数类型     | 必选 | 描述                     |
| -------- | ------------ | ---- | ------------------------ |
| host     | object array | 是   | 待检查主机列表，不能为空 |

#### host[n]

| 参数名称                | 参数类型     | 必选 | 描述                                                                           |
| ----------------------- | ------------ | ---- | ------------------------------------------------------------------------------ |
| bk_host_id              | int64        | 否   | 已有主机 ID；默认 `-1`，表示按新主机检查                                       |
| bk_biz_id               | int64        | 是   | 业务 ID，必须大于等于 `0`                                                      |
| bk_host_innerip_list    | string array | 否   | 内网 IPv4 列表；与 IPv6 列表至少一个非空，元素须为有效 IP                      |
| bk_host_innerip_v6_list | string array | 否   | 内网 IPv6 列表；与 IPv4 列表至少一个非空，元素须为有效 IP                      |
| bk_networkunit_id       | int64        | 否   | 目标管控单元 ID；默认 `-1`。未指定或单元不存在时，返回 `networkunit_not_found` |

检查目标单元是否存在且支持安装，并检查主机业务、管控区域、IP 和节点角色的一致性。新主机可能匹配到已有主机或需要先注册到 CMDB；已有主机更换管控单元时，即使 `status` 为 `normal_install`，`category` 仍可能为 `need_confirm`。应同时读取两字段。

### 调用示例

```json
{
  "host": [
    {
      "bk_host_id": 1001,
      "bk_biz_id": 100,
      "bk_host_innerip_list": ["10.0.0.1"],
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
  "request_id": "req-123456",
  "data": {
    "results": [
      {
        "status": "normal_install",
        "category": "normal_install",
        "message_en": "",
        "message_zh": "",
        "matched": {
          "bk_host_id": 1001,
          "bk_biz_id": 100,
          "bk_networkarea_id": 1,
          "bk_networkunit_id": 1,
          "os_type": "linux",
          "node_role": "agent",
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": []
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称     | 参数类型     | 描述                                               |
| ------------ | ------------ | -------------------------------------------------- |
| code         | int32        | 状态码，`0` 表示检查请求成功，不表示全部主机可安装 |
| message      | string       | 请求信息                                           |
| request_id   | string       | 请求 ID                                            |
| error        | object       | 错误信息                                           |
| permission   | object       | 权限信息                                           |
| data         | object       | 响应数据                                           |
| data.results | object array | 与输入主机顺序一致的检查结果                       |

#### data.results[n]

| 参数名称   | 参数类型 | 描述                                                                                                                  |
| ---------- | -------- | --------------------------------------------------------------------------------------------------------------------- |
| status     | string   | 检查状态，见下表                                                                                                      |
| category   | string   | `normal_install`（正常安装）、`register_to_cmdb_and_install`（注册后安装）、`need_confirm`（需确认）、`error`（错误） |
| message_en | string   | 英文提示                                                                                                              |
| message_zh | string   | 中文提示                                                                                                              |
| matched    | object   | 匹配的已有主机；未匹配时可为空                                                                                        |

#### status

| 值                              | 描述                                           |
| ------------------------------- | ---------------------------------------------- |
| duplicated_inner_ip             | 内网 IPv4 存在冲突                             |
| duplicated_inner_ipv6           | 内网 IPv6 存在冲突                             |
| host_not_found                  | 指定的主机不存在                               |
| networkunit_not_found           | 管控单元不存在或未指定                         |
| mismatched_inner_ip             | 内网 IPv4 与主机不匹配                         |
| mismatched_inner_ipv6           | 内网 IPv6 与主机不匹配                         |
| mismatched_biz_id               | 业务不匹配                                     |
| mismatched_networkarea_id       | 管控区域不匹配                                 |
| invalid_node_role               | 节点角色不允许安装 Agent，例如已有 Proxy       |
| networkunit_not_support_install | 管控单元不支持安装                             |
| normal_install                  | 可执行安装；仍须结合 `category` 判断是否需确认 |
| register_to_cmdb_and_install    | 需先注册到 CMDB 再安装                         |

#### data.results[n].matched

| 参数名称                | 参数类型     | 描述                                           |
| ----------------------- | ------------ | ---------------------------------------------- |
| bk_host_id              | int64        | 匹配主机 ID                                    |
| bk_biz_id               | int64        | 匹配主机业务 ID                                |
| bk_networkarea_id       | int64        | 匹配主机管控区域 ID                            |
| bk_networkunit_id       | int64        | 匹配主机管控单元 ID                            |
| os_type                 | string       | 操作系统                                       |
| node_role               | string       | 节点角色：`agent`、`proxy`；未设置时为空字符串 |
| bk_host_innerip_list    | string array | 匹配主机内网 IPv4 列表                         |
| bk_host_innerip_v6_list | string array | 匹配主机内网 IPv6 列表                         |
