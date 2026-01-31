package database

import (
	"database/sql"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

var JudgeDB *sql.DB

func ConnectJudgeDB() {
	var err error
	JudgeDB, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/judge")
	if err != nil {
		log.Fatal(err)
	}
}
