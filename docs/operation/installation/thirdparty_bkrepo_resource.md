# BKRepo 资源使用说明

`bk-nodemgr` 通过 `repo` 配置直连 `BKRepo`，不通过 `APIGateway`。本文件只记录当前代码已经使用的 `BKRepo` 资源，后续
`BKRepo` 特有资源可在此继续补充。

## 访问范围

| 配置项                                 | 资源语义                                                         | 代码来源                               |
|-------------------------------------|--------------------------------------------------------------|------------------------------------|
| `repo.endpoint`                     | `BKRepo` 服务地址                                                | `internal/file/service/service.go` |
| `repo.projectID`                    | `BKRepo` project；多租户模式下实际请求使用 `<systemTenantID>.<projectID>` | `pkg/thirdparty/bkrepo/bkrepo.go`  |
| `repo.repoName`                     | `BKRepo` repository                                          | `pkg/thirdparty/bkrepo/bkrepo.go`  |
| `repo.accessKey` / `repo.secretKey` | `BKRepo` Basic Auth 凭据                                       | `pkg/thirdparty/bkrepo/bkrepo.go`  |

## 目录资源

`file` 服务启动时会确保以下 `FileGroup` 目录存在。这些目录都位于上表的 `projectID` + `repoName` 下。

| 目录                          | 用途                       |
|-----------------------------|--------------------------|
| `origin/agent`              | Agent 原始包目录              |
| `origin/server`             | Server 原始包目录             |
| `origin/proxy`              | Proxy 原始包目录              |
| `origin/cert`               | 证书原始包目录                  |
| `origin/bintool`            | 二进制工具原始包目录               |
| `origin/v2/plugin`          | V2 插件原始包目录               |
| `origin/v2/external_plugin` | V2 external plugin 原始包目录 |
| `origin/v3/plugin`          | V3 插件原始包目录               |
| `origin/plugin_bintool`     | 插件二进制工具原始包目录             |
| `release/agent`             | Agent 发布包目录              |
| `release/proxy`             | Proxy 发布包目录              |
| `release/cert`              | 证书发布包目录                  |
| `release/bintool`           | 二进制工具发布包目录               |
| `release/plugin_bintool`    | 插件二进制工具发布包目录             |
| `release/plugin`            | 插件发布包目录                  |

## 接口资源

| 操作     | BKRepo resource                                            | 用途             |
|--------|------------------------------------------------------------|----------------|
| `GET`  | `generic/{projectID}/{repoName}/{path}?download=true`      | 下载文件内容         |
| `PUT`  | `generic/{projectID}/{repoName}/{path}`                    | 上传文件内容         |
| `GET`  | `repository/api/node/detail/{projectID}/{repoName}/{path}` | 查询文件或目录节点详情    |
| `GET`  | `repository/api/node/page/{projectID}/{repoName}/{path}`   | 分页列出目录下的文件和子目录 |
| `POST` | `repository/api/node/mkdir/{projectID}/{repoName}/{path}`  | 创建目录           |
| `POST` | `repository/api/node/copy`                                | 复制文件或目录节点      |
| `DELETE` | `repository/api/node/delete/{projectID}/{repoName}/{path}` | 删除文件或目录节点      |

## 代码入口

| 场景接口                                             | 触发的 BKRepo resource                                                      |
|--------------------------------------------------|--------------------------------------------------------------------------|
| `GetFileGroup(path)`                             | `repository/api/node/detail/{projectID}/{repoName}/{path}`               |
| `EnsureFileGroup(path)`                          | 先查询目录；目录不存在时调用 `repository/api/node/mkdir/{projectID}/{repoName}/{path}` |
| `FileGroup.SubGroups()` / `FileGroup.AllFiles()` | `repository/api/node/page/{projectID}/{repoName}/{path}`，并按节点类型区分目录和文件   |
| `FileGroup.GetFile(name)` / `GetFile(path)`      | `repository/api/node/detail/{projectID}/{repoName}/{path}`               |
| `File.Content()`                                 | `generic/{projectID}/{repoName}/{path}?download=true`                    |
| `FileGroup.Store(...)`                           | `generic/{projectID}/{repoName}/{path}`                                  |
| `FileGroup.Copy(...)`                            | `repository/api/node/copy`                                               |
| `FileGroup.Remove(path)`                         | `repository/api/node/delete/{projectID}/{repoName}/{path}`                |

多租户模式下，file manager 的 upstream `FileGroup.Copy` 从 `/system/{basePath}` 复制到请求租户的
`/{tenantID}/{basePath}`；BKRepo project 和 repository 保持不变。

## bkrepo 系统注册

### 创建 Repo

![](docs/operation/installation/img/bkrepo_create_repo_step_1.png)

![](docs/operation/installation/img/bkrepo_create_repo_step_2.png)

### 创建虚拟用户

![](docs/operation/installation/img/bkrepo_create_viture_user_step_1.png)

![](docs/operation/installation/img/bkrepo_create_viture_user_step_2.png)

### 创建Token

![](docs/operation/installation/img/bkrepo_generate_token_step_1.png)

![](docs/operation/installation/img/bkrepo_generate_token_step_2.png)

![](docs/operation/installation/img/bkrepo_generate_token_step_3.png)

### 编写配置文件

```yaml
repo: 
  accessKey: <虚拟用户名称>
  endpoint: <bkrepo域名>
  projectID: <projectID>
  repoName: <repoName>
  secretKey: <虚拟用户Token>
  traceSampleRate: 0
  traceServiceName: file-client-repo
```

配置示例如下：

```yaml
repo: 
  accessKey: g_nodemgr
  endpoint: https://bkrepo.example.com/
  projectID: bk-nodemgr
  repoName: prod
  secretKey: xxxxxxxxxxxxxx
  traceSampleRate: 0
  traceServiceName: file-client-repo
```
