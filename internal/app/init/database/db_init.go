package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "postgres"
	port     = 5432
	user     = "postgres"
	password = "test"
	dbname   = "mydb"
)

func Init() *sql.DB {

	var err error

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		fmt.Printf("Ошибка, %s", err)
	}

	err = db.Ping()

	if err != nil {
		fmt.Printf("Ошибка ping, %s", err)
	}
	return db
}
