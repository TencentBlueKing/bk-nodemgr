# test

## 目录职责

本目录用于存放节点管理系统的集成测试用例，主要测试各个服务的 API 接口功能。

## 目录结构

```
test/
├── cases/              # 测试用例
│   ├── precheck/       # 环境初始化 + 验证（最先执行，确保测试环境就绪，按域组织）
│   │   ├── cmdb/       # CMDB 数据同步验证
│   │   ├── file/       # 初始化包上传与发布
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
├── config.go           # 环境配置结构体和加载逻辑（flag 注册和配置加载）
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
   - 测试环境初始化（precheck 阶段完成，包括初始化包上传、数据同步验证等）
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
make 

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

测试环境必须先准备 `env.yaml`，当前仓库不提供该文件，需手动填充并放置在测试机器上
环境配置统一在 `env.yaml` 中管理，所有字段和默认值见 `config.go` 中的 `EnvConfig` 结构体。

添加新参数：在 `config.go` 加结构体字段和默认值，测试中通过 `test.Env.XXX` 读取。

## 注意事项

1. 测试用例应保持独立性，不依赖执行顺序
2. 测试数据应使用随机后缀避免冲突
3. 测试失败时应输出清晰的错误信息
4. 避免硬编码，使用配置或生成器
5. 预设数据统一维护在 `data/` 目录，不要在测试代码中硬编码
6. 预设数据采用增量策略：不删除、不修改已有条目，只追加新条目（修 bug 除外）
7. 新增测试组需要在 `Makefile` 的 `TEST_GROUPS` 中注册

## 演进方向

1. 完善压力测试
2. 支持并发测试
3. 增加测试报告和覆盖率统计
