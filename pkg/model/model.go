// Packge model 应用模型数据层
package model

import (
	"goblog/pkg/database"
	"goblog/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB gorm.DB 对象
var DB *gorm.DB

// ConnectDB 初始化模型
func ConnectDB() *gorm.DB {
	var err error

	// 使用已经通过 SSH 配置的 DSN
	config := mysql.New(mysql.Config{
		//DSN: "root:zzl19961120...@tcp(127.0.0.1:3306)/goblog?charset=utf8&parseTime=True&loc=Local",
		// 使用自定义的 Dialector，它会通过 SSH 隧道连接
		Conn: database.DB,
	})

	// 准备数据库连接池
	DB, err = gorm.Open(config, &gorm.Config{})

	logger.LogError(err)

	return DB
}
