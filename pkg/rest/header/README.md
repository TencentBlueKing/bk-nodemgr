# header

## 设计意图

此包用于存储整个蓝鲸通用的 header, 比如租户的 header

## 注意事项

1. 请勿在此包中定义 http 的标准 header;
2. 请勿将 apigw 的 header 放入此 pkg;
3. 如果是一些第三方系统特有的 header, 即使那些 header 是 http 的标准 header 也请勿放入此包中;