### 描述

- 该接口提供版本：v3.0.1-alpha.14+。
- 该接口所需权限：package_manage（管理资源包）。
- 该接口功能描述：禁用指定的 Agent 安装包版本。

### URL

POST /api/v3/package/release/agent/disable

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 | 是 | 安装包代次（枚举值：2） |
| release_type | string | 是 | 安装包类型，该接口固定为 agent |
| platform | object | 是 | 目标平台 |
| version | string | 是 | 安装包版本号 |

#### platform

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| os_type | string | 是 | 操作系统类型（与 cpu_arch 组成受支持的平台组合） |
| cpu_arch | string | 是 | CPU 架构（与 os_type 组成受支持的平台组合） |

### 调用示例

```json
{
  "generation": 2,
  "release_type": "agent",
  "platform": {"os_type": "linux", "cpu_arch": "amd64"},
  "version": "2.1.6-rc.5"
}
```

### 响应示例

```json
{"code": 0, "message": "ok", "request_id": "req-1234567890", "data": {}}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| data | object | 空对象，表示操作成功 |
