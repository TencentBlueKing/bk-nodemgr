# 导入包格式 FAQ

## 1. 包管理上传的包必须是什么压缩格式？

包管理上传的原始包必须是 `.tgz` 文件，并且必须与包类型对应。不要把 Agent、Proxy、Server、插件、证书、工具包互相混用。

| 包类型        | 关键目录或文件                                                                                |
| ------------- | --------------------------------------------------------------------------------------------- |
| Agent 包      | 见第 2 节，按 `agent_<os>_<pkg_arch>` 区分平台                                                |
| Proxy 包      | 见第 2 节，单包只能包含一个 Linux 平台的 `server/bin`                                         |
| Server 包     | 见第 2 节，单包只能包含一个 Linux 平台的 `server/bin`                                         |
| 证书包        | 见第 4 节，只支持两种固定布局                                                                 |
| 公共工具包    | `bintool/<os>_<arch>/...`                                                                     |
| 插件工具包    | 见第 3 节，按 `plugin_bintool/v2/<platform>` 或 `plugin_bintool/v3/<platform>` 区分平台和版本 |
| V2 官方插件包 | 见第 3 节，按 `plugins_<os>_<pkg_arch>` 区分平台                                              |
| V2 外部插件包 | 见第 3 节，按 `external_plugins_<os>_<pkg_arch>` 区分平台                                     |
| V3 插件包     | 见第 3 节，按 `<plugin>/plugins_<os>_<pkg_arch>` 区分平台                                     |

## 2. GSE 包各平台目录结构是什么？

GSE 包包括 Agent、Proxy、Server。三类原始导入包都必须有一层外层目录，下例统一写作 `<pkg>/`。

### Agent 原始导入包

Agent 支持在同一个 `.tgz` 中放多个平台目录。每个平台目录名必须是 `agent_<os>_<pkg_arch>`。

| 平台           | 目录名                  | 二进制文件          |
| -------------- | ----------------------- | ------------------- |
| Linux x86_64   | `agent_linux_x86_64/`   | `bin/gse_agent`     |
| Linux aarch64  | `agent_linux_aarch64/`  | `bin/gse_agent`     |
| Windows x86_64 | `agent_windows_x86_64/` | `bin/gse_agent.exe` |
| Darwin x86_64  | `agent_darwin_x86_64/`  | `bin/gse_agent`     |

```text
<pkg>/VERSION
<pkg>/DESCRIPTION
<pkg>/DESCRIPTION_EN
<pkg>/support-files/templates/#etc#gse#gse_agent.conf
<pkg>/support-files/templates/gse_agent.conf.template
<pkg>/support-files/env/gse_agent.env
<pkg>/agent_linux_x86_64/bin/gse_agent
<pkg>/agent_linux_aarch64/bin/gse_agent
<pkg>/agent_windows_x86_64/bin/gse_agent.exe
<pkg>/agent_darwin_x86_64/bin/gse_agent
```

`#etc#gse#gse_agent.conf` 和 `gse_agent.conf.template` 是可识别的 Agent 配置模板路径；平台目录按实际需要提供，但每个有效平台目录下必须有对应平台的 `gse_agent` 二进制。

### Proxy 原始导入包

Proxy 原始导入包的二进制固定放在 `<pkg>/server/bin/` 下。单个 Proxy 原始包只能包含一个 Linux 平台，平台由 `gse_agent`、`gse_data`、`gse_file` 三个 ELF 二进制判定，三者必须是同一平台。

| 平台          | 二进制目录          | 必需二进制                          |
| ------------- | ------------------- | ----------------------------------- |
| Linux x86_64  | `<pkg>/server/bin/` | `gse_agent`、`gse_data`、`gse_file` |
| Linux aarch64 | `<pkg>/server/bin/` | `gse_agent`、`gse_data`、`gse_file` |

```text
<pkg>/VERSION
<pkg>/DESCRIPTION
<pkg>/DESCRIPTION_EN
<pkg>/server/bin/gse_agent
<pkg>/server/bin/gse_data
<pkg>/server/bin/gse_file
<pkg>/support-files/templates/#etc#gse#gse_agent.conf
<pkg>/support-files/templates/gse_agent.conf.template
<pkg>/support-files/templates/#etc#gse#gse_data_proxy.conf
<pkg>/support-files/templates/gse_data_proxy.conf.template
<pkg>/support-files/templates/#etc#gse#gse_file_proxy.conf
<pkg>/support-files/templates/gse_file_proxy.conf.template
<pkg>/support-files/env/gse_proxy.env
<pkg>/support-files/env/gse_agent.env
```

`#etc#...conf` 和 `*.conf.template` 是可识别的 Proxy 配置模板路径；`gse_proxy.env` 和 `gse_agent.env` 必须存在。

### Server 原始导入包

Server 原始导入包的二进制也固定放在 `<pkg>/server/bin/` 下。单个 Server 原始包只能包含一个 Linux 平台，平台由 `gse_file`、`gse_data` 两个 ELF 二进制判定。

| 平台          | 二进制目录          | 必需二进制             |
| ------------- | ------------------- | ---------------------- |
| Linux x86_64  | `<pkg>/server/bin/` | `gse_file`、`gse_data` |
| Linux aarch64 | `<pkg>/server/bin/` | `gse_file`、`gse_data` |

```text
<pkg>/VERSION
<pkg>/server/bin/gse_file
<pkg>/server/bin/gse_data
<pkg>/support-files/templates/#etc#gse#gse_file_proxy.conf
<pkg>/support-files/templates/gse_file_proxy.conf.template
<pkg>/support-files/templates/#etc#gse#gse_data_proxy.conf
<pkg>/support-files/templates/gse_data_proxy.conf.template
<pkg>/support-files/env/gse_proxy.env
```

`#etc#...conf` 和 `*.conf.template` 是可识别的 Server 配置模板路径；`VERSION`、`gse_file`、`gse_data` 必须存在。

### 发布后的平台包

发布后，系统会按平台生成 release 包。release 包不再保留原始外层目录和 `support-files/`，而是统一为 `bin/` 和 `cert/` 两类目录。

Agent release 包结构：

```text
bin/gse_agent
bin/<bintool files>
cert/gseca.crt
cert/gse_agent.crt
cert/gse_agent.key
cert/cert_encrypt.key
```

Proxy / Server release 包结构：

```text
bin/gse_agent
bin/gse_data
bin/gse_file
bin/<bintool files>
cert/gseca.crt
cert/gse_agent.crt
cert/gse_agent.key
cert/gse_server.crt
cert/gse_server.key
cert/gse_api_client.crt
cert/gse_api_client.key
cert/cert_encrypt.key
```

## 3. 插件包各平台目录结构是什么？

插件包包括 V2 官方插件、V2 外部插件、V3 插件和插件工具包。V2 与 V3 的目录结构不同，上传入口也不能混用。

### V2 官方插件原始导入包

V2 官方插件按顶层平台目录组织。平台目录名必须是 `plugins_<os>_<pkg_arch>`，例如 `plugins_linux_x86_64/`、`plugins_linux_aarch64/`、`plugins_windows_x86_64/`、`plugins_darwin_x86_64/`。

```text
plugins_linux_x86_64/<plugin>/project.yaml
plugins_linux_x86_64/<plugin>/bin/<plugin files>
plugins_linux_x86_64/<plugin>/etc/<config>.tpl
plugins_linux_aarch64/<plugin>/project.yaml
plugins_linux_aarch64/<plugin>/bin/<plugin files>
plugins_linux_aarch64/<plugin>/etc/<config>.tpl
```

严格限制：

1. 每个平台目录下必须有 `<plugin>/project.yaml`。
2. 配置模板只识别 `<plugin>/etc/*.tpl`。
3. 发布时会按平台生成 release 包，并把 `<plugin>/bin/` 写入 release 包的 `bin/`。

### V2 外部插件原始导入包

V2 外部插件也按顶层平台目录组织。平台目录名必须是 `external_plugins_<os>_<pkg_arch>`，例如 `external_plugins_linux_x86_64/`、`external_plugins_linux_aarch64/`、`external_plugins_windows_x86_64/`、`external_plugins_darwin_x86_64/`。

```text
external_plugins_linux_x86_64/<plugin>/project.yaml
external_plugins_linux_x86_64/<plugin>/bin/<plugin files>
external_plugins_linux_x86_64/<plugin>/etc/<config>.tpl
external_plugins_linux_x86_64/<plugin>/<other plugin files>
external_plugins_linux_aarch64/<plugin>/project.yaml
external_plugins_linux_aarch64/<plugin>/bin/<plugin files>
external_plugins_linux_aarch64/<plugin>/etc/<config>.tpl
```

严格限制：

1. 每个平台目录下必须有 `<plugin>/project.yaml`。
2. 配置模板只识别 `<plugin>/etc/*.tpl`。
3. 发布时会保留该平台 `<plugin>/` 下的插件文件结构，并把 `bin/` 下文件按可执行文件写入。

### V3 插件原始导入包

V3 插件按插件名作为顶层目录，平台目录放在插件目录下。平台目录名必须是 `plugins_<os>_<pkg_arch>`。

```text
<plugin>/project.yaml
<plugin>/plugins_linux_x86_64/definition.yaml
<plugin>/plugins_linux_x86_64/bin/<plugin files>
<plugin>/plugins_linux_x86_64/templates/<config>.template
<plugin>/plugins_linux_aarch64/definition.yaml
<plugin>/plugins_linux_aarch64/bin/<plugin files>
<plugin>/plugins_linux_aarch64/templates/<config>.template
```

严格限制：

1. 插件顶层目录下必须有 `<plugin>/project.yaml`。
2. 每个平台目录下必须有 `definition.yaml`。
3. 配置模板只识别 `templates/*.template`。
4. 发布时会按平台生成 release 包，并把对应平台目录下的 `bin/` 写入 release 包的 `bin/`。

### 插件工具包原始导入包

插件工具包按 generation 和平台组织。原始导入包中必须先分 `v2` / `v3`，再放平台目录。

```text
plugin_bintool/v2/linux_amd64/<tool files>
plugin_bintool/v2/linux_arm64/<tool files>
plugin_bintool/v2/darwin_amd64/<tool files>
plugin_bintool/v2/windows_amd64/<tool files>
plugin_bintool/v2/aix6_ppc64/<tool files>
plugin_bintool/v2/aix7_ppc64/<tool files>
plugin_bintool/v3/linux_amd64/<tool files>
plugin_bintool/v3/linux_arm64/<tool files>
plugin_bintool/v3/darwin_amd64/<tool files>
plugin_bintool/v3/windows_amd64/<tool files>
plugin_bintool/v3/aix6_ppc64/<tool files>
plugin_bintool/v3/aix7_ppc64/<tool files>
```

发布后，V2 和 V3 插件工具包会分别生成 release 包，包内不再保留 `v2` / `v3` 层级，统一为：

```text
plugin_bintool/linux_amd64/<tool files>
plugin_bintool/linux_arm64/<tool files>
plugin_bintool/darwin_amd64/<tool files>
plugin_bintool/windows_amd64/<tool files>
plugin_bintool/aix6_ppc64/<tool files>
plugin_bintool/aix7_ppc64/<tool files>
```

## 4. `cert` 证书包支持哪两种格式？

`cert` 原始导入包只支持以下两种格式，二选一。

格式一：八个证书文件直接放在 `.tgz` 根目录。

```text
gseca.crt
gse_agent.crt
gse_agent.key
gse_server.crt
gse_server.key
gse_api_client.crt
gse_api_client.key
cert_encrypt.key
```

格式二：八个证书文件全部放在 `.tgz` 内的 `cert/` 目录下。

```text
cert/gseca.crt
cert/gse_agent.crt
cert/gse_agent.key
cert/gse_server.crt
cert/gse_server.key
cert/gse_api_client.crt
cert/gse_api_client.key
cert/cert_encrypt.key
```

严格限制：

1. 只能使用上述两种布局之一。
2. 不能把部分证书放在根目录、部分证书放在 `cert/` 目录下。
3. 八个证书文件必须全部存在，文件名必须完全一致。
4. `cert` 导入包不要再额外套一层业务目录，例如不要使用 `package/cert/gseca.crt`。

发布后生成的 release 证书包会统一为 `cert.tgz`，包内路径统一为 `cert/<file>`。

## 5. 插件包 V2 和 V3 可以混用吗？

不能混用。包管理上传 V2 官方插件、V2 外部插件、V3 插件时，会按各自目录结构读取 `project.yaml`、`definition.yaml`、模板文件和平台目录。

如果 V2 包使用 V3 上传入口，或 V3 包使用 V2 上传入口，目录校验和元数据解析会失败。
