### Description

- API Version: v3.0.1-alpha.119+.
- Required Permissions: Both system_access (System Access) and platform_manage (Platform Management). The platform resource type is `platform_resource`, and the resource ID is `system_config` (System Configuration).
- Function: Query the business configuration whitelist loaded at startup by the responding Backend instance through the application-state API. Credentials, connection endpoints, logging and tracing settings are excluded. This API cannot modify configuration.
- Calls require API Gateway application authorization and application-state authentication. User-state authentication is not required, but the application-state authentication chain still requires a valid username and tenant context. Identity cannot be specified through the request body. This API returns the same business configuration structure and performs the same permission checks as the user-state entrypoint.
- The result is process-wide, not filtered by the request tenant, and does not establish configuration consistency across Backend replicas. This API does not query global settings stored in the database.
- Default event Data IDs do not resolve business-specific overrides. Installation defaults do not represent target-host state; network-unit configuration or deployment-specific settings, among others, may override them. Default direct network-unit settings do not establish that the network unit exists.
- Directory and file-path fields in installation settings are not resolved effective paths. Only the installer HTTP proxy policy and firewall features are returned, not proxy addresses or allowed firewall IPs/CIDRs.

### URL

POST /api/v3/system_config/get

### Request Parameters

No business parameters. Send the JSON object `{}`.

### Request Example

```json
{}
```

### Response Example

```json
{
  "code": 0,
  "message": "OK",
  "request_id": "req-123456",
  "data": {
    "tenant_mode": "single",
    "system": { "env": "dev", "edition": "ce" },
    "iam": { "v3_enabled": false, "v4_enabled": false },
    "node_event_data_id": {
      "agent_base_alarm_event_data_id": 1000,
      "task_proc_event_data_id": 1100008
    },
    "default_direct_network_unit": { "enabled": false, "name": "default" },
    "gse_deploy_configs": [
      {
        "generation": 2,
        "os_type": "linux",
        "base_work_dir": "/tmp/bknm/",
        "base_deploy_dir": "/usr/local/",
        "extra_config_dir": "",
        "host_id_path": "",
        "v2_run_dir": "",
        "v2_data_dir": "",
        "zone_id": "",
        "city_id": "",
        "http_proxy_mode": "inherit",
        "firewall_features": []
      }
    ]
  }
}
```

The example illustrates the schema, not the current environment's configuration.

### Response Parameters

| Parameter  | Type   | Description                                    |
| ---------- | ------ | ---------------------------------------------- |
| code       | int32  | Status code; `0` indicates success             |
| message    | string | Request message                                |
| request_id | string | Request ID                                     |
| error      | object | Error details; omitted on success              |
| permission | object | Permission application information; omitted on success |
| data       | object | Backend business configuration view            |

#### data

| Parameter                   | Type   | Description                                                                                |
| --------------------------- | ------ | ------------------------------------------------------------------------------------------ |
| tenant_mode                 | string | Tenant mode: `single` (single-tenant) or `multiple` (multi-tenant)                           |
| system                      | object | Deployment environment and edition                                                         |
| iam                         | object | IAM permission-management switches                                                         |
| node_event_data_id          | object | Global fallback Data IDs for node events                                                   |
| default_direct_network_unit | object | Settings for automatic creation of the default direct network unit                         |
| gse_deploy_configs          | array  | Installation defaults by GSE generation and OS; an empty array when no entries are configured |

#### data.system

| Parameter | Type   | Description                                                            |
| --------- | ------ | ---------------------------------------------------------------------- |
| env       | string | Configured deployment environment                                      |
| edition   | string | Edition: `ce` (Community), `ee` (Enterprise), or `inner` (Internal)     |

#### data.iam

| Parameter  | Type | Description               |
| ---------- | ---- | ------------------------- |
| v3_enabled | bool | Whether IAM V3 is enabled |
| v4_enabled | bool | Whether IAM V4 is enabled |

#### data.node_event_data_id

| Parameter                     | Type  | Description                                         |
| ----------------------------- | ----- | --------------------------------------------------- |
| agent_base_alarm_event_data_id | int64 | Global fallback Data ID for Agent base alarm events |
| task_proc_event_data_id        | int64 | Global fallback Data ID for task process events     |

#### data.default_direct_network_unit

| Parameter | Type   | Description                                                                          |
| --------- | ------ | ------------------------------------------------------------------------------------ |
| enabled   | bool   | Whether automatic creation of the default direct network unit is enabled             |
| name      | string | Name used during automatic creation; the configured value is returned even when disabled |

#### data.gse_deploy_configs[n]

| Parameter         | Type         | Description                                                                                                                                                                                        |
| ----------------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| generation        | int64        | Configured GSE generation; currently only `2` is supported                                                                                                                                         |
| os_type           | string       | Configured OS type, such as `linux`, `windows`, or `aix`; a configured value does not guarantee that installation is supported on that OS                                                              |
| base_work_dir     | string       | Default installation working directory                                                                                                                                                             |
| base_deploy_dir   | string       | Default installation deployment root directory                                                                                                                                                     |
| extra_config_dir  | string       | Agent extra configuration directory override; an empty string means no override is configured                                                                                                      |
| host_id_path      | string       | Host ID file-path override read by plugins; an empty string means no override is configured                                                                                                        |
| v2_run_dir        | string       | Runtime directory override for plugin V2-compatible installation; an empty string means no override is configured                                                                                  |
| v2_data_dir       | string       | Data directory override for plugin V2-compatible installation; an empty string means no override is configured                                                                                     |
| zone_id           | string       | Region identifier override for node deployment; an empty string means no override is configured                                                                                                    |
| city_id           | string       | City identifier override for node deployment; an empty string means no override is configured                                                                                                      |
| http_proxy_mode   | string       | Installer HTTP proxy policy: `inherit` uses the host's environment proxy settings, `direct` forces a direct connection, and `specified` uses a specified proxy; the proxy address is not returned  |
| firewall_features | string array | Firewall features managed by the Windows installer; effective only for the Windows installer. Supports `gse` and `nodemgr`; an empty array means no rules are managed. Allowed IPs/CIDRs are not returned |
