# winapi

## 设计意图

本包主要提供一些封装好的 windows api封装，方便使用。
本包提供解析后的数据，不直接返回windows api的原始数据。

## 功能边界

1. 此包负责：
    - windows api的封装
    - 将 windows api的原始数据，转换为结构化的, 可读性的数据
2. 此包不负责：
    - 提供 windows api的原始数据

## 设计考量

本包主要是为了减少对 windows api的直接调用, 提高这类文本解析代码的复用率。
