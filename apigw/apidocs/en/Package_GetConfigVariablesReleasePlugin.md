### Description

- API Version: v3.0.1-alpha.23+.
- Required Permission: None.
- API Function: Query plugin configuration variable templates, supporting grouped return of configuration template information by platform.

### URL

POST /api/v3/package/release/plugin/get_config_variables

### Request Parameters

| Parameter  | Type       | Required | Description                                                                     |
| ---------- | ---------- | -------- | ------------------------------------------------------------------------------- |
| generation | int64      | Yes      | Plugin package generation                               |
| name       | string     | Yes      | Plugin name                                                                     |
| platforms  | Platform[] | Yes      | Platform list, supports querying configuration variables for multiple platforms |
| version    | string     | Yes      | Plugin version number                                                           |

#### Platform

| Parameter | Type   | Required | Description                                      |
| --------- | ------ | -------- | ------------------------------------------------ |
| os_type   | string | Yes      | Operating system type; must form a supported platform combination with cpu_arch, e.g., linux, windows, aix |
| cpu_arch  | string | Yes      | CPU architecture; must form a supported platform combination with os_type, e.g., amd64, arm64, ppc64 |

### Request Example

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

### Response Example

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
                "title": "HTTP Port",
                "type": "integer",
                "required": true,
                "default": 9273,
                "description": "Exporter HTTP 服务端口",
                "description_en": "Exporter HTTP service port"
              },
              "log_level": {
                "title": "Log Level",
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
                "title": "HTTP Port",
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

### Response Parameters

| Parameter | Type   | Description                      |
| --------- | ------ | -------------------------------- |
| code      | int32  | Status code, 0 indicates success |
| message   | string | Request message                  |
| request_id | string | Request ID                      |
| data      | object | Response data                    |

#### data

| Parameter | Type                             | Description                                                                                                                                             |
| --------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| config_variables | map<string, ConfigVariablesList> | Configuration variable mapping, key is platform identifier (format: `{os_type}/{cpu_arch}`), value is the configuration variable list for that platform |

#### ConfigVariablesList

| Parameter | Type              | Description                                                                                                   |
| --------- | ----------------- | ------------------------------------------------------------------------------------------------------------- |
| items     | ConfigVariables[] | Configuration variable array, containing variable definitions for all configuration files under this platform |

#### ConfigVariables

| Parameter      | Type                  | Description                                                                          |
| -------------- | --------------------- | ------------------------------------------------------------------------------------ |
| name           | string                | Configuration file name                                                              |
| file_path      | string                | Relative path of the configuration file in the plugin package                        |
| source_path    | string                | Configuration file source path                                                       |
| is_main_config | bool                  | Whether it is the main configuration file                                            |
| source_content | string                | Configuration file source content (template content)                                 |
| variables      | map<string, Property> | Configuration variable definitions, key is variable name, value is variable property |

#### Property

| Parameter      | Type                  | Description                                                                     |
| -------------- | --------------------- | ------------------------------------------------------------------------------- |
| title          | string                | Variable title (for UI display)                                                 |
| type           | string                | Variable type, e.g., string, integer, boolean, object, array                    |
| required       | bool                  | Whether it is required                                                          |
| default        | any                   | Default value, type depends on the type field                                   |
| description    | string                | Variable description (Chinese)                                                  |
| description_en | string                | Variable description (English)                                                  |
| properties     | map<string, Property> | Nested properties (used when type is object), structure is the same as Property |

### Use Cases

1. **Configuration Preview Before Plugin Deployment**: Query the variables that need to be configured for the plugin on different platforms before deployment, for users to fill in.
2. **Configuration Template Management**: Get the template structure of plugin configuration files, used to generate configuration forms or validate user input.
3. **Multi-Platform Configuration Comparison**: Query configuration variables for multiple platforms simultaneously to compare configuration differences across platforms.

### Notes

1. **Platform Identifier Format**: The key format in the response `config_variables` is `{os_type}/{cpu_arch}`, e.g., `linux/amd64`.
2. **Nested Variables**: When the variable type is `object`, the `properties` field contains nested variable definitions with the same structure as `Property`, supporting multi-level nesting.
3. **Variable Types**: The `type` field supports values including: `string`, `integer`, `boolean`, `object`, `array`, etc. The `default` field should be parsed according to the actual type.
4. **Empty Result Handling**: If the specified plugin version does not exist on a certain platform, the key for that platform will not appear in `config_variables`.
