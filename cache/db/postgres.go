package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5438
	user     = "postgres"
	password = "postges"
	ssl = "disable" 
)

func ConnectPg() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+ "password=%s " + "sslmode=%s ",host, port, user, password, ssl)
	db, err := sql.Open("postgres",psqlInfo )

	if err != nil {
		panic(err.Error())
	}

	return db
}
