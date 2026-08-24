# import_v2_plugin

## 目的与适用场景

当第三方平台希望把一个由外部下载链接指向的 v2 插件包（官方插件V2包）导入到 bk-nodemgr package 服务、使其作为可发布条目出现时，使用 `import_v2_plugin`。

`import_v2_plugin` 描述的是导入入口；它不直接证明最终注册结果。第三方平台需要在导入之后通过 [import_result](import_result.md) 读取 workflow 执行结果来确认导入是否已经完成。

> 与 `install` / `upgrade` 等按 host 范围的 workflow 不同，`import_package` 不带 host scope，因此通用 `plugin/workflow/*` 列表查询无法直接命中。本接口不引导使用通用 workflow 列表查询。

## 输入

请求字段以 [pkg_workflow.proto `PackageImportPluginV2PkgReq`](../../../proto/backend/api/v3/pkg_workflow.proto) 为准：

| 字段           | Required | 含义                                                      |
| -------------- | -------- | --------------------------------------------------------- |
| `filename`     | yes      | 插件V2包文件名                                            |
| `download_url` | yes      | 插件V2包可下载 URL；导入过程中 bk-nodemgr 会从该 URL 拉取 |
| `md5`          | yes      | 插件V2包 MD5 校验值；用于校验下载内容                     |

调用方需要保证 `download_url` 在导入过程中对 bk-nodemgr 可达，且 `md5` 与所下载文件内容一致。

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与插件包来源信息。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export PKG_FILENAME="gse_plugin-2.0.0.tgz"
export PKG_DOWNLOAD_URL="https://files.example.com/packages/gse_plugin-2.0.0.tgz"
export PKG_MD5="d41d8cd98f00b204e9800998ecf8427e"

IMPORT_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/package/workflow/import/v2/plugin" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "filename": "${PKG_FILENAME}",
    "download_url": "${PKG_DOWNLOAD_URL}",
    "md5": "${PKG_MD5}"
  }
EOF
)"

WORKFLOW_ID="$(printf '%s' "${IMPORT_RESPONSE}" | jq -r '.data.workflow_id')"
```

## 系统解释

平台会把该请求解释为一次 v2 插件包导入：

- `download_url` 决定插件包来源；`md5` 决定导入后用于校验下载内容是否完整的标准。
- `filename` 决定在 release 记录中显示的插件包名。

`workflow_id` 是本次导入的标识。导入是否最终完成，需要通过 [import_result](import_result.md) 读取 workflow 状态与各 operation 的源信息确认。

## 即时输出

import 响应包含 `data.workflow_id`，用于标识本次导入。

`workflow_id` 不证明插件包已经下载、已经校验、已经写入 release 记录，也不证明插件包已经在 package 服务注册为可发布条目。

## 最终或机器侧可见产物

导入期望产生的可观察效果：

| 产物              | 说明                                                                              |
| ----------------- | --------------------------------------------------------------------------------- |
| workflow 执行结果 | 通过 [import_result](import_result.md) 读取；进入终态后 `data.is_finish = true`   |
| release 记录      | 独立验证项；通过 `package/release/plugin/list` 按 `name` + `version` 命中查询确认 |

公开 contract 不定义：

- 导入的最大等待时长上限（待确认）
- 重复 import 同一 `filename` + `md5` 组合的合并、替换或去重行为
- 导入过程中下载失败或 MD5 不匹配的具体错误返回

## 重复行为

公开 contract 不定义重复 `import_v2_plugin` 调用的合并、替换或去重行为，也不定义 idempotency key、retry window。

## 失败情况与限制

- 请求校验要求 `filename` 非空。
- 请求校验要求 `download_url` 非空。
- 请求校验要求 `md5` 非空。
- 成功的 import 响应只表示导入请求已接受，不证明导入已经完成。
- `download_url` 在导入过程中必须对 bk-nodemgr 可达；不可达或下载失败的具体错误返回不在本文档 contract 中定义（待确认）。
- `md5` 与下载内容不一致时的具体错误返回不在本文档 contract 中定义（待确认）。
- 导入是否最终完成，需要通过 [import_result](import_result.md) 读取 workflow 执行结果确认。

## Contract 参考

- [接入总览](README.md)
- [读取导入结果](import_result.md)
- [Swagger contract](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/pkg_workflow.proto)
- [术语表](../../concepts/glossary.md)
