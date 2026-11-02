|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/comments/docs/tech discussions
|Compression Rule:Follow pipe-index format; keep concise; no prose/code blocks
|Scope:internal/adminclient/jwttoken
|Overview:adminclient jwt_token command; prints a jwt token for one service http server, signed with the service jwt symmetric key; key source (kms global envelope vs plaintext config server section) is resolved inside pkg/config so consumers never perceive kms
|Structure:./:{jwttoken.go,jwttoken_test.go}
|Where to look:command wiring:jwttoken.go:NewJWTTokenCMD|registered in cmd/adminclient/main.go
|Where to look:service/server→config mapping:jwttoken.go:jwtServiceConfigs|service arg selects valid servers + load<Service>JWTKey loader(server-name switch lives here, not in pkg/config);application={info,admin,basic};backend={+callback,proxy};file={+download,export}
|Where to look:signing:pkg/rest/header:NewNodeMgrAuthorizationManager.Generate(loginName,bkUserName)|HS256;fixed 24h expiration
|Where to look:config/decrypt:pkg/config:AdminClient{Application,Backend,File}.LoadFromFile(path)|mirrors service config server sections;kms enabled→KMS.ParseAdminClient<Service>JWT overrides JWTServerConfig.SymmetricKey from global envelope;kms disabled→plaintext yaml as-is
|Where to look:config file resolution:internal/adminclient/configfile:ResolveForService(path,dir,service)|-f explicit;else <conf-dir>/<service>_conf.yaml,legacy fallback bk-nodemgr-<service>.yml;resolved file is trusted as-is(no content probing)
|Where to look:consumer:support-files/initpackage/init_package.py:generate_jwt_token|FileClient caches the token per client instance;script never passes --jwt-key
|Where to look:helm wiring:install/helm/bk-nodemgr/templates/init-packages-job.yaml|mounts global envelope only + BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL env via bk-nodemgr.bkappKMSEnv (kms.enabled gated)
|Where to look:user docs:docs/operation/adminclient/jwt_token.md
|Flow:args(service server)→validateServiceServer→{--jwt-key set: sign directly} else {configfile.ResolveForService→jwtServiceConfigs[service].loadKey(path,server)}→restheader.Generate→stdout single-line token
|Conventions:stdout carries only the token;errors go to stderr;SilenceUsage=true so runtime failures never dump usage into job-captured output|argument errors rejected in Args validation before RunE
|Conventions:empty loaded key is an error;never output an empty token|new service/server=extend jwtServiceConfigs + AdminClient<Service> struct field/JWTServer + envelope item + docs table
|Conventions:flags via configfile.AddPersistentFlags(-f/--file,--conf-dir) + command-local --login-name/--bk-username default admin|claims identify the caller
|Conventions:--jwt-key is a manual-troubleshooting escape hatch that skips config parsing entirely|script consumers (init_package.py, helm jobs) must not use it;they always take the config-resolved default path
|Anti-patterns:printing the decrypted key or envelope content;reading module envelope/private keys (global-only);letting consumers branch on kms.enabled themselves (that branching belongs inside pkg/config loaders)
|Peer AGENTS:pkg/config/AGENTS.md|pkg/config/kms/AGENTS.md
