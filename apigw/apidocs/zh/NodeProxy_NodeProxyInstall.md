### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：proxy_operate（操作Proxy）、networkunit_use_for_proxy（使用网络单元部署Proxy）。
- 该接口功能描述：批量安装节点Proxy。支持手动安装和离线安装模式。

### URL

POST /api/v3/node/proxy/install

### 输入参数

| 参数名称                      | 参数类型  | 必选 | 描述                              |
|---------------------------|-------|----|---------------------------------|
| host                      | array | 是  | 主机列表，详见下方 host 参数说明             |
| target_version            | array | 否  | 目标版本列表，详见下方 target_version 参数说明 |
| is_manual                 | bool  | 否  | 是否手动安装模式，默认 false               |
| is_offline                | bool  | 否  | 是否离线安装模式，默认 false               |

**host[n]**

| 参数名称                         | 参数类型          | 必选 | 描述                                                                           |
|------------------------------|---------------|----|------------------------------------------------------------------------------|
| bk_biz_id                    | int64         | 是  | 业务ID                                                                         |
| bk_networkunit_id            | int64         | 是  | 网络单元ID                                                                       |
| bk_host_id                   | int64         | 否  | 主机ID，-1 表示未指定（自动注册时不填）                                                       |
| bk_addressing                | string        | 是  | 寻址方式：dynamic（动态）/ static（静态）                                                 |
| bk_host_innerip              | array[string] | 否  | 内网IPv4地址列表                                                                   |
| bk_host_innerip_v6           | array[string] | 否  | 内网IPv6地址列表                                                                   |
| os_type                      | string        | 是  | 操作系统类型：linux / windows / darwin                                              |
| cpu_arch                     | string        | 否  | CPU架构：386 / arm / arm64 / amd64                                              |
| login_ip                     | string        | 是  | 登录IP                                                                         |
| login_port                   | int64         | 否  | 登录端口                                                                         |
| login_user                   | string        | 是  | 登录用户名                                                                        |
| login_mode                   | string        | 是  | 登录方式：password_vault（密码库）/ password（密码）/ keyfile（密钥文件）                        |
| login_password               | string        | 否  | 登录密码，login_mode 为 password 时必填                                               |
| login_key_file               | string        | 否  | 密钥文件内容，login_mode 为 keyfile 时必填                                              |
| export_ip                    | string        | 否  | 对外IPv4地址                                                                     |
| export_ip_v6                 | string        | 否  | 对外IPv6地址                                                                     |
| advertise_ip                 | string        | 否  | 广播IPv4地址                                                                     |
| advertise_ip_v6              | string        | 否  | 广播IPv6地址                                                                     |
| re_register                  | bool          | 否  | 是否重新注册，默认 false                                                              |
| install_pre_ordered_plugins  | bool          | 否  | 是否安装预设插件，默认 true                                                             |
| renew_gse_task               | bool          | 否  | 是否重新生成 GSE .task runtime 文件，默认 false 表示保留已有 .task 文件                         |
| renew_gse_proc               | bool          | 否  | 是否重新生成 GSE .proc runtime 文件，默认 false 表示保留已有 .proc 文件                         |
| proxy_tags                   | array[string] | 否  | Proxy标签，可选值：dedicated_installer / cluster_tunnel / file_tunnel / data_tunnel |
| proxy_install_origin_unit_id | int64         | 否  | 安装来源网络单元ID                                                                   |
| credit_expired_interval_sec  | int64         | 否  | 凭证有效期（秒），默认 86400                                                            |
| relay_download_port          | int64         | 否  | 中转下载端口                                                                       |
| relay_callback_port          | int64         | 否  | 中转回调端口                                                                       |

**target_version[n]**

| 参数名称     | 参数类型   | 必选 | 描述                              |
|----------|--------|----|---------------------------------|
| version  | string | 是  | 版本号                             |
| cpu_arch | string | 是  | CPU架构：386 / arm / arm64 / amd64 |
| os_type  | string | 是  | 操作系统类型：linux / windows / darwin |

### 调用示例

```json
{
  "host": [
    {
      "bk_biz_id": 2,
      "bk_networkunit_id": 1,
      "bk_addressing": "dynamic",
      "bk_host_innerip": [
        "10.0.0.1"
      ],
      "os_type": "linux",
      "cpu_arch": "amd64",
      "login_ip": "10.0.0.1",
      "login_port": 22,
      "login_user": "root",
      "login_mode": "password",
      "login_password": "your_password",
      "export_ip": "10.0.0.1",
      "install_pre_ordered_plugins": true,
      "renew_gse_task": false,
      "renew_gse_proc": false
    }
  ],
  "target_version": [
    {
      "version": "2.1.0",
      "cpu_arch": "amd64",
      "os_type": "linux"
    }
  ],
  "is_manual": false,
  "is_offline": false
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "abc123",
  "data": {
    "workflow_id": "wf-20240101-001"
  }
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述             |
|------------|--------|----------------|
| code       | int32  | 状态码，0 表示成功     |
| message    | string | 请求信息           |
| request_id | string | 请求ID           |
| error      | object | 错误信息，成功时为 null |
| data       | object | 响应数据           |

**data**

| 参数名称        | 参数类型   | 描述    |
|-------------|--------|-------|
| workflow_id | string | 工作流ID |

**error**

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| system  | string | 错误来源系统 |
| message | string | 错误信息   |
| details | array  | 详细错误列表 |

**error.details[n]**

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | string | 错误码  |
| message | string | 错误详情 |
