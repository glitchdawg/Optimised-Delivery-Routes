package db

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    sslmode := os.Getenv("DB_SSLMODE")

    if sslmode == "" {
        sslmode = "disable"
    }

    connStr := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        user, password, host, port, dbname, sslmode,
    )

    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        return err
    }
    return DB.Ping()
}