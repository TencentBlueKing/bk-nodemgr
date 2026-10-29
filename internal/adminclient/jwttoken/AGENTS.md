|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/comments/docs/tech discussions
|Compression Rule:Follow pipe-index format; keep concise; no prose/code blocks
|Scope:internal/adminclient/jwttoken
|Overview:adminclient jwt_token command; prints a jwt token for one service http server, signed with the symmetric key decrypted from the kms global envelope; install/upgrade scripts call it instead of reading plaintext jwt keys from values/config
|Structure:./:{jwttoken.go,jwttoken_test.go}
|Where to look:command wiring:jwttoken.go:NewJWTTokenCMD|registered in cmd/adminclient/main.go
|Where to look:service/server→key mapping:jwttoken.go:jwtKeyPickers|application={info,admin,basic};backend={+callback,proxy};file={+download,export}
|Where to look:signing:pkg/rest/header:NewNodeMgrAuthorizationManager.Generate(loginName,bkUserName)|HS256;fixed 24h expiration
|Where to look:config/decrypt:pkg/config:AdminClientJWT.LoadFromFile→KMS.ParseAdminClientJWT|global envelope only;kms disabled is an error
|Where to look:config file resolution:internal/adminclient/configfile:Resolve(path,dir,"kms")|-f explicit;else first *_conf.yaml containing `kms:` under --conf-dir
|Where to look:consumer:support-files/initpackage/init_package.py:generate_jwt_token_kms|FileClient caches the token per client instance
|Where to look:helm wiring:install/helm/bk-nodemgr/templates/init-packages-job.yaml|mounts global envelope only + BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL env via bk-nodemgr.bkappKMSEnv
|Where to look:user docs:docs/operation/adminclient/jwt_token.md
|Flow:args(service server)→pickJWTKey→configfile.Resolve→AdminClientJWT.LoadFromFile(kms global envelope)→pick key→restheader.Generate→stdout single-line token
|Conventions:stdout carries only the token;errors go to stderr;SilenceUsage=true so runtime failures never dump usage into job-captured output|argument errors rejected in Args validation before RunE
|Conventions:empty decrypted key is an error;never output an empty token|new service/server=extend jwtKeyPickers + AdminClientJWT field + envelope item + docs table
|Conventions:flags via configfile.AddPersistentFlags(-f/--file,--conf-dir) + command-local --login-name/--bk-username default admin|claims identify the caller
|Anti-patterns:printing the decrypted key or envelope content;reading module envelope/private keys (global-only);adding plaintext jwt-key fallback flags (plaintext deployments do not use this command)
|Peer AGENTS:pkg/config/AGENTS.md|pkg/config/kms/AGENTS.md
