|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/comments/docs/tech discussions
|Compression Rule:Follow pipe-index format|keep concise|no prose/code blocks
|Scope:pkg/dao/mongo/upload
|Overview:Mongo DAO for uploaded artifact records|one collection per types.UploadCategory|owns persistence, soft deletion, and the distinct saved-name projection that artifact cleanup uses to find unbound files
|In scope:category-to-collection mapping|lazy per-category dao cache|Create/Get/DeleteMany|DistinctSavedName|pkg/types.Upload conversion
|Out of scope:artifact storage in BKRepo|upload router/HTTP handling|release and export persistence|artifact cleanup scheduling
|Structure:pkg/dao/mongo/upload:{handler.go,upload.go,table.go,constants.go}
|Where to look:public API+validation+conversion:handler.go:{IHandler,IDistinctor,Handler,New,Create,Get,DeleteMany,DistinctSavedName,categoryDao}
|Where to look:dao+base.IOrm+indexes:upload.go:{newDao,dao,GetClient,GetTableName,GetIndexes}
|Where to look:collection+document contract+unique key:table.go:{TableName,Data,UniqueFields,UniqueKey}
|Where to look:BSON field keys:constants.go:{FieldKeyUploadID,FieldKeySavedName,FieldKeyCategory}
|Where to look:shared Mongo rules:pkg/dao/mongo/base/AGENTS.md|peer refs:pkg/dao/mongo/{release,host,packageevent}/AGENTS.md
|Where to look:callers:internal/file/storage/upload:{storage.go,iface.go,dao_*.go}|internal/file/periodictask/clean_packages.go binding check
|Conventions:upload collections are category-wide and shared across tenants|never filter them by tenant or infer record ownership from saved names
|Conventions:distinct projections live in IDistinctor and are named `Distinct<Field>`|mirrored from pkg/dao/mongo/{release,host,packageevent}
|Conventions:DistinctSavedName skips an empty candidate list instead of widening the filter into a full scan
|Conventions:one handler serves every category through categoryDao|do not add per-category handlers or bypass base.IOrm
|Conventions:exported Go symbols require English godoc comments|wrap database failures with `%w`
|Tests:no package suite yet|behavior is exercised through the file-service cleanup caller
|Commands:compile check=GOTOOLCHAIN=local go vet ./pkg/dao/mongo/upload
|Anti-patterns:no tenant filters on upload collections|no proto/HTTP structs|no direct mongo.Connect|no package-global clients|no hard delete replacing soft-delete semantics
