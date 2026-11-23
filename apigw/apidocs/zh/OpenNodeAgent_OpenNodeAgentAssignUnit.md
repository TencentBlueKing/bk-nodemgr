### 描述

- 该接口提供版本：v3.0.1-alpha.105+
- 该接口所需权限：system_access（系统访问）、agent_operate（操作 Agent）、networkunit_use_for_agent（使用管控单元部署 Agent）。
- 该接口功能描述：以用户身份为尚未分配管控单元的主机批量指定同一个管控单元。只更新主机的管控单元归属，不执行 Agent 安装、升级或重新配置。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。`agent_operate` 按目标主机所属业务校验，`networkunit_use_for_agent` 按目标管控单元校验。

### URL

POST /api/v3/open/node/agent/assign_unit

### 输入参数

| 参数名称          | 参数类型    | 必选 | 描述                                      |
| ----------------- | ----------- | ---- | ----------------------------------------- |
| bk_host_id        | int64 array | 是   | 主机 ID 列表，不能为空；重复 ID 会去重    |
| bk_networkunit_id | int64       | 是   | 已存在的目标管控单元 ID，必须大于等于 `0` |

所有已找到的主机须属于同一管控区域，且与目标单元的管控区域一致；不满足时请求失败。已分配管控单元的主机不能通过本接口更换单元，即使当前单元与目标相同也计为失败。不存在的主机计入失败数量。响应可部分成功，应读取统计和失败原因，而不仅检查 `code`。

### 调用示例

```json
{
  "bk_host_id": [1001, 1002],
  "bk_networkunit_id": 1
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "success_count": 2,
    "failed_count": 0,
    "failed_reasons": []
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                                   |
| ---------- | -------- | -------------------------------------- |
| code       | int32    | 状态码；`0` 不保证所有主机均已分配成功 |
| message    | string   | 请求信息                               |
| request_id | string   | 请求 ID                                |
| error      | object   | 错误信息                               |
| permission | object   | 权限信息                               |
| data       | object   | 同步分配结果，不返回工作流 ID          |

#### data

| 参数名称       | 参数类型     | 描述                                               |
| -------------- | ------------ | -------------------------------------------------- |
| success_count  | int64        | 分配成功的主机数量                                 |
| failed_count   | int64        | 分配失败的主机数量                                 |
| failed_reasons | string array | 失败原因；可能按批次汇总，不保证与失败主机逐一对应 |
