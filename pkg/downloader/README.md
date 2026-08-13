# downloader

## 设计意图

1. 提供统一的远程制品下载入口，让调用方通过 `Downloader` 接口获取已校验的 `fileiface.File`。
2. 将下载安全边界集中在一个包内：URL 白名单/黑名单、文件名校验、跳转拒绝、大小限制和 checksum 校验都在下载链路内完成。
3. 将远程响应转换为临时本地文件，调用方只处理 `fileiface.File` 和 `io.ReadCloser`，不直接管理下载目录或中间文件路径。

## 功能边界

此包负责：

- 基于 `config.Downloader` 构造下载器，包括 `AllowHosts`、`BlockHosts`、`MaxBytes` 和 tracing 配置。
- 校验下载 URL：仅允许 absolute `http` / `https` URL，并按 host 执行策略。
- 校验下载文件名：`DownloadOptions.Filename` 只能是 basename，不能包含路径、绝对路径、`.`、`..` 或 NUL。
- 通过 `pkg/rest/client` 拉取远程 stream，并拒绝 HTTP redirect。
- 将响应体写入临时文件，校验文件大小和 MD5，再返回 `fileiface.File`。
- 在错误路径清理临时文件；在成功路径把清理动作绑定到 `Content()` 返回的 reader 上。

此包不负责：

- 下载文件的业务解析、解压、安装或分发。
- 持久化保存下载结果；需要复用文件时应在调用方组合 `pkg/filecache` 等缓存能力。
- 调用方目标路径管理；本包不会把远程内容直接写入业务指定目录。
- checksum 的真实性保证；当前 MD5 只用于制品完整性兼容校验，不等价于安全认证。
- `pkg/rest/client` 的 URL escaped path 保真能力；如果需要修复 `%2F` 等 escaped path 的请求语义，应在 rest client 层统一支持。

## 核心流程

```mermaid
flowchart TD
    A[New(config.Downloader)] --> B[Download(ctx, rawURL, opts)]
    B --> C[validate filename]
    C --> D[validate URL policy]
    D --> E[GET raw stream]
    E --> F[write to temporary file]
    F --> G[check max bytes]
    G --> H[verify MD5]
    H --> I[return fileiface.File]
    I --> J[caller reads Content]
    J --> K[reader.Close cleans temporary file]
```

## 使用方式

### 构造下载器

```go
dl, err := downloader.New(config.Downloader{
    AllowHosts:       []string{"repo.example.com"},
    BlockHosts:       []string{"metadata.google.internal"},
    MaxBytes:         512 << 20,
    TraceServiceName: "downloader",
    TraceSampleRate:  1,
})
if err != nil {
    return fmt.Errorf("failed to create downloader: %w", err)
}
```

`AllowHosts` 为空时表示 allowlist 不限制 host，即允许所有 host 进入后续策略；`AllowHosts` 非空时只允许精确匹配的 hostname。`BlockHosts` 的优先级高于 `AllowHosts`：只要命中 blocklist 就始终拒绝。因此即使 `AllowHosts` 为空表示“全开”，也可以通过 `BlockHosts` 拦截某一个 host。匹配前会 trim、lowercase，并去掉末尾 `.`；不支持 wildcard，也不会自动允许子域名。

### 下载并读取文件

```go
file, err := dl.Download(nCtx, "https://repo.example.com/packages/agent.tgz", downloader.DownloadOptions{
    Filename: "agent.tgz",
    Checksum: downloader.Checksum{
        Algorithm: downloader.ChecksumAlgorithmMD5,
        Value:     "d41d8cd98f00b204e9800998ecf8427e",
    },
})
if err != nil {
    return fmt.Errorf("failed to download package: %w", err)
}

reader, err := file.Content(nCtx)
if err != nil {
    return fmt.Errorf("failed to open downloaded package: %w", err)
}
defer reader.Close()

// Consume reader here, for example upload it to another component or copy it to a controlled path.
```

调用方通常只需要依赖 `Downloader` 接口：

```go
type Installer struct {
    downloader downloader.Downloader
}
```

## 文件处理语义

1. `Download` 会先把响应体写入由 `pkg/runtime/tmp` 创建的临时文件，文件名来自 `DownloadOptions.Filename`。
2. 临时文件写入后会通过 `pkg/filex/local` 重新打开为 `fileiface.File`，并读取 `Info().Size` 与 `Info().MD5` 做校验。
3. 如果任一步失败，下载器会清理临时文件和临时目录；清理失败只记录 `System` 日志，不覆盖原始错误。
4. 如果下载成功，返回的 `fileiface.File` 仍然指向临时文件；调用 `Content(nCtx)` 会打开 reader。
5. 调用方必须关闭 `Content()` 返回的 `io.ReadCloser`。该 `Close()` 会关闭文件句柄，并触发一次临时文件和临时目录清理。
6. 不要把返回文件当作长期文件引用；需要持久化或缓存时，应由调用方读取 reader 后写入自己的受控存储。

## URL 策略

1. 只支持 `http` 和 `https`。
2. URL 必须是 absolute URL，不能包含 credentials，不能使用 opaque URL。
3. fragment 会被移除，不参与请求。
4. IPv6 地址必须使用 `[]` 包裹；带 zone 的 IPv6 不作为支持目标。
5. host 策略先判断 `BlockHosts`，命中 blocklist 直接拒绝；未命中时再判断 `AllowHosts`。`AllowHosts` 为空表示 allowlist 全开，不限制 host，但仍会受 `BlockHosts` 约束；`AllowHosts` 非空时，下载 host 必须命中 allowlist。
6. HTTP redirect 始终被拒绝，避免请求在校验后跳出初始 trust boundary。
7. URL escaped path 是否被原样发送由 `pkg/rest/client` 的 URL 构造能力决定；该能力不在 downloader 包内局部修复。

## Checksum 与大小限制

1. 当前只支持 `ChecksumAlgorithmMD5`。
2. `Checksum.Algorithm` 必须显式设置；零值或未知算法会返回错误，不会跳过校验。
3. `Checksum.Value` 应为预期 MD5 hex 串；下载文件的 `Info().MD5` 必须与其完全一致。
4. `MaxBytes <= 0` 时使用 `DefaultMaxBytes`，即 `1 GiB`。
5. `Content-Length` 大于限制时会提前失败；最终仍以落盘后文件大小为准，因为 `Content-Length` 可能缺失或不可信。

## 错误语义

此包提供三个 sentinel errors，调用方应使用 `errors.Is` 判断：

- `ErrDownloadFailed`：URL、HTTP 请求、临时文件、checksum 算法等通用下载失败。
- `ErrDownloadTooLarge`：响应声明或实际落盘文件超过 `MaxBytes`。
- `ErrChecksumMismatch`：下载内容 MD5 与期望值不一致。

示例：

```go
file, err := dl.Download(nCtx, rawURL, opts)
if err != nil {
    switch {
    case errors.Is(err, downloader.ErrChecksumMismatch):
        return fmt.Errorf("downloaded package checksum mismatch: %w", err)
    case errors.Is(err, downloader.ErrDownloadTooLarge):
        return fmt.Errorf("downloaded package is too large: %w", err)
    default:
        return fmt.Errorf("failed to download package: %w", err)
    }
}
```

## 使用限制

1. `DownloadOptions.Filename` 只能传 basename；不要传入业务路径或远程 URL 中未清洗的文件名。
2. `DownloadOptions.Checksum` 必须来自可信的制品元数据；不要为了兼容临时跳过 checksum。
3. 不要依赖 `file.AbsDirs()` 或临时文件路径做业务持久化；该路径会在 reader close 后被删除。
4. 不要重复关闭同一个 `Content()` reader；也不要在把 reader 交给会接管关闭权的函数后再 `defer reader.Close()`。
5. 不要在此包中加入业务重试、解压、缓存或安装逻辑；这些属于调用方 workflow 或 storage/cache 层。
