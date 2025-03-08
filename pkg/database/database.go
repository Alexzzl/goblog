// Package database 数据库相关
package database

import (
	"context"
	"database/sql"
	"goblog/pkg/logger"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"

	"golang.org/x/crypto/ssh"
)

// 定义数据库连接变量
var DB *sql.DB
var sshClient *ssh.Client // 添加全局 SSH 客户端变量

func Initialize() {
	initDB()
	// createTables()
}

func initDB() {
	// 设置 SSH 配置
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("zzl19961120..."), // 也可以使用密钥认证
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境建议使用具体的主机密钥
	}

	// 建立 SSH 连接
	var err error
	sshClient, err = ssh.Dial("tcp", "119.8.108.55:22", sshConfig)
	logger.LogError(err)

	// 通过 SSH 建立 MySQL 连接
	mysqlAddr := "127.0.0.1:3306"
	tunnel, err := sshClient.Dial("tcp", mysqlAddr)
	logger.LogError(err)

	config := mysql.Config{
		User:                 "root",
		Passwd:               "zzl19961120...",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}

	// 使用 RegisterDialContext 注册自定义连接器
	mysql.RegisterDialContext("tcp", func(ctx context.Context, addr string) (net.Conn, error) {
		return tunnel, nil
	})

	// 准备数据库连接池
	DB, err = sql.Open("mysql", config.FormatDSN())
	//fmt.Println(config.FormatDSN())
	logger.LogError(err)

	// 设置最大连接数
	DB.SetMaxOpenConns(25)
	// 设置最大空闲连接数
	DB.SetMaxIdleConns(25)
	// 设置每个链接的过期时间
	DB.SetConnMaxLifetime(5 * time.Minute)

	// 尝试连接，失败会报错
	err = DB.Ping()
	logger.LogError(err)
}

// func createTables() {
// 	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
// 	id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
// 	title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
// 	body longtext COLLATE utf8mb4_unicode_ci);`

// 	_, err := DB.Exec(createArticlesSQL)
// 	logger.LogError(err)
// }
