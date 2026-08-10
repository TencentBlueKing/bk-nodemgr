### 描述

- 该接口提供版本：v3.0.1-alpha.67+。
- 该接口所需权限：无。
- 该接口功能描述：获取 Agent 安装包可选的操作系统类型、CPU 架构、安装包名称和版本号的去重列表。

### URL

POST /api/v3/package/release/agent/distinct

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                    |
| ------------------------ | -------- | ---- | ----------------------- |
| generation               | int64    | 是   | 安装包代次（枚举值：2） |
| exact_include_conditions | object   | 否   | 精确过滤条件            |
| distinct_field           | object   | 否   | 指定需要去重的字段      |

#### exact_include_conditions

| 参数名称   | 参数类型     | 必选 | 描述                             |
| ---------- | ------------ | ---- | -------------------------------- |
| platform   | object array | 否   | 平台（操作系统+CPU架构）过滤条件 |
| version    | string array | 否   | 版本号过滤条件                   |
| as_default | bool array   | 否   | 是否为默认版本过滤条件           |
| enabled    | bool array   | 否   | 是否启用过滤条件                 |
| name       | string array | 否   | 安装包名称过滤条件               |
| file_name  | string array | 否   | 安装包文件名过滤条件             |
| is_hidden  | bool array   | 否   | 是否隐藏过滤条件                 |

#### exact_include_conditions.platform[n]

| 参数名称 | 参数类型 | 必选 | 描述                                             |
| -------- | -------- | ---- | ------------------------------------------------ |
| os_type  | string   | 是   | 操作系统类型（与 cpu_arch 组成受支持的平台组合） |
| cpu_arch | string   | 是   | CPU 架构（与 os_type 组成受支持的平台组合）      |

#### distinct_field

| 参数名称 | 参数类型 | 必选 | 描述                               |
| -------- | -------- | ---- | ---------------------------------- |
| os_type  | bool     | 否   | 是否对操作系统类型去重，默认 false |
| cpu_arch | bool     | 否   | 是否对 CPU 架构去重，默认 false    |
| name     | bool     | 否   | 是否对安装包名称去重，默认 false   |
| version  | bool     | 否   | 是否对版本号去重，默认 false       |

### 调用示例

获取代次为 2 的 Agent 安装包所有可用操作系统类型、CPU 架构、安装包名称和版本号。

```json
{
  "generation": 2,
  "distinct_field": {
    "os_type": true,
    "cpu_arch": true,
    "name": true,
    "version": true
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
    "os_type": ["linux", "windows"],
    "cpu_arch": ["amd64", "arm64"],
    "name": ["gse_agent"],
    "version": ["1.0.0", "1.1.0"]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述               |
| ---------- | -------- | ------------------ |
| code       | int32    | 状态码，0 表示成功 |
| message    | string   | 请求信息           |
| request_id | string   | 请求 ID            |
| data       | object   | 响应数据           |

#### data

| 参数名称 | 参数类型     | 描述                                       |
| -------- | ------------ | ------------------------------------------ |
| os_type  | string array | 去重后的操作系统类型列表（由匹配数据生成） |
| cpu_arch | string array | 去重后的 CPU 架构列表（由匹配数据生成）    |
| name     | string array | 去重后的安装包名称列表（由匹配数据生成）   |
| version  | string array | 去重后的版本号列表（由匹配数据生成）       |
