### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：proxy_operate（操作Proxy）。
- 该接口功能描述：批量更新 Proxy 主机的带外管理字段，并同步到 CMDB。

### URL

POST /api/v3/node/proxy/update_ops_fields

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| hosts | object array | 是 | 主机信息列表 |

#### hosts[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 | 是 | 主机 ID |
| ops_console_host_id | int64 | 否 | 带外管理控制台主机 ID |
| ops_out_band_type | string | 否 | 带外管理类型 |
| ops_out_band_protocol | string | 否 | 带外管理协议 |
| ops_bmc_ip | string | 否 | BMC IP 地址 |
| ops_bmc_port | int64 | 否 | BMC 端口 |

### 调用示例

```json
{
  "hosts": [
    {
      "bk_host_id": 1,
      "ops_console_host_id": 1,
      "ops_out_band_type": "string",
      "ops_out_band_protocol": "string",
      "ops_bmc_ip": "10.0.0.1",
      "ops_bmc_port": 22
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
  "error": null,
  "permission": null,
  "data": {}
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
