# Package Conventions

## Meaning Before Layout

A capability package owns an outcome and the details needed to produce it. Its consumers should not have to know the internal sequence, protocol, file layout, or implementation type to perform the promised operation.

For example, a content package can own embedded resource lookup and return original content. An HTTP host owns routes, response codes, headers, and rendering. A worker can consume the same content capability without pretending to be that HTTP host.

Do not rename `ApplicationHandler` to `ContentHandler` while retaining server configuration, response writing, or host policy inside the capability. The responsibility, not just the word, must change.

## Predictable Names

| Role                                  | Default convention                                                         | Boundary                                                          |
| ------------------------------------- | -------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Primary capability                    | `IHandler`                                                                 | The package qualifier supplies the domain                         |
| Cohesive sub-capability               | `IReader`, `IWriter`, or a meaningful `IHandler<Domain>`                   | Introduce for a real consumer or cohesive split, not every method |
| One concrete implementation           | `Handler`                                                                  | Export only when construction or direct use needs it              |
| Multiple same-package implementations | `EmbeddedHandler`, `RemoteHandler`, or private equivalents                 | Qualify the implementation mechanism, not the first caller        |
| Implementation subpackage             | `mechanism.Handler`                                                        | Use only when the boundary warrants a subpackage                  |
| Constructor                           | `New`, or `NewEmbedded` / `NewRemote` when selecting named implementations | Do not force unrelated existing names to migrate                  |
| Configuration                         | `Config`, with cohesive source-specific settings when needed               | No empty Config or unused options for symmetry                    |
| Consumer field                        | A role name with an interface type                                         | The declaration exposes the required capability                   |

Treat this as the default for new work, not permission to rename stable existing interfaces. Two independent packages conventionally expose `manifest.IHandler` and `template.IHandler`; they need not invent `ManifestReader` and `TemplateRenderer` as their main entrypoints. Methods still name the actual operations, such as `Read` and `Render`.

Uniformity reduces reader effort only if the responsibilities are coherent. Do not combine unrelated domains just to obtain one `IHandler`.

## Minimal Relationship Example

These are Go declaration sketches; imports and method bodies are intentionally omitted. `context.Context` is illustrative: use the target project's context contract in real code.

```go
// Package content exposes original document contents.
type IHandler interface {
    Read(ctx context.Context, key string) ([]byte, error)
}

type Handler struct {
    files fs.FS
}

var _ IHandler = (*Handler)(nil)

func New(files fs.FS) *Handler
func (h *Handler) Read(ctx context.Context, key string) ([]byte, error)
```

An independent worker accepts `content.IHandler`; it does not require `*content.Handler` or inspect `files`. The shared capability resolves its own keys and lookup failures. The worker owns what to do with the returned bytes.

The concrete return from `New` is compatible with abstract consumption. A factory may return `IHandler` when selecting among implementations; that does not make interface returns a universal constructor rule.

## Contracts Include Behavior

State the relevant semantics instead of expecting callers to infer them:

- Are keys logical identifiers or arbitrary filesystem paths? Which inputs are rejected?
- Does not-found differ from an empty successful result?
- Does a collection mean one page or the complete promised set?
- Who owns returned bytes, readers, handles, goroutines, or connections?
- Must context remain valid after the method returns?
- Which error identities are preserved and where are they mapped?
- What side effects, ordering, concurrency, or retry assumptions matter?

Only specify constraints relevant to the capability. Do not add every possible option or method to anticipate hypothetical consumers.

## File Responsibilities, Not Mandatory Scaffolding

Small packages can keep the contract, handler, constructor, and implementation together. Split files when there is a real reason to navigate or change them independently:

- `handler.go`: primary contract, implementation state, construction, capability-level operations.
- `handler_<domain>.go`: cohesive capability methods after domain size justifies a split.
- Mechanism-specific files: details of genuinely different implementations.
- A client/protocol file and named request/response types: only when external transport details actually exist.

Do not create empty `iface`, `factory`, `client`, `types`, or `options` files to satisfy a diagram. Do not expose a raw client and wire DTOs merely because they are already convenient internal objects.

## What Not to Copy From Examples

A mature package can demonstrate good boundary organization while retaining oversized interfaces, stale comments, or unnecessary exports. Extract the useful convention, not the whole package. Generic packages do not inherit external API-specific pagination, envelopes, authentication, or client layers without a matching requirement.
