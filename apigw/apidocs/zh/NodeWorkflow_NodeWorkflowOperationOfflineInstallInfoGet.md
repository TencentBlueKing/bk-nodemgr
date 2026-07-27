### 描述

- 该接口提供版本：v3.0.1-alpha.56+。
- 该接口所需权限：无。
- 该接口功能描述：获取离线安装 Proxy 操作所需的安装脚本、配置、预检查和安装包信息。

**权限说明**：当前后端 handler 未执行直接鉴权；handler 会校验该操作是否为离线安装 Proxy 操作且处于等待手动安装动作。

### URL

POST /api/v3/node/workflow/operation/offline/info

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| operation_id | string | 是 | 离线安装 Proxy 操作 ID |

### 调用示例

```json
{
  "operation_id": "op-20240601-0001"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "install_script": "bash install_proxy.sh",
    "metadata": "{\"operation_id\":\"op-20240601-0001\"}",
    "configs": {
      "proxy.conf": "server=example.com"
    },
    "precheck": "bash precheck.sh",
    "release_info": {
      "generation": "v2",
      "os_type": "linux",
      "cpu_arch": "x86_64",
      "version": "2.0.0"
    },
    "installer_info": {
      "generation": "v2",
      "os_type": "linux",
      "cpu_arch": "x86_64"
    },
    "package_name": "bk-nodemgr-proxy-linux-x86_64.tgz"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| install_script | string | 离线安装脚本内容 |
| metadata | string | 离线安装元数据 |
| configs | object | 配置文件映射，键为文件名，值为文件内容 |
| precheck | string | 预检查脚本内容 |
| release_info | object | 节点包发布信息 |
| installer_info | object | 安装器信息 |
| package_name | string | 安装包名称 |

#### release_info

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| generation | string | 包代际 |
| os_type | string | 操作系统类型 |
| cpu_arch | string | CPU 架构 |
| version | string | 版本号 |

#### installer_info

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| generation | string | 安装器代际 |
| os_type | string | 操作系统类型 |
| cpu_arch | string | CPU 架构 |
