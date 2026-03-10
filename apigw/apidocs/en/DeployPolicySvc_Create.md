### Description

- API Version: v3.0.1+
- Required Permission:
- Function: Create a deploy policy.

### URL

POST /api/v3/deploy_policy/create

### Request Parameters

| Parameter   | Type   | Required | Description               |
|-------------|--------|----------|---------------------------|
| name        | string | Yes      | Deploy policy name        |
| description | string | No       | Deploy policy description |
| enabled     | bool   | Yes      | Whether enabled           |
| specs       | array  | Yes      | Deploy specification list |
| scopes      | array  | Yes      | Deploy scope list         |

#### specs[n]

Deploy specifications define the desired final state.

| Parameter | Type   | Required | Description                                                                      |
|-----------|--------|----------|----------------------------------------------------------------------------------|
| type      | string | Yes      | Spec type (enum:  specify_plugin, specify_plugin_pkg, specify_plugin_sub_config) |
| param     | object | Yes      | Spec parameters, structure depends on type field                                 |

**When type is specify_plugin, param structure:**

Specify plugin version to ensure the target nodes have the specified plugin name and version installed. If the plugin
does not exist, it will be installed; if the version does not match, it will be upgraded.

| Parameter             | Type   | Required | Description                  |
|-----------------------|--------|----------|------------------------------|
| plugin_name           | string | Yes      | Plugin name                  |
| version               | string | Yes      | Plugin version               |
| custom_config_context | object | No       | Custom configuration context |

**When type is specify_plugin_pkg, param structure:**

Specify plugin package version to ensure the target nodes have the specified plugin package name and version installed.
The plugin name will be automatically generated based on the deploy policy ID and module ID. If the plugin does not
exist, it will be installed; if the version does not match, it will be upgraded.

| Parameter             | Type   | Required | Description                  |
|-----------------------|--------|----------|------------------------------|
| plugin_pkg_name       | string | Yes      | Plugin package name          |
| version               | string | Yes      | Plugin package version       |
| custom_config_context | object | No       | Custom configuration context |

**When type is specify_plugin_sub_config, param structure:**

Specify plugin sub-configuration to update the configuration file content of installed plugins. Only updates
configuration, does not involve plugin version installation or upgrade.

| Parameter             | Type   | Required | Description                     |
|-----------------------|--------|----------|---------------------------------|
| plugin_name           | string | Yes      | Plugin name                     |
| config_files_detail   | array  | Yes      | Configuration file details list |
| custom_config_context | object | No       | Custom configuration context    |

**config_files_detail[n] structure:**

| Parameter      | Type   | Required | Description                               |
|----------------|--------|----------|-------------------------------------------|
| name           | string | Yes      | Configuration file name                   |
| content        | string | Yes      | Configuration file content                |
| is_main_config | bool   | Yes      | Whether it is the main configuration file |

#### scopes[n]

Deploy scopes define the target range where the policy applies.

| Parameter | Type   | Required | Description                                                                      |
|-----------|--------|----------|----------------------------------------------------------------------------------|
| type      | string | Yes      | Scope type (enum: topo, service_template, set_template, instance, dynamic_group) |
| scope     | object | Yes      | Scope details, structure depends on type field                                   |

**When type is topo, scope structure:**

Specify target range by topology path.

| Parameter   | Type   | Required | Description                                       |
|-------------|--------|----------|---------------------------------------------------|
| granularity | string | Yes      | Target granularity (enum: host, service_instance) |
| bk_biz_id   | int64  | Yes      | Business ID                                       |
| filter      | object | No       | Target filter                                     |
| paths       | array  | Yes      | Topology path list                                |

**paths[n] structure:**

| Parameter    | Type   | Required | Description          |
|--------------|--------|----------|----------------------|
| topo_obj_id  | string | Yes      | Topology object ID   |
| topo_inst_id | int64  | Yes      | Topology instance ID |

**When type is service_template, scope structure:**

Specify target range by service template.

| Parameter            | Type        | Required | Description                                       |
|----------------------|-------------|----------|---------------------------------------------------|
| granularity          | string      | Yes      | Target granularity (enum: host, service_instance) |
| bk_biz_id            | int64       | Yes      | Business ID                                       |
| filter               | object      | No       | Target filter                                     |
| service_template_ids | int64 array | No       | Service template ID list                          |
| module_ids           | int64 array | No       | Module ID list                                    |

**When type is set_template, scope structure:**

Specify target range by set template.

| Parameter        | Type        | Required | Description                                       |
|------------------|-------------|----------|---------------------------------------------------|
| granularity      | string      | Yes      | Target granularity (enum: host, service_instance) |
| bk_biz_id        | int64       | Yes      | Business ID                                       |
| filter           | object      | No       | Target filter                                     |
| set_template_ids | int64 array | No       | Set template ID list                              |
| set_ids          | int64 array | No       | Set ID list                                       |

**When type is instance, scope structure:**

Specify target range directly by instance ID.

| Parameter    | Type        | Required | Description                                       |
|--------------|-------------|----------|---------------------------------------------------|
| granularity  | string      | Yes      | Target granularity (enum: host, service_instance) |
| bk_biz_id    | int64       | Yes      | Business ID                                       |
| filter       | object      | No       | Target filter                                     |
| instance_ids | int64 array | Yes      | Instance ID list (host_id or service_instance_id) |

**When type is dynamic_group, scope structure:**

Specify target range by dynamic group. Note: dynamic_group type only supports host granularity.

| Parameter         | Type         | Required | Description                              |
|-------------------|--------------|----------|------------------------------------------|
| granularity       | string       | Yes      | Target granularity (only supports: host) |
| bk_biz_id         | int64        | Yes      | Business ID                              |
| filter            | object       | No       | Target filter                            |
| dynamic_group_ids | string array | Yes      | Dynamic group ID list                    |

### Request Example

Create a deploy policy to upgrade agents in the production environment to version 2.1.5.

```json
{
  "name": "Production Agent Upgrade Policy",
  "description": "Unified agent version upgrade for production environment",
  "enabled": true,
  "specs": [
    {
      "type": "specify_agent",
      "param": {
        "node_version": "2.1.5"
      }
    }
  ],
  "scopes": [
    {
      "type": "topo",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "paths": [
          {
            "topo_obj_id": "biz",
            "topo_inst_id": 100
          }
        ]
      }
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "deploy_policy_id": 1001
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                          |
|------------|--------|--------------------------------------|
| code       | int32  | Status code, 0 for success           |
| message    | string | Response message                     |
| request_id | string | Request ID                           |
| error      | object | Error information (empty on success) |
| data       | object | Response data                        |

#### data

| Parameter        | Type  | Description              |
|------------------|-------|--------------------------|
| deploy_policy_id | int64 | Created deploy policy ID |
