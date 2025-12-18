# 手动安装脚本执行先决条件分析

本文档分析了 `script_tools/manual` 目录下各平台脚本的执行先决条件。

## 目录
- [Linux 平台](#linux-平台)
- [Darwin (macOS) 平台](#darwin-macos-平台)
- [Windows 平台](#windows-平台)
- [通用要求](#通用要求)

---

## Linux 平台

### 脚本文件
- `linux/install.sh`

### 必需软件/工具

#### 1. **Shell 环境**
- **Bash Shell**: 脚本使用 `#!/bin/bash`，需要 Bash 解释器
- **最低版本**: Bash 3.0+（支持基本的变量操作和函数）
- **兼容性**: 大多数现代 Linux 发行版默认包含

#### 2. **系统检测工具**
- **`uname`**: 用于检测操作系统类型和 CPU 架构
  - `uname -s`: 检测操作系统类型
  - `uname -m`: 检测 CPU 架构
  - 通常位于 `/bin/uname` 或 `/usr/bin/uname`
  - **可用性**: 所有 POSIX 兼容系统都包含

#### 3. **网络工具**
- **`curl`**: 用于 HTTP 请求（下载安装程序和与服务器通信）
  - 必需功能: 支持 `-s`（静默模式）、`-w`（写入格式）、`-X POST`、`-H`（请求头）、`-d`（POST 数据）、`-o`（输出文件）
  - **最低版本**: curl 7.0+（支持上述所有参数）
  - **安装**: 
    - Ubuntu/Debian: `apt-get install curl`
    - CentOS/RHEL: `yum install curl` 或 `dnf install curl`
    - Alpine: `apk add curl`

#### 4. **文本处理工具**
- **`sed`**: 用于文本处理和字符串提取
  - 必需功能: 支持 `-n`（不打印）、`'1p'`（打印第一行）、正则表达式替换
  - **可用性**: 所有 Linux 发行版默认包含
- **`head`**: 用于提取文件前几行
- **`tail`**: 用于提取文件后几行
- **`echo`**: 标准输出命令
- **`tr`**: 字符转换工具（用于大小写转换）

#### 5. **文件系统工具**
- **`pwd`**: 获取当前工作目录
- **`mkdir`**: 创建目录（支持 `-p` 参数递归创建）
- **`chmod`**: 修改文件权限（用于使安装程序可执行）
- **`cat`**: 读取文件内容
- **`rm`**: 删除文件

#### 6. **时间工具**
- **`date`**: 用于获取时间戳（`date +%s`）
  - 必需功能: 支持 `+%s` 格式（Unix 时间戳）
  - **可用性**: 所有 Linux 发行版默认包含

### 网络要求
- 能够访问回调服务器（`CALLBACK_SERVER_URL`）
- 能够访问下载服务器（`DOWNLOAD_SERVER_URL`）
- 支持 HTTP/HTTPS 协议
- 防火墙允许出站连接

### 权限要求
- 对当前工作目录有读写权限
- 能够创建临时目录和文件
- 能够执行下载的安装程序（可能需要执行权限）

---

## Darwin (macOS) 平台

### 脚本文件
- `darwin/install.sh`

### 必需软件/工具

#### 1. **Shell 环境**
- **Bash Shell**: 脚本使用 `#!/bin/bash`
- **注意**: macOS 10.15+ 默认使用 zsh，但 Bash 仍然可用
- **最低版本**: Bash 3.2+（macOS 默认版本）

#### 2. **系统检测工具**
- **`uname`**: 与 Linux 相同
  - macOS 系统默认包含

#### 3. **网络工具**
- **`curl`**: 与 Linux 相同
  - macOS 系统默认包含（但版本可能较旧）
  - **推荐**: 使用 Homebrew 安装最新版本: `brew install curl`

#### 4. **文本处理工具**
- **`sed`**: macOS 版本为 BSD sed，与 GNU sed 略有差异，但脚本使用的功能兼容
- **`head`**, **`tail`**, **`echo`**, **`tr`**: 与 Linux 相同

#### 5. **文件系统工具**
- 与 Linux 相同（`pwd`, `mkdir`, `chmod`, `cat`, `rm`）

#### 6. **时间工具**
- **`date`**: macOS 版本支持 `+%s` 格式

### 网络要求
- 与 Linux 相同

### 权限要求
- 与 Linux 相同
- macOS 可能需要授予终端完全磁盘访问权限（某些版本）

---

## Windows 平台

### 脚本文件
- `windows/install.bat`

### 必需软件/工具

#### 1. **命令行环境**
- **CMD.exe (命令提示符)**: 脚本使用批处理语法
- **最低版本**: Windows XP+ 的 CMD.exe
- **推荐**: Windows 7+（更好的错误处理）

#### 2. **PowerShell**
- **必需**: PowerShell 用于执行 HTTP 请求
- **最低版本**: 
  - PowerShell 2.0+（Windows 7/Server 2008 R2 默认）
  - PowerShell 3.0+（推荐，Windows 8/Server 2012+）
- **必需功能**:
  - `Invoke-WebRequest` (PowerShell 3.0+)
  - `Get-Content`
  - `Set-Content`
  - `Test-Path`
  - `Remove-Item`
  - `New-Object System.IO.StreamReader`
- **执行策略**: 可能需要设置执行策略
  ```powershell
  Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```
  或使用 `-ExecutionPolicy Bypass` 参数

#### 3. **系统检测工具**
- **PowerShell**: 用于检测 CPU 架构（`$env:PROCESSOR_ARCHITECTURE`）
- **WMIC** (Windows Management Instrumentation Command-line): 作为 CPU 架构检测的备选方案
  - Windows XP+ 包含
  - Windows 10 1903+ 可能已弃用，但脚本有 PowerShell 备选方案

#### 4. **网络工具**
- **PowerShell `Invoke-WebRequest`**: 用于所有 HTTP 请求
  - 替代 curl（Windows 10 1803+ 包含 curl，但脚本使用 PowerShell）
  - 支持 HTTPS（需要有效的 SSL 证书）

#### 5. **文件系统工具**
- **CMD 内置命令**: `mkdir`, `del`, `cd`, `for`, `if`, `set`
- **PowerShell**: 用于文件操作（`Set-Content`, `Get-Content`, `Test-Path`）

#### 6. **其他工具**
- **`ping`**: 用于实现延迟（`ping 127.0.0.1 -n N`）
  - 所有 Windows 版本包含

### 网络要求
- 与 Linux/macOS 相同
- 可能需要配置代理（通过 PowerShell 的代理设置）

### 权限要求
- 对当前目录有读写权限
- 能够创建临时文件（`%TEMP%` 目录）
- 可能需要管理员权限（取决于安装程序要求）

### 特殊注意事项
- **执行策略**: PowerShell 执行策略可能阻止脚本运行
- **防病毒软件**: 可能拦截下载的文件或 PowerShell 脚本
- **防火墙**: Windows 防火墙可能阻止出站连接
- **编码问题**: 脚本使用 UTF-8 编码，确保终端支持

---

## 通用要求

### 网络连接
- **必需**: 所有平台都需要网络连接
- **协议**: HTTP/HTTPS
- **超时**: 脚本有 300 秒（5分钟）的超时机制

### 服务器可访问性
- 回调服务器必须可访问（用于报告系统信息和获取命令）
- 下载服务器必须可访问（用于下载安装程序）

### 临时文件
- 需要能够创建临时文件
- Linux/macOS: 使用当前工作目录或 `/tmp`
- Windows: 使用 `%TEMP%` 目录

### 环境变量
- 脚本不依赖特定的环境变量（除了 Windows 的 `%TEMP%` 和 `%CD%`）

---

## 总结

### 最低要求对比

| 平台        | Shell/解释器 | 网络工具              | 系统检测        | 其他关键工具                  |
| ----------- | ------------ | --------------------- | --------------- | ----------------------------- |
| **Linux**   | Bash 3.0+    | curl 7.0+             | uname           | sed, date, mkdir, chmod       |
| **macOS**   | Bash 3.2+    | curl (系统默认或更新) | uname           | sed (BSD), date, mkdir, chmod |
| **Windows** | CMD.exe      | PowerShell 3.0+       | PowerShell/WMIC | ping (用于延迟)               |

### 常见问题排查

1. **curl 未找到** (Linux/macOS): 安装 curl 包
2. **PowerShell 执行策略错误** (Windows): 使用 `-ExecutionPolicy Bypass`
3. **网络连接失败**: 检查防火墙和代理设置
4. **权限不足**: 确保对目标目录有写权限
