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