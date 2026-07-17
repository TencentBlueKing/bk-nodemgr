### 描述

- 该接口提供版本：v3.0.1-alpha.15+。
- 该接口所需权限：package_view（查看资源包）。
- 该接口功能描述：查询插件安装包列表，支持分页和条件过滤。接口仅返回当前用户有权限查看的插件安装包。

### URL

POST /api/v3/package/release/plugin/list

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
| name | string array | 否 | 插件名称过滤条件，若不传则返回所有已授权的插件 |
| file_name | string array | 否 | 安装包文件名过滤条件 |

#### exact_include_conditions.platform[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| os_type | string | 是 | 操作系统类型（与 cpu_arch 组成受支持的平台组合） |
| cpu_arch | string | 是 | CPU 架构（与 os_type 组成受支持的平台组合） |

**参数说明**：
- `exact_include_conditions.name`：接口将请求的插件名称与用户已授权的插件名称取交集。若用户对指定插件无权限，则相应结果会被过滤，不返回该插件的数据。

### 调用示例

查询代次为 2、已启用的 Linux amd64 平台插件安装包列表。

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
        "release": {
          "name": "bkmonitorbeat",
          "generation": 2,
          "release_type": "plugin",
          "os_type": "linux",
          "cpu_arch": "amd64",
          "version": "3.20.2305",
          "file_name": "bkmonitorbeat-3.20.2305.tgz",
          "labels": [],
          "enabled": true,
          "as_default": true,
          "md5": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
          "updated_at": 1712000000000,
          "operator": "admin"
        }
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
| items | object array | 插件安装包列表 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| release | object | 安装包基础信息 |

#### data.items[n].release

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| name | string | 插件名称 |
| generation | int64 | 安装包代次 |
| release_type | string | 安装包类型，固定为 plugin |
| os_type | string | 操作系统类型（与 cpu_arch 组成受支持的平台组合） |
| cpu_arch | string | CPU 架构（与 os_type 组成受支持的平台组合） |
| version | string | 版本号 |
| file_name | string | 安装包文件名 |
| labels | string array | 标签列表 |
| enabled | bool | 是否启用 |
| as_default | bool | 是否为默认版本 |
| md5 | string | 安装包 MD5 校验值 |
| updated_at | uint64 | 最后更新时间（Unix 毫秒级时间戳） |
| operator | string | 最后操作人 |
