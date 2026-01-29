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

## 验证

运行 `make pre` 准备构建环境。
