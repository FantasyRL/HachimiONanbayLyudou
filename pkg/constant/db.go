package constant

const (
	DBMaxConnections   = 1000            // (DB) 最大连接数
	DBMaxIdleConns     = 10              // (DB) 最大空闲连接数
	DBConnMaxLifetime  = 10 * ONE_SECOND // (DB) 最大可复用时间
	DBConnMaxIdleTime  = 5 * ONE_MINUTE  // (DB) 最长保持空闲状态时间
	DBDefaultBatchSize = 100             // (DB) 默认批量插入大小
)
