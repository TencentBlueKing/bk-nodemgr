|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/runtime/criteria
|OVERVIEW:Defines portable runtime criteria constants/enums for system-level decisions:{OSType,CPUArch,NetType,NetEndpointType,AdminUser}
|OVERVIEW:Boundary=business-agnostic runtime concepts only; no service policy, domain workflow state, product feature flags, or UI/API-specific enums
|Structure:pkg/runtime/criteria:{criteria.go,os_type.go,cpu_arch.go,network.go,admin_user.go,README.md}
|WHERE TO LOOK:package doc:criteria.go|intent/boundary:README.md
|WHERE TO LOOK:OS criteria:os_type.go:{OSType,OS*,Validate,String,StringListToOSTypeList,OSTypeListToStringList}
|WHERE TO LOOK:CPU criteria:cpu_arch.go:{CPUArch,CPUArch*,Validate,String,ToPkgArch,StringListToCPUArchList,CPUArchListToStringList}
|WHERE TO LOOK:network criteria:network.go:{NetType,NetEndpointType,Validate}
|WHERE TO LOOK:admin defaults:admin_user.go:{AdminUser,DefaultAdminUser,Validate,String}
|CONVENTIONS:Keep values portable and aligned with Go/runtime or platform-neutral naming before adding new constants
|CONVENTIONS:Every exported type/constant/function keeps English godoc; validation methods return wrapped/explicit errors with invalid value context
|CONVENTIONS:Preserve stable string values because callers may persist/compare them across modules and releases
|CONVENTIONS:Prefer adding enum values + Validate coverage in the owning file; keep conversion helpers backward-compatible and mark legacy helpers deprecated instead of deleting
|CONVENTIONS:Package must remain dependency-light; use stdlib only unless a runtime-primitive need is proven
|ANTI-PATTERNS:Do not add business constants, workflow statuses, service-specific policy, tenant/product concepts, or frontend-only options here
|ANTI-PATTERNS:Do not rename/remove existing constants or change string values without coordinated migration of all persisted/API consumers
|ANTI-PATTERNS:Do not introduce logging, configuration loading, storage, network calls, or cross-service imports in this package
|ANTI-PATTERNS:Do not duplicate runtime criteria in callers; extend this package when the concept is truly system-level and reusable
