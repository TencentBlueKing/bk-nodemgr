### 描述

- 该接口提供版本：v3.0.1-alpha.105+
- 该接口所需权限：system_access（系统访问）、agent_operate（操作 Agent）。
- 该接口功能描述：以用户身份批量检查 Agent 升级条件，检查主机状态、目标管控单元及平台版本映射；不会执行升级。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。`agent_operate` 按目标主机所属业务校验，不额外检查管控单元使用权限。

### URL

POST /api/v3/open/node/agent/upgrade_check

### 输入参数

| 参数名称       | 参数类型     | 必选 | 描述                                                                                                                               |
| -------------- | ------------ | ---- | ---------------------------------------------------------------------------------------------------------------------------------- |
| host           | object array | 是   | 待检查主机列表，不能为空                                                                                                           |
| target_version | object array | 否   | 以 `os_type` + `cpu_arch` 为键的目标版本映射；同一平台重复时采用最后一项。提供时检查主机平台是否有对应条目，不验证软件包实际可用性 |

#### host[n]

| 参数名称          | 参数类型 | 必选 | 描述                                                                                      |
| ----------------- | -------- | ---- | ----------------------------------------------------------------------------------------- |
| bk_host_id        | int64    | 是   | 主机 ID，必须大于等于 `0`                                                                 |
| bk_networkunit_id | int64    | 否   | 目标管控单元 ID；默认 `-1`，保留当前管控单元                                              |
| cpu_arch          | string   | 否   | 待升级主机的 CPU 架构，例如 `amd64`、`arm64`；为空时返回 `cpu_arch_missing`，不会自动补全 |

#### target_version[n]

| 参数名称 | 参数类型 | 必选 | 描述                                |
| -------- | -------- | ---- | ----------------------------------- |
| version  | string   | 是   | 对应平台的目标版本                  |
| cpu_arch | string   | 是   | CPU 架构，须匹配 `host[n].cpu_arch` |
| os_type  | string   | 是   | 操作系统，须匹配主机记录的操作系统  |

显式指定目标单元时，单元须存在且与主机处于同一管控区域。`busy`、`upgrade`、`stopping` 状态不允许升级。检查只反映当前条件，不锁定主机状态；成功后仍需调用升级接口。

### 调用示例

```json
{
  "host": [{ "bk_host_id": 1001, "bk_networkunit_id": 1, "cpu_arch": "amd64" }],
  "target_version": [
    { "version": "2.1.0", "cpu_arch": "amd64", "os_type": "linux" }
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
        "status": "normal_upgrade",
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
| code         | int32        | 状态码，`0` 表示检查请求成功，不表示全部主机可升级 |
| message      | string       | 请求信息                                           |
| request_id   | string       | 请求 ID                                            |
| error        | object       | 错误信息                                           |
| permission   | object       | 权限信息                                           |
| data         | object       | 响应数据                                           |
| data.results | object array | 与输入主机顺序一致的检查结果                       |

#### data.results[n]

| 参数名称 | 参数类型 | 描述                           |
| -------- | -------- | ------------------------------ |
| status   | string   | 检查状态，见下表               |
| matched  | object   | 匹配的已有主机；未匹配时可为空 |

#### status

| 值                      | 描述                                                                 |
| ----------------------- | -------------------------------------------------------------------- |
| host_not_found          | 主机不存在                                                           |
| networkunit_not_found   | 当前或目标管控单元不存在或未分配                                     |
| networkunit_mismatch    | 目标单元与主机不在同一管控区域                                       |
| version_not_found       | 提供了版本映射，但没有主机平台对应的条目；不等同于已验证软件包不存在 |
| node_status_not_allowed | 当前节点状态不允许升级                                               |
| networkunit_changed     | 升级将更换管控单元，需确认                                           |
| normal_upgrade          | 检查正常，可继续提交升级                                             |
| cpu_arch_missing        | 未提供 CPU 架构                                                      |

#### data.results[n].matched

| 参数名称                | 参数类型     | 描述                                           |
| ----------------------- | ------------ | ---------------------------------------------- |
| bk_host_id              | int64        | 匹配主机 ID                                    |
| bk_biz_id               | int64        | 匹配主机业务 ID                                |
| bk_networkarea_id       | int64        | 匹配主机管控区域 ID                            |
| bk_networkunit_id       | int64        | 匹配主机当前管控单元 ID                        |
| os_type                 | string       | 操作系统                                       |
| node_role               | string       | 节点角色：`agent`、`proxy`；未设置时为空字符串 |
| bk_host_innerip_list    | string array | 匹配主机内网 IPv4 列表                         |
| bk_host_innerip_v6_list | string array | 匹配主机内网 IPv6 列表                         |
