### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：agent_operate（操作Agent）、networkunit_use_for_agent（使用网络单元部署Agent）。
- 该接口功能描述：批量安装节点Agent，支持指定目标版本和手动安装模式。

### URL

POST /api/v3/node/agent/install

### 输入参数

| 参数名称           | 参数类型         | 必选 | 描述                                          |
|----------------|--------------|----|---------------------------------------------|
| info           | object array | 是  | 待安装Agent的主机信息列表                             |
| target_version | object array | 否  | 目标版本列表，用于指定不同平台（os_type + cpu_arch）的Agent版本 |
| is_manual      | bool         | 否  | 是否为手动安装模式，默认false                           |

#### info[n]

| 参数名称                        | 参数类型          | 必选 | 描述                                               |
|-----------------------------|---------------|----|--------------------------------------------------|
| bk_addressing               | string        | 是  | 寻址方式（枚举值：dynamic、static）                         |
| bk_biz_id                   | int64         | 否  | 业务ID，-1表示未指定                                     |
| bk_host_innerip             | array[string] | 是  | 主机内网IPv4地址列表，与bk_host_innerip_v6至少填写一个           |
| bk_host_innerip_v6          | array[string] | 否  | 主机内网IPv6地址列表，与bk_host_innerip至少填写一个              |
| login_ip                    | string        | 是  | 登录IP地址                                           |
| login_port                  | int64         | 否  | 登录端口，-1表示未指定，必须大于0                               |
| login_user                  | string        | 是  | 登录用户名                                            |
| login_mode                  | string        | 是  | 登录方式（枚举值：password_vault、password、keyfile）        |
| login_password              | string        | 否  | 登录密码，当login_mode为password时必填                     |
| login_key_file              | string        | 否  | 登录密钥文件内容，当login_mode为keyfile时必填                  |
| bk_networkunit_id           | int64         | 否  | 网络单元ID，-1表示未指定                                   |
| os_type                     | string        | 是  | 操作系统类型（枚举值：linux、windows、darwin）                 |
| bk_host_id                  | int64         | 否  | 主机ID，-1表示未指定                                     |
| re_register                 | bool          | 否  | 是否重新注册，默认false                                   |
| install_pre_ordered_plugins | bool          | 否  | 是否安装预设插件，默认true                                  |
| renew_gse_task              | bool          | 否  | 是否重新生成 GSE .task runtime 文件，默认false表示保留已有.task文件 |
| renew_gse_proc              | bool          | 否  | 是否重新生成 GSE .proc runtime 文件，默认false表示保留已有.proc文件 |
| install_method              | string        | 否  | Agent 安装方式，合法值：空字符串、ssh、wmi；空字符串表示按 OS 自动选择安装逻辑  |

**参数说明**：

- `bk_addressing`：寻址方式
    - `dynamic`：动态寻址
    - `static`：静态寻址
- `login_mode`：登录方式
    - `password_vault`：从密码库自动获取密码
    - `password`：使用密码登录，需要提供`login_password`
    - `keyfile`：使用密钥文件登录，需要提供`login_key_file`
- `os_type`：操作系统类型，常用值包括`linux`、`windows`、`darwin`等
- `install_method`：Agent 安装方式，作用于单台主机。
    - 空字符串：按 OS 自动选择安装逻辑。
    - `ssh`：指定 SSH 安装。Linux/Darwin/unknown 支持，Windows 当前不支持。
    - `wmi`：指定 WMI 安装。仅 Windows 支持，Linux/Darwin/unknown 不支持。

#### target_version[n]

| 参数名称     | 参数类型   | 必选 | 描述                               |
|----------|--------|----|----------------------------------|
| version  | string | 是  | Agent版本号                         |
| cpu_arch | string | 是  | CPU架构（枚举值：386、arm、arm64、amd64）   |
| os_type  | string | 是  | 操作系统类型（枚举值：linux、windows、darwin） |

### 调用示例

批量安装Linux系统的Agent，使用密码登录方式，并显式指定 SSH 安装方式。

```json
{
  "info": [
    {
      "bk_addressing": "static",
      "bk_biz_id": 100,
      "bk_host_innerip": ["127.0.0.1"],
      "login_ip": "127.0.0.1",
      "login_port": 22,
      "login_user": "root",
      "login_mode": "password",
      "login_password": "your_password",
      "bk_networkunit_id": 1,
      "os_type": "linux",
      "re_register": false,
      "install_pre_ordered_plugins": true,
      "renew_gse_task": false,
      "renew_gse_proc": false,
      "install_method": "ssh"
    }
  ],
  "target_version": [
    {
      "version": "2.0.0",
      "cpu_arch": "amd64",
      "os_type": "linux"
    }
  ],
  "is_manual": false
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "workflow_id": "workflow-abc123def456"
  }
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述            |
|------------|--------|---------------|
| code       | int32  | 状态码，0表示成功     |
| message    | string | 请求信息          |
| request_id | string | 请求ID          |
| error      | object | 错误信息，成功时为null |
| data       | object | 响应数据          |

#### data

| 参数名称        | 参数类型   | 描述                |
|-------------|--------|-------------------|
| workflow_id | string | 工作流ID，可用于查询安装任务状态 |

#### error

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| system  | string | 错误系统标识 |
| message | string | 错误消息   |
| details | array  | 错误详情列表 |

#### error.details[n]

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | string | 错误代码 |
| message | string | 错误消息 |
