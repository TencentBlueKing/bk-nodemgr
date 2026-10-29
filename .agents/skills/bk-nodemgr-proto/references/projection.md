# Proto → Swagger → APIGW projection

## Source and version boundaries

Use the current checkout first. `proto/Makefile` owns generation, `.mise.toml` selects tool versions, and `proto/convert_swagger_type.sh` owns Swagger type post-processing. The reference below was checked against **protoc-gen-openapiv2 v2.27.2**, the configured version; recheck the relevant rule when the toolchain changes.

APIGW rules below refer to upstream commit **`9cbe8cc5f1c1030080bd3088999ad9917aab79c4`**, not an assertion about the deployed gateway version. When a task depends on an import detail, compare the target version's implementation or have an authorized operator verify its import/export result. Do not infer the dashboard's version from the apigw-sync image: they are different components.

The project maintenance flow is **generate Swagger locally → manually import it in APIGW → export resource definitions → update `apigw/resources.yaml` through the existing maintenance process**. No local automatic Swagger-to-resources converter is implied. Gateway edits, export timing, and import behavior can all explain differences between the files; a difference alone does not prove its cause.

## Proto to Swagger

| Source                                                         | Generated result with current defaults                                                      | Writing consequence                                                                                      |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| RPC leading comment, first paragraph                           | Operation `summary`                                                                         | Write a complete, concise Chinese purpose statement                                                      |
| Remaining RPC comment paragraphs                               | Operation `description`                                                                     | Include only necessary details; ensure the description makes sense without the summary                   |
| Message/field comment, first paragraph not ending in ASCII `.` | Schema `title`; later paragraphs become `description`                                       | Chinese `。` does not trigger the ASCII-period branch                                                    |
| Message/field comment, first paragraph ending in ASCII `.`     | Whole comment becomes schema `description` rather than populating `title` from that comment | This is a schema rule, not a general punctuation rule for operations                                     |
| `google.api.http`                                              | Operation method/path and parameter/body binding                                            | Preserve in documentation-only work; inspect the actual operation                                        |
| Service name                                                   | Operation `tags: [ServiceName]` and top-level tag name                                      | Keep existing names/categories; do not add tag overrides just to improve prose                           |
| Service/RPC names                                              | Default `operationId: ServiceName_RPCName`                                                  | Preserve identity; verify generated IDs for additional HTTP bindings rather than assuming one ID per RPC |
| Proto field name                                               | Swagger uses snake_case names because `json_names_for_fields=false`                         | Do not invent camelCase names in descriptions/examples                                                   |
| `int64` generation                                             | Plugin output is post-processed from string/int64 to integer/int64                          | Inspect the final result of `make all`, not the plugin's unprocessed output                              |

The comment rules above assume ordinary comments without explicit OpenAPI overrides. Line wrapping alone is not a paragraph break. An RPC first paragraph ending in either `.` or `。` still enters `summary`; the schema `title` fallback is a different branch. No punctuation-only normalization is required.

### Why a later paragraph matters

Given this comment:

```protobuf
// 查询节点 Agent 状态。
//
// 也支持批量查询。
```

Swagger contains `summary: 查询节点 Agent 状态。` and `description: 也支持批量查询。`. The checked APIGW importer uses the latter as its resource description. For a confirmed batch-query contract, a complete single paragraph such as `按主机 ID 列表查询节点 Agent 状态。` is sufficient. If more explanation is needed, keep the first paragraph concise and make the later description independently meaningful; do not ban multiple paragraphs or mechanically duplicate both fields.

For schema comments, `// 蓝鲸主机 ID。` yields a `title`, while `// 蓝鲸主机 ID.` yields a `description` under the checked generator. Both must be meaningful to the reader. If a consumer specifically needs one key, verify that requirement and the generated result instead of guessing from punctuation or patching JSON.

## Swagger to APIGW and exported resources

| Swagger/import input                                                | APIGW import behavior at the cited commit                                                  | Exported OpenAPI resource                                                                     |
| ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- |
| `operationId`                                                       | Resource `name`, replacing `-` with `_`                                                    | `operationId` from the stored resource name                                                   |
| Nonempty `description`, otherwise `summary`, otherwise empty string | One resource description; the fields are not concatenated                                  | `description`; a separate original `summary` is not reconstructed                             |
| Operation `tags`                                                    | Resource label names                                                                       | Operation `tags` from stored labels                                                           |
| Body parameter schema                                               | Stored as OpenAPI 3 `requestBody.content[content-type].schema`                             | Stored request-body schema, not the original Swagger 2 parameter form                         |
| Response schema                                                     | Stored under OpenAPI 3 response `content`                                                  | Stored response schema                                                                        |
| `definitions` / `$ref`                                              | References resolved before resource parsing                                                | The exporter rebuilds paths; it does not reconstruct the original reusable definitions layout |
| `x-bk-apigateway-resource.descriptionEn`                            | Separate English resource description                                                      | `descriptionEn` in the gateway extension; not an automatic translation of Chinese comments    |
| Gateway extension settings                                          | Authentication, visibility, backend and related settings; missing values may have defaults | Supported `x-bk-apigateway-resource` settings, not arbitrary extension round-tripping         |

An existing stored/exported label may differ from the default generated Service tag; explain it, do not automatically overwrite it. Preserve the current `NodeAgent`, `Topo`, `Ops`, and other actual tag values. Top-level Swagger tag metadata is not an instruction to rename gateway labels.

### Limits that affect a Proto change

- **Comments are not runtime validation.** Do not infer requiredness, valid enums, or defaults from primitive types or encode them as unsupported promises. A required body parameter is distinct from required fields inside that object.
- **Not all constraints survive Swagger 2 conversion.** The cited `convert_openapi2_parameters_to_openapi` copies selected non-body parameter fields; top-level constraints such as `minimum`, `maximum`, and `pattern` are not all carried into the new schema. Body/response schemas take a different path. Check the parameter location and actual consumer before claiming a constraint is preserved or enforced.
- **References are not a lossless round trip.** The cited importer resolves internal references and rejects external references; recursive types may hit the resolver's recursion limit. A successful local generation does not prove that a recursive schema imports successfully. Do not redesign an existing model solely because of this upstream reference; first establish the target version's behavior and obtain approval for contract changes.
- **Gateway description is not the full API reference.** The cited resource validator limits descriptions to 2048 characters. Keep resource prose focused; use `api-doc` for full bilingual calling instructions. Do not truncate essential caller-facing limitations or infer an introduction version/permission from Proto alone.

## Local evidence to inspect on demand

- `proto/backend/api/v3/node_agent.proto` and `docs/api/swagger/backend/api/v3/node_agent.swagger.json`: `NodeAgentInstall`, its default ID/tag and single-paragraph RPC summary.
- `proto/backend/api/v3/iam.proto` and `docs/api/swagger/backend/api/v3/iam.swagger.json`: schema comments with and without a trailing ASCII period.
- `apigw/resources.yaml`: exported descriptions, tags, schemas, and gateway extensions; not proof of the current deployed state.
- `apigw/apidocs/{zh,en}`: full endpoint references named by `operationId`.
- `apigw/definition.yaml`: resource-name consumers, including MCP `resource_names` where configured.
- `install/images/bk-nodemgr-apigw-sync/support-files/bin/sync-apigateway.sh`: consumes exported `resources.yaml`; it is not a Swagger converter. Do not execute it for documentation verification.

## Version-pinned implementation references

Read only the relevant function when resolving a mapping question; these are not extra mandatory steps for every comment edit.

- [gRPC-Gateway v2.27.2 template.go](https://github.com/grpc-ecosystem/grpc-gateway/blob/v2.27.2/protoc-gen-openapiv2/internal/genopenapi/template.go): `renderServices`, `renderServiceTags`, `updateOpenAPIDataFromComments` (default tags/IDs, comment paragraphs and schema punctuation, option overrides).
- [APIGW parser.py](https://github.com/TencentBlueKing/blueking-apigateway/blob/9cbe8cc5f1c1030080bd3088999ad9917aab79c4/src/dashboard/apigateway/apigateway/biz/openapi/parser.py): `BaseParser.get_resources`, `_adapt_description`, `_get_request_body`, `_get_responses`.
- [APIGW openapi_export.py](https://github.com/TencentBlueKing/blueking-apigateway/blob/9cbe8cc5f1c1030080bd3088999ad9917aab79c4/src/dashboard/apigateway/apigateway/service/resource_version/openapi_export.py): `BaseExporter._generate_paths`, `_generate_bk_apigateway_resource`.
- [APIGW schema.py](https://github.com/TencentBlueKing/blueking-apigateway/blob/9cbe8cc5f1c1030080bd3088999ad9917aab79c4/src/dashboard/apigateway/apigateway/biz/openapi/schema.py): `convert_openapi2_parameters_to_openapi`.
- [APIGW openapi.py](https://github.com/TencentBlueKing/blueking-apigateway/blob/9cbe8cc5f1c1030080bd3088999ad9917aab79c4/src/dashboard/apigateway/apigateway/biz/openapi/openapi.py): `OpenAPIImportManager.parse`, `_validate_refs`.
- [APIGW description-priority tests](https://github.com/TencentBlueKing/blueking-apigateway/blob/9cbe8cc5f1c1030080bd3088999ad9917aab79c4/src/dashboard/apigateway/apigateway/tests/biz/resource/importer/test_parser.py): `test_adapt_description`.
- [APIGW import tests](https://github.com/TencentBlueKing/blueking-apigateway/blob/9cbe8cc5f1c1030080bd3088999ad9917aab79c4/src/dashboard/apigateway/apigateway/tests/biz/resource/importer/test_openapi.py): `test_validate_returns_schema_err_for_recursive_ref`.
- [APIGW resource-validation tests](https://github.com/TencentBlueKing/blueking-apigateway/blob/9cbe8cc5f1c1030080bd3088999ad9917aab79c4/src/dashboard/apigateway/apigateway/tests/biz/resource/importer/test_validate.py): `test_validate_resource_field_length`.
