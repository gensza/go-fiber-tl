package config

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func ConnectDB() {
	dsn := "host=localhost user=postgres password=postgres dbname=tablelink_db port=5432 sslmode=disable"
	var err error
	DB, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("Gagal konek ke database:", err)
	}
	fmt.Println("Connected to PostgreSQL!")
}
