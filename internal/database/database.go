package database

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open() (*sql.DB, error) {
	db, err := sql.Open("mysql", os.Getenv("DB_DSN"))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, err
}
