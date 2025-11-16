// Package db 提供了一个基于 GORM Gen 的事务管理包装器
//
// GORM Gen 的事务机制：
// 1. GORM Gen 通过 query.Use(db) 生成 Query 实例，该实例包含了所有数据库操作的构建器
// 2. Query 实例提供了 Transaction 方法，用于执行事务操作
// 3. 在事务中，所有数据库操作都会使用同一个事务连接
//
// 我们的包装设计：
// 1. 泛型抽象：
//   - 定义 Transactional[T] 接口，要求类型 T 必须实现 Transaction 方法
//   - 这个接口是对 GORM Gen 的 Query 实例的抽象，因为 Query 实例本身就实现了 Transaction 方法
//
// 2. 事务传递：
//   - 利用 context.Context 存储和传递事务实例
//   - 在事务开始时，将事务实例存储到 context 中
//   - 通过 Get 方法从 context 中获取事务实例
//   - 支持事务嵌套，确保在嵌套调用中使用同一个事务
//
// 3. 统一接口：
//   - DB[T] 结构体封装了具体的数据库操作类型 T
//   - NewDBWithQuery 工厂方法接收两个参数：
//   - gormDB：底层的 GORM 数据库连接
//   - newQueryFunc：用于创建 Query 实例的函数，如 query.Use
//   - 这种设计允许我们：
//   - 统一管理数据库连接和事务
//   - 在编译时确保类型安全
//   - 支持不同的数据库操作类型（只要它们实现了 Transactional 接口）
//
// 使用示例：
// ```
// // 1. 创建 GORM Gen 的 Query 实例
// query := query.Use(gormDB)
//
// // 2. 创建我们的数据库实例
// db := NewDBWithQuery(gormDB, query.Use)
//
// // 3. 执行事务
//
//	err := Transaction(ctx, func(ctx context.Context) error {
//	    // 获取事务实例
//	    tx := db.Get(ctx)
//	    // 执行数据库操作
//	    return nil
//	})
//
// ```
//
// 注意事项：
// 1. 事务实例通过 context 传递，确保在函数调用链中使用相同的 context
// 2. 支持事务嵌套，在任何嵌套调用中都能获取到同一个事务
// 3. 使用空结构体作为 context key，避免内存分配
package db

import (
	"context"
	"database/sql"

	"gorm.io/gen"
	"gorm.io/gorm"
)

// key 用于在 context 中存储数据库事务的键
// 使用空结构体作为键类型可以避免内存分配
var key keyType = struct{}{}

type keyType struct{}

// Transactional 是一个泛型接口，定义了事务操作的基本行为
// T 表示具体的数据库操作类型（如 GORM 的查询构建器）
type Transactional[T any] interface {
	// Transaction 方法用于执行事务操作
	// fn 是事务中要执行的函数
	// options 是可选的 SQL 事务选项
	Transaction(fn func(tx T) error, options ...*sql.TxOptions) error
}

// DB 是一个泛型结构体，用于管理数据库操作
// T 必须实现 Transactional 接口
type DB[T Transactional[T]] struct {
	query T // 存储数据库查询实例
}

// instance 用于存储全局唯一的数据库实例
var instance any

// newDB 创建一个新的数据库实例
// 参数 query 是实现了 Transactional 接口的数据库查询实例
func newDB[T Transactional[T]](query T) *DB[T] {
	database := &DB[T]{query: query}
	instance = database
	return database
}

// Get 从上下文中获取数据库实例
// 如果上下文中存在事务，则返回事务实例；否则返回默认查询实例
func (d *DB[T]) Get(ctx context.Context) T {
	db, ok := ctx.Value(key).(T)
	if !ok {
		return d.query
	}
	return db
}

// Transaction 执行事务操作
// 支持事务嵌套，在任何嵌套调用中只要传递 context，就能获取到同一个事务
// fn 是要在事务中执行的函数
// options 是可选的 SQL 事务选项
func Transaction[T Transactional[T]](ctx context.Context, fn func(ctx context.Context) error, options ...*sql.TxOptions) error {
	db := instance.(*DB[T])
	return db.query.Transaction(func(tx T) error {
		// 将事务实例存储到上下文中
		ctx = context.WithValue(ctx, key, tx)
		return fn(ctx)
	}, options...)
}

// NewDBWithQuery 创建一个新的数据库实例
// dbClient 是 GORM 数据库客户端
// newQueryFunc 是用于创建查询构建器的函数
// gorm-gen的query包里，query.Use(db, opts...) 会返回一个 Query 实例，Query 实例里包含了各种数据库操作的构建器,就是我们传入的newQueryFunc
// gorm-gen的 Query 实例有一个Transaction方法，可以执行事务操作,我们这里进行了一个包装
func NewDBWithQuery[T Transactional[T]](dbClient *gorm.DB,
	newQueryFunc func(db *gorm.DB, opts ...gen.DOOption) T) *DB[T] {
	return newDB[T](newQueryFunc(dbClient))
}
