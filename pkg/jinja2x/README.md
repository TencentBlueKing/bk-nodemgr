# jinja2x

## 设计意图

本包为 bk-nodemgr 项目提供 Jinja2 模板渲染能力，解决 Go 程序中需要复杂模板语法（特别是配置文件生成场景）的问题。

## 功能边界

1. 此包负责：Jinja2 模板渲染、支持自定义循环语法、临时文件管理
2. 此包不负责：Jinja2 引擎的具体实现、模板语法的标准化


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