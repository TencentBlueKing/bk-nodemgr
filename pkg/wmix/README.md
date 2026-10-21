# wmix

`wmix` 是供 Linux 服务端连接 Windows 主机的共享远程执行组件，封装 WMI/DCOM 命令执行和 SMB 文件传输。

此包只提供传输与执行能力，不定义上层业务策略，也不判定远程任务的业务结果。

调用端支持 Linux amd64 和 arm64，目标端为 Windows。执行依赖目标主机开放相应的 WMI/DCOM 与 SMB 访问，以及调用端可执行临时文件。
