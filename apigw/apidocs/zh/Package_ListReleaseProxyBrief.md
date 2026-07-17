### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：package_view（查看资源包）。
- 该接口功能描述：查询 Proxy 安装包精简列表，返回核心字段及变更日志，支持分页和条件过滤。

### URL

POST /api/v3/package/release/proxy/list/brief

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 | 是 | 安装包代次（枚举值：2） |
| page | object | 是 | 普通列表查询必传；仅当 only_count=true 时可省略 |
| only_count | bool | 否 | 仅返回总数，不返回列表数据，默认 false |
| exact_include_conditions | object | 否 | 精确过滤条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| offset | int32 | 否 | 偏移量，起始值为 0 |
| limit | int32 | 是 | 每页条数，取值范围为 1～1000 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| platform | object array | 否 | 平台（操作系统+CPU架构）过滤条件 |
| version | string array | 否 | 版本号过滤条件 |
| as_default | bool array | 否 | 是否为默认版本过滤条件 |
| enabled | bool array | 否 | 是否启用过滤条件 |
| name | string array | 否 | 安装包名称过滤条件 |
| file_name | string array | 否 | 安装包文件名过滤条件 |

#### exact_include_conditions.platform[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| os_type | string | 是 | 操作系统类型（与 cpu_arch 组成受支持的平台组合） |
| cpu_arch | string | 是 | CPU 架构（与 os_type 组成受支持的平台组合） |

### 调用示例

查询代次为 2、已启用的 Linux amd64 平台 Proxy 安装包精简列表。

```json
{
  "generation": 2,
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "platform": [
      {
        "os_type": "linux",
        "cpu_arch": "amd64"
      }
    ],
    "enabled": [true]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "total": 1,
    "items": [
      {
        "generation": 2,
        "os_type": "linux",
        "cpu_arch": "amd64",
        "version": "2.1.6-rc.5",
        "enabled": true,
        "as_default": true,
        "change_log_en": "Bug fixes and performance improvements",
        "change_log_zh": "修复已知问题并提升性能"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 符合条件的记录总数 |
| items | object array | Proxy 安装包精简列表 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| generation | int64 | 安装包代次 |
| os_type | string | 操作系统类型（与 cpu_arch 组成受支持的平台组合） |
| cpu_arch | string | CPU 架构（与 os_type 组成受支持的平台组合） |
| version | string | 版本号 |
| enabled | bool | 是否启用 |
| as_default | bool | 是否为默认版本 |
| change_log_en | string | 英文变更日志 |
| change_log_zh | string | 中文变更日志 |
