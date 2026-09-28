|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A and README|English for code and agent guidance
|Compression Rule:pipe-index format; concise; no prose/code blocks|Reference:.agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md
|Scope:pkg/filex/local
|Parent rules:pkg/filex/AGENTS.md owns shared stream ownership and close-once conventions; apply them without redefining ownership here
|Overview:local filesystem implementations of fileiface.File and fileiface.FileGroup|no service-specific upload, publication, tenant, or remote-backend policy
|Structure:pkg/filex/local:{local.go,local_file.go,local_file_group.go,utils_unix.go,utils_windows.go,local_file_test.go,local_file_group_test.go,local_file_group_unix_test.go,README.md}
|Where to look:filesystem access:local.go:{rFs,wFs}:lazily initialized afero.OsFs instances
|Where to look:file metadata/content:local_file.go:{NewLocalFile,MD5SumWithBuffer,LocalFile.Content,LocalFile.Info}
|Where to look:directory/storage:local_file_group.go:{NewLocalDir,LocalDir.GetFile,LocalDir.Store,writeDataToFile,writeChunk}
|Where to look:copy/remove boundaries:local_file_group.go:{LocalDir.Copy,copyDirectory,copyFile,LocalDir.Remove,resolvePathWithRoot}|directory methods:local_file_group.go:{LocalDir.IsDir,LocalDir.GetSubGroup,LocalDir.EnsureSubGroup}
|Where to look:directory path validation:local_file_group.go:resolveDirectoryPath|native overlap checks:local_file_group.go:{resolveTargetPath,isSameOrSubPath}
|Where to look:platform paths:{utils_unix.go,utils_windows.go}:GetLocalFileAbsFilePath/GetLocalFileGroupAbsDirPath/GetLocalFileGroupAbsFilePath
|Construction:NewLocalDir requires an existing directory; NewLocalFile requires an existing non-directory file and computes metadata including MD5|do not treat these constructors as creation operations
|Metadata:LocalFile.Info returns a construction-time snapshot; reacquire the File after changes when fresh metadata is required|MD5 is artifact checksum metadata, not a security authenticity guarantee
|Content:each successful LocalFile.Content opens a new OS handle and returns filex.OnceReadCloser; caller owns the returned stream and immediately defers Close|do not cache or reuse one open handle across Content calls
|Store:input is borrowed and must remain open after success or failure; Store owns only the destination handle it opens and defers its Close|writeDataToFile and writeChunk must not close caller-provided streams
|Overwrite:Store uses O_WRONLY/O_CREATE/O_EXCL when overwrite=false and O_WRONLY/O_CREATE/O_TRUNC when true|preserve atomic exclusive creation; do not replace it with an existence check followed by Create
|Write failures:Store is not transactional or atomic content publication; failed writes may leave partial files and overwrite truncates old content|do not claim rollback or old-content preservation
|Buffered writes:writeDataToFile checks cancellation between chunks and flushes its bufio.Writer before the owning caller closes the destination|writeChunk must preserve bytes returned together with io.EOF
|Copy:LocalDir destination uses native copy; other supported destinations use pkg/filex/transfer.CopyStream with FileGroup directory methods|before writes check actual target with os.SameFile (including hard links) and reject directory overlap in either direction after resolving symlinks
|Directory paths:resolveDirectoryPath validates group-relative slash paths and checks existing nodes from the group root with Lstat; missing paths are allowed for creation|IsDir/GetSubGroup report absence with fs.ErrNotExist; GetSubGroup(".") queries live state; EnsureSubGroup creates missing parents
|Directory methods:{IsDir,GetSubGroup,EnsureSubGroup} reject symlinks and special files encountered on resolved paths|AllFiles follows regular-file symlinks|containment is lexical and assumes directory is not concurrently modified by an untrusted process
|Paths:native Copy source and Remove use resolvePath; native Copy destination uses resolveDestinationPath|reject empty, absolute, volume-qualified, and escaping relative paths; group root is allowed only as a copy destination
|Path limits:resolvePathWithRoot performs lexical containment checks, not symlink sandboxing|Store and GetFile do not use this resolver; do not assume all methods enforce identical path validation
|Platforms:use filepath and existing platform helpers; Unix helpers restore the leading separator while Windows helpers preserve volume components|retain build tags and validate platform-specific behavior when changing path construction
|Errors:retain wrapped operation errors and destination Close failures in Store; cleanup must run on early returns and panic unwinding|do not restore implicit input closure or bypass shared close-once wrappers
|Tests:local_file_group_unix_test.go covers FIFOs with Unix build tags|local_file_group_test.go uses t.TempDir for constructor/copy/remove cases; local_file_test.go has environment-backed setup|do not report a selected subset as full-package coverage
|Commands:environment-independent subset:GOTOOLCHAIN=go1.25.12 go test ./pkg/filex/local -run 'TestNewLocalDir|TestLocalDir'|full package after environment setup:GOTOOLCHAIN=go1.25.12 go test ./pkg/filex/local
