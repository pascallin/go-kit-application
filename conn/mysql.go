package conn

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/pascallin/go-kit-application/config"
)

var (
	mOnce               sync.Once
	mysqlSingleInstance *gorm.DB
)

func GetMysqlDB() *gorm.DB {
	mOnce.Do(func() {
		db, err := openMysql()
		if err != nil {
			log.Error(err)
		}
		mysqlSingleInstance = db
	})
	return mysqlSingleInstance
}

func openMysql() (*gorm.DB, error) {
	c := config.GetMysqlConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Password, c.Host, c.Port, c.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info("Mysql database connected")

	return db, nil
}
