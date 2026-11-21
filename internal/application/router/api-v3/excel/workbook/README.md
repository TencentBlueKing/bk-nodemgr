# workbook

`api-v3/excel` 的 Excel 模板工作簿实现：列定义驱动的三行表头渲染、样式与下拉校验，并按列顺序解析回数据行。仅服务 application 的 Excel 模板接口，依赖标准库与 excelize。

## 职责

- 渲染模板：三行表头、列宽与行高、表头样式、下拉校验
- 解析回读：校验表头与必填标注行后，按列顺序返回数据行
- 类型化模板：`TypedColumn[T]` 把列定义绑定到行类型，渲染与解析分别调用 `ValueOf`、`Parse`

## 非职责

- 列间约束、取值合法性等校验，由调用方在 `Parse` 或调用方代码中完成
- 多语言：当前只输出 zh-CN 文案
- HTTP 绑定、业务字段转换与主机解析：留在上级 `excel` 处理器包

## 布局契约

| 行  | 内容                                              |
| --- | ------------------------------------------------- |
| 1   | `Column.Name`                                     |
| 2   | `Column.Required` 渲染出的 `必填` / `可选`        |
| 3   | `Column.Description`                              |
| 4+  | 数据行，`IndexedRow.RowIndex` 即此处的 Excel 行号 |

列宽按表头、标注、说明的显示宽度估算，单列限制在 24–50，非 ASCII 字符按 2 个宽度单位计算；行高按换行后的行数估算。

## 下拉校验

- `Options` 为空时不生成校验，单元格自由填写
- 选项总长不超过 250 个 UTF-16 单元时写内联公式；超长时写入隐藏 sheet `_options_<n>` 并用区域引用
- `AllowCustom=false` 用 stop 级校验（只能选列表值），`true` 用 warning 级（允许自定义值）
- 校验范围是整列，从第 4 行到工作表末行

## 解析与错误

- 表头行与列定义不一致 → `invalid excel template header`
- 必填标注行不一致 → `outdated excel template layout, please download the latest template`
- 无数据行 → `no data found in excel file`
- `ParseIndexed` 不中断整表解析，把行级错误写入 `IndexedRow.Err`，由调用方决定是否逐行返回
- `Required: true` 的列为空白时，行级错误为 `column "列名" is required`
- 所有单元格均为空白的行会被跳过
- 错误文案统一英文（含返回给调用方的文件级与行级错误）；zh-CN 仅用于渲染进模板的文案（表头/必填标注/下拉提示）

第二行文案变化会让旧模板被判定为 `outdated`，需要用户重新下载模板。

## 文案

`constants.go` 集中存放布局、下拉与渲染文案常量（含必填/可选、下拉提示），当前固定 zh-CN。

## 用法

Pass the sheet name and `[]workbook.TypedColumn[T]` to `workbook.NewTemplate`. Use the returned
`workbook.IHandler[T]` to write `[]T` with `Write` or read `[]workbook.IndexedRow[T]` with `ParseIndexed`.

## 放置约束

本包是 application 的内部实现，不放入 `pkg`；只有第二个服务真正复用时，才按 `pkg/AGENTS.md` 评估上移。
