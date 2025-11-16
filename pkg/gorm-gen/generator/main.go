package main

import (
	"flag"
	"fmt"
	"gorm.io/driver/postgres"

	"github.com/spf13/viper"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const postgresTcpDSN = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai"

var configPath = flag.String("f", "gorm-gen/generator/etc/config.yaml", "config file path")

func main() {
	// 解析命令行-f参数，否则返回默认值
	flag.Parse()

	// 加载数据库配置
	cfg := MustLoadConfig(*configPath)

	// 构建数据库连接字符串
	dsn := fmt.Sprintf(postgresTcpDSN, cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		panic(fmt.Errorf("cannot establish db connection: %w", err))
	}

	// 配置代码生成器
	genCfg := gen.Config{
		OutPath:      "pkg/gorm-gen/query", // 生成的查询代码输出路径
		ModelPkgPath: "pkg/gorm-gen/model", // 生成的模型代码输出路径

		Mode: gen.WithDefaultQuery | gen.WithQueryInterface, // 生成默认查询方法和查询接口

		FieldNullable:     true,  // 字段可为空
		FieldCoverable:    false, // 不生成字段覆盖相关代码
		FieldSignable:     true,  // 生成字段符号相关代码
		FieldWithIndexTag: false, // 不生成索引标签
		FieldWithTypeTag:  true,  // 生成类型标签
	}

	// 创建代码生成器实例
	g := gen.NewGenerator(genCfg)
	g.UseDB(db)

	// 配置数据类型映射
	dataMap := map[string]func(columnType gorm.ColumnType) (dataType string){
		"numeric": func(columnType gorm.ColumnType) (dataType string) {
			return "decimal.Decimal" // 将数据库 numeric 类型映射为 decimal.Decimal
		},
	}
	g.WithDataTypeMap(dataMap)

	// 配置特殊字段处理
	autoUpdateTimeField := gen.FieldGORMTag("updated_at", func(tag field.GormTag) field.GormTag {
		return tag.Append("autoUpdateTime") // 自动更新时间字段
	})
	autoCreateTimeField := gen.FieldGORMTag("created_at", func(tag field.GormTag) field.GormTag {
		return tag.Append("autoCreateTime") // 自动创建时间字段
	})
	softDeleteField := gen.FieldType("deleted_at", "gorm.DeletedAt") // 软删除字段

	// 组合所有字段选项
	fieldOpts := []gen.ModelOpt{autoCreateTimeField, autoUpdateTimeField, softDeleteField}

	// 生成所有表的模型
	allModel := g.GenerateAllTable(fieldOpts...)

	// 应用基本查询方法
	g.ApplyBasic(allModel...)

	// 执行代码生成
	g.Execute()
}

// MustLoadConfig 从指定路径加载数据库配置
// 如果加载失败则会 panic
func MustLoadConfig(path string) *DatabaseConfig {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed to read config file: %s", err))
	}

	var c DatabaseConfig
	if err := v.Unmarshal(&c); err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %s", err))
	}

	return &c
}
