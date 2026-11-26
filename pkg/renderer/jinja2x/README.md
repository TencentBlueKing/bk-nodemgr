# jinja2x

## 设计意图

本包为 bk-nodemgr 项目提供 Jinja2 模板渲染能力，解决 Go 程序中需要复杂模板语法（特别是配置文件生成场景）的问题。

## 功能边界

1. 此包负责：Jinja2 模板渲染、支持自定义循环语法、内嵌渲染进程管理
2. 此包不负责：Jinja2 引擎的具体实现、模板语法的标准化

## 运行方式

每个 Go 进程按需启动一个内嵌 Jinja2 HTTP Server，通过 `127.0.0.1` 通信并复用连接。Server 退出后，下一次渲染自动启动新的 Server。当次失败的请求不会自动重试。

默认端口范围为 `28150–28159`，Backend 可通过 `jinja2.portRange` 配置。进程间请求携带随机认证令牌，端口不需要对外开放。

## API 使用

### 基本渲染

```go
// 包级别函数
result, err := jinja2x.Render("Hello {{ name }}!", map[string]any{"name": "World"})

// 或使用 Handler
handler := jinja2x.New()
result, err := handler.Render("Server: {{ host }}:{{ port }}", map[string]any{
    "host": "localhost",
    "port": 8080,
})
```
