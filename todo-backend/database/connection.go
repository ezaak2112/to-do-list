package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func Connect() (*sql.DB, error) {
	// Load file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Tidak menemukan file .env, menggunakan environment default")
	}

	// Ambil konfigurasi database dari .env
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	// Menyusun DSN (Data Source Name) untuk koneksi MySQL
	dsn := mysqlUser + ":" + mysqlPassword + "@tcp(" + mysqlHost + ":" + mysqlPort + ")/" + mysqlDatabase + "?parseTime=true"

	// Membuka koneksi ke database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Memastikan koneksi berhasil
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Berhasil terhubung ke database!")
	return db, nil
}
