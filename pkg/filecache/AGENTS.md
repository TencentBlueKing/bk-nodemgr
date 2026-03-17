# FILECACHE KNOWLEDGE BASE

## OVERVIEW

`pkg/filecache` provides a generic local file cache with MD5-based deduplication, atomic writes,
per-filename download serialization, optional startup restore, and background GC.

It is used by services that need to cache remote files locally (e.g. installer binaries downloaded
from the `file` service) to avoid redundant network downloads across workflow retries or concurrent requests.

Key capabilities:
- `IFileCache` interface with `GetOrFetch`, `FileExists`, `Close`
- MD5-keyed subdirectory isolation: `{cacheBaseDir}/{md5}/{filename}`
- two-tier locking: global `sync.RWMutex` for the in-memory index + per-filename `sync.Mutex` for download serialization
- double-check locking to prevent redundant downloads under concurrent requests
- `pendingDirs sync.Map` to guard GC against TOCTOU races with in-flight writes
- MD5 format validation in `isMD5Hex` to prevent path traversal via `expectedMD5`
- `safeRemoveAll` that refuses to delete paths outside `baseDir`
- startup restore via `restore()` — disabled by default (`RestoreOnStart: false`)
- background GC loop: removes expired entries (by last-access time) and orphaned MD5 dirs

## WHERE TO LOOK

- `IFileCache` interface and `FetchFn` type alias: `iface.go`
- Core implementation, `Options`, `New`, `GetOrFetch`, `download`, GC, restore: `cache.go`
- Unit tests (concurrency, MD5 mismatch, GC, cache hit/miss): `cache_test.go`

## CONVENTIONS

- Always inject via `IFileCache`; never depend on the concrete `*fileCache` type in callers.
- `fetchFn` is called at most once per (filename, MD5) pair per cache miss. It must return an `io.ReadCloser`
  that the cache layer will read and close — callers must not close it after passing it in.
- `GetOrFetch` returns `(fileiface.File, dirPath string, error)`. The caller obtains content by calling
  `file.Content(ctx)`, which opens a new `*os.File` handle. The caller owns this handle and must close it
  exactly once. **Do not close the handle a second time** (e.g. via a deferred close if the handle is also
  passed to a function that closes it internally, such as `sshx.Client.TransferFile` or
  `tmp.NewTempFileWithSpecialName`).
- `expectedMD5` must be a valid 32-character lowercase or uppercase hex string. Pass it through before use
  to avoid silent path traversal bugs.
- Log all cache events (hit, miss, download start/complete/fail, MD5 mismatch, GC) using `pkg/logger`.
- Use `New(nCtx, baseDir, Options{...})` to construct; pass the service-level context so the GC goroutine
  is bound to the service lifetime.
- Call `Close()` in `GracefulShutdown` to stop the GC goroutine cleanly before process exit.

## CALLER INTEGRATION PATTERN

```go
// 1. Construct once at service startup (e.g. in service.go).
fc, err := filecache.New(ctx, cfg.FileCache.Dir, filecache.Options{
    ExpirationTime: time.Duration(cfg.FileCache.ExpirationHours) * time.Hour,
    GCInterval:     time.Duration(cfg.FileCache.GCIntervalHours) * time.Hour,
    RestoreOnStart: cfg.FileCache.RestoreOnStart,
})

// 2. Inject via Capability / OptionFn.

// 3. Use in a workflow action (SSH example — TransferFile closes the reader):
cachedFile, _, err := fc.GetOrFetch(ctx, fileInfo.Name, fileInfo.MD5,
    func(nCtx contextx.IContext) (io.ReadCloser, error) {
        resp, err := fileHandler.DownloadInstaller(nCtx, osType, arch)
        if err != nil {
            return nil, err
        }
        return resp.Data, nil
    })
reader, err := cachedFile.Content(ctx)
// TransferFile owns and closes reader — do NOT defer reader.Close() here.
err = sshClient.TransferFile(reader, remotePath)

// 4. Use in a workflow action (WMI example — NewTempFileWithSpecialName closes the reader):
reader, err := cachedFile.Content(ctx)
// NewTempFileWithSpecialName owns and closes reader — do NOT defer reader.Close() here.
tmpFile, err := tmp.NewTempFileWithSpecialName(reader, toolName)
defer tmpFile.CleanUp()
err = wmiClient.UploadFile(ctx, tmpFile.Path(), destDir)

// 5. Shut down with the service.
_ = fc.Close()
```

## ANTI-PATTERNS

- Do not close the `io.ReadCloser` returned by `cachedFile.Content()` more than once; passing it to
  `TransferFile` or `NewTempFileWithSpecialName` already transfers ownership (double-close causes
  "file already closed" errors on Linux).
- Do not close the `io.ReadCloser` returned by `fetchFn` — the cache layer (via `LocalDir.Store`) owns it.
- Do not use an unvalidated string as `expectedMD5`; always ensure it comes from a trusted source
  (e.g. `InfoInstaller` response) and passes `isMD5Hex` before reaching `GetOrFetch`.
- Do not read from or write to the cache `baseDir` directly; the cache manages its own subdirectory layout.
- Do not share a single `IFileCache` instance across services with different `baseDir` configurations.
- Do not call `Close()` more than once; the underlying `stopCh` channel will panic on a second close.
- Do not set `RestoreOnStart: true` without considering startup latency — scanning large directories
  with many cached files will block `New()` until the scan completes.
