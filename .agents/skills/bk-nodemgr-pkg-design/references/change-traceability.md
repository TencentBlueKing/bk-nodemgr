# Change Locality and Traceability

## Two Equal Success Criteria

- **Locality:** changes to implementation details do not force unrelated consumers to understand or adapt to them.
- **Traceability:** a necessary capability change has an explicit route through contracts, implementations, compositions, assembly, and consumers.

End-to-end changes are not inherently design failures. A needless cross-layer change, or an inability to determine the affected chain, is a failure. Do not promise that good abstraction eliminates contract migration.

## Follow Both Directions

Start at a real operation and trace:

`consumer -> required interface -> implementation -> delegated capability`

Then follow assembly to determine which implementations and compositions are selected. From a changed interface, enumerate all implementations, wrapping implementations, injected consumers, and callers of the affected operation.

Use semantic references, implementation search, interface assertions, constructor signatures, and call graphs where available. Check structural implementations without explicit assertions, composed/embedded interfaces, factory return values, and adapters too. A compiler and reference index are complementary: neither proves all semantic relationships by itself.

Record a compact impact map with the actual symbols or locations relevant to the task. Do not replace it with a directory inventory or “search all usages.” Keep calls and typed dependencies explicit rather than introducing string dispatch, global discovery, or concrete assertions.

## Worked Change: Bytes to a Caller-Owned Stream

Prefer compatible additive extension for a stable contract. Approval to support streaming alone does not authorize replacing an existing signature. This example assumes explicit approval of the in-place signature migration and its compatibility scope, with applicable project and scoped rules satisfied. Without that approval, retain the stable operation and propose a compatible streaming capability instead.

The explicitly approved migration changes:

```go
Read(ctx context.Context, key string) ([]byte, error)
```

into:

```go
Read(ctx context.Context, key string) (io.ReadCloser, error)
```

The affected graph includes two implementations, tenant and cache compositions, assembly, an HTTP adapter, and a worker. The change is genuinely contract-wide; do not keep the old contract and discover optional `Stream` methods from concrete types.

### Agree on the behavioral contract

- A successful open transfers a non-nil stream to the caller, including for empty content. The caller closes it on every exit path.
- Opening failure leaves no owned stream; the implementation cleans up partially acquired resources.
- Reading can fail after opening succeeds, including after partial content. Do not equate successful open with complete success.
- Context must remain valid during consumption. A returned stream must not depend on a child context canceled by a deferred cancel at method return.
- Define how close errors interact with primary read/operation errors, preserving useful error identity.

### Assign each necessary change

| Role            | Required adaptation                                                                                                                         |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `IHandler`      | Define the new return type, lifetime, context, and error contract                                                                           |
| `MemoryHandler` | Return independent reader state over stable content; preserve mutation/snapshot guarantees                                                  |
| `RemoteHandler` | Transfer the remote body instead of eagerly buffering; preserve context and late errors; clean up rejected responses                        |
| `TenantHandler` | Keep resolving and validating before delegation; do not consume or close the returned stream prematurely                                    |
| `CachedHandler` | Give each hit an independent reader; define miss population without silently restoring eager buffering; publish only complete valid entries |
| Assembly        | Verify selected implementations and composition order still satisfy the contract; avoid unrelated constructor migrations                    |
| HTTP adapter    | Keep context valid while copying, close on all exits, and distinguish opening errors from failures after committing the response            |
| Worker          | Consume incrementally, handle late errors and cancellation, close promptly per operation rather than deferring across a long loop           |

If cache storage cannot support the required streaming behavior, explain that concrete obstacle and agree on a compatible adaptation. Do not silently disable promised caching, cache partial data as complete, or introduce an unapproved cache subsystem.

### Verify the chain, not just the signature

Compile the affected graph and verify all required implementations and wrappers still satisfy `IHandler`. Then check the requirement's resource lifetime, cancellation, partial-read errors, complete-cache publication, and tenant delegation through applicable behavior checks.

A compile-time assertion identifies a signature mismatch. It does not identify a stream closed too early or a partial cache entry incorrectly advertised as complete.

Preserve unrelated names, constructors, backends, and operations. Plan coherent migration or compatibility steps appropriate to the actual release boundary; do not invent rollout machinery for an in-process change.

## Review Questions

1. What change is being isolated, and what contract remains stable?
2. If the contract changes, which implementations, compositions, assembly points, and consumers must follow?
3. Can a reader follow those relationships directly from declarations and construction?
4. Are wrappers still substitutable in result, tenant, error, resource, and cancellation semantics?
5. Which checks prove signatures, and which prove behavior?

More layers or interfaces are not evidence of success. Clear change responsibility and discoverable links are.
