### 描述

- 该接口提供版本：v3.0.1-alpha.23+。
- 该接口所需权限：无。
- 该接口功能描述：查询插件配置变量模板，支持按平台分组返回配置模板信息。

### URL

POST /api/v3/package/release/plugin/get_config_variables

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述                                       |
| ---------- | ---------- | ---- | ------------------------------------------ |
| generation | int64      | 是   | 插件安装包代次 |
| name       | string     | 是   | 插件名称                                   |
| platforms  | Platform[] | 是   | 平台列表，支持查询多个平台的配置变量       |
| version    | string     | 是   | 插件版本号                                 |

#### Platform

| 参数名称 | 参数类型 | 必选 | 描述                                  |
| -------- | -------- | ---- | ------------------------------------- |
| os_type  | string   | 是   | 操作系统类型；须与 cpu_arch 组成受支持的平台组合，如 linux、windows、aix  |
| cpu_arch | string   | 是   | CPU 架构；须与 os_type 组成受支持的平台组合，如 amd64、arm64、ppc64 |

### 调用示例

```json
{
  "generation": 2,
  "name": "bkmonitorbeat",
  "platforms": [
    {
      "os_type": "linux",
      "cpu_arch": "amd64"
    },
    {
      "os_type": "windows",
      "cpu_arch": "amd64"
    }
  ],
  "version": "1.13.0"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "config_variables": {
      "linux/amd64": {
        "items": [
          {
            "name": "bkmonitorbeat_exporter.yaml",
            "file_path": "etc/bkmonitorbeat_exporter.yaml",
            "source_path": "etc/bkmonitorbeat_exporter.yaml",
            "is_main_config": false,
            "source_content": "...",
            "variables": {
              "http_port": {
                "title": "HTTP 端口",
                "type": "integer",
                "required": true,
                "default": 9273,
                "description": "Exporter HTTP 服务端口",
                "description_en": "Exporter HTTP service port"
              },
              "log_level": {
                "title": "日志级别",
                "type": "string",
                "required": false,
                "default": "info",
                "description": "日志输出级别",
                "description_en": "Log output level"
              }
            }
          }
        ]
      },
      "windows/amd64": {
        "items": [
          {
            "name": "bkmonitorbeat_exporter.yaml",
            "file_path": "etc\\bkmonitorbeat_exporter.yaml",
            "source_path": "etc\\bkmonitorbeat_exporter.yaml",
            "is_main_config": false,
            "source_content": "...",
            "variables": {
              "http_port": {
                "title": "HTTP 端口",
                "type": "integer",
                "required": true,
                "default": 9273,
                "description": "Exporter HTTP 服务端口",
                "description_en": "Exporter HTTP service port"
              }
            }
          }
        ]
      }
    }
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述               |
| -------- | -------- | ------------------ |
| code     | int32    | 状态码，0 表示成功 |
| message  | string   | 请求信息           |
| request_id | string | 请求 ID            |
| data     | object   | 响应数据           |

#### data

| 参数名称 | 参数类型                         | 描述                                                                                         |
| -------- | -------------------------------- | -------------------------------------------------------------------------------------------- |
| config_variables | map<string, ConfigVariablesList> | 配置变量映射表，key 为平台标识（格式：`{os_type}/{cpu_arch}`），value 为该平台的配置变量列表 |

#### ConfigVariablesList

| 参数名称 | 参数类型          | 描述                                             |
| -------- | ----------------- | ------------------------------------------------ |
| items    | ConfigVariables[] | 配置变量数组，包含该平台下所有配置文件的变量定义 |

#### ConfigVariables

| 参数名称       | 参数类型              | 描述                                         |
| -------------- | --------------------- | -------------------------------------------- |
| name           | string                | 配置文件名称                                 |
| file_path      | string                | 配置文件在插件包中的相对路径                 |
| source_path    | string                | 配置文件源路径                               |
| is_main_config | bool                  | 是否为主配置文件                             |
| source_content | string                | 配置文件源内容（模板内容）                   |
| variables      | map<string, Property> | 配置变量定义，key 为变量名，value 为变量属性 |

#### Property

| 参数名称       | 参数类型              | 描述                                                       |
| -------------- | --------------------- | ---------------------------------------------------------- |
| title          | string                | 变量标题（用于 UI 展示）                                   |
| type           | string                | 变量类型，如 string, integer, boolean, object, array       |
| required       | bool                  | 是否必填                                                   |
| default        | any                   | 默认值，类型根据 type 字段决定                             |
| description    | string                | 变量描述（中文）                                           |
| description_en | string                | 变量描述（英文）                                           |
| properties     | map<string, Property> | 嵌套属性（当 type 为 object 时使用），结构与 Property 相同 |

### 使用场景

1. **插件部署前配置预览**：在部署插件前，查询该插件在不同平台上需要配置的变量，供用户填写。
2. **配置模板管理**：获取插件配置文件的模板结构，用于生成配置表单或验证用户输入。
3. **多平台配置对比**：同时查询多个平台的配置变量，对比不同平台的配置差异。

### 注意事项

1. **平台标识格式**：响应中的 `config_variables` 的 key 格式为 `{os_type}/{cpu_arch}`，如 `linux/amd64`。
2. **嵌套变量**：当变量类型为 `object` 时，`properties` 字段包含嵌套的变量定义，结构与 `Property` 相同，支持多层嵌套。
3. **变量类型**：`type` 字段支持的值包括：`string`、`integer`、`boolean`、`object`、`array` 等，需根据实际类型解析 `default` 字段。
4. **空结果处理**：如果指定的插件版本在某个平台上不存在，该平台的 key 不会出现在 `config_variables` 中。
