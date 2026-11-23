### Description

- API Version: v3.0.1-alpha.105+
- Required Permissions: system_access (System Access), agent_operate (Operate Agent).
- Function: Check Agent upgrade prerequisites in batches as the authenticated user, including host state, target network units, and platform-version mappings. Does not perform upgrades.
- The API Gateway authenticates and forwards the user identity. Do not include a username or tenant parameter in the request body. `agent_operate` is checked against the target hosts' businesses; no additional network unit use permission is checked.

### URL

POST /api/v3/open/node/agent/upgrade_check

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                                                                                                                                        |
| -------------- | -------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| host           | object array   | Yes      | Nonempty host list to check                                                                                                                                                        |
| target_version | object array   | No       | Target versions keyed by `os_type` + `cpu_arch`; the last entry wins for duplicate platforms. When provided, checks for a matching platform entry, not actual package availability |

#### host[n]

| Parameter Name    | Parameter Type | Required | Description                                                                                                                           |
| ----------------- | -------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_id        | int64          | Yes      | Host ID, greater than or equal to `0`                                                                                                 |
| bk_networkunit_id | int64          | No       | Target network unit ID; default `-1` retains the current unit                                                                         |
| cpu_arch          | string         | No       | Target host CPU architecture, for example `amd64` or `arm64`; empty returns `cpu_arch_missing` rather than inferring the architecture |

#### target_version[n]

| Parameter Name | Parameter Type | Required | Description                                      |
| -------------- | -------------- | -------- | ------------------------------------------------ |
| version        | string         | Yes      | Target version for the platform                  |
| cpu_arch       | string         | Yes      | CPU architecture matching `host[n].cpu_arch`     |
| os_type        | string         | Yes      | OS matching the host's recorded operating system |

An explicitly selected target unit must exist and belong to the same network area as the host. Hosts in `busy`, `upgrade`, or `stopping` state cannot be upgraded. This check reflects current conditions without locking the host state; call the upgrade API separately afterward.

### Request Example

```json
{
  "host": [{ "bk_host_id": 1001, "bk_networkunit_id": 1, "cpu_arch": "amd64" }],
  "target_version": [
    { "version": "2.1.0", "cpu_arch": "amd64", "os_type": "linux" }
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
    "results": [
      {
        "status": "normal_upgrade",
        "matched": {
          "bk_host_id": 1001,
          "bk_biz_id": 100,
          "bk_networkarea_id": 1,
          "bk_networkunit_id": 1,
          "os_type": "linux",
          "node_role": "agent",
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": []
        }
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                                                             |
| -------------- | -------------- | --------------------------------------------------------------------------------------- |
| code           | int32          | Status code; `0` means the check request succeeded, not that every host can be upgraded |
| message        | string         | Request message                                                                         |
| request_id     | string         | Request ID                                                                              |
| error          | object         | Error information                                                                       |
| permission     | object         | Permission information                                                                  |
| data           | object         | Response data                                                                           |
| data.results   | object array   | Check results in the same order as the input hosts                                      |

#### data.results[n]

| Parameter Name | Parameter Type | Description                                            |
| -------------- | -------------- | ------------------------------------------------------ |
| status         | string         | Check status, listed below                             |
| matched        | object         | Matched existing host; may be empty if no host matches |

#### status

| Value                   | Description                                                                                                         |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------- |
| host_not_found          | Host does not exist                                                                                                 |
| networkunit_not_found   | Current or target network unit does not exist or is unassigned                                                      |
| networkunit_mismatch    | Target unit and host belong to different network areas                                                              |
| version_not_found       | A version mapping was provided without an entry for the host platform; does not prove that a package is unavailable |
| node_status_not_allowed | Current node state does not allow upgrade                                                                           |
| networkunit_changed     | Upgrade changes the network unit and requires confirmation                                                          |
| normal_upgrade          | Checks passed; upgrade may be submitted                                                                             |
| cpu_arch_missing        | CPU architecture was not supplied                                                                                   |

#### data.results[n].matched

| Parameter Name          | Parameter Type | Description                                            |
| ----------------------- | -------------- | ------------------------------------------------------ |
| bk_host_id              | int64          | Matched host ID                                        |
| bk_biz_id               | int64          | Matched host business ID                               |
| bk_networkarea_id       | int64          | Matched host network area ID                           |
| bk_networkunit_id       | int64          | Matched host current network unit ID                   |
| os_type                 | string         | Operating system                                       |
| node_role               | string         | Node role: `agent` or `proxy`; empty string when unset |
| bk_host_innerip_list    | string array   | Matched host inner IPv4 list                           |
| bk_host_innerip_v6_list | string array   | Matched host inner IPv6 list                           |
