package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const postgresTcpDSN = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai"

// InitDBClient 初始化db连接(mysql)
func InitDBClient() (*gorm.DB, error) {

	dsn := fmt.Sprintf(postgresTcpDSN, config.PgSQL.Host, config.PgSQL.User, config.PgSQL.Password, config.PgSQL.Database, config.PgSQL.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,  // 在执行任何 SQL 时都会创建一个 prepared statement 并将其缓存，以提高后续的效率
		SkipDefaultTransaction: false, // 不禁用默认事务(即单个创建、更新、删除时使用事务)
		TranslateError:         true,  // 允许翻译错误
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: glogger.New(
			log.Default(),
			glogger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  glogger.Info,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      true,
				Colorful:                  true,
			}),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB() // 尝试获取 DB 实例对象
	if err != nil {
		return nil, fmt.Errorf("get generic database object error: %w", err)
	}

	sqlDB.SetMaxIdleConns(constant.DBMaxIdleConns)       // 最大闲置连接数
	sqlDB.SetMaxOpenConns(constant.DBMaxConnections)     // 最大连接数
	sqlDB.SetConnMaxLifetime(constant.DBConnMaxLifetime) // 最大可复用时间
	sqlDB.SetConnMaxIdleTime(constant.DBConnMaxIdleTime) // 最长保持空闲状态时间
	db = db.WithContext(context.Background())

	// 进行连通性测试
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database error: %w", err)
	}

	logger.Info("database connected successfully")
	return db, nil
}
