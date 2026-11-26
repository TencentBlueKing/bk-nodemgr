|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow pipe-index format; keep concise; no prose/code blocks
|Scope:pkg/renderer/jinja2x

|Overview:staticx-packaged Python jinja2 binary wrapped as embedded subprocess renderer|implements renderer.IHandler alongside gotemplate
|Structure:jinja2x/:jinja2x.go|server.go|client.go|jinja2_bin_linux.go|jinja2x_test.go|jinja2/:jinja2.go|jinja2_linux_amd64.go|jinja2_linux_arm64.go|jinja2_exec.py|Dockerfile|jinja2_builder.sh|binaries/
|Where to look:public API (Handler,Render,New):jinja2x.go|server lifecycle:server.go,jinja2_bin_linux.go|HTTP IPC:client.go|embedded binary + build tag:jinja2/jinja2_linux_{amd64,arm64}.go|test cases:jinja2x_test.go,server_test.go|rebuild pipeline:jinja2/{Dockerfile,jinja2_builder.sh,jinja2_exec.py}|pre-built binaries:jinja2/binaries/

|Conventions:Process Group Management (staticx-specific):Setpgid:true→Pdeathsig:SIGKILL→Cancel kills -pgid→WaitDelay=0→zombie reaping via syscall.Wait4(-pgid,WNOHANG) with maxWaitAttempts=100|ref:pkg/wmix/wmi_bin.go
|Conventions:Binary Lifecycle:mutex-protected cached extraction to temp file|chmod 0700|re-extract if removed
|Conventions:Server Lifecycle:one lazily started server per Go process|cmd.Wait detects exit; next render replaces exited server|stdin EOF detects parent exit|Close stops server at backend shutdown
|Conventions:IPC:authenticated HTTP over 127.0.0.1|reuse client connections|bounded render timeout|preserve renderer.IHandler contract and template semantics

|Anti-patterns:removing Setpgid/Pdeathsig/Cancel/Wait4 causes staticx zombie leaks|adding service-specific rendering logic here (shared renderer, orchestration→internal/)|editing jinja2/binaries/ directly (rebuild via staticx pipeline)|replacing context.Background() without understanding zombie-reaping implications
