package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func DB(connStr string) (*sql.DB, error) {
	// DB, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/information_schema?charset=utf8")
	DB, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}
	DB.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	DB.SetMaxIdleConns(10)
	return DB, nil
}
