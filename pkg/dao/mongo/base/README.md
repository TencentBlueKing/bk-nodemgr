# base

## 设计意图
base 包是一个用于提供MongoDB操作的基础功能的包。它封装了MongoDB的基本操作，简化了与数据库的交互，使得其他模块可以更方便地进行数据存取。
该包的设计目标是提供一个易于使用的接口，同时保持对MongoDB强大功能的访问。通过使用此包，开发者可以更专注于业务逻辑，而不必过多关注底层数据库操作的细节。


## 功能边界
1. 此包负责：
    - 提供MongoDB的基本操作接口，如连接、查询、插入、更新和删除等。
    - 封装MongoDB的连接配置和操作逻辑，简化数据库交互。
2. 此包不负责：
    - 业务逻辑的实现。

## 设计考量
1. 提高代码的可读性和可维护性，减少重复代码。

## ScanAll 索引规则

`ScanAll` 使用 MongoDB `_id` 作为 cursor key，按 `_id` 升序分批扫描：第一批使用调用方传入的 filter，后续批次会追加 `_id > lastID`。因此，`ScanAll` 的性能取决于调用方 filter 与 `_id` cursor 是否能被同一个 compound index 支持。

新增或修改 `ScanAll` 使用点时，调用方必须为该查询模式评估并补充特化索引：

1. 将稳定的 equality filter 字段放在索引前缀。
2. 将 `_id` 放在索引最后，用于 cursor range 与 sort。
3. 如果查询始终带有 `basic.is_deleted = false`，优先使用 partial index 缩小索引范围。
4. partial index 的 filter 必须能被 `ScanAll` 查询条件完整蕴含；不要依赖查询条件没有包含的 partial filter 字段。

例如，按业务扫描未删除 Host 时，查询条件为 `basic.is_deleted = false` 与 `data.static.biz_id = <biz_id>`，应在 Host DAO 的 `GetIndexes()` 中提供匹配索引：

```go
mongo.IndexModel{
	Keys: bson.D{
		{Key: "data.static.biz_id", Value: 1},
		{Key: "_id", Value: 1},
	},
	Options: options.Index().SetPartialFilterExpression(bson.D{
		{Key: "basic.is_deleted", Value: false},
	}),
}
```

不要为了解决 `ScanAll` 慢查询优先扩展通用 cursor/sort 能力；除非业务已经确认需要新的扫描顺序，否则保持 `_id` cursor 语义，并通过匹配当前 filter 的特化索引解决扫描放大。
