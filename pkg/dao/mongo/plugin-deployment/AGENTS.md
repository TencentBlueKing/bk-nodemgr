|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/plugin-deployment
|Overview:Mongo DAO for plugin deployment artifacts|persists install/transfer/runtime context keyed by token|converts db models↔pkg/types plugin deployment structs
|Structure:pkg/dao/mongo/plugin-deployment:{plugindeploy.go,table.go,constant.go,options.go,handler.go,error.go,handler_test.go,README.md}
|Where to look:collection+indexes:plugindeploy.go:{newDao,GetTableName,ExpireTimeSec,GetIndexes}:TTL index key=data.expire_at
|Where to look:document schema:table.go:{Data,Info,PluginConf,configDetail,Table}:collection=plugin_deployment|unique=data.token
|Where to look:field paths:constant.go:{FieldKeyToken,FieldKeyInfo,FieldKeyPluginConf,FieldKeyPluginConfConfigFilesDetail,FieldKeyExpireAt,FieldKeyHostID,FieldKeyPluginName,FieldKeyPluginVersion}
|Where to look:query builders:options.go:{WithToken,WithHostID,WithPluginName,WithPluginVersion}:compose over base.AliveFilter()
|Where to look:handler API:handler.go:IHandler{Create,List,GetInfo,UpdateInfo,GetPluginConf,UpdatePluginConf,GetPluginConfConfigFilesDetail,UpsertPluginConfConfigFilesDetail}
|Where to look:package errors:error.go:ErrInvalidToken
|Where to look:upstream orchestration:internal/backend/storage/plugin/storage.go:daoPluginDeployment
|Where to look:shared DTOs:pkg/types:{plugin_deployment.go,condition.go}
|Type Flow:pkg/types.PluginDeployment↔handler converters↔table.go BSON structs↔Mongo plugin_deployment
|Conventions:guard nCtx/token/nil params in every exported DAO method and return base/package sentinel errors consistently
|Conventions:keep DAO storage-oriented only; workflow policy and permission logic belong to internal/backend/{manager,storage}
|Conventions:follow base.TableBroker+base.IOrm pattern; build filters from base.AliveFilter() + OptFn chain
|Conventions:preserve merge-by-name behavior for config details in UpsertPluginConfConfigFilesDetail
|Conventions:keep json/bson tags, field-key constants, and converter mappings synchronized when schema evolves
|Dependencies:pkg/dao/mongo/base (ORM/filter/page/errors)|pkg/types (domain DTOs)|pkg/contextx (DAO context contract)
|Anti-patterns:no business policy/retry/permission decisions in this package
|Anti-patterns:no exposure of Data/Info/PluginConf outside package boundary; callers consume pkg/types via IHandler
|Anti-patterns:no raw context.Context in handler signatures; keep contextx.IContext contract
|Anti-patterns:no handwritten dotted Mongo field paths in handlers when constants already exist
|Anti-patterns:no schema field additions without synchronized updates to {table.go,constant.go,handler.go converters,handler_test.go}
|Tests:handler_test.go uses Mongo integration env (.env); keep fixtures aligned with converter semantics
