### Description

- API Version: v3.0.0+.
- Required Permission: agent_operate (Operate Agent), networkunit_use_for_agent (Use Network Unit for Agent Deployment).
- Function: Batch install node agents with support for specifying target versions and manual installation mode.

### URL

POST /api/v3/node/agent/install

### Input Parameters

| Parameter Name            | Parameter Type | Required | Description                                                                                    |
|---------------------------|----------------|----------|------------------------------------------------------------------------------------------------|
| info                      | object array   | Yes      | List of host information for agent installation                                                |
| target_version            | object array   | No       | Target version list for specifying agent versions for different platforms (os_type + cpu_arch) |
| is_manual                 | bool           | No       | Whether it is manual installation mode, default is false                                       |
| enable_compatibility_mode | bool           | No       | Whether to enable compatibility mode, default is false.                                        |

#### info[n]

| Parameter Name              | Parameter Type | Required | Description                                                                                              |
|-----------------------------|----------------|----------|----------------------------------------------------------------------------------------------------------|
| bk_addressing               | string         | Yes      | Addressing mode (enum values: dynamic, static)                                                           |
| bk_biz_id                   | int64          | No       | Business ID, -1 means not specified                                                                      |
| bk_host_innerip             | string         | Yes      | Host inner network IPv4 address, at least one of bk_host_innerip and bk_host_innerip_v6 must be provided |
| bk_host_innerip_v6          | string         | No       | Host inner network IPv6 address, at least one of bk_host_innerip and bk_host_innerip_v6 must be provided |
| login_ip                    | string         | Yes      | Login IP address                                                                                         |
| login_port                  | int64          | No       | Login port, -1 means not specified, must be greater than 0                                               |
| login_user                  | string         | Yes      | Login username                                                                                           |
| login_mode                  | string         | Yes      | Login method (enum values: password_vault, password, keyfile)                                            |
| login_password              | string         | No       | Login password, required when login_mode is password                                                     |
| login_key_file              | string         | No       | Login key file content, required when login_mode is keyfile                                              |
| bk_networkunit_id           | int64          | No       | Network unit ID, -1 means not specified                                                                  |
| os_type                     | string         | Yes      | Operating system type (enum values: linux, windows, darwin)                                              |
| bk_host_id                  | int64          | No       | Host ID, -1 means not specified                                                                          |
| re_register                 | bool           | No       | Whether to re-register, default is false                                                                 |
| install_pre_ordered_plugins | bool           | No       | Whether to install pre-ordered plugins, default is true                                                  |

**Parameter Notes**:

- `bk_addressing`: Addressing mode
    - `dynamic`: Dynamic addressing
    - `static`: Static addressing
- `login_mode`: Login method
    - `password_vault`: Automatically retrieve password from password vault
    - `password`: Use password login, requires `login_password`
    - `keyfile`: Use key file login, requires `login_key_file`
- `os_type`: Operating system type, common values include `linux`, `windows`, `darwin`, etc.

#### target_version[n]

| Parameter Name | Parameter Type | Required | Description                                                 |
|----------------|----------------|----------|-------------------------------------------------------------|
| version        | string         | Yes      | Agent version number                                        |
| cpu_arch       | string         | Yes      | CPU architecture (enum values: 386, arm, arm64, amd64)      |
| os_type        | string         | Yes      | Operating system type (enum values: linux, windows, darwin) |

### Request Example

Batch install agents for Linux systems using password login.

```json
{
  "info": [
    {
      "bk_addressing": "static",
      "bk_biz_id": 100,
      "bk_host_innerip": "127.0.0.1",
      "login_ip": "127.0.0.1",
      "login_port": 22,
      "login_user": "root",
      "login_mode": "password",
      "login_password": "your_password",
      "bk_networkunit_id": 1,
      "os_type": "linux",
      "re_register": false,
      "install_pre_ordered_plugins": true
    }
  ],
  "target_version": [
    {
      "version": "2.0.0",
      "cpu_arch": "amd64",
      "os_type": "linux"
    }
  ],
  "is_manual": false,
  "enable_compatibility_mode": false
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

| Parameter Name | Parameter Type | Description                        |
|----------------|----------------|------------------------------------|
| code           | int32          | Status code, 0 indicates success   |
| message        | string         | Request message                    |
| request_id     | string         | Request ID                         |
| error          | object         | Error information, null on success |
| data           | object         | Response data                      |

#### data

| Parameter Name | Parameter Type | Description                                                |
|----------------|----------------|------------------------------------------------------------|
| workflow_id    | string         | Workflow ID, can be used to query installation task status |

#### error

| Parameter Name | Parameter Type | Description             |
|----------------|----------------|-------------------------|
| system         | string         | Error system identifier |
| message        | string         | Error message           |
| details        | array          | Error details list      |

#### error.details[n]

| Parameter Name | Parameter Type | Description   |
|----------------|----------------|---------------|
| code           | string         | Error code    |
| message        | string         | Error message |
