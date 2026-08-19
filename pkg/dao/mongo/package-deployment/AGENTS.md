|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/package-deployment
|Overview:Storage-oriented Mongo DAO for package deployment records keyed by token|owns collection/BSON contract, base.IOrm wiring, indexes, CRUD, partial info access, and pkg/types.PackageDeployment conversion
|In scope:collection naming|base.IOrm wiring|Data BSON contract|Info payload fields|OptFn query builders|indexes|types conversion
|Out of scope:package import/workflow orchestration|status derivation|scheduler policy|HTTP/proto conversion|cross-collection joins|upload/download behavior
|Structure:pkg/dao/mongo/package-deployment:{packagedeploy.go,handler.go,table.go,constant.go,options.go}
|Where to look:public API+types conversion+partial info access:handler.go:{IHandler,Handler,New,CreatePackageDeployment,ListPackageDeployment,GetPackageDeploymentInfo,UpdatePackageDeploymentInfo,convertPackageDeploymentFromTypes,convertPackageDeploymentToTypes,convertInfoFromTypes,convertInfoToTypes,convertPlatformsFromTypes,convertPlatformsToTypes}
|Where to look:dao+base.IOrm+indexes:packagedeploy.go:{newDao,dao,GetClient,GetTableName,GetIndexes}
|Where to look:collection name+document contract+unique key:table.go:{TableName,Data,UniqueFields,UniqueKey,Table}
|Where to look:BSON field keys+filters:{constant.go,options.go}:{FieldKeyToken,FieldKeyInfo,FieldKeyInfoUploadID,WithToken,WithUploadID}
|Where to look:domain type:pkg/types/package_deployment.go:{PackageDeployment,PackageDeploymentInfo}
|Where to look:upstream storage orchestration:internal/backend/storage/pkg:{storage.go,dao_pkg_deployment.go,iface.go}
|Where to look:peer refs:pkg/dao/mongo/{node-deployment,package-workflow}/AGENTS.md|shared Mongo rules:pkg/dao/mongo/base/AGENTS.md
|Conventions:collection=fixed `package_deployment` via TableName|not tenant-suffixed|unique key=Token (UniqueFields:`data.token`)
|Conventions:document envelope=base.TableBroker with Data payload under basic+data|Data implements base.IData
|Conventions:public boundary accepts/returns pkg/types.PackageDeployment / types.PackageDeploymentInfo|Mongo Data structs, BSON paths, and base.IOrm details stay package-private
|Conventions:handler validates nil nCtx/nil deployment/nil info and empty token via base sentinel errors|no package-local error sentinels
|Conventions:read filters=base.AliveFilter()→apply OptFn chain|List computes total with Count then applies base.ParsePage(page)
|Conventions:Info field alignment=table.go Info fields (ImportPluginPkgOptions{FileSourceType,FileSource,FileName,MD5},Upload{UploadID,Name,Version,Platforms},Release) map 1:1 to types.PackageDeploymentInfo|platform conversion via closest package-private helpers
|Conventions:Mongo field paths must align with table.go BSON tags|reuse FieldKey constants for new filters/indexes instead of duplicating strings
|Conventions:new indexes use mongo.IndexModel in dao.GetIndexes() and must match actual query filters|current explicit indexes=FieldKeyInfoUploadID
|Conventions:exported Go symbols require English godoc comments|errors preserve base sentinel semantics and wrapping
|Anti-patterns:no package import/workflow orchestration, status aggregation, scheduler logic, proto/HTTP structs, or upload/download validation in this DAO
|Anti-patterns:no TTL/expire_at reintroduction unless collection contract changes explicitly|no cross-tenant collection access or cross-collection joins
|Anti-patterns:no bypass of Handler/base.IOrm to access arbitrary collections|TableBroker envelope must remain centralized
|Anti-patterns:no change to Token uniqueness, BSON tags, TableName format, or FieldKey_ paths without checking indexes and all storage callers
