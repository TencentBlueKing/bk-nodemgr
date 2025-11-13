# jinja2

## 设计意图

本包基于 Python 的 Jinja2 模板引擎，为 Go 程序提供灵活的配置文件模板渲染支持，专门用于蓝鲸节点管理系统中的动态配置生成。

## 功能边界

1. 此包负责：Jinja2 模板渲染、支持自定义循环语法、文件参数处理
2. 此包不负责：Jinja2 引擎的具体实现、模板语法的标准化

## 使用限制

- 本包需要 Python 3.x 环境
- 依赖 jinja2 和 six 库
- 模板文件需使用 UTF-8 编码

## 演进方向

- 支持更多模板引擎（如 Go template）
- 支持模板文件热重载
- 支持模板语法验证

## jinja2_exec.py 脚本介绍

jinja2_exec.py 是一个命令行工具，用于渲染 Jinja2 模板文件。它支持标准的 Jinja2 语法以及自定义的循环语法，特别适用于配置文件的动态生成。

### 核心功能

- **模板渲染**: 支持字符串和结构化数据的递归渲染
- **自定义循环语法**: 支持 `$for`、`$item`、`$body` 的循环结构
- **模板缓存**: 内置模板缓存机制提高性能
- **安全沙箱**: 使用 Jinja2 的 SandboxedEnvironment
- **错误处理**: 渲染失败时返回原始数据并输出错误信息

### 基本语法

jinja2_exec.py 的基本命令行语法如下：

```
python jinja2_exec.py [-h] [-t TEMPLATE_FILE] [-c CONTEXT_FILE] [-o OUTPUT_FILE] [-v] [--version] [template] [context] [output]
```

参数说明：

- `template`: 模板文件路径（必需）
- `context`: 上下文文件路径，JSON 格式（必需）
- `output`: 输出文件路径（可选，默认输出到标准输出）

### 基本使用示例

#### 使用位置参数

最基本的使用方式是使用位置参数：

```bash
python jinja2_exec.py template.yaml context.json output.yaml
```

或输出到标准输出：

```bash
python jinja2_exec.py template.yaml context.json
```

#### 使用缩写选项

使用缩写选项参数：

```bash
python jinja2_exec.py -t template.yaml -c context.json -o output.yaml
```

#### 使用长选项

使用长选项参数：

```bash
python jinja2_exec.py --template template.yaml --context context.json --output output.yaml
```

### 模板语法支持

#### 标准 Jinja2 语法

支持标准的 Jinja2 模板语法：

```yaml
Hello {{ name }}!

Server: {{ host.ip }}:{{ host.port }}

{% if features %}
Features:
{% for feature in features %}
- {{ feature }}
{% endfor %}
{% endif %}
```

#### 自定义循环语法

支持自定义的循环语法，用于结构化数据模板：

```yaml
servers:
  $for: "servers"
  $item: "server"
  $body:
    name: "{{ server.name }}"
    ip: "{{ server.ip }}"
    port: {{ server.port }}
    enabled: true
```

或 JSON 格式：

```json
{
  "$for": "servers",
  "$item": "server",
  "$body": {
    "name": "{{ server.name }}",
    "ip": "{{ server.ip }}",
    "port": {{ server.port }},
    "enabled": true
  }
}
```

### 命令行选项详解

- `-t TEMPLATE_FILE, --template TEMPLATE_FILE`: 指定模板文件路径
- `-c CONTEXT_FILE, --context CONTEXT_FILE`: 指定上下文文件路径（JSON 格式）
- `-o OUTPUT_FILE, --output OUTPUT_FILE`: 指定输出文件路径
- `-v, --verbose`: 启用详细输出模式
- `--version`: 显示版本信息（当前版本：1.0.0）

### 上下文文件格式

上下文文件必须使用 JSON 格式：

```json
{
  "name": "张三",
  "app_name": "蓝鲸节点管理",
  "host": {
    "ip": "192.168.1.100",
    "port": 8080
  },
  "features": ["监控", "日志", "告警"],
  "servers": [
    {
      "name": "web-server",
      "ip": "10.0.1.10",
      "port": 80
    },
    {
      "name": "api-server",
      "ip": "10.0.1.11",
      "port": 8080
    }
  ]
}
```

### 高级功能

#### 模板缓存

脚本内置了模板缓存机制，相同的模板内容只会编译一次，提高渲染性能。

#### 安全沙箱

使用 Jinja2 的安全沙箱环境，防止模板中的恶意代码执行。

#### 错误处理

当模板渲染失败时，会输出详细的错误信息到 stderr，并返回原始数据。

#### 递归渲染

支持对字典、列表等复杂数据结构进行递归渲染，确保所有模板表达式都被正确处理。

### 使用示例

#### 示例 1：生成服务配置

模板文件 `service.yaml`：

```yaml
service:
  name: {{ service_name }}
  version: {{ version }}
  host: {{ host.ip }}
  port: {{ host.port }}
  environment: {{ environment }}
```

上下文文件 `production.json`：

```json
{
  "service_name": "api-service",
  "version": "1.0.0",
  "host": {
    "ip": "10.0.1.100",
    "port": 8080
  },
  "environment": "production"
}
```

执行渲染：

```bash
python jinja2_exec.py service.yaml production.json -o service-production.yaml
```

#### 示例 2：批量生成监控配置

模板文件 `monitor.yaml`：

```yaml
monitoring:
  services:
    $for: "services"
    $item: "service"
    $body:
      name: "{{ service.name }}"
      endpoint: "{{ service.host }}:{{ service.port }}"
      metrics: ["cpu", "memory", "disk"]
      interval: 60
```

上下文文件 `services.json`：

```json
{
  "services": [
    {
      "name": "web-service",
      "host": "10.0.1.10",
      "port": 80
    },
    {
      "name": "api-service",
      "host": "10.0.1.11",
      "port": 8080
    }
  ]
}
```

执行渲染：

```bash
python jinja2_exec.py monitor.yaml services.json -o monitor-config.yaml
```

### 核心实现说明

#### nested_render_data 函数

这是核心的渲染函数，支持：
- 字符串模板渲染
- 字典数据递归处理
- 自定义循环语法解析
- 列表数据遍历渲染

#### find_element 函数

支持点号分隔的路径访问，如 `a.b.c.d` 来获取嵌套字典中的值。

### 集成到 Go 程序

可以将 jinja2_exec.py 脚本编译为二进制文件，嵌入到 Go 程序中使用：

```go
// 在 Go 程序中调用
cmd := exec.Command("jinja2", "-t", "template.yaml", "-c", "context.json", "-o", "output.yaml")
output, err := cmd.CombinedOutput()
if err != nil {
    log.Fatal(err)
}
```