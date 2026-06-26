|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/wmix

|Overview:WMI client abstraction for Linux→Windows remote command execution and file upload|wraps embedded staticx wmiexec subprocess|shared infra package, not service orchestration
|Structure:wmix/:README.md|wmix.go|wmi_bin.go|wmix_test.go|wmi_bin_test.go|wmiexec/{wmiexec.go,wmiexec_linux_amd64.go,wmiexec_linux_arm64.go,wmiexec.py,wmiexec-builder.sh,Dockerfile,binaries/}
|Where to look:public API + config validation:wmix.go:{IClient,Config,Client,NewClient,RunCommand,RunSilentCommand,UploadFile}
|Where to look:binary lifecycle + subprocess:wmi_bin.go:{wmiBinaryPath,extractWMIBinary,wmiRunCmd}
|Where to look:embedded binary + build tags:wmiexec/{wmiexec.go,wmiexec_linux_amd64.go,wmiexec_linux_arm64.go}|rebuild pipeline:wmiexec/{Dockerfile,wmiexec-builder.sh,wmiexec.py}|pre-built binaries:wmiexec/binaries/
|Where to look:binary path regression tests:wmi_bin_test.go|remote integration-style tests:wmix_test.go (requires .env)

|Conventions:Target identity flow:Config{Domain,IP,User,Password,AuthMethod,Timeout}→Client.target/envs→wmiexec args/env|preserve existing target/env construction unless wmiexec contract changes
|Conventions:Timeout flow:Config.Timeout default→context.WithTimeout in RunCommand|caller ctx is parent; subprocess cancellation handled by wmiRunCmd
|Conventions:Binary Lifecycle:singleton cached path protected by mutex|extract embedded binary to temp file|chmod 0700|if cached file is missing cleanup then re-extract|cleanup ignores already-removed files
|Conventions:Process Group Management (staticx-specific):Setpgid:true→Pdeathsig:SIGKILL→Cancel kills -pgid→WaitDelay=0→syscall.Wait4(-pgid,WNOHANG) zombie reaping|maxWaitAttempts=100
|Conventions:Output Handling:wmiexec license banner occupies first two stdout lines|wmiRunCmd strips those lines before returning stdout
|Conventions:File Upload:UploadFile builds wmiexec shell command `lput <src> <dst>` and delegates to RunCommand|wrap errors with src/dst context
|Conventions:Tests:use wmi_bin_test.go for binary cache/create/recreate/cleanup/concurrent behavior|wmix_test.go depends on .env and real/mock remote target setup

|Anti-patterns:removing Setpgid/Pdeathsig/Cancel/Wait4 causes staticx child/zombie leaks|do not simplify subprocess cleanup without proving process-tree behavior
|Anti-patterns:editing wmiexec/binaries/ directly without rebuilding via wmiexec pipeline|do not hand-edit generated embedded binary payloads
|Anti-patterns:adding service-specific Windows/WMI policy here|service orchestration belongs in internal/\*, wmix stays shared transport wrapper
|Anti-patterns:bypassing wmiBinaryPath cache/self-heal path or calling wmiexec binary paths directly|do not duplicate binary extraction helpers
|Anti-patterns:logging credentials or embedding password/domain/user in stable log messages|errors may carry operational context only when needed and must not expose secrets
