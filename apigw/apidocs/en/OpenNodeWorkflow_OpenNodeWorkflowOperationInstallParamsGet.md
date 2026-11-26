### Description

- API Version: v3.0.1-alpha.120+.
- Required Permission: system_access (System Access), agent_history_view (View Agent History), proxy_history_view (View Proxy History). Permissions are checked for the businesses associated with the historical workflow according to its node roles: `blank` and `agent` require Agent history permission, `proxy` requires Proxy history permission, and mixed, unknown, or missing roles require both.
- Function: Read recoverable parameters from selected historical Agent / Proxy installation operations for execution diagnosis (for example, to check historical target versions, authentication modes, and connection parameters). This API does not create installations, retry operations, or modify historical tasks.
- The API gateway authenticates and forwards the user identity; no username or tenant parameter is needed in the request body.

**Usage Boundaries**:

- Only `install_agent` and `install_proxy` workflows are supported. Operations do not have to be failed. Use the [operation list API](OpenNodeWorkflow_OpenNodeWorkflowOperationList.md) to obtain operation IDs.
- Authentication mode can be copied, but the response contains no passwords, private keys, password-vault credential IDs, deployment tokens, or instance runtime state.
- Parameters come from historical deployment records, not a complete snapshot of the original request. `missing_fields` identifies inputs that could not be recovered; empty values for those fields must not be treated as previous selections. Operations without usable deployment data appear in `unavailable_operation_ids` instead of `items`.
- Historical host IDs, businesses, network areas, network units, and target versions are not guaranteed to remain valid. These parameters do not establish the actual runtime configuration or installation result; correlate them with execution instances and step logs.

### URL

POST /api/v3/open/node/workflow/operation/install_params/get

### Request Parameters

| Parameter     | Type         | Required | Description                                                                                                                                                                                                                              |
| ------------- | ------------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| workflow_id   | string       | Yes      | Historical installation workflow ID in the current tenant; must not be empty or whitespace-only                                                                                                                                          |
| operation_ids | string array | Yes      | Selected operation IDs belonging to the workflow. The list must be nonempty, contain no empty or whitespace-only elements, and have no duplicates. A nonexistent operation or one belonging to another workflow fails the entire request |

### Request Example

```json
{
  "workflow_id": "wf:00000000000000000000000000000001",
  "operation_ids": ["oper:00000000000000000000000000000001", "oper:00000000000000000000000000000002"]
}
```

### Response Example

The first operation has partially recoverable parameters; the second operation no longer has usable deployment data.

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "type": "install_agent",
    "items": [
      {
        "operation_id": "oper:00000000000000000000000000000001",
        "host": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "bk_addressing": "static",
          "bk_host_innerip": ["10.0.0.1"],
          "bk_host_innerip_v6": [],
          "os_type": "linux",
          "login_ip": "10.0.0.1",
          "login_user": "root",
          "login_mode": "password",
          "install_method": "",
          "login_port": 22,
          "is_sudo": false,
          "re_register": false,
          "install_pre_ordered_plugins": true,
          "renew_gse_task": false,
          "renew_gse_proc": false,
          "export_ip": "",
          "export_ip_v6": "",
          "advertise_ip": "",
          "advertise_ip_v6": "",
          "proxy_tags": [],
          "proxy_install_origin_unit_id": 0,
          "relay_download_port": 0,
          "relay_callback_port": 0,
          "cpu_arch": "x86_64"
        },
        "target_version": [
          { "os_type": "linux", "cpu_arch": "x86_64", "version": "2.0.0" }
        ],
        "is_manual": false,
        "is_offline": false,
        "missing_fields": ["bk_host_innerip_v6", "install_method"]
      }
    ],
    "unavailable_operation_ids": ["oper:00000000000000000000000000000002"]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                    |
| ---------- | ------ | -------------------------------------------------------------- |
| code       | int32  | Status code; `0` means success                                 |
| message    | string | Request message                                                |
| request_id | string | Request ID                                                     |
| error      | object | Error information, usually empty on success                    |
| permission | object | Permission-application information when access is denied       |
| data       | object | Recoverable installation parameters and unavailable operations |

#### data

| Parameter                 | Type         | Description                                                                                                                                                                    |
| ------------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| type                      | string       | Historical workflow type: `install_agent` or `install_proxy`                                                                                                                   |
| items                     | object array | Recoverable operation parameters in the requested `operation_ids` order, excluding unavailable operations; may be empty                                                        |
| unavailable_operation_ids | string array | IDs of operations that exist and belong to this workflow but lack deployment records or static/dynamic host information, in request order. Empty when no such operations exist |

#### items[n]

| Parameter      | Type         | Description                                                                                           |
| -------------- | ------------ | ----------------------------------------------------------------------------------------------------- |
| operation_id   | string       | Historical operation ID, corresponding to an operation in the request                                 |
| host           | object       | Copyable historical host installation parameters, without credential contents                         |
| target_version | object array | Historical installation target versions, not current defaults or versions detected after installation |
| is_manual      | bool         | Whether manual installation was selected                                                              |
| is_offline     | bool         | Whether offline installation was selected                                                             |
| missing_fields | string array | Names of unrecovered fields; see below                                                                |

#### items[n].host

| Parameter                    | Type         | Description                                                                                                                                                                                                                  |
| ---------------------------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_id                   | int64        | Historical host ID, for historical context only; handle host identity in the new request according to normal installation rules                                                                                              |
| bk_biz_id                    | int64        | Historical business ID                                                                                                                                                                                                       |
| bk_networkarea_id            | int64        | Historical network area ID; `0` is valid and must not be treated as missing                                                                                                                                                  |
| bk_networkunit_id            | int64        | Historical network unit ID                                                                                                                                                                                                   |
| bk_addressing                | string       | Addressing mode: `dynamic` or `static`                                                                                                                                                                                       |
| bk_host_innerip              | string array | Historical internal IPv4 addresses                                                                                                                                                                                           |
| bk_host_innerip_v6           | string array | Internal IPv6 addresses; not persisted in current historical records. See `missing_fields`                                                                                                                                   |
| os_type                      | string       | Historical OS type, such as `linux`, `windows`, or `darwin`                                                                                                                                                                  |
| login_ip                     | string       | Login address                                                                                                                                                                                                                |
| login_user                   | string       | Login user                                                                                                                                                                                                                   |
| login_mode                   | string       | Authentication mode: `password`, `keyfile`, or `password_vault`; credential contents are not included                                                                                                                        |
| install_method               | string       | Connection method. Normal installation accepts an empty string (OS-dependent automatic selection), `ssh`, or `wmi`. An empty historical value is marked missing and does not prove automatic selection was previously chosen |
| login_port                   | int64        | Login port                                                                                                                                                                                                                   |
| is_sudo                      | bool         | Whether to use sudo                                                                                                                                                                                                          |
| re_register                  | bool         | Whether to re-register the host                                                                                                                                                                                              |
| install_pre_ordered_plugins  | bool         | Whether to install pre-ordered plugins                                                                                                                                                                                       |
| renew_gse_task               | bool         | Whether to regenerate GSE `.task` runtime files                                                                                                                                                                              |
| renew_gse_proc               | bool         | Whether to regenerate GSE `.proc` runtime files                                                                                                                                                                              |
| export_ip                    | string       | External service IPv4 address                                                                                                                                                                                                |
| export_ip_v6                 | string       | External service IPv6 address                                                                                                                                                                                                |
| advertise_ip                 | string       | Advertised IPv4 address                                                                                                                                                                                                      |
| advertise_ip_v6              | string       | Advertised IPv6 address                                                                                                                                                                                                      |
| proxy_tags                   | string array | Proxy tags: `dedicated_installer` (dedicated installer), `cluster_tunnel` (cluster tunnel), `file_tunnel` (file tunnel), or `data_tunnel` (data tunnel)                                                                      |
| proxy_install_origin_unit_id | int64        | Network unit from which the Proxy is installed                                                                                                                                                                               |
| relay_download_port          | int64        | Relay download port                                                                                                                                                                                                          |
| relay_callback_port          | int64        | Relay callback port                                                                                                                                                                                                          |
| cpu_arch                     | string       | Historical host CPU architecture, such as `x86_64`                                                                                                                                                                           |

#### items[n].target_version[n]

| Parameter | Type   | Description                                                                                                                   |
| --------- | ------ | ----------------------------------------------------------------------------------------------------------------------------- |
| os_type   | string | OS type for the target version                                                                                                |
| cpu_arch  | string | CPU architecture for the target version; preserve the historical identifier rather than rewriting it to an architecture alias |
| version   | string | Historical target version; confirm it is currently available before submitting a new installation                             |

#### missing_fields

The following field names may currently be returned:

| Field              | Meaning                                                                                                                                                                                     |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_innerip_v6 | Current historical deployment records do not persist internal IPv6 addresses, so this field is always marked. An empty list does not prove the original request contained no IPv6 addresses |
| install_method     | Installation connection method was not recovered. Ask the user to confirm it instead of interpreting an empty string as a recovered automatic-mode selection                                |
| target_version     | Historical target versions were not recovered and must be selected again. This API does not substitute the latest version                                                                   |

`missing_fields` does not describe credential completion and is not a guarantee that all other historical parameters are complete. Normal form validation still applies before creating an installation.

### Error Scenarios

- Empty required IDs, an empty operation list, or empty, whitespace-only, or duplicate operation IDs fail request validation.
- Missing historical-view permissions produce a permission error.
- A nonexistent workflow, a workflow outside the current tenant, an unsupported workflow type, or an operation that does not exist or belong to the specified workflow fails the entire request rather than returning partial success.
- Unreadable operation parameters, missing deployment linkage, or storage errors fail the entire request. Unavailable deployment records are instead reported through `unavailable_operation_ids` and may be returned successfully alongside recoverable operations.
