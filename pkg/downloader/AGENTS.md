|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/downloader
|Overview:shared remote artifact downloader|validates URL policy→fetches HTTP/HTTPS stream→stores temporary file→verifies checksum→returns fileiface.File
|Structure:pkg/downloader:{iface.go,downloader.go,common.go,temp_file.go,downloader_test.go,README.md}
|Where to look:public interface:iface.go:Downloader.Download returns verified temporary-backed fileiface.File
|Where to look:construction/config:downloader.go:New consumes config.Downloader allow/block/maxBytes/tracing settings
|Where to look:URL and filename policy:common.go:{validateDownloadURL,validateDownloadFilename}
|Where to look:file lifecycle:temp_file.go:{temporaryFile.Content,temporaryFileReader.Close}|downloader.go:storeVerifiedContent
|Where to look:behavior tests:downloader_test.go:URL policy, filename safety, checksum/size/temp cleanup expectations
|Dependencies:uses pkg/config.Downloader|pkg/contextx.IContext|pkg/filex/iface.File|pkg/filex/local|pkg/rest/client RawStream|pkg/runtime/tmp|pkg/runtime/ssl|pkg/tracing|pkg/logger
|Conventions:inject via Downloader interface; do not depend on concrete *client outside package
|Conventions:DownloadOptions.Filename must be a basename only; never accept path, slash, backslash, absolute path, dot, dotdot, or NUL
|Conventions:URL policy allows only http/https absolute URLs; rejects credentials, fragments, opaque URLs, bare IPv6, IPv6 zones, invalid ports, blocklisted hosts, and non-allowlisted hosts when allowHosts exists
|Conventions:allowHosts matching is exact after trim+lowercase+trailing-dot removal; no wildcard or implicit subdomain matching
|Conventions:redirects are rejected; trust boundary stays on initially validated URL
|Conventions:Checksum currently supports MD5 only; unsupported/zero algorithm must fail instead of skipping verification
|Conventions:escaped path preservation belongs to pkg/rest/client URL construction; do not patch it locally in downloader
|Conventions:MaxBytes<=0 falls back to DefaultMaxBytes(1GiB); enforce both advertised Content-Length fast-fail and stored file size check
|Conventions:Download returns temporary-backed fileiface.File; caller must call Content(ctx), consume reader, and close returned io.ReadCloser to remove temp file+dir
|Conventions:response body is owned and closed by downloader storage path; caller owns only the reader returned by file.Content(ctx)
|Conventions:error wrapping must preserve sentinel matching with errors.Is for ErrDownloadFailed, ErrDownloadTooLarge, ErrChecksumMismatch
|Conventions:lower-level failures return wrapped errors; only temp cleanup failure is logged with logger.G.Sys().WithErr(err).With("file", path)
|Anti-patterns:do not write downloaded content to caller-chosen destination path; this package exposes verified temp-backed fileiface.File only
|Anti-patterns:do not add checksum algorithms or auth semantics without updating Checksum contract, tests, README, and callers
|Anti-patterns:do not silently skip checksum when Checksum.Value is empty or Algorithm is zero
|Anti-patterns:do not allow redirects, credentials, wildcard hosts, path-like filenames, or direct temp directory access
|Anti-patterns:do not log-and-return ordinary download failures inside package; caller boundary decides logging
|Commands:test package:go test ./pkg/downloader/...
