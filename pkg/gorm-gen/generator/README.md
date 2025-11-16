# GORM Gen
数据库模型/查询生成

## 生成模型/查询

在项目根目录下执行 (因为使用的是相对路径，所以需要在项目根目录下执行)

```shell
go run ./cmd/gen
```

## 默认行为

- 可空字段不使用 sql.NullXXX 类型，而是直接使用指针类型
- `decimal/numeric` 类型默认映射为 `decimal.Decimal`