# wmiexec

## 设计意图

本包基于 python 的 wmiexec 脚本，为Linux 下 Golang 使用 WMI 执行命令提供支持。

## 功能边界

1. 此包负责：通过 WMI 在 Windows 系统上执行命令，并返回结果。
2. 此包不负责：WMI 协议的具体实现

## 使用限制

- 本包仅支持在 Linux 系统的 AMD64 和 ARM64 架构上运行。

## 演进方向

- 支持在 Windows 系统上执行命令
- 支持在 MacOS 系统上执行命令

## wmiexec 脚本介绍

wmiexec 是 Impacket 工具集中的一个脚本，它使用 Windows Management Instrumentation (WMI) 接口在远程 Windows 系统上执行命令。它通过
Windows Management Instrumentation (WMI) 在 Windows 主机上提供交互式 shell

### 基本语法

wmiexec.py 的基本命令行语法如下：

```
wmiexec.py [-h] [-share SHARE] [-nooutput] [-ts] [-silentcommand] [-debug] [-codec CODEC] [-shell-type {cmd,powershell}] [-com-version MAJOR_VERSION:MINOR_VERSION] [-hashes LMHASH:NTHASH] [-no-pass] [-k] [-aesKey hex key] [-dc-ip ip address] [-target-ip ip address] [-A authfile] [-keytab KEYTAB] target [command ...]
```

其中：

- `target` 是必需的参数，格式为 `[[domain/]username[:password]@]<targetName or address>`
- `command` 是要在目标主机上执行的命令

### 基本使用示例

#### 使用用户名和密码

最基本的使用方式是指定域、用户名和密码：

```bash
wmiexec test.local/john:password123@10.10.10.1
```

或者：

```bash
wmiexec ignite/administrator:Ignite@987@192.168.1.105
```

执行后会获得一个交互式 shell，可以直接输入命令，例如：

```bash
wmiexec ignite/administrator:Ignite@987@192.168.1.105 dir
```

#### 使用哈希值（Pass-the-Hash）

如果你有 NTLM 哈希但没有明文密码，可以使用 `-hashes` 选项：

```bash
wmiexec administrator@10.10.10.1 -hashes :0e0363213e37b94221497260b0bcb4fc
```

执行后将启动半交互式 shell。

### 命令行选项详解

以下是一些重要的命令行选项：

- `-hashes LMHASH:NTHASH`: 使用 NTLM 哈希进行认证（Pass-the-Hash）
- `-no-pass`: 不提供密码（用于空密码）
- `-k`: 使用 Kerberos 认证
- `-aesKey hex key`: 用于 Kerberos 认证的 AES 密钥（128 或 256 位）
- `-dc-ip ip address`: 域控制器的 IP 地址。如果省略，将使用目标参数中指定的域部分（FQDN）
- `-target-ip ip address`: 目标机器的 IP 地址。如果省略，将使用指定为目标的任何内容
- `-A authfile`: smbclient/mount.cifs 风格的认证文件
- `-keytab KEYTAB`: 从 keytab 文件读取 SPN 的密钥

其他可用选项：

- `-share SHARE`: 指定用于输出的共享（默认为 ADMIN$）
- `-nooutput`: 不获取命令输出
- `-ts`: 向每个输出行添加时间戳
- `-silentcommand`: 执行命令而不输出
- `-debug`: 启用调试输出
- `-codec CODEC`: 设置目标编码（默认为 UTF-8）
- `-shell-type {cmd,powershell}`: 指定要使用的 shell 类型（cmd 或 powershell）
- `-com-version MAJOR_VERSION:MINOR_VERSION`: DCOM 版本，格式为"主版本号:次版本号"

### 交互式 Shell 功能

启动 wmiexec 后，你将获得一个半交互式 shell，可以在其中执行各种 Windows 命令。例如：

```
C:\>whoami
offsec\administrator
```

shell 中还支持一些额外的命令，例如 `put` 可以用来上传文件：

```
C:\>put SharpHound.exe
[*] Uploading SharpHound.exe to C:\SharpHound.exe
```

wmiexec 支持多种认证方法，还可以在半交互模式下运行 cmd 或 powershell 命令。所有 wmiexec 的命令都会根据使用的 shell
类型预先添加特定参数。