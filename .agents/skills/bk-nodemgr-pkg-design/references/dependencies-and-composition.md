# Dependencies and Composition

## Connect Through Capabilities

An interface is useful when it describes a real consumer need and hides implementation decisions. Multiple backends are not the only reason: independent callers can share a stable capability while the implementation evolves. Conversely, an interface mechanically mirroring every concrete method does not establish a useful boundary.

Make the relationship explicit in three places:

1. Implementation: a conformance assertion names the contract it fulfills.
2. Consumer: fields or parameters declare the required abstract capability.
3. Assembly: constructors select concrete implementations and inject them into consumers.

```go
// In an application-owned assembly location.
source := content.New(embeddedFiles)
httpAdapter := NewHTTPAdapter(source) // Constructor accepts content.IHandler.
worker := NewWorker(source)           // Constructor accepts content.IHandler.
```

Assembly knows the concrete source; ordinary operations do not. Dependency injection here means ordinary constructor arguments, not a new container or registry.

## Unified Entry, Narrow Consumers

When a package has cohesive read and write domains, its entrypoint can compose them:

```go
type IReader interface {
    Read(ctx context.Context, key string) ([]byte, error)
}
type IWriter interface {
    Write(ctx context.Context, key string, body []byte) error
}
type IHandler interface {
    IReader
    IWriter
}
```

A read-only consumer can accept `IReader`; the full handler still provides the package's unified entry. Only introduce subinterfaces when consumers or a genuine domain boundary justify them. If an operation is required, express that dependency at construction or compilation, not through `handler.(SomeOptionalInterface)` inside the operation.

## Parallel Implementation Versus Composition

| Situation                                                 | Approach                                                | What stays stable                                  |
| --------------------------------------------------------- | ------------------------------------------------------- | -------------------------------------------------- |
| A different storage mechanism fulfills the same operation | A parallel implementation                               | Contract and operation consumers                   |
| Existing operations need a coherent policy around them    | A composition accepting and implementing the capability | Underlying implementations and operation consumers |
| The required operation's meaning or result changes        | An explicit, authorized contract migration              | Unrelated APIs, not the affected signatures        |

Do not force incompatible behavior under one interface just to avoid a contract change.

## Worked Composition: Tenant-Scoped Object Access

Assume an object capability already provides `Read` and `Write`, and the service owns tenant identity and namespace policy. A service-private composition depends on the object interface and a narrow injected resolver. It also implements that same object interface.

```go
// In a service-private package, not the shared object capability.
type ITenantResolver interface {
    Resolve(ctx context.Context) (string, error)
}
type Handler struct {
    inner    object.IHandler
    resolver ITenantResolver
}

var _ object.IHandler = (*Handler)(nil)

func New(inner object.IHandler, resolver ITenantResolver) *Handler
func (h *Handler) Read(ctx context.Context, key string) ([]byte, error)
func (h *Handler) Write(ctx context.Context, key string, body []byte) error
```

The declarations illustrate relationships, not runnable method bodies. Both operations resolve the tenant, validate and scope the logical key, then delegate with the same context. Do not resolve the current tenant once at construction or store mutable request identity on a shared handler.

- An empty or invalid tenant fails before delegation; never silently fall back to an unscoped operation.
- Reads and writes use one consistent mapping. Respect the underlying key grammar and confinement guarantees; string concatenation alone does not prove isolation.
- Preserve the result, error identity, cancellation, and ownership contract unless an explicit policy says otherwise.
- The composition owns tenant policy. The underlying handler continues owning storage and protocol details.

Assembly chooses `object.NewMemory()` or `object.NewDisk(root)`, wraps it with `tenantobject.New(base, resolver)`, and injects the result as `object.IHandler`. Existing concrete-returning constructors remain valid. Callers continue expressing logical operations without per-caller tenant branches.

A future audit layer can likewise accept and implement `object.IHandler`. Do not build it before needed. When requested, choose the order deliberately: outside the tenant layer it can observe logical keys and rejected attempts; inside it observes accepted operations after namespace mapping. Same signatures do not make these orders behaviorally equivalent.

## Avoid False Abstractions

- Do not make callers select, unwrap, or type-assert the backend to use a required capability.
- Do not add one type whose name field switches between unrelated responsibilities when separate implementations would make those responsibilities explicit.
- Do not create subpackages, alternate implementations, or global configuration just to demonstrate extensibility.
- Do not move service policy into a shared package merely because its wrapper implements a shared interface.
- Do not prescribe `cli` or protocol DTO layers for content lookup, pure calculation, or other capabilities without an external protocol.
