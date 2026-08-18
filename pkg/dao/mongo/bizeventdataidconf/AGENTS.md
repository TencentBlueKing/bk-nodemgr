|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/dao/mongo/bizeventdataidconf
|Overview:Tenant-scoped Mongo DAO for optional Node event data IDs keyed by bk_biz_id|routes operations to biz_event_data_id_conf_<tenantID>|owns schema,index,read,taskProcEventDataID upsert,and pkg/types conversion
|In scope:tenant collection naming|base.IOrm wiring|unique index|OptFn filter|tenant dao cache|taskProcEventDataID atomic upsert|types conversion
|Out of scope:runtime defaults|Monitor API calls|workflow resolution|HTTP/proto conversion|migration workers/runbooks|old shared collection cleanup
|Structure:pkg/dao/mongo/bizeventdataidconf:{handler.go,bizeventdataidconf.go,table.go,constant.go,options.go,AGENTS.md}
|Where to look:public API+tenant routing+types conversion:handler.go:{IHandler,Handler,New,tenantDao,Get,UpdateTaskProcEventDataID,convertBizEventDataIDConfToTypes}
|Where to look:base.IOrm+atomic upsert:bizeventdataidconf.go:{newDao,dao,upsertTaskProcEventDataID}
|Where to look:collection+document+unique key:table.go:{TableName,BizEventDataIDConf,UniqueFields,UniqueKey,TableBizEventDataIDConf}
|Where to look:BSON paths+filters:{constant.go,options.go}:{FieldKeyBizID,fieldTaskProcEventDataID,WithBizID}
|Conventions:collection=biz_event_data_id_conf_<tenantID> via TableName(tenantID)|resolve tenant only from contextx.IContext after CheckTenantID()|per-tenant uniqueness=data.biz_id|BizID and persisted event data IDs must be positive
|Conventions:tenantDao owns daoMap and EnsureIndexes() bootstrap|do not cache tenant dao entries outside Handler
|Conventions:document envelope=base.TableBroker with basic fields and data payload|inner type implements base.IData
|Conventions:nil agentBaseAlarmEventDataID/taskProcEventDataID means unconfigured|never coerce absence to zero or a default
|Conventions:reads=base.AliveFilter()→WithBizID|storage maps mongo.ErrNoDocuments to found=false without default synthesis
|Conventions:UpdateTaskProcEventDataID upserts only data.biz_id+data.task_proc_event_data_id and restores basic.is_deleted=false
|Conventions:handler boundary accepts/returns pkg/types.BizEventDataIDConf|Mongo document types and BSON details stay package-local
|Conventions:exported Go symbols require English godoc comments|errors retain operation context and wrapped causes
|Dependencies:pkg/dao/mongo/base:{TableBroker,IOrm,AliveFilter,WithValues}|pkg/types:BizEventDataIDConf|pkg/contextx:IContext
|Parent rules:pkg/dao/mongo/base/AGENTS.md|pkg/AGENTS.md
|Anti-patterns:no runtime fallback,Monitor invocation,or workflow policy in this DAO
|Anti-patterns:no shared-collection tenantFilter, runtime dual-read/dual-write, or cross-tenant collection access
|Anti-patterns:no tenantID in unique key|bk_biz_id uniqueness is scoped by the tenant collection
|Anti-patterns:no persistence of global defaults|only explicit/local or externally returned values belong in the collection
|Anti-patterns:no raw context.Context,proto types,HTTP payloads,or direct Mongo document exposure
|Commands:go test ./pkg/dao/mongo/bizeventdataidconf|compile=go test ./pkg/dao/mongo/bizeventdataidconf -run '^$'
