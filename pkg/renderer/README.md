# renderer

## 设计意图

本包为 bk-nodemgr 项目提供模板渲染能力，解决 Go 程序中需要复杂模板语法（特别是配置文件生成场景）的问题。

## 功能边界

1. 此包负责：生成所需模板解析器、模板渲染
2. 此包不负责：Jinja2 与 go-template 引擎的具体实现、模板语法的标准化，仅作为调用入口封装

## 可用引擎

- jinja2x: 基于 Jinja2 的模板渲染引擎
- gotemplate: 基于 Go 标准库 text/template 的模板渲染引擎
