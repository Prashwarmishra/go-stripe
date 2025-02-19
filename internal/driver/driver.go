package driver

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func OpenDB(dsn string) (*sql.DB, error) {
	DB, err := sql.Open("mysql", dsn)

	if err != nil {
		fmt.Println("error connecting to sql database:", err)
		return nil, err
	}

	err = DB.Ping()

	if err != nil {
		fmt.Println("failed to receive ping from sql database:", err)
		return nil, err
	}

	return DB, nil
}
