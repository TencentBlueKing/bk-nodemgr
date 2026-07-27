### Description

- API Version: v3.0.1-alpha.56+.
- Required Permission: None.
- Function: Get installation script, configs, precheck, and package information for an offline Proxy installation operation.

**Permission Notes**: The current backend handler does not perform a direct permission check. The handler verifies that the operation is an offline Proxy installation operation and is waiting for the manual installation action.

### URL

POST /api/v3/node/workflow/operation/offline/info

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| operation_id | string | Yes | Offline Proxy installation operation ID |

### Request Example

```json
{
  "operation_id": "op-20240601-0001"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "install_script": "bash install_proxy.sh",
    "metadata": "{\"operation_id\":\"op-20240601-0001\"}",
    "configs": {
      "proxy.conf": "server=example.com"
    },
    "precheck": "bash precheck.sh",
    "release_info": {
      "generation": "v2",
      "os_type": "linux",
      "cpu_arch": "x86_64",
      "version": "2.0.0"
    },
    "installer_info": {
      "generation": "v2",
      "os_type": "linux",
      "cpu_arch": "x86_64"
    },
    "package_name": "bk-nodemgr-proxy-linux-x86_64.tgz"
  }
}
```

### Response Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| code | int32 | Status code, `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter | Type | Description |
| --- | --- | --- |
| install_script | string | Offline installation script content |
| metadata | string | Offline installation metadata |
| configs | object | Config file map. The key is filename and the value is file content |
| precheck | string | Precheck script content |
| release_info | object | Node package release information |
| installer_info | object | Installer information |
| package_name | string | Package name |

#### release_info

| Parameter | Type | Description |
| --- | --- | --- |
| generation | string | Package generation |
| os_type | string | OS type |
| cpu_arch | string | CPU architecture |
| version | string | Version |

#### installer_info

| Parameter | Type | Description |
| --- | --- | --- |
| generation | string | Installer generation |
| os_type | string | OS type |
| cpu_arch | string | CPU architecture |
