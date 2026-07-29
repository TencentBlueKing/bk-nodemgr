# 开发环境准备

## 必需工具

### 可选: 使用 mise 管理项目所需工具

> 如果使用vscode等工具进行开发时找不到go的二进制可以使用mise全局安装go1.25.12
> `mise use -g go@1.25.12`
> 安装完成后重启vscode即可

```bash
curl https://mise.run | sh
# use bash
echo 'eval "$(~/.local/bin/mise activate bash)"' >> ~/.bashrc
# use zsh
echo 'eval "$(~/.local/bin/mise activate zsh)"' >> ~/.zshrc
mise install
pnpm setup
# use bash
source ~/.bashrc
# use zsh
source ~/.zshrc
mise run setup-extra-tools
```

### Go 1.25.12

```bash
go get golang.org/dl/go1.25.12@latest
go install golang.org/dl/go1.25.12@latest
go1.25.12 download
```

### protoc-gen-go v1.36.5

```bash
go1.25.12 install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5
protoc-gen-go --version
```

### protoc-gen-openapiv2 v2.27.2

```bash
go1.25.12 install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.2
protoc-gen-openapiv2 --version
```

### golangci-lint v2.4.0

```bash
go1.25.12 install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0
golangci-lint --version
```

### protoc（Protocol Buffers 编译器）

用于编译 `.proto` 文件生成 Go 代码和 Swagger 文档。

```bash
# 下载 protoc
PROTOC_VERSION=29.3
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip
unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /tmp/protoc
sudo cp /tmp/protoc/bin/protoc /usr/local/bin/
sudo cp -r /tmp/protoc/include/* /usr/local/include/
rm -rf /tmp/protoc protoc-${PROTOC_VERSION}-linux-x86_64.zip

# 验证安装
protoc --version
```

### Node.js 和 pnpm

前端开发必需工具。项目使用 pnpm@9.8.0 作为包管理器。

**安装 Node.js（推荐使用 nvm）：**

```bash
# 安装 nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash

# 重新加载 shell 配置
source ~/.bashrc  # 或 source ~/.zshrc

# 安装 Node.js LTS 版本
nvm install --lts
nvm use --lts

# 验证安装
node --version
npm --version
```

**安装 pnpm：**

```bash
# 使用 npm 安装 pnpm
npm install -g pnpm@9.8.0

# 初始化 pnpm 全局环境
pnpm setup

# 验证安装
pnpm --version
```

### UPX v5.1.0（可选，用于二进制压缩）

构建 tools 时启用 `UPX_ENABLED=1` 需要安装 UPX。

```bash
curl -L https://github.com/upx/upx/releases/download/v5.1.0/upx-5.1.0-amd64_linux.tar.xz -o /tmp/upx-5.1.0-amd64_linux.tar.xz
tar -xf /tmp/upx-5.1.0-amd64_linux.tar.xz -C /tmp
cp /tmp/upx-5.1.0-amd64_linux/upx /usr/local/bin/upx
upx --version
```

安装后，构建 tools 时会自动对 linux/windows 平台的二进制进行压缩（macOS/arm 平台不支持，会跳过）：

```bash
make tools   # 默认启用 UPX_ENABLED=1
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
go1.25.12 install golang.org/x/tools/gopls@latest
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

## 本地开发环境

### MongoDB

后端服务依赖 MongoDB 数据库。可以使用 Docker 快速启动：

```bash
# 启动 MongoDB 容器
docker run -d \
  --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongo:7.0

# 验证连接
docker exec -it mongodb mongosh -u admin -p admin
```

或使用项目提供的 docker-compose 配置：

```bash
cd install/docker-compose/bk-nodemgr
./deploy.sh
```

### 配置文件

后端服务需要配置文件才能启动。参考模板位于 `install/docker-compose/bk-nodemgr/templates/config/`：

- `bk-nodemgr-backend.yml.tpl` - backend 服务配置模板
- `bk-nodemgr-application.yml.tpl` - application 服务配置模板
- `bk-nodemgr-file.yml.tpl` - file 服务配置模板

根据实际环境修改配置文件中的数据库连接、端口等信息。

## 本地开发流程

### 1. 克隆代码并安装依赖

```bash
# 克隆仓库
git clone https://github.com/TencentBlueKing/bk-nodemgr.git
cd bk-nodemgr

# 准备构建环境（会自动安装 Go 1.25.12 并执行 go mod tidy）
make pre
```

### 2. 编译 Proto 文件

如果修改了 `.proto` 文件，需要重新生成代码：

```bash
cd proto
make clean
make all
cd ..
```

### 3. 编译后端服务

```bash
# 编译所有后端服务
make backend application file relay

# 或编译单个服务
make backend      # 编译 backend 服务
make application  # 编译 application 服务
make file         # 编译 file 服务
make relay        # 编译 relay 服务
```

编译产物位于 `build/<version>/` 目录。

### 4. 启动服务

```bash
# 启动 backend 服务（需要先准备配置文件）
./build/<version>/bk-nodemgr-backend -f /path/to/config.yaml

# 启动 application 服务
./build/<version>/bk-nodemgr-application webserver -f /path/to/config.yaml

# 启动 file 服务
./build/<version>/bk-nodemgr-file -f /path/to/config.yaml
```

### 5. 前端开发

```bash
cd front

# 安装依赖
pnpm install

# 启动开发服务器（默认端口 5008）
pnpm dev

# 构建生产版本
pnpm build

# 类型检查
pnpm typecheck

# 代码检查
pnpm lint
```

前端开发服务器启动后，访问 `http://localhost:5008` 即可查看页面。

### 6. 运行测试

```bash
# 后端测试
go1.25.12 test ./...

# 前端单元测试
cd front
pnpm test:unit

# 前端 E2E 测试
pnpm test:e2e
```

### 7. 代码检查

```bash
# 后端代码检查
golangci-lint run

# 前端代码检查
cd front
pnpm lint
```

## 验证

完成所有工具安装后，运行以下命令验证环境：

```bash
# 验证 Go 工具链
go1.25.12 version
protoc-gen-go --version
protoc-gen-openapiv2 --version
golangci-lint --version

# 验证 Proto 工具
protoc --version
clang-format --version

# 验证前端工具
node --version
pnpm --version

# 验证 LSP 服务器
gopls version
pyright --version
yaml-language-server --version
bash-language-server --version

# 准备构建环境
make pre
```

## 常见问题

### pnpm 全局安装失败

如果遇到 `ERR_PNPM_NO_GLOBAL_BIN_DIR` 错误，需要先运行：

```bash
pnpm setup
```

该命令会自动创建全局目录、设置 `PNPM_HOME` 环境变量并更新 `PATH`。

### protoc 找不到

确保 `/usr/local/bin` 在 `PATH` 环境变量中：

```bash
export PATH="/usr/local/bin:$PATH"
```

建议将此行添加到 `~/.bashrc` 或 `~/.zshrc` 中。

### MongoDB 连接失败

检查 MongoDB 是否正常运行：

```bash
docker ps | grep mongodb
```

确认配置文件中的连接信息正确，包括主机、端口、用户名、密码等。
