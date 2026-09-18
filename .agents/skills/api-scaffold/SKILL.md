---
name: api-scaffold
description: Scaffold API protocol layer infrastructure for bk-nodemgr. Use when users want to add a new API endpoint, create interface scaffolding, set up proto definitions, or build API infrastructure. Triggers on phrases like "新增 API", "创建接口", "搭建协议层", "帮我搭建 xxx 接口的基础设施", "先把 proto 和转换层搞好", or any request to create a new REST API endpoint in the backend/file/relay/application services.
---

# API Scaffold Skill

This skill helps scaffold the protocol layer infrastructure for new API endpoints in bk-nodemgr. It handles proto definitions, code generation, conversion layer implementation, route registration, and handler scaffolding — everything except the core business logic.

## What This Skill Does

When a user wants to add a new API endpoint, this skill automates the repetitive infrastructure work:

1. **Proto Definition** - Add Request/Response messages and RPC method to the appropriate proto file
2. **Proto Generation** - Run `make clean && make all` to regenerate pb.go files
3. **Conversion Layer** - Implement Validate() and Convert methods in pkg/proto
4. **Route Registration** - Register the new endpoint in the router's Load() function
5. **Handler Scaffold** - Create a handler file with the standard structure (BindJSON, error handling, response conversion) but leave business logic as TODO

The user then fills in the business logic (service calls, data processing, etc.) and error handling specific to their use case.

## When to Use This Skill

Use this skill when the user:
- Wants to add a new API endpoint to backend/file/relay/application services
- Says "新增 API", "创建接口", "搭建协议层"
- Asks to "set up the infrastructure for a new endpoint"
- Wants proto definitions and conversion layer without business logic
- Mentions creating a REST API handler in internal/*/router/api-v3

## Information Gathering

Before starting, collect these details from the user:

### Required Information

1. **Service name**: Which service? (backend, file, relay, application)
2. **Module name**: Which module? (auth, node, package, networkarea, etc.)
3. **Endpoint name**: What's the API called? (e.g., "Authorized", "ListNodes")
4. **Request structure**: What fields does the request need?
   - Field names and types
   - Whether it's a single item or array (批量操作)
   - Validation rules (required fields, constraints)
5. **Response structure**: What does the response return?
   - Field names and types
   - Whether it's a single result or array
   - Nested structures if any

### Optional Information

- **Reusable types**: Can we reuse existing proto messages? (e.g., AuthResource)
- **Similar endpoints**: Is there a similar endpoint to reference for patterns?
- **Special requirements**: Any special validation, error handling, or response format needs?

If the user's request is vague, ask clarifying questions. For example:
- "Should this be a batch operation (array input) or single item?"
- "What fields do you need in the request? Any required validations?"
- "What should the response contain?"

## Implementation Workflow

Follow these steps in order. Create a task list at the start to track progress.

### Step 0: Create Initial API Documentation (Optional but Recommended)

Before implementing, create an initial version of the API documentation. This helps clarify the interface contract and serves as a placeholder that will be completed later using the api-doc skill.

**Files**:
- `apigw/apidocs/zh/{OperationId}.md` (Chinese)
- `apigw/apidocs/en/{OperationId}.md` (English)

**Naming**: Use swagger operationId format (e.g., `NodeAgent_NodeAgentInstall.md`)

**Template** (simplified version, fill only what you know):
```markdown
### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：[权限名称，暂时可留空]。
- 该接口功能描述：[一句话说明接口功能]。

### URL

[HTTP_METHOD] /api/v3/[service]/[resource]/[action]

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| param1 | string | 是 | [参数描述] |
| param2 | int64 | 否 | [参数描述] |

### 调用示例

```json
{
  "param1": "value",
  "param2": 123
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "field1": "value",
    "field2": true
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| field1 | string | [字段描述] |
| field2 | bool | [字段描述] |
```

**Key points**:
- This is an **initial version** - fill only what you know from the user's requirements
- Use the standard API doc template structure from `docs/developer/api_doc_template.md`
- Leave placeholders (like permission name) if information is not yet available
- After implementation is complete, use the **api-doc skill** to generate the full documentation with:
  - Complete field descriptions from proto comments
  - Business logic details from code
  - Error cases and edge conditions
  - Usage scenarios and examples

**When to skip**: If the user wants to implement first and document later, or if the interface is extremely simple.

### Step 1: Modify Proto Definition

**File**: `proto/{service}/api/v3/{module}.proto`

Add the Request, Response, and RPC method definitions. Follow these patterns:

**For batch operations** (array input/output):
```protobuf
message AuthorizedReq {
  repeated AuthorizedItem items = 1;
}

message AuthorizedItem {
  string action = 1;
  string resource_type = 2;
}

message AuthorizedResp {
  message Data {
    repeated AuthorizedResult results = 1;
  }

  string code = 1;
  string message = 2;
  string request_id = 3;
  google.protobuf.Struct error = 4;
  Data data = 5;
}

message AuthorizedResult {
  string action = 1;
  string resource_type = 2;
  bool is_any = 3;
  repeated AuthResource resources = 4;
}
```

**For single operations**:
```protobuf
message GetNodeReq {
  string node_id = 1;
}

message GetNodeResp {
  message Data {
    NodeInfo node = 1;
  }

  string code = 1;
  string message = 2;
  string request_id = 3;
  google.protobuf.Struct error = 4;
  Data data = 5;
}
```

**Add RPC method to service**:
```protobuf
service Auth {
  rpc Verify(AuthVerifyReq) returns (AuthVerifyResp) {}
  rpc Authorized(AuthorizedReq) returns (AuthorizedResp) {}  // New method
}
```

**Key patterns**:
- Response always has the standard structure: code, message, request_id, error, data
- Data is a nested message containing the actual payload
- Use `repeated` for arrays
- Reuse existing messages when possible (e.g., AuthResource)

### Step 2: Regenerate Proto Files

Run proto generation from the proto directory:

```bash
cd proto && make clean && make all
```

This generates:
- `pkg/proto/{service}/api/v3/{module}.pb.go` - Generated proto code
- `docs/api/swagger/{service}/api/v3/{module}.swagger.json` - Swagger docs

Verify the generation succeeded (exit code 0, no errors).

### Step 3: Implement Conversion Layer

**File**: `pkg/proto/{service}/api/v3/{module}.go`

Implement these methods for the Request:

**Validate()** - Check required fields:
```go
func (x *AuthorizedReq) Validate() error {
    for i, item := range x.GetItems() {
        if item.GetAction() == "" {
            return fmt.Errorf("items[%d].action is required", i)
        }
        if item.GetResourceType() == "" {
            return fmt.Errorf("items[%d].resource_type is required", i)
        }
    }
    return nil
}
```

**AutoConvert()** - Usually empty unless there's automatic conversion logic:
```go
func (x *AuthorizedReq) AutoConvert() {
    // Usually empty
}
```

**Convert methods** - Transform between proto and internal types:
```go
// Convert proto request to internal types
func (x *AuthorizedReq) ConvertItemsToInternal() []InternalType {
    return conv.SliceToSlice(x.GetItems(), func(item *AuthorizedItem) InternalType {
        return InternalType{
            Action:       item.GetAction(),
            ResourceType: item.GetResourceType(),
        }
    })
}
```

Implement these methods for the Response:

**Convert methods** - Transform internal results to proto:
```go
func (x *AuthorizedResp) ConvertResultsFromScopes(scopes []auth.AuthorizedScope, items []*AuthorizedItem) {
    x.Data = &AuthorizedResp_Data{
        Results: make([]*AuthorizedResult, len(scopes)),
    }

    for i, scope := range scopes {
        x.Data.Results[i] = &AuthorizedResult{
            Action:       items[i].GetAction(),
            ResourceType: items[i].GetResourceType(),
            IsAny:        scope.IsAny,
            Resources:    ConvertInternalToAuthResources(scope.Resources),
        }
    }
}
```

**Helper functions** - Add any needed conversion helpers:
```go
func convertAuthResourceFromTypes(resources []types.AuthResource) []*AuthResource {
	return conv.SliceToSlice(resources, func(resource types.AuthResource) *AuthResource {
		return &AuthResource{
			SystemId: resource.SystemID,
			Type:     string(resource.Type),
			Id:       resource.ID,
		}
	})
}
```

**Key patterns**:
- Use `conv.SliceToSlice` for array transformations
- Validate all required fields with descriptive error messages
- Convert methods should handle nil/empty cases gracefully
- Reuse existing conversion helpers when possible

### Step 4: Register Route

**File**: `internal/{service}/router/api-v3/{module}/{module}.go`

Add the route registration in the `Load()` function:

```go
func Load(rg *gin.RouterGroup, capability *options.Capability) {
    h := newHandler(rg, capability)
    h.rg.POST("/verify", restserver.Handler(h.Verify))
    h.rg.POST("/authorized", restserver.Handler(h.Authorized))  // New route
}
```

**Pattern**: Use `restserver.Handler()` wrapper for all routes.

### Step 5: Create Handler Scaffold

**File**: `internal/{service}/router/api-v3/{module}/{endpoint_name}.go`

Create a new file with this structure:

```go
package {module}

import (
    "github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
    protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/{service}/api/v3"
    resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
    "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/restserver"
)

// {EndpointName} handles the {endpoint_name} request
func (h *handler) {EndpointName}(rCtx restserver.IContext) (interface{}, error) {
    req := new(protoBackend.{EndpointName}Req)
    if err := rCtx.BindJSON(req); err != nil {
        logger.G.Biz(rCtx).WithErr(err).Error("failed to {endpoint_name}, failed to decode request body")
        return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
    }

    // TODO: Implement business logic here
    // Call service/storage layer to get data
    // Example:
    // results, err := h.service{Something}.Query{Something}(rCtx, req.GetItems())
    // if err != nil {
    //     logger.G.Biz(rCtx).WithErr(err).Error("failed to {endpoint_name}")
    //     return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
    // }

    resp := new(protoBackend.{EndpointName}Resp)
    // TODO: Convert results to response
    // resp.ConvertResultsFrom...(results)

    return resp.GetData(), nil
}
```

**What's included**:
- Package and imports
- Handler function signature
- BindJSON with error handling and logging
- TODO markers for business logic
- Response structure initialization
- Return statement

**What's NOT included (user fills in)**:
- Business logic (service/storage calls)
- Error handling for business operations
- Response conversion method call

**IMPORTANT - Scaffold Boundaries**:

This is a SCAFFOLD ONLY. Do NOT implement:
- ❌ Calls to `h.storage*` or `h.service*` methods
- ❌ Database queries or data processing logic
- ❌ Business rules or domain logic
- ❌ Actual error handling beyond BindJSON

The handler should compile and return an empty/nil response. The user will fill in the business logic later.

### Step 6: Verify Implementation

Run LSP diagnostics on all modified files:

```bash
# Check for type errors, undefined references, etc.
lsp_diagnostics internal/{service}/router/api-v3/{module}/
lsp_diagnostics pkg/proto/{service}/api/v3/{module}.go
```

Ensure:
- No type errors
- No undefined references
- All imports are correct
- Proto generation succeeded

If there are errors, fix them before proceeding.

**Verify scaffold boundaries**:
```bash
# Check that handler does NOT call storage/service layers
grep -n "h\.storage" internal/{service}/router/api-v3/{module}/{endpoint_name}.go
grep -n "h\.service" internal/{service}/router/api-v3/{module}/{endpoint_name}.go
```

These greps should return NO matches. If they do, remove the business logic calls - this is a scaffold only.


### Step 7: Commit Changes

Create a commit with all the infrastructure changes:

```bash
git add proto/{service}/api/v3/{module}.proto
git add pkg/proto/{service}/api/v3/{module}.pb.go
git add pkg/proto/{service}/api/v3/{module}.go
git add docs/api/swagger/{service}/api/v3/{module}.swagger.json
git add internal/{service}/router/api-v3/{module}/{module}.go
git add internal/{service}/router/api-v3/{module}/{endpoint_name}.go

git commit -m "feat({module}): add {endpoint_name} API scaffold

- Add proto definitions for {EndpointName}Req/Resp
- Implement conversion layer with validation
- Register route in {module} router
- Create handler scaffold with TODO markers"
```

## Common Patterns and Reusable Components

### Reusable Proto Messages

Check if these existing messages can be reused:
- `AuthResource` - For IAM resources (system_id, type, id)
- `PageReq` / `PageResp` - For paginated requests/responses
- Common error structures

### Conversion Utilities

Use these helpers from `pkg/conv`:
- `SliceToSlice` - Transform arrays
- `PtrTo` - Create pointers
- `ValueOr` - Handle nil with defaults

### Error Codes

Standard error codes from `pkg/rest/errf`:
- `InvalidParameter` (400) - Request validation failed
- `PermissionDenied` (403) - Authorization failed
- `BackendOperateFailed` (500) - Internal error

### Logging Pattern

Always log errors with context:
```go
logger.G.Biz(rCtx).WithErr(err).Error("failed to {operation}, {specific_reason}")
```

## Edge Cases and Considerations

### Batch vs Single Operations

**Batch** (use when user wants to process multiple items):
- Request: `repeated Item items = 1`
- Response: `repeated Result results = 1`
- Iterate over items in business logic

**Single** (use for one-item operations):
- Request: Direct fields (no repeated)
- Response: Single object in data

### Nested Structures

For complex responses with nested data:
```protobuf
message ComplexResp {
  message Data {
    message NestedInfo {
      string field1 = 1;
      int32 field2 = 2;
    }
    NestedInfo info = 1;
    repeated string tags = 2;
  }
  Data data = 5;
}
```

### Optional Fields

Use proto3 optional for nullable fields:
```protobuf
message MyReq {
  string required_field = 1;
  optional string optional_field = 2;
}
```

### Validation Complexity

For complex validation (cross-field, regex, ranges), implement in Validate():
```go
func (x *MyReq) Validate() error {
    if x.GetStartTime() > x.GetEndTime() {
        return fmt.Errorf("start_time must be before end_time")
    }
    if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(x.GetName()) {
        return fmt.Errorf("name must be lowercase alphanumeric with hyphens")
    }
    return nil
}
```

## What the User Does Next

After this skill completes, the user should:

1. **Implement business logic** in the handler's TODO section
2. **Add error handling** specific to their use case
3. **Call the response conversion method** with actual results
4. **Test the endpoint** with real requests
5. **Add tests** if needed (usually in pkg/ layer for conversion logic)

The skill provides a working skeleton that compiles and has proper structure — the user just fills in the domain-specific logic.

## Example: Full Workflow

User says: "我要加个查询用户权限的接口，叫 authorized，批量查询，输入是 action 和 resource_type 数组，输出是每个的授权范围"

**Step 1: Gather info**
- Service: backend
- Module: auth
- Endpoint: Authorized
- Request: Array of {action, resource_type}
- Response: Array of {action, resource_type, is_any, resources}

**Step 2-7: Execute workflow**
- Modify proto/backend/api/v3/auth.proto
- Run proto generation
- Implement conversion in pkg/proto/backend/api/v3/auth.go
- Register route in internal/backend/router/api-v3/auth/auth.go
- Create internal/backend/router/api-v3/auth/authorized.go
- Verify with LSP diagnostics
- Commit changes

**Result**: User has a working API scaffold and can now implement the business logic (calling authorizer.ListAuthorizedInstances).

## Tips for Success

1. **Ask before assuming** - If the user's request is ambiguous, clarify before generating code
2. **Follow existing patterns** - Look at similar endpoints in the same module for consistency
3. **Reuse when possible** - Don't duplicate proto messages or conversion logic
4. **Validate thoroughly** - LSP diagnostics must pass before committing
5. **Clear TODOs** - Make it obvious what the user needs to fill in
6. **Commit atomically** - One commit for the complete scaffold

This skill saves significant time by automating the repetitive protocol layer work while leaving the interesting business logic to the user.
