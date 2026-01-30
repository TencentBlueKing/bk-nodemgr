# 开发环境准备

## 必需工具

### Go 1.23.10

```bash
go get golang.org/dl/go1.23.10@latest
go install golang.org/dl/go1.23.10@latest
go1.23.10 download
```

### protoc-gen-go v1.36.5

```bash
go1.23.10 install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5
protoc-gen-go --version
```

### protoc-gen-openapiv2 v2.27.2

```bash
go1.23.10 install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.2
protoc-gen-openapiv2 --version
```

## LSP 服务器

用于 Vibe Coding 等 IDE 的语言服务器安装。

> **注意**：首次使用 pnpm 全局安装前，需要先运行 `pnpm setup` 初始化全局环境。否则会出现 `ERR_PNPM_NO_GLOBAL_BIN_DIR` 错误。
>
> ```bash
> pnpm setup
> ```
>
> 该命令会自动创建全局目录、设置 `PNPM_HOME` 环境变量并更新 `PATH`。

### gopls (Go)

Go 语言服务器。

```bash
go1.23.10 install golang.org/x/tools/gopls@latest
gopls version
```

### pyright (Python)

Python 静态类型检查器和语言服务器。

```bash
pnpm install -g pyright
pyright --version
```

### yaml-language-server (YAML)

YAML 语言服务器。

```bash
pnpm install -g yaml-language-server
```

### bash-language-server (Bash)

Bash 语言服务器。

```bash
pnpm install -g bash-language-server
bash-language-server --version
```

### vtsls (TypeScript/JavaScript)

TypeScript 和 JavaScript 语言服务器，支持前端项目的 `.ts`、`.tsx`、`.js`、`.jsx` 文件以及 Vue 文件。

```bash
npm install -g @vtsls/language-server
```

## 验证

运行 `make pre` 准备构建环境。
