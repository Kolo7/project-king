package db

import (
	"sync"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_DB  = &gorm.DB{}
	once sync.Once
)

func GetDB(dsn ...string) *gorm.DB {
	once.Do(func() {
		if len(dsn) == 0 {
			panic("dsn is null")
		}
		var err error
		_DB, err = gorm.Open(mysql.Open(dsn[0]), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		// 给_DB设置连接池参数
		sqlDB, err := _DB.DB()
		if err != nil {
			panic(err)
		}
		// SetMaxIdleConns 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(10)
		// SetMaxOpenConns 设置打开数据库连接的最大数量。
		sqlDB.SetMaxOpenConns(100)
		// SetConnMaxLifetime 设置了连接可复用的最大时间。
		sqlDB.SetConnMaxLifetime(time.Hour)

		if err := _DB.Use(otelgorm.NewPlugin()); err != nil {
			panic(err)
		}
	})
	return _DB
}
