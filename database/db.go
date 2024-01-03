package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"verivista/pt/interfaces"
)

// DBClient 数据库连接
var DBClient *sql.DB

// InitDB 创建数据库连接
func InitDB(DB interfaces.DB) error {

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", DB.User, DB.Pass, DB.IP, DB.Port, DB.Name)

	var err error
	DBClient, err = sql.Open(DB.Driver, dataSource)
	if err != nil {
		return fmt.Errorf("[DB Connect Create Error]: %v", err)
	}

	if err = DBClient.Ping(); err != nil {
		return fmt.Errorf("[DB Ping Fail]: %v", err)
	}

	return nil
}
