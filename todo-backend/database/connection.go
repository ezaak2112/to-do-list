package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() *sql.DB {
	dsn := "root:@tcp(127.0.0.1:3306)/todoapp?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Gagal konek database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database tidak merespon:", err)
	}

	log.Println("Berhasil terhubung ke database!")
	return db
}
