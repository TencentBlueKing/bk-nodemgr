### Description

- API version: v3.0.1-alpha.72+.
- Required permission: ⚠️ To be confirmed (handler not yet implemented).
- Description: Starts a one-shot plugin debug workflow. The debug unit contains a target scope (host or service instance), plugin, version, config templates, and a custom config context.

### URL

POST /api/v3/plugin/start_debug

### Request Parameters

| Parameter  | Type   | Required | Description                                                                                  |
| ---------- | ------ | -------- | -------------------------------------------------------------------------------------------- |
| debug_info | object | Yes      | Plugin debug info describing the target scope, plugin, version, and config of one debug unit |

#### debug_info

| Parameter             | Type         | Required | Description                                                                                                                                                                         |
| --------------------- | ------------ | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| scope                 | object       | Yes      | Debug target scope (host or service instance)                                                                                                                                       |
| plugin_name           | string       | Yes      | Plugin name                                                                                                                                                                         |
| version               | string       | Yes      | Plugin version                                                                                                                                                                      |
| config_template_name  | string array | No       | List of config template names. Templates are recorded in the plugin package (release) of the given version; a template is uniquely located by plugin name + version + template name |
| custom_config_context | object       | No       | Custom config context shared by all config templates of this debug unit, used to override template render variables                                                                 |

#### debug_info.scope

| Parameter    | Type        | Required | Description                                                                                                       |
| ------------ | ----------- | -------- | ----------------------------------------------------------------------------------------------------------------- |
| granularity  | string      | Yes      | Target granularity (enum: host, service_instance)                                                                 |
| bk_biz_id    | int64       | Yes      | Business ID                                                                                                       |
| filter       | object      | No       | Target filter condition (placeholder structure for now)                                                           |
| instance_ids | int64 array | Yes      | Target instance IDs; host IDs when granularity is host, service instance IDs when granularity is service_instance |

### Request Example

Starts debug for one plugin against a host.

```json
{
  "debug_info": {
    "scope": {
      "granularity": "host",
      "bk_biz_id": 2,
      "instance_ids": [1001]
    },
    "plugin_name": "basereport",
    "version": "1.0.0",
    "config_template_name": ["basereport.conf"],
    "custom_config_context": {
      "report_path": "/usr/local/gse/plugins/etc"
    }
  }
}
```

### Response Example

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

### Response Parameters

| Parameter  | Type   | Description                  |
| ---------- | ------ | ---------------------------- |
| code       | int32  | Status code, 0 means success |
| message    | string | Request message              |
| request_id | string | Request ID                   |
| error      | object | Error info, null on success  |
| data       | object | Response data                |

#### data

| Parameter   | Type   | Description                                               |
| ----------- | ------ | --------------------------------------------------------- |
| workflow_id | string | Debug workflow ID, used to stop or query the debug status |

#### error

| Parameter | Type   | Description             |
| --------- | ------ | ----------------------- |
| system    | string | Error system identifier |
| message   | string | Error message           |
| details   | array  | Error detail list       |

#### error.details[n]

| Parameter | Type   | Description   |
| --------- | ------ | ------------- |
| code      | string | Error code    |
| message   | string | Error message |
