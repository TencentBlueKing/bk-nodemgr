### Description

- API Version: v3.0.1-alpha.105+; `is_sudo` is supported from v3.0.1-alpha.106.
- Required Permissions: system_access (System Access), agent_operate (Operate Agent), networkunit_use_for_agent (Use Network Unit to Deploy Agent).
- Function: Install node Agents in batches as the authenticated user, with per-platform target versions and manual installation support.
- The API Gateway authenticates and forwards the user identity. Do not include a username or tenant parameter in the request body. `login_user` is the target host login account, not the API caller.
- `agent_operate` is checked against the requested businesses; `networkunit_use_for_agent` is checked against the target network units.

### URL

POST /api/v3/open/node/agent/install

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                                                                                     |
| -------------- | -------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------- |
| host           | object array   | Yes      | Nonempty installation host list; the field is not `info`                                                                        |
| target_version | object array   | No       | Versions selected by `os_type` + `cpu_arch`; duplicate platforms are not allowed. Omit to select default versions automatically |
| is_manual      | bool           | No       | Manual installation mode, default `false`; required host fields below must still be supplied                                    |

#### host[n]

| Parameter Name              | Parameter Type | Required    | Description                                                                                                                                               |
| --------------------------- | -------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bk_addressing               | string         | Yes         | Addressing mode: `dynamic` or `static`                                                                                                                    |
| bk_biz_id                   | int64          | Yes         | Business ID, greater than or equal to `0`                                                                                                                 |
| bk_host_innerip             | string array   | No          | Inner IPv4 list; this or `bk_host_innerip_v6` must be nonempty; entries must be valid IP addresses                                                        |
| bk_host_innerip_v6          | string array   | No          | Inner IPv6 list; this or `bk_host_innerip` must be nonempty; entries must be valid IP addresses                                                           |
| login_ip                    | string         | Yes         | Nonempty target host login IP                                                                                                                             |
| login_port                  | int64          | Yes         | Login port, greater than or equal to `0`; for example, `22` for SSH                                                                                       |
| login_user                  | string         | Yes         | Nonempty target host login account                                                                                                                        |
| login_mode                  | string         | Yes         | Credential type: `password_vault`, `password`, or `keyfile`                                                                                               |
| login_password              | string         | Conditional | Encrypted password for automatic installation with `password`; may be omitted to reuse an existing available credential. Not a plaintext password         |
| login_key_file              | string         | Conditional | For automatic installation with `keyfile`, Base64-encode the key file content before encryption; may be omitted to reuse an existing available credential |
| bk_networkunit_id           | int64          | Yes         | Existing target network unit ID, greater than or equal to `0`                                                                                             |
| os_type                     | string         | Yes         | Nonempty operating system, for example `linux`, `windows`, or `darwin`; installation requires an available package for the platform                       |
| bk_host_id                  | int64          | No          | Existing host ID; defaults to `-1`, meaning no existing host is specified                                                                                 |
| re_register                 | bool           | No          | Whether to register again, default `false`                                                                                                                |
| install_pre_ordered_plugins | bool           | No          | Whether to install pre-ordered plugins, default `false`                                                                                                   |
| renew_gse_task              | bool           | No          | Regenerate the GSE `.task` runtime file; default `false` preserves the existing file                                                                      |
| renew_gse_proc              | bool           | No          | Regenerate the GSE `.proc` runtime file; default `false` preserves the existing file                                                                      |
| install_method              | string         | No          | `""` (default, select by OS), `ssh`, or `wmi`; `wmi` is only available on Windows                                                                         |
| is_sudo                     | bool           | No          | Set to `true` for non-root login with passwordless sudo to run the installer through `sudo -n`; default `false`, effective only on Unix-like systems      |

Automatic installation credentials must be encrypted using the credential public key and protocol enabled in the deployment. Do not send plaintext passwords or keys. Manual mode does not process login credentials, but still validates required fields. After submitting a manual installation, obtain and execute the installation commands separately; this API does not return commands.

With `install_method=""`, Linux/Darwin use SSH. Direct Windows Agent installation tries SSH first and falls back to WMI if SSH preparation fails before the installer starts. Explicit `ssh` or `wmi` disables this automatic fallback.

#### target_version[n]

| Parameter Name | Parameter Type | Required | Description                                                                               |
| -------------- | -------------- | -------- | ----------------------------------------------------------------------------------------- |
| version        | string         | Yes      | Target Agent version, uploaded and available for the platform                             |
| cpu_arch       | string         | Yes      | Target CPU architecture, for example `amd64` or `arm64`; must match the detected platform |
| os_type        | string         | Yes      | Target OS, for example `linux`, `windows`, or `darwin`                                    |

### Request Example

Install a Linux Agent using a configured password vault.

```json
{
  "host": [
    {
      "bk_addressing": "static",
      "bk_biz_id": 100,
      "bk_host_innerip": ["10.0.0.1"],
      "login_ip": "10.0.0.1",
      "login_port": 22,
      "login_user": "root",
      "login_mode": "password_vault",
      "bk_networkunit_id": 1,
      "os_type": "linux",
      "install_method": "ssh"
    }
  ],
  "target_version": [
    { "version": "2.1.0", "cpu_arch": "amd64", "os_type": "linux" }
  ],
  "is_manual": false
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": { "workflow_id": "workflow-abc123" }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                                                       |
| -------------- | -------------- | --------------------------------------------------------------------------------- |
| code           | int32          | Status code; `0` means the request succeeded, not that installation has completed |
| message        | string         | Request message                                                                   |
| request_id     | string         | Request ID                                                                        |
| error          | object         | Error information                                                                 |
| permission     | object         | Permission information                                                            |
| data           | object         | Response data                                                                     |

#### data

| Parameter Name | Parameter Type | Description                                                                                                                 |
| -------------- | -------------- | --------------------------------------------------------------------------------------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking the asynchronous operation, not a `trigger_id`; installation is not complete when this is returned |

From v3.0.1-alpha.114, use `OpenNodeWorkflow_OpenNodeWorkflowList` with `exact_include_conditions.workflow_id` to query execution status. The query additionally requires the corresponding history-view permission.
