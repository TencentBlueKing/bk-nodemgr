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

### Request Examples

#### Example 1: Using topology path (topo) to specify scope

Create a deploy policy to deploy monitoring collector plugin on all hosts in the entire business.

```json
{
  "name": "Production Monitor Plugin Deploy Policy",
  "description": "Unified deployment of monitoring collector plugin for production environment",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
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

#### Example 2: Using service template to specify scope

Create a deploy policy to deploy plugin on all service instances under specified service templates.

```json
{
  "name": "Web Service Monitor Plugin Deploy Policy",
  "description": "Deploy monitoring plugin for all service instances under web service templates",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "service_template",
      "scope": {
        "granularity": "service_instance",
        "bk_biz_id": 100,
        "filter": {},
        "service_template_ids": [1001, 1002]
      }
    }
  ]
}
```

#### Example 3: Using set template to specify scope

Create a deploy policy to deploy plugin on all hosts under specified set templates.

```json
{
  "name": "Test Cluster Monitor Plugin Deploy Policy",
  "description": "Deploy monitoring plugin for all hosts under test cluster templates",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "set_template",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "set_template_ids": [2001, 2002]
      }
    }
  ]
}
```

#### Example 4: Using instance IDs to directly specify scope

Create a deploy policy to deploy plugin on a specified list of hosts.

```json
{
  "name": "Specific Hosts Monitor Plugin Deploy Policy",
  "description": "Deploy monitoring plugin for specified host list",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "instance",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "instance_ids": [10001, 10002, 10003]
      }
    }
  ]
}
```

#### Example 5: Using dynamic group to specify scope

Create a deploy policy to deploy plugin on all hosts in dynamic groups.

```json
{
  "name": "Dynamic Group Monitor Plugin Deploy Policy",
  "description": "Deploy monitoring plugin for hosts in dynamic groups",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "dynamic_group",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "dynamic_group_ids": ["group-abc123", "group-def456"]
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
