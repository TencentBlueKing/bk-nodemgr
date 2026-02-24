# test

## 目录职责

本目录用于存放节点管理系统的集成测试用例，主要测试各个服务的 API 接口功能。

## 目录结构

```
test/
├── cases/              # 测试用例
│   ├── precheck/       # 预检查（最先执行，验证数据同步，按域组织）
│   │   ├── cmdb/       # CMDB 数据同步检查
│   │   └── ...
│   ├── backend/        # backend 服务测试
│   │   └── topo/       # topo 模块（按路由层级组织）
│   │       ├── networkunit/
│   │       ├── networkarea/
│   │       └── ...
│   ├── application/    # application 服务测试
│   │   └── topo/       # topo 模块（按路由层级组织）
│   │       ├── networkunit/
│   │       ├── networkarea/
│   │       └── ...
│   └── file/           # file 服务测试
│       └── publish/    # publish 模块（按路由层级组织）
│           └── release/
│               ├── agent/
│               └── ...
├── data/               # 预设测试数据（唯一数据源）
│   └── cmdb.yaml
├── tools/              # CI 辅助脚本
│   └── gen-mock-config.sh
├── helper/             # 测试辅助函数
├── test.go             # 测试入口和配置
├── Makefile            # 构建脚本
└── build/              # 编译后的测试二进制（.gitignore）
    ├── precheck/
    ├── backend/
    ├── application/
    └── file/
```

## 功能边界

1. 此目录负责：
   - API 接口的功能性测试
   - 数据同步正确性预检查
   - 数据正确性验证
   - 错误场景覆盖
   - 测试数据的准备(MOCK)

2. 此目录不负责：
   - 单元测试（应在对应代码目录下）
   - 性能测试

## 本地开发

```bash
# 在 test/ 目录下构建所有测试用例
cd test/
make build

# 或在项目根目录构建（会触发 test 目标）
make all
```

### 运行测试

```bash
# 运行所有测试（按 TEST_GROUPS 顺序：precheck → backend → application → file）(推荐)
make test

# 清理测试文件
make clean
```


## 测试配置

测试配置通过命令行参数或环境变量传入：

```go
// test.go
flag.StringVar(&TestFlagBackendServer, "backend", "127.0.0.1:28100", "backend host and port")
flag.StringVar(&TestFlagApplicationServer, "application", "127.0.0.1:28000", "application host and port")
flag.StringVar(&TestFlagFileServer, "file", "127.0.0.1:28200", "file host and port")
flag.StringVar(&TestFlagDataDir, "data-dir", "", "path to test data directory")
```

运行时指定配置：
```bash
cd build/backend/topo
./networkunit.test -backend=ip:port -data-dir=../../data
```

## 注意事项

1. 测试用例应保持独立性，不依赖执行顺序
2. 测试数据应使用随机后缀避免冲突
3. 测试失败时应输出清晰的错误信息
4. 避免硬编码，使用配置或生成器
5. 预设数据统一维护在 `data/cmdb.yaml`，不要在测试代码中硬编码
6. 预设数据采用增量策略：不删除、不修改已有条目，只追加新条目（修 bug 除外）
7. 新增测试组需要在 `Makefile` 的 `TEST_GROUPS` 中注册

## 演进方向

1. 完善压力测试
2. 支持并发测试
3. 增加测试报告和覆盖率统计
