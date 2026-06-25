### Description

- API Version: v3.0.1+.
- Required Permission: proxy_operate (Operate Proxy), networkunit_use_for_proxy (Use Network Unit for Proxy Deployment).
- Function: Batch install node Proxy. Supports manual and offline installation modes.

### URL

POST /api/v3/node/proxy/install

### Input Parameters

| Parameter                 | Type  | Required | Description                                              |
|---------------------------|-------|----------|----------------------------------------------------------|
| host                      | array | Yes      | Host list, see host parameters below                     |
| target_version            | array | No       | Target version list, see target_version parameters below |
| is_manual                 | bool  | No       | Manual installation mode, default false                  |
| is_offline                | bool  | No       | Offline installation mode, default false                 |

**host[n]**

| Parameter                    | Type          | Required | Description                                                                                       |
|------------------------------|---------------|----------|---------------------------------------------------------------------------------------------------|
| bk_biz_id                    | int64         | Yes      | Business ID                                                                                       |
| bk_networkunit_id            | int64         | Yes      | Network unit ID                                                                                   |
| bk_host_id                   | int64         | No       | Host ID, -1 means unspecified (omit for auto-registration)                                        |
| bk_addressing                | string        | Yes      | Addressing mode: dynamic / static                                                                 |
| bk_host_innerip              | array[string] | No       | Inner IPv4 address list                                                                           |
| bk_host_innerip_v6           | array[string] | No       | Inner IPv6 address list                                                                           |
| os_type                      | string        | Yes      | OS type: linux / windows / darwin                                                                 |
| cpu_arch                     | string        | No       | CPU architecture: 386 / arm / arm64 / amd64                                                       |
| login_ip                     | string        | Yes      | Login IP                                                                                          |
| login_port                   | int64         | No       | Login port                                                                                        |
| login_user                   | string        | Yes      | Login username                                                                                    |
| login_mode                   | string        | Yes      | Login mode: password_vault / password / keyfile                                                   |
| login_password               | string        | No       | Password, required when login_mode is password                                                    |
| login_key_file               | string        | No       | Key file content, required when login_mode is keyfile                                             |
| export_ip                    | string        | No       | Export IPv4 address                                                                               |
| export_ip_v6                 | string        | No       | Export IPv6 address                                                                               |
| advertise_ip                 | string        | No       | Advertise IPv4 address                                                                            |
| advertise_ip_v6              | string        | No       | Advertise IPv6 address                                                                            |
| re_register                  | bool          | No       | Re-register host, default false                                                                   |
| install_pre_ordered_plugins  | bool          | No       | Whether to install pre-ordered plugins, default true                                              |
| renew_gse_task               | bool          | No       | Whether to regenerate the GSE .task runtime file; default false preserves the existing .task file |
| renew_gse_proc               | bool          | No       | Whether to regenerate the GSE .proc runtime file; default false preserves the existing .proc file |
| install_method               | string        | No       | Proxy install method. Supported values: empty string, ssh. Empty string means Proxy-scenario auto selection |
| proxy_tags                   | array[string] | No       | Proxy tags: dedicated_installer / cluster_tunnel / file_tunnel / data_tunnel                      |
| proxy_install_origin_unit_id | int64         | No       | Origin network unit ID for installation                                                           |
| credit_expired_interval_sec  | int64         | No       | Credential expiry interval in seconds, default 86400                                              |
| relay_download_port          | int64         | No       | Relay download port                                                                               |
| relay_callback_port          | int64         | No       | Relay callback port                                                                               |

`install_method` notes:

- Empty string: use Proxy-scenario auto selection. In the current version, auto selection chooses SSH. Future versions may evolve the auto selection logic independently for Proxy deployment scenarios.
- `ssh`: explicitly use SSH install logic.
- `wmi`: not supported for Proxy installation in the current version.

**target_version[n]**

| Parameter | Type   | Required | Description                                 |
|-----------|--------|----------|---------------------------------------------|
| version   | string | Yes      | Version string                              |
| cpu_arch  | string | Yes      | CPU architecture: 386 / arm / arm64 / amd64 |
| os_type   | string | Yes      | OS type: linux / windows / darwin           |

### Call Example

```json
{
  "host": [
    {
      "bk_biz_id": 2,
      "bk_networkunit_id": 1,
      "bk_addressing": "dynamic",
      "bk_host_innerip": [
        "10.0.0.1"
      ],
      "os_type": "linux",
      "cpu_arch": "amd64",
      "login_ip": "10.0.0.1",
      "login_port": 22,
      "login_user": "root",
      "login_mode": "password",
      "login_password": "your_password",
      "export_ip": "10.0.0.1",
      "install_pre_ordered_plugins": true,
      "renew_gse_task": false,
      "renew_gse_proc": false,
      "install_method": "ssh"
    }
  ],
  "target_version": [
    {
      "version": "2.1.0",
      "cpu_arch": "amd64",
      "os_type": "linux"
    }
  ],
  "is_manual": false,
  "is_offline": false
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "abc123",
  "data": {
    "workflow_id": "wf-20240101-001"
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                  |
|------------|--------|------------------------------|
| code       | int32  | Status code, 0 means success |
| message    | string | Request message              |
| request_id | string | Request ID                   |
| error      | object | Error info, null on success  |
| data       | object | Response data                |

**data**

| Parameter   | Type   | Description |
|-------------|--------|-------------|
| workflow_id | string | Workflow ID |

**error**

| Parameter | Type   | Description         |
|-----------|--------|---------------------|
| system    | string | Error source system |
| message   | string | Error message       |
| details   | array  | Detailed error list |

**error.details[n]**

| Parameter | Type   | Description  |
|-----------|--------|--------------|
| code      | string | Error code   |
| message   | string | Error detail |
