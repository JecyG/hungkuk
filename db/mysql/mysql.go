package mysql

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type (
	Option struct {
		Host            string `json:"host"`              // HOST
		Port            int    `json:"port"`              // 端口
		DBName          string `json:"db_name"`           // 数据库名称
		Username        string `json:"username"`          // 用户名
		Password        string `json:"password"`          // 密码
		Params          string `json:"params"`            // 连接参数
		MaxIdleConns    int    `json:"max_idle_conns"`    // 连接池：最大空闲连接数量
		MaxOpenConns    int    `json:"max_open_conns"`    // 连接池：最大打开连接数量
		ConnMaxLifetime int    `json:"conn_max_lifetime"` // 连接池：连接最大可复用时间（单位：秒）
	}
)

var (
	_options   = make(map[string]*Option)
	_defaultDB = ""
	_instances sync.Map
)

func Init(options ...Option) {
	for _, option := range options {
		_options[option.DBName] = &option
	}

	if len(options) > 0 {
		_defaultDB = options[0].DBName
	}
}

func DefaultDB() (*gorm.DB, error) {
	return DB(_defaultDB)
}

func DB(dbName string) (*gorm.DB, error) {
	instance, exist := _instances.Load(dbName)
	if !exist {
		option, ok := _options[dbName]
		if !ok {
			return nil, fmt.Errorf("none option of db: %s", dbName)
		}

		var err error
		instance, err = createInstance(option)
		if err != nil {
			return nil, err
		}

		_instances.Store(option.DBName, instance)
	}

	return instance.(*gorm.DB), nil
}

func createInstance(option *Option) (*gorm.DB, error) {
	// user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		option.Username,
		option.Password,
		option.Host,
		option.Port,
		option.DBName,
		option.Params)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(option.MaxIdleConns)
	sqlDB.SetMaxOpenConns(option.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(option.ConnMaxLifetime) * time.Second)

	return db, nil
}
