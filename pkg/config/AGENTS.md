|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/config
|Overview:single source of truth for typed service startup config:{infrastructure,servers,front-end injection,auth}|owns:{loading,defaults,required-field/value-range validation}|not:{business logic,post-startup state mutation}
|Structure:pkg/config:{config.go,application.go,backend.go,file.go,relay.go,adminclient.go,kms.go,kms/}
|Where to look:shared config:config.go:{Etcd,Redis,MongoDB,Log,HTTPServer,AuthIdentity,JWTServerConfig,JWTClientConfig,Front,TraceService,TLSConfig,APIGatewayClient}
|Where to look:shared contracts:config.go:Front={vault switch,ports,URLs};HTTPServer={bind/advertise IP,port};JWT{Server,Client}Config={symmetric,asymmetric};APIGatewayClient={endpoints,appCode,appSecret};TLSConfig={CA/cert/key paths,insecureSkipVerify}
|Where to look:application:application.go:ApplicationService;Load/LoadFromFile/Validate;Front.{WindowsWMIPortDefault,WindowsSSHPortDefault,UnixSSHPortDefault}
|Where to look:services:{backend,file,relay}.go:{Backend,File,Relay}Service|backend={GSE,CMDB,IAM,plugin,workflow workers}|file={download port,mount host dir}|relay=relay config
|Where to look:KMS integration:kms.go:KMS + Parse*ServiceConfig;decrypt via pkg/config/kms then map typed secrets into config
|Where to look:adminclient:adminclient.go:AdminClientDatabase={database check/proxy};AdminClientBKApp={bkapp code/secret}|shared validation tests:config_test.go
|Lifecycle:New*Service→LoadFromFile→Validate|ApplicationService also exposes Load(filePath)|load YAML via yaml tags; deployment env is rendered into YAML, not read by a LoadFromEnv method
|Conventions:one service=one *Service config;embed shared sub-structs by value, not pointer|shared sub-structs belong in config.go;service-only sub-structs stay in their service file
|Conventions:centralize defaults in NewApplicationService/New*Service;do not scatter defaults across loaders|expose Validate for required fields/ranges;call after loading before startup;fail fast with field-specific fmt.Errorf errors
|Conventions:YAML=camelCase|deployment env=BK_NODEMGR_<SERVICE>_<FIELD> in SCREAMING_SNAKE_CASE|env access belongs at explicit config-loading boundaries, not consumers|exported symbols require English comments
|Front port flow:BK_NODEMGR_APPLICATION_{WINDOWS_WMI,WINDOWS_SSH,UNIX_SSH}_PORT_DEFAULT→install/docker-compose/bk-nodemgr/templates/config/bk-nodemgr-application.yml.tpl→rendered YAML→ApplicationService.LoadFromFile→Front.{WindowsWMIPortDefault,WindowsSSHPortDefault,UnixSSHPortDefault}
|Front port flow:Front→internal/application/service→internal/application/frontsetting.Option→internal/application/router/web→index.html→window.PROJECT_CONFIG.{WINDOWS_WMI,WINDOWS_SSH,UNIX_SSH}_PORT_DEFAULT
|Exception:KMS remains in kms.go with Parse*ServiceConfig, not config.go;ApplicationService/BackendService/FileService embed it;keep typed envelope mapping and config together
|Exception:KMS.load/loadGlobal in kms.go centralize BK_NODEMGR_KMS_PRIVATE_KEY_* bootstrap env reads|loadGlobal reads only BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL for global-only consumers;do not duplicate private-key loading across services
|Exception:AdminClientDatabase/AdminClientBKApp are minimal client configs, not *Service;reuse common sections across service files|LoadFromFile skips whole-config Validate;commands validate only consumed sections to tolerate unrelated placeholders such as __ADVERTISE_IPV4__
|Exception:adminclient loaders perform KMS decryption like service loaders|AdminClientBKApp requires kms.enabled=true;plaintext deployments read app credentials from values and do not use bkapp
|Anti-patterns:business logic/derived state in config;post-startup config mutation;service-specific sub-structs in config.go;skipped startup validation;hardcoded business-code ports/URLs instead of config
|Child AGENTS:pkg/config/kms/AGENTS.md|pkg/config/kms/envelope/AGENTS.md
