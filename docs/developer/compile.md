# 编译

项目代码分为两大部分
- 主代码: [go 1.23.10](../../go.mod)
- tool代码: [go 1.20](../../tools/go.mod)

主要产出物包括：
- application 二进制文件
- backend 二进制文件
- file 二进制文件
- relay 二进制文件
- 前端dist 资源文件
- 各平台的tool 二进制文件
- 各平台的gse bintool 工具包
- 各平台的gse plugin bintool 工具包
- 各平台的手动安装 脚本包

### 编译单独资源
编译获取单独的二进制或工具包, `make all`即获取所有产出物
```bash
make application backend
make front
make tools
make all
```

### 编译服务镜像
编译获取节点管理服务镜像(所有模块通用), apigw同步镜像
```bash
make docker-build-server
make docker-build-apigw-sync
```

### 编译插件包
编译获取relay插件包, 可直接用于上传至页面包管理使用
```bash
make plugin-pkg-relay
```

### 产出物一览
```bash
build/
└── build-tag
    ├── bintools
    │   ├── bintool.tgz
    │   └── plugin_bintool.tgz
    ├── bk-nodemgr-application
    ├── bk-nodemgr-backend
    ├── bk-nodemgr-file
    ├── bk-nodemgr-relay
    │   ├── plugins_linux_aarch64
    │   │   ├── bin
    │   │   │   └── bk-nodemgr-relay
    │   │   ├── definition.yaml
    │   │   ├── etc
    │   │   └── templates
    │   │       └── bk-nodemgr-relay.conf.template
    │   ├── plugins_linux_x86_64
    │   │   ├── bin
    │   │   │   └── bk-nodemgr-relay
    │   │   ├── definition.yaml
    │   │   ├── etc
    │   │   └── templates
    │   │       └── bk-nodemgr-relay.conf.template
    │   └── project.yaml
    ├── dist
    ├── scripts
    │   └── manual
    │       ├── darwin
    │       │   └── install.sh
    │       ├── linux
    │       │   └── install.sh
    │       └── windows
    │           └── install.bat
    └── tools
        ├── installer_darwin_amd64
        ├── installer_darwin_arm64
        ├── installer_linux_amd64
        ├── installer_linux_arm
        ├── installer_linux_arm64
        ├── installer_windows_amd64.exe
        ├── installer_windows_arm64.exe
        └── installer_windows_arm.exe
```