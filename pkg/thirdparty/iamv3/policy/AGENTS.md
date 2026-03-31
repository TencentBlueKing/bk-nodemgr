|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow pipe-index format; keep concise; no prose/code blocks
|Scope:pkg/thirdparty/iamv3/policy

|Overview:IAM v2 policy expression parser|transforms iam-go-sdk ExprCell tree into authorized resource ID list
|Purpose:extract authorized instance scope from IAM policy without evaluating against specific resources
|Contract:Parse(expr,systemID,resourceType)→(isAny bool,resources []types.IAMResource,err)

|Structure:policy/:{policy.go,policy_test.go}
|Where to look:expression parsing entry:policy.go:Parse
|Where to look:operator handlers:policy.go:{parseEqPolicy,parseInPolicy,parseCompositePolicy}
|Where to look:set operations:policy.go:{unionAuthorizedInstances,intersectAuthorizedInstances}
|Where to look:resource building:policy.go:{buildAuthorizedResources,deduplicateResources}

|Supported Operators:any→full access|eq→single ID|in→ID list|AND→intersection|OR→union
|Unsupported:_bk_iam_path_ attribute|non-id attribute conditions|other operators

|Conventions:return (true,[],nil) for any-access|return (false,ids,nil) for scoped access|return (false,nil,err) for unsupported
|Conventions:always deduplicate and sort resources before returning
|Conventions:keep parsing logic recursive for composite expressions (AND/OR)
|Conventions:use pkg/types.IAMResource as output type, not raw proto or SDK types

|Anti-patterns:do not evaluate policy against specific resources—use iamv3.IsAllowed for that
|Anti-patterns:do not handle _bk_iam_path_—reject with explicit error
|Anti-patterns:do not expose SDK ExprCell in return types—convert to business types

|Dependencies:github.com/TencentBlueKing/iam-go-sdk/expression|pkg/types.IAMResource
|Parent:pkg/thirdparty/iamv3/AGENTS.md

|Testing:go test ./pkg/thirdparty/iamv3/policy -v
