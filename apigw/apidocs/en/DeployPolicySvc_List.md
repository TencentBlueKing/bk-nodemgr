### Description

- API Version: v3.0.1+
- Required Permission:
- Function: Query deploy policy list with pagination and filtering support.

### URL

POST /api/v3/deploy_policy/list

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| page | object | No | Pagination configuration |
| only_count | bool | No | Whether to return only the total count without details |
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions |
| exact_exclude_conditions | object | No | Exact match exclude conditions |
| fuzzy_exclude_conditions | object | No | Fuzzy match exclude conditions |
| executed_time_range | object | No | Execution time range |

#### page

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| count | bool | Yes | Whether to return total record count |
| start | uint32 | No | Record start position, starting from 0 |
| limit | uint32 | No | Records per page, maximum 500 |
| sort | string | No | Sort field |
| order | string | No | Sort order (ASC, DESC) |

#### exact_include_conditions

Exact match include conditions. Matches if any condition is met.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| deploy_policy_id | int64 array | No | Deploy policy ID list |
| dsu_id | int64 array | No | DSU ID list |
| deploy_policy_name | string array | No | Deploy policy name list |
| operator | string array | No | Operator list |
| enabled | bool array | No | Enabled status list |

#### fuzzy_include_conditions

Fuzzy match include conditions. Matches if any condition is met.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| deploy_policy_name | string array | No | Deploy policy name list (fuzzy match) |
| operator | string array | No | Operator list (fuzzy match) |

#### exact_exclude_conditions

Exact match exclude conditions. Excludes if any condition is met.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| deploy_policy_id | int64 array | No | Deploy policy ID list |
| dsu_id | int64 array | No | DSU ID list |
| deploy_policy_name | string array | No | Deploy policy name list |
| operator | string array | No | Operator list |
| enabled | bool array | No | Enabled status list |

#### fuzzy_exclude_conditions

Fuzzy match exclude conditions. Excludes if any condition is met.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| deploy_policy_name | string array | No | Deploy policy name list (fuzzy match) |
| operator | string array | No | Operator list (fuzzy match) |

#### executed_time_range

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| start | string | No | Start time, format: 2006-01-02T15:04:05Z |
| end | string | No | End time, format: 2006-01-02T15:04:05Z |

### Request Example

Query enabled deploy policies, sorted by creation time in descending order.

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "created_at",
    "order": "DESC"
  },
  "exact_include_conditions": {
    "enabled": [true]
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 25,
    "items": [
      {
        "deploy_policy_id": 1001,
        "dsu_id": 2001,
        "meta": {
          "name": "Production Agent Upgrade Policy",
          "description": "Unified agent version upgrade for production environment"
        },
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
        ],
        "operator": "admin",
        "enabled": true
      }
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, 0 for success |
| message | string | Response message |
| request_id | string | Request ID |
| error | object | Error information (empty on success) |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| total | int64 | Total record count matching the criteria |
| items | array | Query result data |

#### data.items[n]

| Parameter | Type | Description |
|---------|----------|------|
| deploy_policy_id | int64 | Deploy policy ID |
| dsu_id | int64 | DSU ID |
| meta | object | Deploy policy metadata |
| specs | array | Deploy specification list |
| scopes | array | Deploy scope list |
| operator | string | Operator |
| enabled | bool | Whether enabled |

#### data.items[n].meta

| Parameter | Type | Description |
|---------|----------|------|
| name | string | Deploy policy name |
| description | string | Deploy policy description |

#### data.items[n].specs[n]

Deploy specifications define the desired final state.

| Parameter | Type | Description |
|---------|----------|------|
| type | string | Spec type (enum: specify_agent, specify_proxy, specify_plugin, specify_plugin_pkg, specify_plugin_sub_config) |
| param | object | Spec parameters, structure depends on type field |

**When type is specify_agent, param structure:**

| Parameter | Type | Description |
|---------|----------|------|
| node_version | string | Agent version |

**When type is specify_proxy, param structure:**

| Parameter | Type | Description |
|---------|----------|------|
| node_version | string | Proxy version |

**When type is specify_plugin, param structure:**

| Parameter | Type | Description |
|---------|----------|------|
| plugin_name | string | Plugin name |
| version | string | Plugin version |
| custom_config_context | object | Custom configuration context |

**When type is specify_plugin_pkg, param structure:**

| Parameter | Type | Description |
|---------|----------|------|
| plugin_pkg_name | string | Plugin package name |
| version | string | Plugin package version |
| custom_config_context | object | Custom configuration context |

**When type is specify_plugin_sub_config, param structure:**

| Parameter | Type | Description |
|---------|----------|------|
| plugin_name | string | Plugin name |
| config_files_detail | array | Configuration file details list |
| custom_config_context | object | Custom configuration context |

**config_files_detail[n] structure:**

| Parameter | Type | Description |
|---------|----------|------|
| name | string | Configuration file name |
| content | string | Configuration file content |
| is_main_config | bool | Whether it is the main configuration file |

#### data.items[n].scopes[n]

Deploy scopes define the target range where the policy applies.

| Parameter | Type | Description |
|---------|----------|------|
| type | string | Scope type (enum: topo, service_template, set_template, instance, dynamic_group) |
| scope | object | Scope details, structure depends on type field |

**When type is topo, scope structure:**

| Parameter | Type | Description |
|---------|----------|------|
| granularity | string | Target granularity (enum: host, service_instance) |
| bk_biz_id | int64 | Business ID |
| filter | object | Target filter |
| paths | array | Topology path list |

**paths[n] structure:**

| Parameter | Type | Description |
|---------|----------|------|
| topo_obj_id | string | Topology object ID |
| topo_inst_id | int64 | Topology instance ID |

**When type is service_template, scope structure:**

| Parameter | Type | Description |
|---------|----------|------|
| granularity | string | Target granularity (enum: host, service_instance) |
| bk_biz_id | int64 | Business ID |
| filter | object | Target filter |
| service_template_ids | int64 array | Service template ID list |
| module_ids | int64 array | Module ID list |

**When type is set_template, scope structure:**

| Parameter | Type | Description |
|---------|----------|------|
| granularity | string | Target granularity (enum: host, service_instance) |
| bk_biz_id | int64 | Business ID |
| filter | object | Target filter |
| set_template_ids | int64 array | Set template ID list |
| set_ids | int64 array | Set ID list |

**When type is instance, scope structure:**

| Parameter | Type | Description |
|---------|----------|------|
| granularity | string | Target granularity (enum: host, service_instance) |
| bk_biz_id | int64 | Business ID |
| filter | object | Target filter |
| instance_ids | int64 array | Instance ID list (host_id or service_instance_id) |

**When type is dynamic_group, scope structure:**

| Parameter | Type | Description |
|---------|----------|------|
| granularity | string | Target granularity (only supports: host) |
| bk_biz_id | int64 | Business ID |
| filter | object | Target filter |
| dynamic_group_ids | string array | Dynamic group ID list |

**Note**: dynamic_group type does not support service_instance granularity.
