package database

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
)

// ConnectDB 負責連線到 SQLite 資料庫並回傳一個資料庫連線實例。
func ConnectDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// InitializeDatabase 負責讀取並執行 SQL 腳本來初始化資料庫結構。
func InitializeDatabase(db *sql.DB) error {
	schema, err := os.ReadFile("scripts/schema.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return err
	}

	return nil
}
