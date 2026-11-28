# installer

`installer` defines shared installer command tokens, file conventions, status values, and node/plugin script rendering for `tools/cmd/installer` consumers.

## Command rendering

Callers provide installer parameters and select the command form required by the execution channel:

- `RenderUnixLaunchScript` renders a script, creates and enters its workspace, writes the script with `echo`, and launches it with `sh`.
- `RenderWindowsLaunchScript` renders a batch script, creates and enters its workspace, writes the script with `echo`, and launches it with `cmd /c`.
- `RenderWindowsCommand` enters an existing workspace and runs the rendered command without writing a batch file or creating a directory.
- `WindowsSSHLaunchCommand` launches a transferred batch file asynchronously through an encoded PowerShell `Start-Process` command. `EnsureWindowsWorkDirCommand` prepares its workspace with PowerShell `New-Item`.
- `WindowsShellLaunchCommand` and `EnsureWindowsShellWorkDirCommand` render the corresponding Cygwin shell commands.
- Transport helpers render Unix `mkdir`/`chmod` commands and WMI `mkdir`/batch-path commands. WMI commands retain their separate command-channel form; they are not PowerShell commands.

Render callbacks return the script name, content, and render error. The launch helpers return a complete command for the selected form. Relay proxy wire commands retain only the script-write and launch steps from `RenderUnixScript`; the receiver prepares the workspace with `UnixWorkspaceCommand`, which accepts only trusted commands already rendered by this package.

## Responsibility boundary

This package renders commands only. Callers retain connection management, file transfer, remote execution, workflow ordering, logging, and installation-result handling. Preparing a transferred script and rendering its launch command are separate operations.
