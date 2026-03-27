### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：proxy_operate（操作Proxy）。
- 该接口功能描述：批量更新节点Proxy的动态配置信息，包括登录信息、网络地址、标签等。此操作为元数据更新，不触发远程操作。

### URL

POST /api/v3/node/proxy/update

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|---------|------|------|
| host | array | 是 | 主机列表，详见下方 host 参数说明 |

**host[n]**

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|---------|------|------|
| bk_host_id | int64 | 是 | 主机ID |
| login_ip | string | 否 | 登录IP |
| login_port | int64 | 否 | 登录端口 |
| login_user | string | 否 | 登录用户名 |
| login_mode | string | 否 | 登录方式：password_vault（密码库）/ password（密码）/ keyfile（密钥文件） |
| export_ip | string | 否 | 对外IPv4地址 |
| export_ip_v6 | string | 否 | 对外IPv6地址 |
| advertise_ip | string | 否 | 广播IPv4地址 |
| advertise_ip_v6 | string | 否 | 广播IPv6地址 |
| proxy_tags | array[string] | 否 | Proxy标签，可选值：dedicated_installer / cluster_tunnel / file_tunnel / data_tunnel |
| relay_download_port | int64 | 否 | 中转下载端口 |
| relay_callback_port | int64 | 否 | 中转回调端口 |

### 调用示例

```json
{
    "host": [
        {
            "bk_host_id": 1001,
            "login_ip": "10.0.0.1",
            "login_port": 22,
            "login_user": "root",
            "login_mode": "password",
            "export_ip": "10.0.0.1",
            "proxy_tags": ["cluster_tunnel"]
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
    "data": {}
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|---------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求ID |
| error | object | 错误信息，成功时为 null |
| data | object | 响应数据（成功时为空对象） |

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
